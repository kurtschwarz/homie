import SwiftUI

private enum PersonEditorRoute: Identifiable {
    case add
    case edit(Person)

    var id: String {
        switch self {
        case .add: "add"
        case .edit(let person): "edit-\(person.id ?? person.name)"
        }
    }

    var person: Person? {
        if case .edit(let person) = self { return person }
        return nil
    }
}

struct SettingsView: View {
    @Environment(PeopleManager.self) private var peopleManager
    @State private var editorRoute: PersonEditorRoute?

    var body: some View {
        NavigationStack {
            List {
                Section {
                    ForEach(peopleManager.people) { person in
                        Button {
                            editorRoute = .edit(person)
                        } label: {
                            PersonRow(person: person)
                        }
                        .buttonStyle(.plain)
                        .alignmentGuide(.listRowSeparatorLeading) { _ in -16 }
                        .alignmentGuide(.listRowSeparatorTrailing) { dimensions in dimensions.width + 16 }
                    }
                    Button {
                        editorRoute = .add
                    } label: {
                        Label("Add someone", systemImage: "plus.circle")
                            .font(.headline).foregroundStyle(.tint).padding(.vertical, 6)
                    }
                    .alignmentGuide(.listRowSeparatorLeading) { _ in -16 }
                    .alignmentGuide(.listRowSeparatorTrailing) { dimensions in dimensions.width + 16 }
                } header: { Text("FAMILY") }
                Section {
                    Text("Everyone’s emoji and colour mark their events on the kitchen screen. Tap someone to change them.")
                        .font(.callout).foregroundStyle(.secondary)
                }
                if let errorMessage = peopleManager.errorMessage {
                    Section { Text(errorMessage).foregroundStyle(.red) }
                }
            }
            .overlay {
                if peopleManager.isLoading && peopleManager.people.isEmpty {
                    ProgressView("Loading people…")
                }
            }
            .navigationTitle("Settings")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button { Task { await peopleManager.loadPeople() } } label: { Image(systemName: "arrow.clockwise") }
                        .accessibilityLabel("Refresh people")
                }
            }
            .sheet(item: $editorRoute) { route in
                EditPersonView(person: route.person)
            }
            .task { await peopleManager.loadPeople() }
        }
    }
}
