import SwiftUI

extension Color {
    static let homieGreen = Color(red: 0.22, green: 0.73, blue: 0.53)
    static let homieGreenDark = Color(red: 0.17, green: 0.45, blue: 0.33)
    static let homieGreenLight = Color(red: 0.47, green: 0.84, blue: 0.68)
    static let homieButtonText = Color(red: 11 / 255, green: 51 / 255, blue: 36 / 255)
    static let homieError = Color(red: 1, green: 0.55, blue: 0.55)

    static let homieWindowTint = Color(hex: 0x363A48)
    static let homieText = Color(hex: 0xF2F2F4)
    static let homieSectionHeading = Color(hex: 0xA9A9B0)
    static let homieBadgeText = Color(hex: 0xC9C9CE)
    static let homieDivider = Color.white.opacity(0.08)
    static let homieRowSelected = Color.white.opacity(0.12)
    static let homieRowHover = Color.white.opacity(0.06)
    static let homieBadgeBorder = Color.white.opacity(0.25)

    init(hex: UInt32) {
        self.init(
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255
        )
    }
}
