import Testing
@testable import PhotoSync

struct ArtworkAlbumsTests {
    private static func album(_ id: String, count: Int = 3, shared: Bool = false, canAdd: Bool = true) -> PhotoAlbum {
        PhotoAlbum(id: id, title: id, count: count, isShared: shared, canAddContent: canAdd, recency: nil, cover: nil)
    }

    private let library = [
        Self.album("pets"),
        Self.album("Art"),
        Self.album("empty", count: 0),
        Self.album("family", shared: true),
        Self.album("locked", canAdd: false)
    ]

    @Test func listsOnlyWritableOwnedAlbumsAlphabetically() {
        let artwork = ArtworkAlbums(albums: library, storedID: nil, enabled: true)

        #expect(artwork.candidates.map(\.id) == ["Art", "empty", "pets"])
    }

    @Test func emptyAlbumsAreCandidates() {
        let artwork = ArtworkAlbums(albums: library, storedID: "empty", enabled: true)

        #expect(artwork.candidates.contains { $0.id == "empty" })
        #expect(artwork.status == .ready)
    }

    @Test func searchNarrowsCandidatesIgnoringCase() {
        let artwork = ArtworkAlbums(albums: library, storedID: nil, enabled: true, search: "PE")

        #expect(artwork.candidates.map(\.id) == ["pets"])
    }

    @Test(arguments: [(nil as String?, ArtworkAlbums.Status.needsChoice), ("pets", .ready), ("gone", .unavailable), ("family", .unavailable), ("locked", .unavailable)])
    func statusReflectsTheStoredAlbumWhenEnabled(storedID: String?, expected: ArtworkAlbums.Status) {
        #expect(ArtworkAlbums(albums: library, storedID: storedID, enabled: true).status == expected)
    }

    @Test(arguments: [nil as String?, "pets", "gone"])
    func switchedOffIsAlwaysQuiet(storedID: String?) {
        #expect(ArtworkAlbums(albums: library, storedID: storedID, enabled: false).status == .off)
    }
}
