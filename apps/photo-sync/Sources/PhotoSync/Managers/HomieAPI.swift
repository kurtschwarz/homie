import Foundation

struct PhotoUploadInput: Encodable {
    let assetID: String
    let version: String
    let takenAt: Int
    let width: Int
    let height: Int
    let latitude: Double?
    let longitude: Double?
    let place: String?
    let focus: FocalRegionInput?
}

struct UploadedPhoto: Decodable {
    let id: String
    let url: String
}

struct HomieDrawing: Decodable, Equatable {
    struct Coordinate: Decodable, Equatable {
        let latitude: Double
        let longitude: Double
    }

    let id: String
    let url: String
    let takenAt: Int
    let coordinate: Coordinate?
}

struct DrawingsPayload: Decodable {
    let unpushedDrawings: [HomieDrawing]?
}

struct PushedIDsPayload: Decodable {
    let pushedDrawingAssetIDs: [String]?
}

struct MarkPushedPayload: Decodable {
    let markDrawingPushed: Bool?
}

/// Closed set: the server's role enum only has ADMIN, so an unrecognised role fails decoding
/// rather than being treated as a non-admin. Add a case here when the server adds a role.
enum HomieRole: String, Decodable, Equatable {
    case admin = "ADMIN"
}

struct HomieUser: Decodable, Equatable {
    let id: String
    let username: String
    let role: HomieRole
}

struct GraphQLResponse<Payload: Decodable>: Decodable {
    let data: Payload?
    let errors: [GraphQLErrorItem]?
}

struct GraphQLErrorItem: Decodable { let message: String }

struct UploadPayload: Decodable {
    let uploadPhoto: UploadedPhoto?
}

struct LoginPayload: Decodable { let login: HomieUser? }

struct SessionPayload: Decodable {
    struct Session: Decodable { let user: HomieUser? }
    let session: Session?
}

struct HomieHealth: Decodable, Equatable {
    let status: String
    let version: String
}

struct HealthPayload: Decodable { let health: HomieHealth? }

struct HomieAPI {
    // A bundle launched by `open` does not inherit the shell environment, so
    // UserDefaults is the command-line override that works for the app bundle:
    //   defaults write ar.schw.homie.photo-sync HomieAPIURL http://localhost:3030
    static let defaultEndpoint = URL(string: "http://homie/graphql")!

    let endpoint: URL
    private static let sharedSession: URLSession = {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.httpShouldSetCookies = false
        return URLSession(configuration: configuration)
    }()

    private let session: URLSession
    private let sessionToken: () -> String?

    init(
        endpoint: URL = HomieAPI.configuredEndpoint(),
        session: URLSession? = nil,
        sessionToken: (() -> String?)? = nil
    ) {
        self.endpoint = endpoint
        self.session = session ?? Self.sharedSession
        self.sessionToken = sessionToken ?? { HomieSession.token(for: endpoint) }
    }

    // A bare origin gets /graphql appended so the override can be spelled the
    // same way as HOMIE_API_URL is for the web apps.
    static func configuredEndpoint(
        defaults: UserDefaults = .standard,
        environment: [String: String] = ProcessInfo.processInfo.environment
    ) -> URL {
        for override in [defaults.string(forKey: "HomieAPIURL"), environment["HOMIE_API_URL"]] {
            guard let override, let url = URL(string: override), url.host() != nil else { continue }
            return url.path().isEmpty || url.path() == "/" ? url.appending(path: "graphql") : url
        }
        return defaultEndpoint
    }

    func login(username: String, password: String) async throws -> (user: HomieUser, token: String) {
        let (data, response) = try await perform(
            query: "mutation Login($username: String!, $password: String!) { login(username: $username, password: $password) { id username role } }",
            variables: ["username": username, "password": password],
            token: nil
        )
        let payload = try parse(LoginPayload.self, data: data, response: response)
        guard let user = payload?.login else { throw HomieAPIError.invalidResponse }
        guard let token = Self.sessionToken(from: response) else { throw HomieAPIError.missingSessionCookie }
        return (user, token)
    }

    func currentUser(token: String) async throws -> HomieUser? {
        let (data, response) = try await perform(
            query: "query Session { session { user { id username role } } }",
            variables: [:],
            token: token
        )
        return try parse(SessionPayload.self, data: data, response: response)?.session?.user
    }

