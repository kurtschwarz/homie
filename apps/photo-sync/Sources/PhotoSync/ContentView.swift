import AppKit
import Photos
import SwiftUI

struct ContentView: View {
    @ObservedObject private var hostStore: HomieHostStore
    @ObservedObject private var session: SyncSession
    @State private var isPickingAlbums = false

    private var auth: AuthModel { session.auth }
    private var library: PhotoLibrary { session.library }
    private var uploader: PhotoUploader { session.uploader }
    private var selection: AlbumSelection { session.selection }

    init(hostStore: HomieHostStore, session: SyncSession) {
        self.hostStore = hostStore
        self.session = session
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
            await session.start()
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
        session.startSync()
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
        #if compiler(>=6.2)
        if #available(macOS 26, *) {
            ToolbarItem(placement: .principal) { title }
                .sharedBackgroundVisibility(.hidden)
        } else {
            ToolbarItem(placement: .principal) { title }
        }
        #else
        ToolbarItem(placement: .principal) { title }
        #endif
    }
}
