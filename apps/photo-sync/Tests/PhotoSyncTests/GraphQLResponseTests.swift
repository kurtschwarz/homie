import XCTest
@testable import PhotoSync

final class GraphQLResponseTests: XCTestCase {
    func testDecodesNullDataAlongsideGraphQLErrors() throws {
        let body = Data(#"{"data":null,"errors":[{"message":"upload failed"}]}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<UploadPayload>.self, from: body)

        XCTAssertNil(response.data)
        XCTAssertEqual(response.errors?.first?.message, "upload failed")
    }

    func testDecodesSuccessfulUpload() throws {
        let body = Data(#"{"data":{"uploadPhoto":{"id":"photo-1","url":"/photos/photo-1.jpg?v=1"}}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<UploadPayload>.self, from: body)

        XCTAssertEqual(response.data?.uploadPhoto?.id, "photo-1")
        XCTAssertEqual(response.data?.uploadPhoto?.url, "/photos/photo-1.jpg?v=1")
    }

    func testDecodesUnpushedDrawings() throws {
        let body = Data(#"{"data":{"unpushedDrawings":[{"id":"d1","url":"/photos/d1.jpg?v=1","takenAt":1761000000,"coordinate":{"latitude":45.4,"longitude":-75.7}},{"id":"d2","url":"/photos/d2.jpg","takenAt":1761000100,"coordinate":null}]}}"#.utf8)

        let drawings = try JSONDecoder().decode(GraphQLResponse<DrawingsPayload>.self, from: body).data?.unpushedDrawings

        XCTAssertEqual(drawings?.map(\.id), ["d1", "d2"])
        XCTAssertEqual(drawings?.first?.coordinate, HomieDrawing.Coordinate(latitude: 45.4, longitude: -75.7))
        XCTAssertNil(drawings?.last?.coordinate)
    }

    func testDecodesGraphQLErrorWithNullUploadField() throws {
        let body = Data(#"{"data":{"uploadPhoto":null},"errors":[{"message":"upload photo: width and height must be positive"}]}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<UploadPayload>.self, from: body)

        XCTAssertNil(response.data?.uploadPhoto)
        XCTAssertEqual(response.errors?.first?.message, "upload photo: width and height must be positive")
    }
}
