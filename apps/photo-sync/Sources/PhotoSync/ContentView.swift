import AppKit
import Photos
import SwiftUI

struct ContentView: View {
    @ObservedObject private var hostStore: HomieHostStore
    @StateObject private var auth: AuthModel
    @StateObject private var library: PhotoLibrary
    @StateObject private var uploader: PhotoUploader
    @StateObject private var selection = AlbumSelection()
    @State private var isPickingAlbums = false

    init(hostStore: HomieHostStore) {
        self.hostStore = hostStore
        let endpoint = hostStore.selected.graphQLURL
        let api = HomieAPI(endpoint: endpoint, sessionToken: { HomieSession.token(for: endpoint) })
        let auth = AuthModel(api: api)
        let library = PhotoLibrary()
        _auth = StateObject(wrappedValue: auth)
        _library = StateObject(wrappedValue: library)
        _uploader = StateObject(wrappedValue: PhotoUploader(
            library: library,
            api: api,
            onAuthenticationRequired: { auth.sessionExpired() }
        ))
    }

    var body: some View {
        Group {
            switch auth.state {
            case .checking:
                ProgressView("Checking your Homie session…")
            case .signedOut:
                SignInView(auth: auth, hostStore: hostStore)
            case .signedIn:
                photoLibraryContent
            }
        }
        .frame(minWidth: 520, minHeight: 620)
        .hidingWindowTitle()
        .toolbar { WindowTitleToolbar() }
        .task {
            await auth.restoreSession()
        }
    }

    @ViewBuilder
    private var photoLibraryContent: some View {
        switch library.authorization {
        case .authorized, .limited:
            albumFlow
                .onAppear {
                    if library.albums.isEmpty { library.loadAlbums() }
                }
        case .denied, .restricted:
            ContentUnavailableView("Photos Access Needed", systemImage: "photo.badge.exclamationmark", description: Text("Allow Photo Sync to access your Photos library in System Settings → Privacy & Security → Photos."))
        case .notDetermined:
            ContentUnavailableView {
                Label("Connect to Photos", systemImage: "photo.on.rectangle.angled")
            } description: {
                Text("Photo Sync reads your Homie album and displays photo metadata locally.")
            } actions: {
                Button("Allow Photo Access") { Task { await library.authorize() } }
                    .buttonStyle(.borderedProminent)
            }
        @unknown default:
            ContentUnavailableView("Photos Unavailable", systemImage: "photo")
        }
    }

    @ViewBuilder
    private var albumFlow: some View {
        if isPickingAlbums || !selection.hasSelection {
            AlbumPickerView(
                library: library,
                selection: selection,
                hostName: hostStore.selected.name,
                username: username,
                onStart: {
                    isPickingAlbums = false
                    startSync()
                },
                onSignOut: { auth.signOut() }
            )
        } else {
            SyncStatusView(
                library: library,
                selection: selection,
                uploader: uploader,
                onSyncNow: startSync,
                onChangeAlbums: { isPickingAlbums = true }
            )
        }
    }

    private var username: String {
        if case .signedIn(let user) = auth.state { return user.username }
        return ""
    }

    private func startSync() {
        let ids = library.albums.map(\.id).filter { selection.contains($0) }
        uploader.sync(albumIDs: ids) { selection.markSynced() }
    }
}

private extension View {
    @ViewBuilder
    func hidingWindowTitle() -> some View {
        if #available(macOS 15, *) {
            toolbar(removing: .title)
        } else {
            self
        }
    }
}

private struct WindowTitleToolbar: ToolbarContent {
    private var title: some View {
        Text("Homie Photos")
            .font(.system(size: 13, weight: .semibold))
            .foregroundStyle(.secondary)
    }

    var body: some ToolbarContent {
        if #available(macOS 26, *) {
            ToolbarItem(placement: .principal) { title }
                .sharedBackgroundVisibility(.hidden)
        } else {
            ToolbarItem(placement: .principal) { title }
        }
    }
}
