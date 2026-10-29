import Foundation
import Testing
@testable import PhotoSync

struct GraphQLResponseTests {
    @Test func decodesNullDataAlongsideGraphQLErrors() throws {
        let body = Data(#"{"data":null,"errors":[{"message":"upload failed"}]}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<UploadPayload>.self, from: body)

        #expect(response.data == nil)
        #expect(response.errors?.first?.message == "upload failed")
    }

    @Test func decodesSuccessfulUpload() throws {
        let body = Data(#"{"data":{"uploadPhoto":{"id":"photo-1","url":"/photos/photo-1.jpg?v=1"}}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<UploadPayload>.self, from: body)

        #expect(response.data?.uploadPhoto?.id == "photo-1")
        #expect(response.data?.uploadPhoto?.url == "/photos/photo-1.jpg?v=1")
    }

    @Test func decodesUnpushedDrawings() throws {
        let body = Data(#"{"data":{"unpushedDrawings":[{"id":"d1","url":"/photos/d1.jpg?v=1","takenAt":1761000000,"coordinate":{"latitude":45.4,"longitude":-75.7}},{"id":"d2","url":"/photos/d2.jpg","takenAt":1761000100,"coordinate":null}]}}"#.utf8)

        let drawings = try JSONDecoder().decode(GraphQLResponse<DrawingsPayload>.self, from: body).data?.unpushedDrawings

        #expect(drawings?.map(\.id) == ["d1", "d2"])
        #expect(drawings?.first?.coordinate == HomieDrawing.Coordinate(latitude: 45.4, longitude: -75.7))
        #expect(drawings?.last?.coordinate == nil)
    }

    @Test func decodesGraphQLErrorWithNullUploadField() throws {
        let body = Data(#"{"data":{"uploadPhoto":null},"errors":[{"message":"upload photo: width and height must be positive"}]}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<UploadPayload>.self, from: body)

        #expect(response.data?.uploadPhoto == nil)
        #expect(response.errors?.first?.message == "upload photo: width and height must be positive")
    }
}
