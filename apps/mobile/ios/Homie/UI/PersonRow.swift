import SwiftUI

struct PersonRow: View {
    let person: Person

    private var color: Color {
        switch person.swatch {
        case "clay": .brown
        case "amber": .orange
        case "moss", "pine", "teal": .teal
        case "sky": .cyan
        case "iris", "orchid", "plum": .purple
        default: .gray
        }
    }

    var body: some View {
        HStack(spacing: 12) {
            Text(person.emoji.isEmpty ? String(person.name.prefix(1)) : person.emoji)
                .font(.title3).frame(width: 40, height: 40)
                .background(color.opacity(0.18), in: Circle())
            Text(person.name).font(.headline).foregroundStyle(.primary)
            Spacer()
            Circle().fill(color).frame(width: 12, height: 12)
            Text(person.swatch.capitalized).foregroundStyle(.secondary)
            Image(systemName: "chevron.right").font(.caption.bold()).foregroundStyle(.tertiary)
        }
        .padding(.vertical, 2)
        .frame(maxWidth: .infinity, alignment: .leading)
        .contentShape(Rectangle())
    }
}
