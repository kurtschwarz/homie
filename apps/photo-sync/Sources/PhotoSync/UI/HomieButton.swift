import SwiftUI

struct HomieButtonStyle: ButtonStyle {
    var prominent = false
    var fullWidth = false
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 14, weight: .semibold))
            .foregroundStyle(prominent ? Color.homieButtonText : Color.primary)
            .padding(.horizontal, 22)
            .padding(.vertical, fullWidth ? 11 : 9)
            .frame(maxWidth: fullWidth ? .infinity : nil)
            .background {
                if fullWidth {
                    RoundedRectangle(cornerRadius: 12).fill(
                        LinearGradient(
                            colors: [Color.homieGreenLight, Color.homieGreen],
                            startPoint: .top,
                            endPoint: .bottom
                        )
                    )
                } else if prominent {
                    RoundedRectangle(cornerRadius: 12).fill(Color.homieGreen.opacity(0.75))
                } else {
                    RoundedRectangle(cornerRadius: 12).fill(Color(nsColor: .controlBackgroundColor))
                        .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.secondary.opacity(0.3)))
                }
            }
            .opacity(isEnabled ? (configuration.isPressed ? 0.8 : 1) : 0.45)
    }
}
