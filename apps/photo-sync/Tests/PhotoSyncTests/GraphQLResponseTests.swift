import XCTest
@testable import PhotoSync

final class GraphQLResponseTests: XCTestCase {
    func testDecodesNullDataAlongsideGraphQLErrors() throws {
        let body = Data(#"{"data":null,"errors":[{"message":"upload failed"}]}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse.self, from: body)

        XCTAssertNil(response.data)
        XCTAssertEqual(response.errors?.first?.message, "upload failed")
    }

    func testDecodesSuccessfulUpload() throws {
        let body = Data(#"{"data":{"uploadPhoto":{"id":"photo-1","url":"/photos/photo-1.jpg?v=1"}}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse.self, from: body)

        XCTAssertEqual(response.data?.uploadPhoto?.id, "photo-1")
        XCTAssertEqual(response.data?.uploadPhoto?.url, "/photos/photo-1.jpg?v=1")
    }

    func testDecodesGraphQLErrorWithNullUploadField() throws {
        let body = Data(#"{"data":{"uploadPhoto":null},"errors":[{"message":"upload photo: width and height must be positive"}]}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse.self, from: body)

        XCTAssertNil(response.data?.uploadPhoto)
        XCTAssertEqual(response.errors?.first?.message, "upload photo: width and height must be positive")
    }
}
