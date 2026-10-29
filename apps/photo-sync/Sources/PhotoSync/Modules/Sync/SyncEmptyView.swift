import SwiftUI

struct SyncEmptyView: View {
    let onChooseAlbums: () -> Void

    var body: some View {
        VStack(spacing: 14) {
            PaneIconTile(pane: .status, size: 64)
            Text("Choose albums to start syncing")
                .font(.system(size: 22, weight: .bold))
                .foregroundStyle(Color.homieText)
            Button("Choose albums", action: onChooseAlbums)
                .buttonStyle(HomieButtonStyle(prominent: true))
                .padding(.top, 6)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
