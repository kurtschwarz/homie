import Combine
import Foundation

@MainActor
final class AuthModel: ObservableObject {
    enum State {
        case checking
        case signedOut(String?)
        case signedIn(HomieUser)
    }

    @Published private(set) var state: State = .checking
    @Published private(set) var isSigningIn = false

    private let api: HomieAPI
    private var didRestore = false

    init(api: HomieAPI = HomieAPI()) {
        self.api = api
    }

    func restoreSession() async {
        guard !didRestore else { return }
        didRestore = true
        guard let token = HomieSession.token else {
            state = .signedOut(nil)
            return
        }
        do {
            guard let user = try await api.currentUser(token: token) else {
                HomieSession.clear()
                state = .signedOut(nil)
                return
            }
            guard user.role == .admin else {
                try? await api.logout(token: token)
                HomieSession.clear()
                state = .signedOut(HomieAPIError.forbidden.localizedDescription)
                return
            }
            state = .signedIn(user)
        } catch {
            if case HomieAPIError.unauthorized = error { HomieSession.clear() }
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
            try HomieSession.save(result.token)
            state = .signedIn(result.user)
        } catch {
            state = .signedOut(error.localizedDescription)
        }
    }

    func signOut() {
        let token = HomieSession.token
        HomieSession.clear()
        state = .signedOut(nil)
        guard let token else { return }
        Task { try? await api.logout(token: token) }
    }

    func sessionExpired() {
        HomieSession.clear()
        state = .signedOut(HomieAPIError.unauthorized.localizedDescription)
    }
}
