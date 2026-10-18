import SwiftUI

struct EditPersonView: View {
    @Environment(PeopleManager.self) private var peopleManager
    @Environment(\.dismiss) private var dismiss
    let person: Person?
    @State private var name: String
    @State private var emoji: String
    @State private var swatch: String

    init(person: Person?) {
        self.person = person
        _name = State(initialValue: person?.name ?? "")
        _emoji = State(initialValue: person?.emoji ?? "")
        _swatch = State(initialValue: person?.swatch ?? Person.defaultSwatch)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("Person") {
                    TextField("Name", text: $name)
                    TextField("Emoji", text: $emoji)
                }
                Section("Colour") {
                    Picker("Colour", selection: $swatch) {
                        ForEach(Person.swatches, id: \.self) { Text($0.capitalized).tag($0) }
                    }
                }
                if let person {
                    Section {
                        Button("Remove \(person.name)", role: .destructive) {
                            Task { if await peopleManager.delete(person) { dismiss() } }
                        }
                    }
                }
            }
            .navigationTitle(person == nil ? "Add someone" : "Edit person")
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
}
