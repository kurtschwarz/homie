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
        }
        .tint(Color(red: 227.0 / 255, green: 96.0 / 255, blue: 42.0 / 255))
    }

    private enum Destination: Hashable {
        case home
        case reminders
        case photos
    }
}

#Preview {
    ContentView()
}
