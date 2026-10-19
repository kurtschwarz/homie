import SwiftUI

struct ContentView: View {
    @State private var selection: Destination = .home

    var body: some View {
        TabView(selection: $selection) {
            Tab("Home", image: "phosphor-house", value: Destination.home) {
                HomeView()
            }

            Tab("Reminders", image: "phosphor-bell", value: Destination.reminders) {
                RemindersView()
            }

            Tab("Photos", image: "phosphor-image", value: Destination.photos) {
                PhotosView()
            }

            Tab("Settings", image: "phosphor-sliders-horizontal", value: Destination.settings) {
                SettingsView()
            }
        }
        .tint(Color(red: 49.0 / 255, green: 85.0 / 255, blue: 199.0 / 255))
    }

    private enum Destination: Hashable {
        case home
        case reminders
        case photos
        case settings
    }
}

#Preview {
    ContentView()
}
