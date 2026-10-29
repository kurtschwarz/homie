import Foundation
import Testing
@testable import PhotoSync

@MainActor
struct OnboardingResetTests {
    private let defaults = TestDefaults.make()

    @Test func resetClearsLocalStateAndRevokesCurrentSession() async {
        defaults.set("http://localhost:3030", forKey: "HomieAPIURL")
        defaults.set("Local Homie", forKey: "HomieAPIName")
        defaults.set(false, forKey: "HomieAPIDiscovered")
        defaults.set(["photo": "version"], forKey: "photo-sync.uploadedVersions")
        defaults.set(["coordinate": "place"], forKey: "photo-sync.geocodedPlaces")
        defaults.set(["album"], forKey: "photo-sync.selectedAlbums")
        defaults.set("artwork", forKey: "photo-sync.artworkAlbum")
        let store = HomieHostStore(defaults: defaults)
        var didClearSessions = false
        var logout: (endpoint: URL, token: String)?
        let reset = OnboardingReset(
            defaults: defaults,
            tokenForEndpoint: { _ in "session-token" },
            clearSessions: { didClearSessions = true },
            logout: { endpoint, token in logout = (endpoint, token) }
        )

        await reset.run(hostStore: store)

        #expect(didClearSessions)
        #expect(logout?.endpoint.absoluteString == "http://localhost:3030/graphql")
        #expect(logout?.token == "session-token")
        #expect(store.selected.graphQLURL == HomieAPI.defaultEndpoint)
        #expect(defaults.object(forKey: "photo-sync.uploadedVersions") == nil)
        #expect(defaults.object(forKey: "photo-sync.geocodedPlaces") == nil)
        #expect(defaults.object(forKey: "photo-sync.selectedAlbums") == nil)
        #expect(defaults.object(forKey: "photo-sync.artworkAlbum") == nil)
    }
}
