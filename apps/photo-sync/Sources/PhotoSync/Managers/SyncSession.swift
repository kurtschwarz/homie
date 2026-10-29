import Combine
import Foundation
import Photos

@MainActor
final class SyncSession: ObservableObject {
    static let pollInterval: Duration = .seconds(300)

    let auth: AuthModel
    let library: PhotoLibrary
    let uploader: PhotoUploader
    let selection: AlbumSelection
    let permissions: PermissionsModel

    private var forwarding: [AnyCancellable] = []

    init(endpoint: URL) {
        let api = HomieAPI(endpoint: endpoint, sessionToken: { HomieSession.token(for: endpoint) })
        let auth = AuthModel(api: api)
        let library = PhotoLibrary()
        self.auth = auth
        self.library = library
        selection = AlbumSelection()
        let permissions = PermissionsModel(library: library, endpoint: endpoint)
        self.permissions = permissions
        uploader = PhotoUploader(library: library, api: api, onAuthenticationRequired: { auth.sessionExpired() })

        let changes = [
            auth.objectWillChange, library.objectWillChange,
            uploader.objectWillChange, selection.objectWillChange,
            permissions.objectWillChange,
        ]
        forwarding = changes.map { publisher in
            publisher.sink { [weak self] _ in self?.objectWillChange.send() }
        }
    }

    var isSignedIn: Bool {
        if case .signedIn = auth.state { return true }
        return false
    }

    var needsSignIn: Bool {
        if case .signedOut = auth.state { return true }
        return false
    }

    var hasPhotosAccess: Bool { library.authorization == .authorized || library.authorization == .limited }

    var needsSetup: Bool { isSignedIn && (!hasPhotosAccess || !selection.hasSelection) }

    var failedCount: Int { uploader.albumProgress.values.reduce(0) { $0 + $1.failed } + (uploader.drawingProgress?.failed ?? 0) }

    var statusLine: String {
        if needsSignIn { return "Sign in needed" }
        if case .checking = auth.state { return "Checking your Homie session…" }
        if uploader.isUploading {
            let done = uploader.completed + (uploader.drawingProgress?.completed ?? 0)
            let total = uploader.total + (uploader.drawingProgress?.total ?? 0)
            return total > 0 ? "Syncing \(done.formatted()) of \(total.formatted())" : "Syncing…"
        }
        if needsSetup { return "Open Homie Photos to finish setup" }
        if failedCount > 0 { return "Some items didn't sync" }
        if let last = selection.lastSyncedAt { return "Last synced \(last.formatted(.relative(presentation: .named)))" }
        return "Not synced yet"
    }

    /// The checks never prompt, so this runs without a window and the menu bar keeps working when the
    /// window is never opened.
    func start() async {
        permissions.refresh()
        guard !permissions.showsScreen else { return }
        await auth.restoreSession()
        loadAlbumsIfNeeded()
    }

    var needsWindowAtLaunch: Bool {
        Self.needsWindow(showsPermissions: permissions.showsScreen, needsSignIn: needsSignIn, hasSelection: selection.hasSelection)
    }

    nonisolated static func needsWindow(showsPermissions: Bool, needsSignIn: Bool, hasSelection: Bool) -> Bool {
        showsPermissions || needsSignIn || !hasSelection
    }

    func startSync() {
        let ids = library.albums.map(\.id).filter { selection.contains($0) }
        uploader.sync(albumIDs: ids, artworkAlbumID: selection.artworkAlbumID) { [selection] in selection.markSynced() }
    }

    func syncIfDue() {
        permissions.refresh()
        loadAlbumsIfNeeded()
        guard Self.shouldPoll(
            signedIn: isSignedIn,
            hasPhotosAccess: hasPhotosAccess,
            hasSelection: selection.hasSelection,
            keepSyncing: selection.keepSyncing,
            isUploading: uploader.isUploading
        ) else { return }
        startSync()
    }

    func poll() async {
        await start()
        while !Task.isCancelled {
            syncIfDue()
            try? await Task.sleep(for: Self.pollInterval)
        }
    }

    nonisolated static func shouldPoll(
        signedIn: Bool,
        hasPhotosAccess: Bool,
        hasSelection: Bool,
        keepSyncing: Bool,
        isUploading: Bool
    ) -> Bool {
        signedIn && hasPhotosAccess && hasSelection && keepSyncing && !isUploading
    }

    private func loadAlbumsIfNeeded() {
        if hasPhotosAccess, library.albums.isEmpty { library.loadAlbums() }
    }
}
