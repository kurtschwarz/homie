import Foundation

@MainActor
struct OnboardingReset {
    private let defaults: UserDefaults
    private let tokenForEndpoint: (URL) -> String?
    private let clearSessions: () -> Void
    private let logout: (URL, String) async -> Void

    init(
        defaults: UserDefaults = .standard,
        tokenForEndpoint: @escaping (URL) -> String? = HomieSession.token(for:),
        clearSessions: @escaping () -> Void = HomieSession.clearAll,
        logout: @escaping (URL, String) async -> Void = { endpoint, token in
            do {
                try await HomieAPI(endpoint: endpoint).logout(token: token)
            } catch {
                Log.api.error("Logout during reset failed: \(error.localizedDescription, privacy: .private)")
            }
        }
    ) {
        self.defaults = defaults
        self.tokenForEndpoint = tokenForEndpoint
        self.clearSessions = clearSessions
        self.logout = logout
    }

    func run(hostStore: HomieHostStore) async {
        let endpoint = hostStore.selected.graphQLURL
        let token = tokenForEndpoint(endpoint)
        clearSessions()
        var ledger = UploadLedger(defaults: defaults)
        ledger.clear()
        AlbumSelection.clear(defaults: defaults)
        hostStore.resetToDefault()
        if let token { await logout(endpoint, token) }
    }
}
