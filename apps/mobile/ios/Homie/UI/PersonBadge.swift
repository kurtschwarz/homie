import SwiftUI

extension Person {
    /// Mirrors the swatch palette in @homie/ui, so a person reads the same here
    /// as they do on the kitchen screen.
    var color: Color {
        switch swatch {
        case "clay": .brown
        case "amber": Color(red: 0.72, green: 0.38, blue: 0.15)
        case "moss": .green
        case "pine", "teal": .teal
        case "sky": .blue
        case "iris": .indigo
        case "orchid", "plum": .purple
        default: .gray
        }
    }

    /// People without an emoji fall back to the first letter of their name.
    var badgeText: String {
        emoji.isEmpty ? String(name.prefix(1)) : emoji
    }
}

/// The round emoji badge that marks a person's events.
struct PersonBadge: View {
    let person: Person
    var diameter: CGFloat = 40

    var body: some View {
        Text(person.badgeText)
            .font(.system(size: diameter * 0.5))
            .frame(width: diameter, height: diameter)
            .background(person.color, in: Circle())
    }
}
