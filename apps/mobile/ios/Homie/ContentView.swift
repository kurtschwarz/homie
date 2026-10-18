import SwiftUI

struct ContentView: View {
    var body: some View {
        TabView {
            Tab("Home", systemImage: "house") {
                HomeView()
            }

            Tab("Reminders", systemImage: "bell") {
                RemindersView()
            }

            Tab("Photos", systemImage: "photo") {
                PhotosView()
            }

            Tab("Settings", systemImage: "slider.horizontal.3") {
                SettingsView()
            }
        }
    }
}

#Preview {
    ContentView()
}
