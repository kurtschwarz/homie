import SwiftUI

extension Color {
    static let homieGreen = Color(red: 0.22, green: 0.73, blue: 0.53)
    static let homieGreenDark = Color(red: 0.17, green: 0.45, blue: 0.33)
    static let homieButtonText = Color(red: 11 / 255, green: 51 / 255, blue: 36 / 255)

    static let homieWindowTint = Color(hex: 0x363A48)
    static let homieText = Color(hex: 0xF2F2F4)
    static let homieSectionHeading = Color(hex: 0xA9A9B0)
    static let homieBadgeText = Color(hex: 0xC9C9CE)

    init(hex: UInt32) {
        self.init(
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255
        )
    }
}
