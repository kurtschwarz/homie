import Combine
import Foundation

final class AuthModel: ObservableObject {
    enum State {
        case checking
        case signedOut(String?)
        case signedIn(HomieUser)
    }

    @Published private(set) var state: State
    @Published private(set) var isSigningIn = false

    private let api: HomieAPI
    private let endpoint: URL
    private var didRestore = false

    init(api: HomieAPI = HomieAPI()) {
        self.api = api
        endpoint = api.endpoint
        state = .checking
    }

    func restoreSession() async {
        guard !didRestore else { return }
        didRestore = true
        guard let token = HomieSession.token(for: endpoint) else {
            state = .signedOut(nil)
            return
        }
        do {
            guard let user = try await api.currentUser(token: token) else {
                HomieSession.clear(for: endpoint)
                state = .signedOut(nil)
                return
            }
            guard user.role == .admin else {
                try? await api.logout(token: token)
                HomieSession.clear(for: endpoint)
                state = .signedOut(HomieAPIError.forbidden.localizedDescription)
                return
            }
            state = .signedIn(user)
        } catch {
            if case HomieAPIError.unauthorized = error { HomieSession.clear(for: endpoint) }
            state = .signedOut(error.localizedDescription)
        }
    }

    func signIn(username: String, password: String) async {
        guard !isSigningIn else { return }
        let username = username.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !username.isEmpty, !password.isEmpty else {
            state = .signedOut("Enter your username and password.")
            return
        }
        isSigningIn = true
        defer { isSigningIn = false }
        do {
            let result = try await api.login(username: username, password: password)
            guard result.user.role == .admin else {
                try? await api.logout(token: result.token)
                throw HomieAPIError.forbidden
            }
            try HomieSession.save(result.token, for: endpoint)
            state = .signedIn(result.user)
        } catch {
            state = .signedOut(error.localizedDescription)
        }
    }

    func signOut() {
        let token = HomieSession.token(for: endpoint)
        HomieSession.clear(for: endpoint)
        state = .signedOut(nil)
        guard let token else { return }
        Task { try? await api.logout(token: token) }
    }

    func sessionExpired() {
        HomieSession.clear(for: endpoint)
        state = .signedOut(HomieAPIError.unauthorized.localizedDescription)
    }
}
