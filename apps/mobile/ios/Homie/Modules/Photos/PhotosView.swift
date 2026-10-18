import SwiftUI

struct PhotosView: View {
    var body: some View {
        NavigationStack {
            Color(uiColor: .systemGroupedBackground)
                .ignoresSafeArea()
                .navigationTitle("Photos")
        }
    }
}
