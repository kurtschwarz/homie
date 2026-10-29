import Foundation
import Testing
@testable import PhotoSync

struct HomieAPIAuthTests {
    @Test func decodesLoginResponse() throws {
        let body = Data(#"{"data":{"login":{"id":"user-1","username":"family","role":"ADMIN"}}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<LoginPayload>.self, from: body)

        #expect(response.data?.login == HomieUser(id: "user-1", username: "family", role: .admin))
    }

    @Test func decodesAuthenticatedSession() throws {
        let body = Data(#"{"data":{"session":{"user":{"id":"user-1","username":"family","role":"ADMIN"}}}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<SessionPayload>.self, from: body)

        #expect(response.data?.session?.user?.username == "family")
    }

    @Test func decodesMissingSession() throws {
        let body = Data(#"{"data":{"session":null}}"#.utf8)

        let response = try JSONDecoder().decode(GraphQLResponse<SessionPayload>.self, from: body)

        #expect(response.data?.session == nil)
    }

    @Test func extractsSessionTokenFromSetCookie() {
        let header = "homie_session=abc123; Path=/; Expires=Wed, 25 Nov 2026 12:00:00 GMT; HttpOnly; SameSite=Lax"

        #expect(HomieAPI.sessionToken(fromSetCookieHeader: header) == "abc123")
    }

    @Test func rejectsAnEmptySessionCookie() {
        #expect(HomieAPI.sessionToken(fromSetCookieHeader: "homie_session=; Path=/; HttpOnly") == nil)
    }

    @Test func ignoresUnrelatedCookies() {
        #expect(HomieAPI.sessionToken(fromSetCookieHeader: "other=value; Path=/") == nil)
    }

    @Test func classifiesUnauthorizedGraphQLError() {
        guard case .unauthorized = HomieAPI.error(forGraphQLMessage: "unauthorized") else {
            Issue.record("Expected an unauthorized error")
            return
        }
    }

    @Test func classifiesForbiddenGraphQLError() {
        guard case .forbidden = HomieAPI.error(forGraphQLMessage: "forbidden") else {
            Issue.record("Expected a forbidden error")
            return
        }
    }

    @Test func preservesOtherGraphQLErrors() {
        guard case .graphQL(let message) = HomieAPI.error(forGraphQLMessage: "invalid username or password") else {
            Issue.record("Expected a GraphQL error")
            return
        }
        #expect(message == "invalid username or password")
    }
}
