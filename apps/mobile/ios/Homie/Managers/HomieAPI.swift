import Foundation

enum HomieAPIError: LocalizedError {
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

private struct GraphQLResponse<Payload: Decodable>: Decodable {
    struct ErrorItem: Decodable { let message: String }
    let data: Payload?
    let errors: [ErrorItem]?
}

/// Posts GraphQL queries to the Homie API and decodes the `data` payload.
enum HomieAPI {
    static func perform<Payload: Decodable>(
        _ query: String,
        variables: [String: Any] = [:]
    ) async throws -> Payload {
        var request = URLRequest(url: HomieEndpoint.graphQLURL)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: ["query": query, "variables": variables])

        let (data, response) = try await URLSession.shared.data(for: request)
        guard let response = response as? HTTPURLResponse else { throw HomieAPIError.invalidResponse }
        guard (200..<300).contains(response.statusCode) else { throw HomieAPIError.httpStatus(response.statusCode) }

        let result = try JSONDecoder().decode(GraphQLResponse<Payload>.self, from: data)
        if let message = result.errors?.first?.message { throw HomieAPIError.server(message) }
        guard let payload = result.data else { throw HomieAPIError.invalidResponse }
        return payload
    }
}
