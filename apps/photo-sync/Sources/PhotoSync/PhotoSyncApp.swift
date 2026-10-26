import SwiftUI

@main
struct PhotoSyncApp: App {
    @StateObject private var hostStore = HomieHostStore()

    var body: some Scene {
        WindowGroup("Homie Photos") {
            ContentView(hostStore: hostStore)
                .id(hostStore.selected.graphQLURL.absoluteString)
        }
        .windowStyle(.titleBar)
    }
}
