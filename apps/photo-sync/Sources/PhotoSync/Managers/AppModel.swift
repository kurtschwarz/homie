import Combine
import Foundation
import ServiceManagement

final class LaunchAtLogin: ObservableObject {
    @Published private(set) var isEnabled = SMAppService.mainApp.status == .enabled
    @Published private(set) var errorMessage: String?

    func set(_ enabled: Bool) {
        do {
            if enabled { try SMAppService.mainApp.register() } else { try SMAppService.mainApp.unregister() }
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
        isEnabled = SMAppService.mainApp.status == .enabled
    }
}

final class AppModel: ObservableObject {
    let hostStore = HomieHostStore()
    let launchAtLogin = LaunchAtLogin()
    @Published private(set) var session: SyncSession

    private var hostChanges: AnyCancellable?
    private var polling: Task<Void, Never>?

    init() {
        session = SyncSession(endpoint: hostStore.selected.graphQLURL)
        hostChanges = hostStore.$contentIdentity.dropFirst().sink { [weak self] _ in
            guard let self else { return }
            self.session = SyncSession(endpoint: self.hostStore.selected.graphQLURL)
            self.startPolling()
        }
        startPolling()
    }

    private func startPolling() {
        polling?.cancel()
        let session = session
        polling = Task { await session.poll() }
    }
}
