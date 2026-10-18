import SwiftUI

struct EditPersonView: View {
    @Environment(PeopleManager.self) private var peopleManager
    @Environment(\.dismiss) private var dismiss

    @State private var name: String
    @State private var emoji: String
    @State private var swatch: String

    let person: Person?
    private let emojis = ["🦊", "🐢", "🦋", "🐻", "🐼", "🐸", "🦁", "🐯", "🐨", "🐙", "🦄", "🐝", "🐬", "🦉", "🐧", "🌻", "🏠", "☕️"]

    init(person: Person?) {
        self.person = person
        _name = State(initialValue: person?.name ?? "")
        _emoji = State(initialValue: person?.emoji ?? "")
        _swatch = State(initialValue: person?.swatch ?? Person.defaultSwatch)
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 28) {
                    eventPreview
                    nameSection
                    badgeSection
                    if let person {
                        Button("Remove \(person.name)", role: .destructive) {
                            Task { if await peopleManager.delete(person) { dismiss() } }
                        }
                        .frame(maxWidth: .infinity)
                    }
                }
                .padding(.horizontal, 20)
                .padding(.top, 20)
                .padding(.bottom, 32)
            }
            .background(Color(uiColor: .systemGroupedBackground))
            .navigationTitle(person.map { "Edit \($0.name)" } ?? "Add someone")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) { Button("Cancel") { dismiss() } }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        let updated = Person(
                            id: person?.id,
                            name: name.trimmingCharacters(in: .whitespacesAndNewlines),
                            emoji: emoji,
                            swatch: swatch,
                            isFamily: person?.isFamily ?? false
                        )
                        Task { if await peopleManager.save(updated) { dismiss() } }
                    }
                    .disabled(name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                }
            }
        }
    }

    private var avatarColor: Color {
        switch swatch {
        case "clay": .brown
        case "amber": .orange
        case "moss": .green
        case "pine", "teal": .teal
        case "sky": .blue
        case "iris": .indigo
        case "orchid", "plum": .purple
        default: .gray
        }
    }

    private var avatarText: String {
        emoji.isEmpty ? String(name.prefix(1)) : emoji
    }

    private var eventPreview: some View {
        RoundedCard {
            VStack(alignment: .leading, spacing: 24) {
                HStack(spacing: 16) {
                    Text(avatarText).font(.system(size: 38)).frame(width: 64, height: 64)
                        .background(avatarColor, in: Circle())
                    VStack(alignment: .leading, spacing: 4) {
                        Text(name.isEmpty ? "Name" : name).font(.title3.bold()).foregroundStyle(.primary)
                        Text(swatch.capitalized).font(.body).foregroundStyle(.secondary)
                    }
                    Spacer(minLength: 0)
                }
                Rectangle().fill(Color(uiColor: .separator)).frame(height: 1)
                HStack(spacing: 18) {
                    Text("5:45 pm").foregroundStyle(.secondary)
                    Text(avatarText).font(.body).frame(width: 40, height: 40)
                        .background(avatarColor, in: Circle())
                    Text("Swim lessons").font(.body.weight(.semibold)).foregroundStyle(.primary)
                    Spacer(minLength: 0)
                }
            }
            .padding(24)
        }
    }

    private var nameSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            sectionLabel("Name")
            RoundedCard {
                TextField("Name", text: $name)
                    .font(.title3)
                    .padding(16)
            }
        }
    }

    private var badgeSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            sectionLabel("Badge")
            RoundedCard {
                VStack(spacing: 18) {
                    HStack(spacing: 8) {
                        ForEach(Person.swatches, id: \.self) { option in
                            let color = color(for: option)
                            Button { swatch = option } label: {
                                Circle().fill(color).frame(maxWidth: .infinity).aspectRatio(1, contentMode: .fit)
                                    .overlay { if swatch == option { Image(systemName: "checkmark").font(.body.bold()).foregroundStyle(.white) } }
                                    .padding(3)
                                    .overlay(Circle().stroke(swatch == option ? Color.primary : .clear, lineWidth: 2.5))
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    Rectangle().fill(Color(uiColor: .separator)).frame(height: 1)
                    LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 8), count: 6), spacing: 8) {
                        ForEach(emojis, id: \.self) { option in
                            Button { emoji = option } label: {
                                Text(option).font(.system(size: 18)).frame(width: 40, height: 40)
                                    .background(avatarColor, in: Circle())
                                    .overlay(Circle().stroke(emoji == option ? Color.primary : .clear, lineWidth: 3))
                                    .padding(3)
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    TextField("Or type any emoji", text: $emoji)
                        .font(.body)
                        .padding(16)
                        .background(Color(uiColor: .systemGroupedBackground), in: RoundedRectangle(cornerRadius: 16))
                }
                .padding(18)
            }
            Text("Choose a colour and emoji that are easy to tell apart from across the kitchen.")
                .font(.body)
                .foregroundStyle(.secondary)
        }
    }

    private func sectionLabel(_ title: String) -> some View {
        Text(title.uppercased()).font(.caption.bold()).tracking(1.5).foregroundStyle(.secondary)
    }

    private func color(for value: String) -> Color {
        switch value {
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
}
