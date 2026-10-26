import XCTest
@testable import PhotoSync

final class HomieHostTests: XCTestCase {
    func testBuildsManualHostFromBareHostname() {
        let host = HomieHost.manual("kitchen.local:3000")

        XCTAssertEqual(host?.graphQLURL.absoluteString, "http://kitchen.local:3000/graphql")
        XCTAssertEqual(host?.address, "kitchen.local:3000")
        XCTAssertEqual(host?.source, .manual)
    }

    func testKeepsExplicitGraphQLPath() {
        let host = HomieHost.manual("https://example.test/api/graphql")

        XCTAssertEqual(host?.graphQLURL.absoluteString, "https://example.test/api/graphql")
        XCTAssertEqual(host?.address, "example.test")
    }

    func testConfiguredHostReadsExistingOverride() {
        let defaults = UserDefaults(suiteName: "HomieHostTests")!
        defaults.removePersistentDomain(forName: "HomieHostTests")
        defaults.set("http://kitchen.local:3000/graphql", forKey: "HomieAPIURL")

        let host = HomieHost.configured(defaults: defaults)

        XCTAssertEqual(host.graphQLURL.absoluteString, "http://kitchen.local:3000/graphql")
        XCTAssertEqual(host.source, .manual)
    }

    func testEquivalentLocalNamesShareASessionAccount() {
        let short = URL(string: "http://homie:3000/graphql")!
        let local = URL(string: "http://homie.local.:3000/graphql")!

        XCTAssertEqual(HomieSession.sessionAccount(for: short), HomieSession.sessionAccount(for: local))
    }

    func testEquivalentLocalNamesHaveTheSameDiscoveryIdentity() {
        let short = HomieHost(name: "Short", graphQLURL: URL(string: "http://homie:3000/graphql")!, source: .fallback)
        let local = HomieHost(name: "Bonjour", graphQLURL: URL(string: "http://homie.local.:3000/graphql")!, source: .discovered)

        XCTAssertEqual(short.id, local.id)
    }

    func testDifferentSchemesDoNotShareASessionAccount() {
        let http = URL(string: "http://homie.local:3000/graphql")!
        let https = URL(string: "https://homie.local:3000/graphql")!

        XCTAssertNotEqual(HomieSession.sessionAccount(for: http), HomieSession.sessionAccount(for: https))
    }

    @MainActor
    func testValidatedSelectionIsPersisted() async throws {
        let defaults = defaults(suite: "HomieHostSelectionTests")
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "OK", version: "1.2.3") }
        let host = HomieHost(name: "Homie · Kitchen", graphQLURL: URL(string: "http://kitchen.local:3000/graphql")!, source: .discovered)

        try await store.select(host)

        XCTAssertEqual(store.selected, host)
        XCTAssertEqual(store.health, HomieHealth(status: "OK", version: "1.2.3"))
        XCTAssertEqual(defaults.string(forKey: "HomieAPIURL"), host.graphQLURL.absoluteString)
        XCTAssertEqual(defaults.string(forKey: "HomieAPIName"), "Homie · Kitchen")
        XCTAssertTrue(defaults.bool(forKey: "HomieAPIDiscovered"))
    }

    @MainActor
    func testSetupRequiredDoesNotReplaceSelection() async {
        let defaults = defaults(suite: "HomieHostSetupTests")
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "SETUP_REQUIRED", version: "1.2.3") }
        let original = store.selected
        let candidate = HomieHost.manual("other.local:3000")!

        do {
            try await store.select(candidate)
            XCTFail("Expected setup-required selection to fail")
        } catch {
            XCTAssertEqual(error.localizedDescription, "This Homie needs to be set up first. Open its manager to finish setup.")
        }
        XCTAssertEqual(store.selected, original)
        XCTAssertNil(defaults.string(forKey: "HomieAPIURL"))
    }

    @MainActor
    func testSetupRequiredIsExposedForEquivalentDiscoveredHost() async {
        let defaults = defaults(suite: "HomieHostFallbackSetupTests")
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "SETUP_REQUIRED", version: "1.2.3") }
        let discovered = HomieHost(name: "Homie", graphQLURL: URL(string: "http://homie.local:3000/graphql")!, source: .fallback)

        try? await store.select(discovered)

        XCTAssertTrue(store.requiresSetup)
        XCTAssertNil(store.health)
    }

    @MainActor
    func testFailedSelectionKeepsExistingHost() async {
        enum TestError: Error { case offline }
        let defaults = defaults(suite: "HomieHostFailureTests")
        let store = HomieHostStore(defaults: defaults) { _ in throw TestError.offline }
        let original = store.selected

        do {
            try await store.select(HomieHost.manual("offline.local:3000")!)
            XCTFail("Expected selection to fail")
        } catch {}

        XCTAssertEqual(store.selected, original)
        XCTAssertNil(defaults.string(forKey: "HomieAPIURL"))
    }

    @MainActor
    func testResetToDefaultClearsPersistedHost() async throws {
        let defaults = defaults(suite: "HomieHostResetTests")
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "OK", version: "1.2.3") }
        try await store.select(HomieHost.manual("other.local:3000")!)
        let previousIdentity = store.contentIdentity

        store.resetToDefault()

        XCTAssertEqual(store.selected.graphQLURL, HomieAPI.defaultEndpoint)
        XCTAssertNotEqual(store.contentIdentity, previousIdentity)
        XCTAssertNil(store.health)
        XCTAssertFalse(store.requiresSetup)
        XCTAssertNil(defaults.string(forKey: "HomieAPIURL"))
        XCTAssertNil(defaults.string(forKey: "HomieAPIName"))
        XCTAssertNil(defaults.object(forKey: "HomieAPIDiscovered"))
    }

    private func defaults(suite: String) -> UserDefaults {
        let defaults = UserDefaults(suiteName: suite)!
        defaults.removePersistentDomain(forName: suite)
        return defaults
    }
}
