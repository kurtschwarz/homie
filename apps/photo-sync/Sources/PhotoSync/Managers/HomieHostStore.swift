import Foundation
import Observation

@Observable
final class HomieHostStore {
    private(set) var selected: HomieHost
    private(set) var health: HomieHealth?
    private(set) var isChecking = false
    private(set) var requiresSetup = false
    private(set) var contentIdentity = UUID()

    let discovery = HomieDiscovery()
    private let defaults: UserDefaults
    private let healthCheck: (URL) async throws -> HomieHealth
    private var checkTask: Task<HomieHealth, Error>?

    init(
        defaults: UserDefaults = .standard,
        healthCheck: @escaping (URL) async throws -> HomieHealth = { try await HomieAPI(endpoint: $0).health() }
    ) {
        self.defaults = defaults
        self.healthCheck = healthCheck
        selected = HomieHost.configured(defaults: defaults)
    }

    private func check(_ url: URL) async throws -> HomieHealth {
        checkTask?.cancel()
        let healthCheck = healthCheck
        let task = Task { try await healthCheck(url) }
        checkTask = task
        isChecking = true
        defer {
            if !task.isCancelled {
                checkTask = nil
                isChecking = false
            }
        }
        let result = await task.result
        if task.isCancelled { throw CancellationError() }
        return try result.get()
    }

    func validateSelected() async {
        let result: HomieHealth?
        do {
            result = try await check(selected.graphQLURL)
        } catch is CancellationError {
            return
        } catch {
            result = nil
        }
        health = result?.status == "OK" ? result : nil
        requiresSetup = result?.status == "SETUP_REQUIRED"
    }

    func select(_ host: HomieHost) async throws {
        let health = try await check(host.graphQLURL)
        guard health.status == "OK" else {
            if health.status == "SETUP_REQUIRED", host.id == selected.id {
                self.health = nil
                requiresSetup = true
            }
            throw HomieHostError.setupRequired
        }
        Log.discovery.info("Switching host: \(host.graphQLURL.host() ?? "unknown", privacy: .private) (\(String(describing: host.source), privacy: .public))")
        defaults.set(host.graphQLURL.absoluteString, forKey: "HomieAPIURL")
        defaults.set(host.name, forKey: "HomieAPIName")
        defaults.set(host.source == .discovered, forKey: "HomieAPIDiscovered")
        self.health = health
        requiresSetup = false
        selected = host
        contentIdentity = UUID()
    }

    func resetToDefault() {
        Log.discovery.info("Resetting host to default")
        checkTask?.cancel()
        checkTask = nil
        isChecking = false
        health = nil
        requiresSetup = false
        defaults.removeObject(forKey: "HomieAPIURL")
        defaults.removeObject(forKey: "HomieAPIName")
        defaults.removeObject(forKey: "HomieAPIDiscovered")
        selected = HomieHost.configured(defaults: defaults)
        contentIdentity = UUID()
    }
}

private enum HomieHostError: LocalizedError {
    case setupRequired

    var errorDescription: String? {
        "This Homie needs to be set up first. Open its manager to finish setup."
    }
}
