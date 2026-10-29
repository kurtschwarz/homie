import AppKit
import Photos
import SwiftUI

struct ContentView: View {
    @ObservedObject private var hostStore: HomieHostStore
    @ObservedObject private var session: SyncSession
    @State private var pane: AppPane = .status

    private var auth: AuthModel { session.auth }
    private var library: PhotoLibrary { session.library }
    private var uploader: PhotoUploader { session.uploader }
    private var selection: AlbumSelection { session.selection }

    init(hostStore: HomieHostStore, session: SyncSession) {
        self.hostStore = hostStore
        self.session = session
    }

    var body: some View {
        HStack(spacing: 0) {
            SidebarView(selection: pane, isSignedIn: session.isSignedIn) { pane = $0 }
            paneContent
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .glassWindowChrome()
        .preferredColorScheme(.dark)
        .task {
            await session.start()
        }
    }

    @ViewBuilder
    private var paneContent: some View {
        if pane.isAvailableSignedOut {
            PanePlaceholder(pane: .about)
        } else {
            switch auth.state {
            case .checking:
                ProgressView("Checking your Homie session…")
            case .signedOut:
                SignInView(auth: auth, hostStore: hostStore)
            case .signedIn:
                switch pane {
                case .status, .albums:
                    photoLibraryContent
                default:
                    PanePlaceholder(pane: pane)
                }
            }
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
        if pane == .albums || !selection.hasSelection {
            AlbumPickerView(
                library: library,
                selection: selection,
                hostName: hostStore.selected.name,
                username: username,
                onStart: {
                    pane = .status
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
                onChangeAlbums: { pane = .albums }
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
