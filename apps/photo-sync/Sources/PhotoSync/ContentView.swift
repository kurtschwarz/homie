import AppKit
import Photos
import SwiftUI

struct ContentView: View {
    private let hostStore: HomieHostStore
    private let session: SyncSession
    @State private var pane: AppPane = .status

    private var auth: AuthModel { session.auth }
    private var library: PhotoLibrary { session.library }
    private var runner: SyncRunner { session.runner }
    private var selection: AlbumSelection { session.selection }

    init(hostStore: HomieHostStore, session: SyncSession) {
        self.hostStore = hostStore
        self.session = session
    }

    var body: some View {
        HStack(spacing: 0) {
            SidebarView(selection: highlightedPane, isSignedIn: session.isSignedIn, isLoading: phase.hidesSidebar) { pane = $0 }
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

    private var highlightedPane: AppPane { pane }

    private var isChecking: Bool {
        if case .checking = auth.state { return true }
        return false
    }

    private var phase: StartupPhase {
        StartupPhase.resolve(
            showsPermissions: showsPermissions,
            isCheckingSession: isChecking,
            isSignedIn: session.isSignedIn
        )
    }

    @ViewBuilder
    private var paneContent: some View {
        switch phase {
        case .permissions:
            PermissionsView(permissions: session.permissions, library: library)
        case .gettingReady:
            GettingReadyView(hostName: hostStore.selected.name)
        case .signedOut:
            if pane.isAvailableSignedOut {
                PanePlaceholder(pane: pane)
            } else {
                SignInView(auth: auth, hostStore: hostStore)
            }
        case .signedIn:
            switch SignedInScreen.resolve(pane: pane, hasSelection: selection.hasSelection) {
            case .placeholder:
                PanePlaceholder(pane: pane)
            case let screen:
                photoLibraryContent(screen)
            }
        }
    }

    @ViewBuilder
    private func photoLibraryContent(_ screen: SignedInScreen) -> some View {
        switch library.authorization {
        case .authorized, .limited:
            albumFlow(screen)
                .onAppear {
                    if library.albums.isEmpty { library.loadAlbums() }
                }
        default:
            Color.clear
        }
    }

    @ViewBuilder
    private func albumFlow(_ screen: SignedInScreen) -> some View {
        switch screen {
        case .picker:
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
        case .emptyStatus:
            SyncEmptyView(onChooseAlbums: { pane = .albums })
        default:
            SyncStatusView(
                library: library,
                selection: selection,
                runner: runner,
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
