import XCTest
@testable import PhotoSync

final class OnboardingResetTests: XCTestCase {
    @MainActor
    func testResetClearsLocalStateAndRevokesCurrentSession() async {
        let defaults = UserDefaults(suiteName: "OnboardingResetTests")!
        defaults.removePersistentDomain(forName: "OnboardingResetTests")
        defaults.set("http://localhost:3030", forKey: "HomieAPIURL")
        defaults.set("Local Homie", forKey: "HomieAPIName")
        defaults.set(false, forKey: "HomieAPIDiscovered")
        defaults.set(["photo": "version"], forKey: "photo-sync.uploadedVersions")
        defaults.set(["coordinate": "place"], forKey: "photo-sync.geocodedPlaces")
        defaults.set(["album"], forKey: "photo-sync.selectedAlbums")
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

        XCTAssertTrue(didClearSessions)
        XCTAssertEqual(logout?.endpoint.absoluteString, "http://localhost:3030/graphql")
        XCTAssertEqual(logout?.token, "session-token")
        XCTAssertEqual(store.selected.graphQLURL, HomieAPI.defaultEndpoint)
        XCTAssertNil(defaults.object(forKey: "photo-sync.uploadedVersions"))
        XCTAssertNil(defaults.object(forKey: "photo-sync.geocodedPlaces"))
        XCTAssertNil(defaults.object(forKey: "photo-sync.selectedAlbums"))
    }
}
