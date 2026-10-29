import SwiftUI

struct PanePlaceholder: View {
    let pane: AppPane

    var body: some View {
        VStack(spacing: 14) {
            PaneIconTile(pane: pane, size: 64)
            Text(pane.title)
                .font(.system(size: 22, weight: .bold))
                .foregroundStyle(Color.homieText)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
