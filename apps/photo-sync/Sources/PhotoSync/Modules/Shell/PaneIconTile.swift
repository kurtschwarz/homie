import SwiftUI

struct PaneIconTile: View {
    let pane: AppPane
    var size: CGFloat = 26

    private var corner: CGFloat { size * 7 / 26 }

    private var gradient: [Color] {
        switch pane.icon {
        case .homie: [Color.white.opacity(0.1), Color.white.opacity(0.1)]
        case .symbol(_, let gradient): gradient
        }
    }

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: corner, style: .continuous)
                .fill(LinearGradient(colors: gradient, startPoint: .top, endPoint: .bottom))
                .shadow(color: .black.opacity(0.3), radius: 1, y: 1)
                .overlay(
                    RoundedRectangle(cornerRadius: corner, style: .continuous)
                        .strokeBorder(
                            LinearGradient(colors: [.white.opacity(0.35), .clear], startPoint: .top, endPoint: .center),
                            lineWidth: 1
                        )
                )
            switch pane.icon {
            case .homie:
                HomieMark().frame(width: size * 0.8, height: size * 0.8)
            case .symbol(let name, _):
                Image(systemName: name)
                    .font(.system(size: size * 0.5, weight: .bold))
                    .foregroundStyle(.white)
            }
        }
        .frame(width: size, height: size)
        .accessibilityHidden(true)
    }
}
