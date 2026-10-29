import Foundation
import Testing
@testable import PhotoSync

@MainActor
struct PushedDrawingSkipTests {
    private struct Asset: Identifiable { let id: String }

    private let album = [Asset(id: "photo-1"), Asset(id: "drawing-1"), Asset(id: "photo-2")]
    private let defaults = TestDefaults.make()

    @Test func skipsAssetsHomieKnowsAreDrawings() {
        let result = PhotoUploader.uploadable(album, excluding: ["drawing-1"])

        #expect(result.map(\.id) == ["photo-1", "photo-2"])
    }

    @Test func skipListDoesNotDependOnTheLocalLedgerSoResetDoesNotDuplicateArtwork() {
        var ledger = UploadLedger(defaults: defaults)
        ledger.savePushedDrawing("d1", assetID: "drawing-1")

        ledger.clear()
        let result = PhotoUploader.uploadable(album, excluding: ["drawing-1"])

        #expect(ledger.pushedAssetID(forDrawing: "d1") == nil)
        #expect(result.map(\.id) == ["photo-1", "photo-2"])
    }

    @Test func uploadsADrawingAsAnOrdinaryPhotoOnceHomieForgetsIt() {
        let result = PhotoUploader.uploadable(album, excluding: [])

        #expect(result.map(\.id) == ["photo-1", "drawing-1", "photo-2"])
    }
}
