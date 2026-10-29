import AppKit
import Combine
import Foundation

@MainActor
final class PermissionsModel: ObservableObject {
    @Published private(set) var keychain: AccessState = .allowed
    @Published private(set) var localNetwork: AccessState = .allowed
    @Published private(set) var showsScreen = false
    @Published private(set) var isRequestingKeychain = false
    @Published private(set) var isRequestingLocalNetwork = false
    @Published private(set) var isRequestingPhotos = false

    private let library: PhotoLibrary
    private let endpoint: URL

    init(library: PhotoLibrary, endpoint: URL) {
        self.library = library
        self.endpoint = endpoint
    }

    var gate: PermissionsGate {
        PermissionsGate(photos: library.authorization, localNetwork: localNetwork, keychain: keychain)
    }

    func refresh() {
        library.refreshAuthorization()
        // Browsing is what prompts, so launch trusts the last Allow rather than asking again.
        if localNetwork != .denied { localNetwork = LocalNetworkProbe.recordedAccess() }
        // A denial is only known from the last Allow; probing again would just report "needs prompt".
        if keychain != .denied { keychain = HomieSession.probeAccess(for: endpoint) }
        showsScreen = PermissionsGate.showsScreen(currentlyShowing: showsScreen, gate: gate)
    }

    /// PhotoKit takes about two seconds to report an answer the system already has, so the row says it is
    /// waiting rather than sitting there unticked.
    func allowPhotos() async {
        guard !isRequestingPhotos else { return }
        isRequestingPhotos = true
        defer { isRequestingPhotos = false }
        await library.authorize()
        Log.permissions.info("Photos answer: \(String(describing: self.library.authorization), privacy: .public)")
    }

    func allowLocalNetwork() async {
        guard !isRequestingLocalNetwork else { return }
        isRequestingLocalNetwork = true
        defer { isRequestingLocalNetwork = false }
        let answer = await LocalNetworkProbe.request(timeout: .seconds(60))
        Log.permissions.info("Local network answer: \(String(describing: answer), privacy: .public)")
        if let answer { localNetwork = answer }
    }

    func allowKeychain() async {
        guard !isRequestingKeychain else { return }
        isRequestingKeychain = true
        defer { isRequestingKeychain = false }
        let endpoint = endpoint
        keychain = await Task.detached { HomieSession.requestAccess(for: endpoint) }.value
    }

    func openPhotosSettings() {
        guard let url = URL(string: "x-apple.systempreferences:com.apple.preference.security?Privacy_Photos") else { return }
        NSWorkspace.shared.open(url)
    }

    func openLocalNetworkSettings() {
        guard let url = URL(string: "x-apple.systempreferences:com.apple.preference.security?Privacy_LocalNetwork") else { return }
        NSWorkspace.shared.open(url)
    }

    func continueIfSatisfied() {
        if gate.isSatisfied { showsScreen = false }
    }
}
