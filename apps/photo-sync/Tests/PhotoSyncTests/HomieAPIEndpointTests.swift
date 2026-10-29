import Foundation
import Testing
@testable import PhotoSync

struct HomieAPIEndpointTests {
    private let suite = TestDefaults.make()

    private func defaults(_ url: String?) -> UserDefaults {
        if let url { suite.set(url, forKey: "HomieAPIURL") }
        return suite
    }

    @Test func defaultsToThePi() {
        let endpoint = HomieAPI.configuredEndpoint(defaults: defaults(nil), environment: [:])

        #expect(endpoint.absoluteString == "http://homie/graphql")
    }

    @Test func userDefaultsOverrideWinsOverTheEnvironment() {
        let endpoint = HomieAPI.configuredEndpoint(
            defaults: defaults("http://localhost:3030"),
            environment: ["HOMIE_API_URL": "http://elsewhere:3000"]
        )

        #expect(endpoint.absoluteString == "http://localhost:3030/graphql")
    }

    @Test func environmentOverrideAppliesWithoutAUserDefault() {
        let endpoint = HomieAPI.configuredEndpoint(
            defaults: defaults(nil),
            environment: ["HOMIE_API_URL": "http://localhost:3030/"]
        )

        #expect(endpoint.absoluteString == "http://localhost:3030/graphql")
    }

    @Test func anExplicitPathIsLeftAlone() {
        let endpoint = HomieAPI.configuredEndpoint(
            defaults: defaults("http://localhost:3030/graphql"),
            environment: [:]
        )

        #expect(endpoint.absoluteString == "http://localhost:3030/graphql")
    }

    @Test func garbageOverrideFallsBackToTheDefault() {
        let endpoint = HomieAPI.configuredEndpoint(defaults: defaults("not a url"), environment: [:])

        #expect(endpoint == HomieAPI.defaultEndpoint)
    }
}
