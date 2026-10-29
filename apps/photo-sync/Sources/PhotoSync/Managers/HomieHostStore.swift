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
    private var checkID = UUID()

    init(
        defaults: UserDefaults = .standard,
        healthCheck: @escaping (URL) async throws -> HomieHealth = { try await HomieAPI(endpoint: $0).health() }
    ) {
        self.defaults = defaults
        self.healthCheck = healthCheck
        selected = HomieHost.configured(defaults: defaults)
    }

    func validateSelected() async {
        let host = selected
        let requestID = UUID()
        checkID = requestID
        isChecking = true
        defer { if checkID == requestID { isChecking = false } }
        let result = try? await healthCheck(host.graphQLURL)
        guard checkID == requestID, selected.id == host.id else { return }
        health = result?.status == "OK" ? result : nil
        requiresSetup = result?.status == "SETUP_REQUIRED"
    }

    func select(_ host: HomieHost) async throws {
        let requestID = UUID()
        checkID = requestID
        isChecking = true
        defer { if checkID == requestID { isChecking = false } }
        let health = try await healthCheck(host.graphQLURL)
        guard checkID == requestID else { throw CancellationError() }
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
        checkID = UUID()
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
