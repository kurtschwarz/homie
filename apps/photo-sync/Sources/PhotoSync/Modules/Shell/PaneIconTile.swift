import SwiftUI

struct PaneIconTile: View {
    let pane: AppPane
    var size: CGFloat = 26

    private var corner: CGFloat { size * 7 / 26 }

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: corner, style: .continuous)
                .fill(LinearGradient(colors: pane.gradient, startPoint: .top, endPoint: .bottom))
                .shadow(color: .black.opacity(0.3), radius: 1, y: 1)
                .overlay(
                    RoundedRectangle(cornerRadius: corner, style: .continuous)
                        .strokeBorder(
                            LinearGradient(colors: [.white.opacity(0.35), .clear], startPoint: .top, endPoint: .center),
                            lineWidth: 1
                        )
                )
            if let symbol = pane.symbol {
                Image(systemName: symbol)
                    .font(.system(size: size * 0.5, weight: .bold))
                    .foregroundStyle(.white)
            } else {
                HomieMark().frame(width: size * 0.8, height: size * 0.8)
            }
        }
        .frame(width: size, height: size)
    }
}

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
