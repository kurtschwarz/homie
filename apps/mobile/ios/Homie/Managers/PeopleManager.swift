import Foundation
import Observation

private struct PersonDTO: Decodable {
    let id: String
    let name: String
    let emoji: String?
    let swatch: String
    let isFamily: Bool

    var person: Person {
        Person(id: id, name: name, emoji: emoji ?? "", swatch: swatch, isFamily: isFamily)
    }
}

private struct PeoplePayload: Decodable { let people: [PersonDTO]? }

@MainActor
@Observable
final class PeopleManager {
    private(set) var people: [Person] = []
    private(set) var errorMessage: String?
    private(set) var isLoading = false

    func loadPeople() async {
        guard !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let response: PeoplePayload = try await HomieAPI.perform("{ people { id name emoji swatch isFamily } }")
            people = (response.people ?? []).map { dto in
                Person(id: dto.id, name: dto.name, emoji: dto.emoji ?? "", swatch: dto.swatch, isFamily: dto.isFamily)
            }
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func save(_ person: Person) async -> Bool {
        do {
            var input: [String: Any] = [
                "name": person.name,
                "emoji": person.emoji,
                "swatch": person.swatch,
                "isFamily": person.isFamily,
            ]
            if let id = person.id { input["id"] = id }
            let _: EmptyPayload = try await HomieAPI.perform(
                "mutation($input: PersonInput!) { savePerson(input: $input) { id } }",
                variables: ["input": input]
            )
            await loadPeople()
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func delete(_ person: Person) async -> Bool {
        guard let id = person.id else { return true }
        do {
            let _: EmptyPayload = try await HomieAPI.perform(
                "mutation($id: ID!) { deletePerson(id: $id) }",
                variables: ["id": id]
            )
            await loadPeople()
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

}

private struct EmptyPayload: Decodable {}
