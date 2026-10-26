import XCTest
@testable import PhotoSync

final class HomieAPIAuthTests: XCTestCase {
    func testDecodesLoginResponse() throws {
        let body = Data(#"{"data":{"login":{"id":"user-1","username":"family","role":"ADMIN"}}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<LoginPayload>.self, from: body)

        XCTAssertEqual(response.data?.login, HomieUser(id: "user-1", username: "family", role: .admin))
    }

    func testDecodesAuthenticatedSession() throws {
        let body = Data(#"{"data":{"session":{"user":{"id":"user-1","username":"family","role":"ADMIN"}}}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<SessionPayload>.self, from: body)

        XCTAssertEqual(response.data?.session?.user?.username, "family")
    }

    func testDecodesMissingSession() throws {
        let body = Data(#"{"data":{"session":null}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<SessionPayload>.self, from: body)

        XCTAssertNil(response.data?.session)
    }

    func testExtractsSessionTokenFromSetCookie() {
        let header = "homie_session=abc123; Path=/; Expires=Wed, 25 Nov 2026 12:00:00 GMT; HttpOnly; SameSite=Lax"

        XCTAssertEqual(HomieAPI.sessionToken(fromSetCookieHeader: header), "abc123")
    }

    func testRejectsAnEmptySessionCookie() {
        XCTAssertNil(HomieAPI.sessionToken(fromSetCookieHeader: "homie_session=; Path=/; HttpOnly"))
    }

    func testIgnoresUnrelatedCookies() {
        XCTAssertNil(HomieAPI.sessionToken(fromSetCookieHeader: "other=value; Path=/"))
    }

    func testClassifiesUnauthorizedGraphQLError() {
        guard case .unauthorized = HomieAPI.error(forGraphQLMessage: "unauthorized") else {
            return XCTFail("Expected an unauthorized error")
        }
    }

    func testClassifiesForbiddenGraphQLError() {
        guard case .forbidden = HomieAPI.error(forGraphQLMessage: "forbidden") else {
            return XCTFail("Expected a forbidden error")
        }
    }

    func testPreservesOtherGraphQLErrors() {
        guard case .graphQL(let message) = HomieAPI.error(forGraphQLMessage: "invalid username or password") else {
            return XCTFail("Expected a GraphQL error")
        }
        XCTAssertEqual(message, "invalid username or password")
    }
}
