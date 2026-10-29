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
            SidebarView(selection: highlightedPane, isSignedIn: session.isSignedIn, isLoading: isChecking || showsPermissions) { pane = $0 }
            paneContent
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .glassWindowChrome()
        .preferredColorScheme(.dark)
        .onChange(of: session.isSignedIn) { _, signedIn in
            if !signedIn { pane = .status }
        }
        .task {
            session.permissions.refresh()
            await session.start()
        }
        .onChange(of: showsPermissions) { _, shows in
            if !shows { Task { await session.start() } }
        }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in
            session.permissions.refresh()
        }
    }

    private var showsPermissions: Bool { session.permissions.showsScreen }

    private var isPickingAlbums: Bool { pane == .albums || !selection.hasSelection }

    private var highlightedPane: AppPane {
        let showsPicker = session.isSignedIn && session.hasPhotosAccess && isPickingAlbums
        return pane == .status && showsPicker ? .albums : pane
    }

    private var isChecking: Bool {
        if case .checking = auth.state { return true }
        return false
    }

    @ViewBuilder
    private var paneContent: some View {
        if showsPermissions {
            PermissionsView(permissions: session.permissions, library: library)
        } else {
            authContent
        }
    }

    @ViewBuilder
    private var authContent: some View {
        switch auth.state {
        case .checking:
            GettingReadyView(hostName: hostStore.selected.name)
        case .signedOut:
            if pane.isAvailableSignedOut {
                PanePlaceholder(pane: pane)
            } else {
                SignInView(auth: auth, hostStore: hostStore)
            }
        case .signedIn:
            switch pane {
            case .status, .albums:
                photoLibraryContent
            default:
                PanePlaceholder(pane: pane)
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
        default:
            Color.clear
        }
    }

    @ViewBuilder
    private var albumFlow: some View {
        if isPickingAlbums {
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
