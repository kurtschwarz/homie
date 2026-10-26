import SwiftUI

struct HomieButtonStyle: ButtonStyle {
    var prominent = false
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 17, weight: .semibold))
            .foregroundStyle(prominent ? Color.homieGreenDark : Color.primary)
            .padding(.horizontal, 28)
            .padding(.vertical, 12)
            .background {
                if prominent {
                    RoundedRectangle(cornerRadius: 12).fill(Color.homieGreen.opacity(0.75))
                } else {
                    RoundedRectangle(cornerRadius: 12).fill(Color(nsColor: .controlBackgroundColor))
                        .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.secondary.opacity(0.3)))
                }
            }
            .opacity(isEnabled ? (configuration.isPressed ? 0.8 : 1) : 0.45)
    }
}
