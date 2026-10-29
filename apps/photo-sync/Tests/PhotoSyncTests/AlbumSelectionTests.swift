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
        first.toggle("a")
        first.setArtworkAlbum("a")

        let second = AlbumSelection(defaults: defaults)
        #expect(second.artworkAlbumID == "a")
        #expect(second.contains("a"))

        second.setArtworkAlbum(nil)
        #expect(AlbumSelection(defaults: defaults).artworkAlbumID == nil)
    }

    @Test func artworkSwitchDefaultsOffForFreshInstalls() {
        #expect(AlbumSelection(defaults: defaults).artworkEnabled == false)
    }

    @Test func artworkSwitchDefaultsOnForAnExistingStoredAlbum() {
        AlbumSelection(defaults: defaults).setArtworkAlbum("a")

        #expect(AlbumSelection(defaults: defaults).artworkEnabled == true)
    }

    @Test func turningTheSwitchOffKeepsTheAlbumAndPersists() {
        let first = AlbumSelection(defaults: defaults)
        first.setArtworkAlbum("a")
        first.setArtworkEnabled(false)

        let second = AlbumSelection(defaults: defaults)
        #expect(second.artworkEnabled == false)
        #expect(second.artworkAlbumID == "a")
    }

    @Test func clearRemovesArtworkSwitch() {
        AlbumSelection(defaults: defaults).setArtworkEnabled(true)

        AlbumSelection.clear(defaults: defaults)

        #expect(AlbumSelection(defaults: defaults).artworkEnabled == false)
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
        selection.toggle("a")
        selection.markSynced()

        AlbumSelection.clear(defaults: defaults)

        let reloaded = AlbumSelection(defaults: defaults)
        #expect(!reloaded.hasSelection)
        #expect(reloaded.lastSyncedAt == nil)
    }
}