    func logout(token: String) async throws {
        let (data, response) = try await perform(query: "mutation Logout { logout }", variables: [:], token: token)
        _ = try parse(EmptyPayload.self, data: data, response: response)
    }

    func health() async throws -> HomieHealth {
        let (data, response) = try await perform(
            query: "query Health { health { status version } }",
            variables: [:],
            token: nil
        )
        guard let health = try parse(HealthPayload.self, data: data, response: response)?.health else {
            throw HomieAPIError.invalidResponse
        }
        return health
    }

    func upload(_ input: PhotoUploadInput, jpeg: Data, filename: String) async throws -> UploadedPhoto {
        guard let token = sessionToken(), !token.isEmpty else { throw HomieAPIError.unauthorized }
        let boundary = "HomiePhotoSync-\(UUID().uuidString)"
        var request = request(token: token)
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        let encoder = JSONEncoder()
        let inputData = try encoder.encode(input)
        let operations: [String: Any] = [
            "query": "mutation UploadPhoto($input: PhotoUploadInput!, $file: Upload!) { uploadPhoto(input: $input, file: $file) { id url } }",
            "variables": ["input": try JSONSerialization.jsonObject(with: inputData), "file": NSNull()]
        ]
        let operationsData = try JSONSerialization.data(withJSONObject: operations)
        let mapData = try JSONSerialization.data(withJSONObject: ["0": ["variables.file"]])

        var body = Data()
        appendField("operations", value: String(decoding: operationsData, as: UTF8.self), boundary: boundary, to: &body)
        appendField("map", value: String(decoding: mapData, as: UTF8.self), boundary: boundary, to: &body)
        body.append(Data("--\(boundary)\r\nContent-Disposition: form-data; name=\"0\"; filename=\"\(filename)\"\r\nContent-Type: image/jpeg\r\n\r\n".utf8))
        body.append(jpeg)
        body.append(Data("\r\n--\(boundary)--\r\n".utf8))
        request.httpBody = body

        let (data, response) = try await send(request)
        guard let photo = try parse(UploadPayload.self, data: data, response: response)?.uploadPhoto else { throw HomieAPIError.invalidResponse }
        return photo
    }

    func unpushedDrawings() async throws -> [HomieDrawing] {
        guard let token = sessionToken(), !token.isEmpty else { throw HomieAPIError.unauthorized }
        let (data, response) = try await perform(
            query: "query UnpushedDrawings { unpushedDrawings { id url takenAt coordinate { latitude longitude } } }",
            variables: [:],
            token: token
        )
        guard let drawings = try parse(DrawingsPayload.self, data: data, response: response)?.unpushedDrawings else {
            throw HomieAPIError.invalidResponse
        }
        return drawings
    }

    func pushedDrawingAssetIDs() async throws -> Set<String> {
        guard let token = sessionToken(), !token.isEmpty else { throw HomieAPIError.unauthorized }
        let (data, response) = try await perform(
            query: "query PushedDrawingAssetIDs { pushedDrawingAssetIDs }",
            variables: [:],
            token: token
        )
        guard let ids = try parse(PushedIDsPayload.self, data: data, response: response)?.pushedDrawingAssetIDs else {
            throw HomieAPIError.invalidResponse
        }
        return Set(ids)
    }

    func downloadDrawing(_ drawing: HomieDrawing) async throws -> Data {
        guard let token = sessionToken(), !token.isEmpty else { throw HomieAPIError.unauthorized }
        guard let url = URL(string: drawing.url, relativeTo: endpoint)?.absoluteURL else { throw HomieAPIError.invalidResponse }
        var request = URLRequest(url: url)
        request.setValue("PhotoSync/1", forHTTPHeaderField: "User-Agent")
        request.setValue("homie_session=\(token)", forHTTPHeaderField: "Cookie")
        let (data, response) = try await send(request)
        if response.statusCode == 401 { throw HomieAPIError.unauthorized }
        guard (200..<300).contains(response.statusCode) else { throw HomieAPIError.http(response.statusCode) }
        return data
    }

    func markDrawingPushed(id: String, assetID: String) async throws {
        guard let token = sessionToken(), !token.isEmpty else { throw HomieAPIError.unauthorized }
        let (data, response) = try await perform(
            query: "mutation MarkDrawingPushed($id: ID!, $assetID: ID!) { markDrawingPushed(id: $id, assetID: $assetID) }",
            variables: ["id": id, "assetID": assetID],
            token: token
        )
        guard try parse(MarkPushedPayload.self, data: data, response: response)?.markDrawingPushed == true else {
            throw HomieAPIError.invalidResponse
        }
    }

