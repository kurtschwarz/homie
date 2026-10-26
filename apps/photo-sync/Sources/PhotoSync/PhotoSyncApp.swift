import SwiftUI

@main
struct PhotoSyncApp: App {
    @StateObject private var hostStore = HomieHostStore()

    var body: some Scene {
        WindowGroup {
            ContentView(hostStore: hostStore)
                .id(hostStore.selected.graphQLURL.absoluteString)
        }
        .windowStyle(.titleBar)
    }
}
