//
//  HomieCard.swift
//  Homie
//
//  The matte "soft-touch" card surface from the Homie app: a barely-there top-to-bottom fade,
//  a bright hairline on the top edge, a faint dark one on the bottom, and a soft shadow.
//  Same values as the prototype's --matte-surface and .ap-matte-s, for the light, dark and night themes.
//
//  Usage:
//      VStack { … }
//          .padding(16)
//          .homieCard()                         // 22 pt corners, follows light/dark mode
//
//      .homieCard(cornerRadius: 16)             // tiles and buttons
//      .homieCard(theme: .night)                // the warm kitchen night theme
//
//      HomieCardBackground(cornerRadius: 22)   // or use the background view directly
//
//  Requires iOS 15 / macOS 12.
//

import SwiftUI

public enum HomieCardTheme: Sendable {
    /// Follows the system light/dark setting.
    case automatic
    case light
    case dark
    /// Homie's warm night palette.
    case night
}

public struct HomieCardBackground: View {
    public var cornerRadius: CGFloat
    public var theme: HomieCardTheme
    @Environment(\.colorScheme) private var colorScheme

    public init(cornerRadius: CGFloat = 22, theme: HomieCardTheme = .automatic) {
        self.cornerRadius = cornerRadius
        self.theme = theme
    }

    public var body: some View {
        let p = palette
        let shape = RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
        shape
            // Surface: from --surface at the top to 96% surface + 4% ink at the bottom.
            .fill(LinearGradient(colors: [p.top, p.bottom], startPoint: .top, endPoint: .bottom))
            // Hairlines: bright on the top edge, faint dark on the bottom edge (the CSS inset shadows).
            .overlay(
                shape
                    .inset(by: 0.5)
                    .stroke(
                        LinearGradient(stops: [
                            .init(color: p.highlight, location: 0),
                            .init(color: p.highlight.opacity(0), location: 0.08),
                            .init(color: p.lowlight.opacity(0), location: 0.92),
                            .init(color: p.lowlight, location: 1)
                        ], startPoint: .top, endPoint: .bottom),
                        lineWidth: 1
                    )
            )
            // Lift: a tight contact shadow plus a soft one underneath.
            .shadow(color: .black.opacity(p.contactShadow), radius: 1, x: 0, y: 1)
            .shadow(color: .black.opacity(p.softShadow), radius: 5, x: 0, y: 4)
    }

    private struct Palette {
        var top: Color, bottom: Color, highlight: Color, lowlight: Color
        var contactShadow: Double, softShadow: Double
    }

    private var resolved: HomieCardTheme {
        if case .automatic = theme { return colorScheme == .dark ? .dark : .light }
        return theme
    }

    private var palette: Palette {
        switch resolved {
        case .dark:
            // surface #17181B → #202123
            return Palette(top: Color(homieCardHex: 0x17181B), bottom: Color(homieCardHex: 0x202123),
                           highlight: .white.opacity(0.07), lowlight: .black.opacity(0.3),
                           contactShadow: 0, softShadow: 0.4)
        case .night:
            // surface #16120E → #1E1A15
            return Palette(top: Color(homieCardHex: 0x16120E), bottom: Color(homieCardHex: 0x1E1A15),
                           highlight: .white.opacity(0.07), lowlight: .black.opacity(0.3),
                           contactShadow: 0, softShadow: 0.4)
        default:
            // surface #FFFFFF → #F6F6F6
            return Palette(top: Color(homieCardHex: 0xFFFFFF), bottom: Color(homieCardHex: 0xF6F6F6),
                           highlight: .white.opacity(0.85), lowlight: .black.opacity(0.05),
                           contactShadow: 0.05, softShadow: 0.10)
        }
    }
}

public extension View {
    /// Puts the Homie matte card surface behind this view and clips it to the same corners.
    func homieCard(cornerRadius: CGFloat = 22, theme: HomieCardTheme = .automatic) -> some View {
        self
            .clipShape(RoundedRectangle(cornerRadius: cornerRadius, style: .continuous))
            .background(HomieCardBackground(cornerRadius: cornerRadius, theme: theme))
    }
}

private extension Color {
    init(homieCardHex hex: UInt32) {
        self.init(.sRGB, red: Double((hex >> 16) & 0xFF) / 255, green: Double((hex >> 8) & 0xFF) / 255, blue: Double(hex & 0xFF) / 255, opacity: 1)
    }
}

#if DEBUG
struct HomieCard_Previews: PreviewProvider {
    static var sample: some View {
        HStack(spacing: 12) {
            Image(systemName: "tray")
                .font(.system(size: 18, weight: .medium))
                .frame(width: 40, height: 40)
                .background(Color.gray.opacity(0.12), in: RoundedRectangle(cornerRadius: 12))
            VStack(alignment: .leading, spacing: 2) {
                Text("You’re in offline mode").font(.system(size: 15, weight: .semibold))
                Text("You can still add events and reminders. They’ll sync as soon as Homie’s back.")
                    .font(.system(size: 13)).foregroundColor(.secondary)
            }
        }
        .padding(16)
    }

    static var previews: some View {
        VStack(spacing: 20) {
            sample.homieCard(theme: .light)
            sample.homieCard(theme: .dark).environment(\.colorScheme, .dark)
        }
        .padding(20)
        .background(Color(homieCardHex: 0xF4F6F8))
    }
}
#endif
