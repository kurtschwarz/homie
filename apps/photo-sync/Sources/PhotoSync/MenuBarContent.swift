import AppKit
import SwiftUI

struct MenuBarIcon: View {
    let session: SyncSession

    private var symbol: String? {
        if session.uploader.isUploading { return "arrow.triangle.2.circlepath" }
        if session.needsSignIn || session.needsSetup || session.failedCount > 0 { return "exclamationmark.triangle" }
        return nil
    }

    private static func template(_ image: NSImage) -> NSImage {
        image.isTemplate = true
        return image
    }

    var body: some View {
        if let symbol {
            Image(systemName: symbol)
        } else if let icon = Bundle.main.image(forResource: "HomieMenuTemplate") {
            Image(nsImage: Self.template(icon))
        } else {
            Image(systemName: "photo.on.rectangle")
        }
    }
}

/// The window is suppressed at launch, so this is the only view that exists then: it runs the startup
/// checks and opens the window when there is something the user has to finish.
struct MenuBarLabel: View {
    let session: SyncSession
    @Environment(\.openWindow) private var openWindow

    var body: some View {
        MenuBarIcon(session: session)
            .task {
                await session.start()
                guard session.needsWindowAtLaunch else { return }
                openWindow(id: "main")
                NSApp.activate(ignoringOtherApps: true)
            }
    }
}

struct MenuBarContent: View {
    let session: SyncSession
    let launchAtLogin: LaunchAtLogin
    let onAbout: () -> Void
    let onReset: () -> Void
    @Environment(\.openWindow) private var openWindow

    var body: some View {
        Text(session.statusLine)
        Divider()
        Button("Open Homie Photos…") {
            openWindow(id: "main")
            NSApp.activate(ignoringOtherApps: true)
        }
        Button("Sync Now") { session.startSync() }
            .disabled(!session.isSignedIn || !session.hasPhotosAccess || !session.selection.hasSelection || session.uploader.isUploading)
        Divider()
        Toggle("Start at Login", isOn: Binding(get: { launchAtLogin.isEnabled }, set: { launchAtLogin.set($0) }))
        if let error = launchAtLogin.errorMessage {
            Text(error)
        }
        Divider()
        Button("About Homie Photos", action: onAbout)
        Button("Reset Homie Photos…", action: onReset)
        Divider()
        Button("Quit Homie Photos") { NSApp.terminate(nil) }
            .keyboardShortcut("q")
    }
}
