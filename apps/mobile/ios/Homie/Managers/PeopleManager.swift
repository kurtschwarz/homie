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

private struct PeopleResponse: Decodable {
    struct Data: Decodable { let people: [PersonDTO]? }
    struct ErrorItem: Decodable { let message: String }
    let data: Data?
    let errors: [ErrorItem]?
}

private enum PeopleManagerError: LocalizedError {
    case invalidResponse
    case httpStatus(Int)
    case server(String)

    var errorDescription: String? {
        switch self {
        case .invalidResponse: "The API returned an invalid response."
        case .httpStatus(let status): "The API returned HTTP \(status)."
        case .server(let message): message
        }
    }
}

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
            let response = try await perform("{ people { id name emoji swatch isFamily } }")
            people = (response.data?.people ?? []).map { dto in
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
            _ = try await perform(
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
            _ = try await perform(
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

    private func perform(_ query: String, variables: [String: Any] = [:]) async throws -> PeopleResponse {
        var request = URLRequest(url: HomieEndpoint.graphQLURL)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: ["query": query, "variables": variables])
        let (data, response) = try await URLSession.shared.data(for: request)
        guard let response = response as? HTTPURLResponse else { throw PeopleManagerError.invalidResponse }
        guard (200..<300).contains(response.statusCode) else { throw PeopleManagerError.httpStatus(response.statusCode) }
        let result = try JSONDecoder().decode(PeopleResponse.self, from: data)
        if let message = result.errors?.first?.message { throw PeopleManagerError.server(message) }
        return result
    }

}
