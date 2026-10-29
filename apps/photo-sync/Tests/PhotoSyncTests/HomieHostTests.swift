import Foundation
import Testing
@testable import PhotoSync

struct HomieHostTests {
    private let defaults = TestDefaults.make()

    @Test(arguments: [
        ("kitchen.local:3000", "http://kitchen.local:3000/graphql", "kitchen.local:3000"),
        ("https://example.test/api/graphql", "https://example.test/api/graphql", "example.test"),
    ])
    func buildsManualHost(input: String, graphQLURL: String, address: String) {
        let host = HomieHost.manual(input)

        #expect(host?.graphQLURL.absoluteString == graphQLURL)
        #expect(host?.address == address)
        #expect(host?.source == .manual)
    }

    @Test func configuredHostReadsExistingOverride() {
        defaults.set("http://kitchen.local:3000/graphql", forKey: "HomieAPIURL")

        let host = HomieHost.configured(defaults: defaults)

        #expect(host.graphQLURL.absoluteString == "http://kitchen.local:3000/graphql")
        #expect(host.source == .manual)
    }

    @Test func equivalentLocalNamesShareASessionAccount() {
        let short = URL(string: "http://homie/graphql")!
        let local = URL(string: "http://homie.local./graphql")!

        #expect(HomieSession.sessionAccount(for: short) == HomieSession.sessionAccount(for: local))
    }

    @Test func equivalentLocalNamesHaveTheSameDiscoveryIdentity() {
        let short = HomieHost(name: "Short", graphQLURL: URL(string: "http://homie/graphql")!, source: .fallback)
        let local = HomieHost(name: "Bonjour", graphQLURL: URL(string: "http://homie.local./graphql")!, source: .discovered)

        #expect(short.id == local.id)
    }

    @Test func differentSchemesDoNotShareASessionAccount() {
        let http = URL(string: "http://homie.local/graphql")!
        let https = URL(string: "https://homie.local/graphql")!

        #expect(HomieSession.sessionAccount(for: http) != HomieSession.sessionAccount(for: https))
    }

    @Test func validatedSelectionIsPersisted() async throws {
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "OK", version: "1.2.3") }
        let host = HomieHost(name: "Homie · Kitchen", graphQLURL: URL(string: "http://kitchen.local:3000/graphql")!, source: .discovered)

        try await store.select(host)

        #expect(store.selected == host)
        #expect(store.health == HomieHealth(status: "OK", version: "1.2.3"))
        #expect(defaults.string(forKey: "HomieAPIURL") == host.graphQLURL.absoluteString)
        #expect(defaults.string(forKey: "HomieAPIName") == "Homie · Kitchen")
        #expect(defaults.bool(forKey: "HomieAPIDiscovered"))
    }

    @Test func setupRequiredDoesNotReplaceSelection() async {
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "SETUP_REQUIRED", version: "1.2.3") }
        let original = store.selected
        let candidate = HomieHost.manual("other.local:3000")!

        do {
            try await store.select(candidate)
            Issue.record("Expected setup-required selection to fail")
        } catch {
            #expect(error.localizedDescription == "This Homie needs to be set up first. Open its manager to finish setup.")
        }
        #expect(store.selected == original)
        #expect(defaults.string(forKey: "HomieAPIURL") == nil)
    }

    @Test func setupRequiredIsExposedForEquivalentDiscoveredHost() async {
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "SETUP_REQUIRED", version: "1.2.3") }
        let discovered = HomieHost(name: "Homie", graphQLURL: URL(string: "http://homie.local/graphql")!, source: .fallback)

        try? await store.select(discovered)

        #expect(store.requiresSetup)
        #expect(store.health == nil)
    }

    @Test func failedSelectionKeepsExistingHost() async {
        enum TestError: Error { case offline }
        let store = HomieHostStore(defaults: defaults) { _ in throw TestError.offline }
        let original = store.selected

        do {
            try await store.select(HomieHost.manual("offline.local:3000")!)
            Issue.record("Expected selection to fail")
        } catch {}

        #expect(store.selected == original)
        #expect(defaults.string(forKey: "HomieAPIURL") == nil)
    }

    @Test func resetToDefaultClearsPersistedHost() async throws {
        let store = HomieHostStore(defaults: defaults) { _ in HomieHealth(status: "OK", version: "1.2.3") }
        try await store.select(HomieHost.manual("other.local:3000")!)
        let previousIdentity = store.contentIdentity

        store.resetToDefault()

        #expect(store.selected.graphQLURL == HomieAPI.defaultEndpoint)
        #expect(store.contentIdentity != previousIdentity)
        #expect(store.health == nil)
        #expect(!store.requiresSetup)
        #expect(defaults.string(forKey: "HomieAPIURL") == nil)
        #expect(defaults.string(forKey: "HomieAPIName") == nil)
        #expect(defaults.object(forKey: "HomieAPIDiscovered") == nil)
    }
}
