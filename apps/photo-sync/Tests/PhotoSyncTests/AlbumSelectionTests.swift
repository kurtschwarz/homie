import XCTest
@testable import PhotoSync

@MainActor
final class AlbumSelectionTests: XCTestCase {
    private func makeDefaults() -> UserDefaults {
        let defaults = UserDefaults(suiteName: "AlbumSelectionTests")!
        defaults.removePersistentDomain(forName: "AlbumSelectionTests")
        return defaults
    }

    func testSelectionPersistsAcrossInstances() {
        let defaults = makeDefaults()
        let first = AlbumSelection(defaults: defaults)
        first.toggle("a")
        first.toggle("b")
        first.toggle("a")
        first.setKeepSyncing(false)
        let date = Date(timeIntervalSince1970: 1_000)
        first.markSynced(at: date)

        let second = AlbumSelection(defaults: defaults)
        XCTAssertEqual(second.albumIDs, ["b"])
        XCTAssertFalse(second.keepSyncing)
        XCTAssertEqual(second.lastSyncedAt, date)
    }

    func testArtworkAlbumPersistsAndIsIndependentOfSources() {
        let defaults = makeDefaults()
        let first = AlbumSelection(defaults: defaults)
        first.select(["a"])
        first.setArtworkAlbum("a")

        let second = AlbumSelection(defaults: defaults)
        XCTAssertEqual(second.artworkAlbumID, "a")
        XCTAssertTrue(second.contains("a"))

        second.setArtworkAlbum(nil)
        XCTAssertNil(AlbumSelection(defaults: defaults).artworkAlbumID)
    }

    func testClearRemovesArtworkAlbum() {
        let defaults = makeDefaults()
        AlbumSelection(defaults: defaults).setArtworkAlbum("a")

        AlbumSelection.clear(defaults: defaults)

        XCTAssertNil(AlbumSelection(defaults: defaults).artworkAlbumID)
    }

    func testKeepSyncingDefaultsToOn() {
        XCTAssertTrue(AlbumSelection(defaults: makeDefaults()).keepSyncing)
    }

    func testClearRemovesEverything() {
        let defaults = makeDefaults()
        let selection = AlbumSelection(defaults: defaults)
        selection.select(["a"])
        selection.markSynced()

        AlbumSelection.clear(defaults: defaults)

        let reloaded = AlbumSelection(defaults: defaults)
        XCTAssertFalse(reloaded.hasSelection)
        XCTAssertNil(reloaded.lastSyncedAt)
    }
}
