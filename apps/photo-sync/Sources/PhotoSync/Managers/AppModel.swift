import Foundation
import Observation
import ServiceManagement

@Observable
final class LaunchAtLogin {
    private(set) var isEnabled = SMAppService.mainApp.status == .enabled
    private(set) var errorMessage: String?

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

@Observable
final class AppNavigation {
    var requestedPane: AppPane?
}

@Observable
final class AppModel {
    let hostStore = HomieHostStore()
    let navigation = AppNavigation()
    let launchAtLogin = LaunchAtLogin()
    private(set) var session: SyncSession

    @ObservationIgnored private var polling: Task<Void, Never>?

    init() {
        session = SyncSession(endpoint: hostStore.selected.graphQLURL)
        observeHost()
        startPolling()
    }

    private func observeHost() {
        withObservationTracking {
            _ = hostStore.contentIdentity
        } onChange: { [weak self] in
            Task { @MainActor [weak self] in
                guard let self else { return }
                session = SyncSession(endpoint: hostStore.selected.graphQLURL)
                startPolling()
                observeHost()
            }
        }
    }

    private func startPolling() {
        polling?.cancel()
        let session = session
        polling = Task { await session.poll() }
    }
}
