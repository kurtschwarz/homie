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
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("FAMILY").font(.caption.bold()).tracking(1.5).foregroundStyle(.secondary)
                RoundedCard {
                    ForEach(Array(peopleManager.people.enumerated()), id: \.element.id) { index, person in
                        Button { editorRoute = .edit(person) } label: {
                            PersonRow(person: person)
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .padding(.horizontal, 20)
                                .padding(.vertical, 12)
                        }
                        .buttonStyle(.plain)
                        if index < peopleManager.people.count - 1 || !peopleManager.people.isEmpty {
                            Divider()
                        }
                    }
                    Button { editorRoute = .add } label: {
                        Label("Add someone", systemImage: "plus.circle")
                            .font(.headline).foregroundStyle(.tint)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(.horizontal, 20)
                            .padding(.vertical, 16)
                    }
                    .buttonStyle(.plain)
                }
                Text("Everyone’s emoji and colour mark their events on the kitchen screen. Tap someone to change them.")
                    .font(.caption)
                if let errorMessage = peopleManager.errorMessage {
                    Text(errorMessage).foregroundStyle(.red)
                }
            }
            .padding(.horizontal, 16)
            .padding(.top, 8)
        }
        .background(Color(uiColor: .systemGroupedBackground))
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
                .presentationBackground(Color(uiColor: .systemGroupedBackground))
        }
        .task { await peopleManager.loadPeople() }
    }
}
