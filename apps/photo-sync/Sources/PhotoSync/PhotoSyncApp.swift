import AppKit
import SwiftUI

@main
struct PhotoSyncApp: App {
    @StateObject private var model = AppModel()

    var body: some Scene {
        MenuBarExtra {
            MenuBarContent(
                session: model.session,
                launchAtLogin: model.launchAtLogin,
                onAbout: showAbout,
                onReset: confirmReset
            )
        } label: {
            MenuBarLabel(session: model.session)
        }
        .menuBarExtraStyle(.menu)

        MainWindow(model: model)
    }

    private func showAbout() {
        // The standard panel renders "Version <applicationVersion> (<version>)".
        var options: [NSApplication.AboutPanelOptionKey: Any] = [:]
        if let commit = Bundle.main.object(forInfoDictionaryKey: "HomieGitCommit") as? String {
            options[.version] = commit
        }
        NSApp.activate(ignoringOtherApps: true)
        NSApp.orderFrontStandardAboutPanel(options: options)
    }

    private func confirmReset() {
        NSApp.activate(ignoringOtherApps: true)
        let alert = NSAlert()
        alert.messageText = "Reset Homie Photos?"
        alert.informativeText = "This removes saved hosts, sessions, upload history, and cached places. Your macOS Photos permission is unchanged."
        alert.alertStyle = .warning
        alert.addButton(withTitle: "Reset").hasDestructiveAction = true
        alert.addButton(withTitle: "Cancel")
        guard alert.runModal() == .alertFirstButtonReturn else { return }
        let hostStore = model.hostStore
        Task { await OnboardingReset().run(hostStore: hostStore) }
    }
}

private struct MainWindow: Scene {
    @ObservedObject var model: AppModel

    var body: some Scene {
        Window("Homie Photos", id: "main") {
            ContentView(hostStore: model.hostStore, session: model.session)
                .id(model.hostStore.contentIdentity)
        }
        .windowStyle(.hiddenTitleBar)
        .windowResizability(.contentSize)
        .defaultLaunchBehavior(.suppressed)
    }
}
