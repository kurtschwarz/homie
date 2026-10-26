import AppKit
import SwiftUI

@main
struct PhotoSyncApp: App {
    @StateObject private var hostStore = HomieHostStore()

    var body: some Scene {
        WindowGroup("Homie Photos") {
            ContentView(hostStore: hostStore)
                .id(hostStore.contentIdentity)
        }
        .windowStyle(.titleBar)
        .defaultSize(width: 560, height: 680)
        .commands {
            CommandGroup(after: .appInfo) {
                Divider()
                Button("Reset Homie Photos…") {
                    confirmReset()
                }
            }
        }
    }

    private func confirmReset() {
        let alert = NSAlert()
        alert.messageText = "Reset Homie Photos?"
        alert.informativeText = "This removes saved hosts, sessions, upload history, and cached places. Your macOS Photos permission is unchanged."
        alert.alertStyle = .warning
        alert.addButton(withTitle: "Reset").hasDestructiveAction = true
        alert.addButton(withTitle: "Cancel")
        guard alert.runModal() == .alertFirstButtonReturn else { return }
        Task { await OnboardingReset().run(hostStore: hostStore) }
    }
}