    static func sessionToken(fromSetCookieHeader header: String) -> String? {
        header.split(separator: ";", omittingEmptySubsequences: true)
            .lazy
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .first { $0.hasPrefix("homie_session=") }
            .map { String($0.dropFirst("homie_session=".count)) }
            .flatMap { $0.isEmpty ? nil : $0 }
    }

    static func error(forGraphQLMessage message: String) -> HomieAPIError {
        switch message.lowercased() {
        case "unauthorized": .unauthorized
        case "forbidden": .forbidden
        default: .graphQL(message)
        }
    }

    private static func sessionToken(from response: HTTPURLResponse) -> String? {
        guard let header = response.value(forHTTPHeaderField: "Set-Cookie") else { return nil }
        return sessionToken(fromSetCookieHeader: header)
    }

    private func perform(query: String, variables: [String: String], token: String?) async throws -> (Data, HTTPURLResponse) {
        var request = request(token: token)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: ["query": query, "variables": variables])
        return try await send(request)
    }

    private func request(token: String?) -> URLRequest {
        var request = URLRequest(url: endpoint)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue("PhotoSync/1", forHTTPHeaderField: "User-Agent")
        if let token, !token.isEmpty {
            request.setValue("homie_session=\(token)", forHTTPHeaderField: "Cookie")
        }
        return request
    }

    private func send(_ request: URLRequest) async throws -> (Data, HTTPURLResponse) {
        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: request)
        } catch {
            throw HomieAPIError.network(message: error.localizedDescription, endpoint: endpoint)
        }
        guard let response = response as? HTTPURLResponse else { throw HomieAPIError.invalidResponse }
        guard !data.isEmpty else { throw HomieAPIError.emptyResponse(status: response.statusCode, endpoint: endpoint) }
        return (data, response)
    }

    private func parse<Payload: Decodable>(_ type: Payload.Type, data: Data, response: HTTPURLResponse) throws -> Payload? {
        let decoded: GraphQLResponse<Payload>
        do {
            decoded = try JSONDecoder().decode(GraphQLResponse<Payload>.self, from: data)
        } catch {
            guard (200..<300).contains(response.statusCode) else { throw HomieAPIError.http(response.statusCode) }
            let body = String(decoding: data.prefix(512), as: UTF8.self)
            throw HomieAPIError.unexpectedResponse(status: response.statusCode, body: body, reason: error.localizedDescription)
        }
        if let message = decoded.errors?.first?.message {
            throw Self.error(forGraphQLMessage: message)
        }
        guard (200..<300).contains(response.statusCode) else { throw HomieAPIError.http(response.statusCode) }
        return decoded.data
    }

    private func appendField(_ name: String, value: String, boundary: String, to data: inout Data) {
        data.append(Data("--\(boundary)\r\nContent-Disposition: form-data; name=\"\(name)\"\r\n\r\n\(value)\r\n".utf8))
    }
}

private struct EmptyPayload: Decodable {}

enum HomieAPIError: LocalizedError {
    case invalidResponse
    case missingSessionCookie
    case unauthorized
    case forbidden
    case emptyResponse(status: Int, endpoint: URL)
    case http(Int)
    case graphQL(String)
    case network(message: String, endpoint: URL)
    case unexpectedResponse(status: Int, body: String, reason: String)

    var errorDescription: String? {
        switch self {
        case .invalidResponse: "The server returned an invalid response."
        case .missingSessionCookie: "The server signed in successfully but did not return a session."
        case .unauthorized: "Your Homie session has expired. Sign in again."
        case .forbidden: "This account is not allowed to sync photos. Use an admin account."
        case .emptyResponse(let status, let endpoint):
            "The server returned an empty response (HTTP \(status)). Is the API running at \(endpoint.absoluteString)?"
        case .http(let status): "The server returned HTTP \(status)."
        case .graphQL(let message): message
        case .network(let message, let endpoint): "Could not reach \(endpoint.absoluteString): \(message)"
        case .unexpectedResponse(let status, let body, let reason):
            "Unexpected server response (HTTP \(status)): \(reason). Body: \(body)"
        }
    }
}
