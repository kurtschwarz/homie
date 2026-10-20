import SwiftUI

/// The standard primary action button style used across Homie.
public struct HomieButton<Icon: View>: View {
    public var title: String
    public var action: () -> Void
    @ViewBuilder public var icon: () -> Icon
    public var iconSize: CGFloat
    /// Stretches the button to its container's width instead of hugging its label.
    public var fillsWidth: Bool

    public init(
        title: String,
        iconSize: CGFloat = 16,
        fillsWidth: Bool = false,
        action: @escaping () -> Void,
        @ViewBuilder icon: @escaping () -> Icon = { EmptyView() }
    ) {
        self.title = title
        self.action = action
        self.iconSize = iconSize
        self.fillsWidth = fillsWidth
        self.icon = icon
    }

    public var body: some View {
        SwiftUI.Button(action: action) {
            HStack(spacing: 8) {
                icon()
                    .frame(width: iconSize, height: iconSize)
                    .foregroundColor(.white)
                    .accessibilityHidden(true)
                Text(title)
            }
            .font(.system(size: 14, weight: .semibold))
            .foregroundColor(.white)
            .padding(.horizontal, 16)
            .frame(maxWidth: fillsWidth ? .infinity : nil)
            .frame(height: 40)
            .background(HomieButtonBackground())
            .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
        }
        .buttonStyle(.plain)
    }
}

private struct HomieButtonBackground: View {
    var body: some View {
        LinearGradient(
            colors: [Color(glitchHex: 0xE8784A), Color(glitchHex: 0xD85A25)],
            startPoint: .top,
            endPoint: .bottom
        )
        .overlay {
            RoundedRectangle(cornerRadius: 14, style: .continuous)
                .strokeBorder(
                    LinearGradient(
                        colors: [.white.opacity(0.28), .clear, .black.opacity(0.12)],
                        startPoint: .top,
                        endPoint: .bottom
                    ),
                    lineWidth: 1
                )
        }
    }
}

private extension Color {
    init(glitchHex hex: UInt32) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255,
            opacity: 1
        )
    }
}
