import Combine
import Foundation

@MainActor
final class HomieHostStore: ObservableObject {
    @Published private(set) var selected: HomieHost
    @Published private(set) var health: HomieHealth?
    @Published private(set) var isChecking = false
    @Published private(set) var requiresSetup = false

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
        defaults.set(host.graphQLURL.absoluteString, forKey: "HomieAPIURL")
        defaults.set(host.name, forKey: "HomieAPIName")
        defaults.set(host.source == .discovered, forKey: "HomieAPIDiscovered")
        self.health = health
        requiresSetup = false
        selected = host
    }
}

private enum HomieHostError: LocalizedError {
    case setupRequired

    var errorDescription: String? {
        "This Homie needs to be set up first. Open its manager to finish setup."
    }
}
