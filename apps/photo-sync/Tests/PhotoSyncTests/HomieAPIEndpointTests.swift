import XCTest
@testable import PhotoSync

final class HomieAPIEndpointTests: XCTestCase {
    private func defaults(_ url: String?) -> UserDefaults {
        let suite = UserDefaults(suiteName: "HomieAPIEndpointTests")!
        suite.removePersistentDomain(forName: "HomieAPIEndpointTests")
        if let url { suite.set(url, forKey: "HomieAPIURL") }
        return suite
    }

    func testDefaultsToThePi() {
        let endpoint = HomieAPI.configuredEndpoint(defaults: defaults(nil), environment: [:])

        XCTAssertEqual(endpoint.absoluteString, "http://homie/graphql")
    }

    func testUserDefaultsOverrideWinsOverTheEnvironment() {
        let endpoint = HomieAPI.configuredEndpoint(
            defaults: defaults("http://localhost:3030"),
            environment: ["HOMIE_API_URL": "http://elsewhere:3000"]
        )

        XCTAssertEqual(endpoint.absoluteString, "http://localhost:3030/graphql")
    }

    func testEnvironmentOverrideAppliesWithoutAUserDefault() {
        let endpoint = HomieAPI.configuredEndpoint(
            defaults: defaults(nil),
            environment: ["HOMIE_API_URL": "http://localhost:3030/"]
        )

        XCTAssertEqual(endpoint.absoluteString, "http://localhost:3030/graphql")
    }

    func testAnExplicitPathIsLeftAlone() {
        let endpoint = HomieAPI.configuredEndpoint(
            defaults: defaults("http://localhost:3030/graphql"),
            environment: [:]
        )

        XCTAssertEqual(endpoint.absoluteString, "http://localhost:3030/graphql")
    }

    func testGarbageOverrideFallsBackToTheDefault() {
        let endpoint = HomieAPI.configuredEndpoint(defaults: defaults("not a url"), environment: [:])

        XCTAssertEqual(endpoint, HomieAPI.defaultEndpoint)
    }
}
