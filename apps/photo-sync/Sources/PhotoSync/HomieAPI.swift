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

struct GraphQLResponse: Decodable {
    struct Payload: Decodable {
        let uploadPhoto: UploadedPhoto?

        private enum CodingKeys: String, CodingKey { case uploadPhoto }

        init(from decoder: Decoder) throws {
            let container = try decoder.container(keyedBy: CodingKeys.self)
            uploadPhoto = try container.decodeIfPresent(UploadedPhoto.self, forKey: .uploadPhoto)
        }
    }

    struct ErrorItem: Decodable { let message: String }
    let data: Payload?
    let errors: [ErrorItem]?

    private enum CodingKeys: String, CodingKey { case data, errors }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        data = try container.decodeIfPresent(Payload.self, forKey: .data)
        errors = try container.decodeIfPresent([ErrorItem].self, forKey: .errors)
    }
}

struct HomieAPI {
    // Local development endpoint for this test app.
    private let endpoint = URL(string: "http://localhost:3030/graphql")!

    func upload(_ input: PhotoUploadInput, jpeg: Data, filename: String) async throws -> UploadedPhoto {
        let boundary = "HomiePhotoSync-\(UUID().uuidString)"
        var request = URLRequest(url: endpoint)
        request.httpMethod = "POST"
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")

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

        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await URLSession.shared.data(for: request)
        } catch {
            throw HomieAPIError.network(error.localizedDescription)
        }
        guard let response = response as? HTTPURLResponse else { throw HomieAPIError.invalidResponse }
        guard !data.isEmpty else { throw HomieAPIError.emptyResponse(status: response.statusCode) }
        let result: GraphQLResponse
        do {
            result = try JSONDecoder().decode(GraphQLResponse.self, from: data)
        } catch {
            let body = String(decoding: data.prefix(512), as: UTF8.self)
            throw HomieAPIError.unexpectedResponse(status: response.statusCode, body: body, reason: error.localizedDescription)
        }
        if let message = result.errors?.first?.message { throw HomieAPIError.graphQL(message) }
        guard (200..<300).contains(response.statusCode), let photo = result.data?.uploadPhoto else {
            throw HomieAPIError.http(response.statusCode)
        }
        return photo
    }

    private func appendField(_ name: String, value: String, boundary: String, to data: inout Data) {
        data.append(Data("--\(boundary)\r\nContent-Disposition: form-data; name=\"\(name)\"\r\n\r\n\(value)\r\n".utf8))
    }
}

enum HomieAPIError: LocalizedError {
    case invalidResponse
    case emptyResponse(status: Int)
    case http(Int)
    case graphQL(String)
    case network(String)
    case unexpectedResponse(status: Int, body: String, reason: String)

    var errorDescription: String? {
        switch self {
        case .invalidResponse: "The server returned an invalid response."
        case .emptyResponse(let status): "The server returned an empty response (HTTP \(status)). Is the API running at http://localhost:3030?"
        case .http(let status): "The server returned HTTP \(status)."
        case .graphQL(let message): message
        case .network(let message): "Could not reach http://localhost:3030/graphql: \(message)"
        case .unexpectedResponse(let status, let body, let reason):
            "Unexpected server response (HTTP \(status)): \(reason). Body: \(body)"
        }
    }
}
