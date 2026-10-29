import Foundation
import Testing
@testable import PhotoSync

struct AlbumSelectionTests {
    private let defaults = TestDefaults.make()

    @Test func selectionPersistsAcrossInstances() {
        let first = AlbumSelection(defaults: defaults)
        first.toggle("a")
        first.toggle("b")
        first.toggle("a")
        first.setKeepSyncing(false)
        let date = Date(timeIntervalSince1970: 1_000)
        first.markSynced(at: date)

        let second = AlbumSelection(defaults: defaults)
        #expect(second.albumIDs == ["b"])
        #expect(!second.keepSyncing)
        #expect(second.lastSyncedAt == date)
    }

    @Test func artworkAlbumPersistsAndIsIndependentOfSources() {
        let first = AlbumSelection(defaults: defaults)
        first.select(["a"])
        first.setArtworkAlbum("a")

        let second = AlbumSelection(defaults: defaults)
        #expect(second.artworkAlbumID == "a")
        #expect(second.contains("a"))

        second.setArtworkAlbum(nil)
        #expect(AlbumSelection(defaults: defaults).artworkAlbumID == nil)
    }

    @Test func clearRemovesArtworkAlbum() {
        AlbumSelection(defaults: defaults).setArtworkAlbum("a")

        AlbumSelection.clear(defaults: defaults)

        #expect(AlbumSelection(defaults: defaults).artworkAlbumID == nil)
    }

    @Test func keepSyncingDefaultsToOn() {
        #expect(AlbumSelection(defaults: defaults).keepSyncing)
    }

    @Test func clearRemovesEverything() {
        let selection = AlbumSelection(defaults: defaults)
        selection.select(["a"])
        selection.markSynced()

        AlbumSelection.clear(defaults: defaults)

        let reloaded = AlbumSelection(defaults: defaults)
        #expect(!reloaded.hasSelection)
        #expect(reloaded.lastSyncedAt == nil)
    }
}
