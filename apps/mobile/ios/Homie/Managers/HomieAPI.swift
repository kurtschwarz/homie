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
    private struct HealthPayload: Decodable {
        struct Health: Decodable {
            let status: String
        }

        let health: Health
    }

    static func checkHealth() async throws {
        let payload: HealthPayload = try await perform("{ health { status } }")
        guard payload.health.status.lowercased() == "ok" else {
            throw HomieAPIError.server("The Homie API reported an unhealthy status.")
        }
    }

    static func perform<Payload: Decodable>(
        _ query: String,
        variables: [String: Any] = [:]
    ) async throws -> Payload {
        var request = URLRequest(url: HomieEndpoint.graphQLURL)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if let token = HomieCredential.token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        // These are small queries against a machine on the same network. The
        // 60 second default leaves pull to refresh spinning for a minute when
        // the API is unreachable, which reads as a hang.
        request.timeoutInterval = 10
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
