import XCTest
@testable import PhotoSync

final class PushedDrawingSkipTests: XCTestCase {
    private struct Asset: Identifiable { let id: String }

    private let album = [Asset(id: "photo-1"), Asset(id: "drawing-1"), Asset(id: "photo-2")]

    func testSkipsAssetsHomieKnowsAreDrawings() {
        let result = PhotoUploader.uploadable(album, excluding: ["drawing-1"])

        XCTAssertEqual(result.map(\.id), ["photo-1", "photo-2"])
    }

    @MainActor
    func testSkipListDoesNotDependOnTheLocalLedgerSoResetDoesNotDuplicateArtwork() {
        let defaults = UserDefaults(suiteName: "PushedDrawingSkipResetTests")!
        defaults.removePersistentDomain(forName: "PushedDrawingSkipResetTests")
        let ledger = UploadLedger(defaults: defaults)
        ledger.savePushedDrawing("d1", assetID: "drawing-1")

        ledger.clear()
        let result = PhotoUploader.uploadable(album, excluding: ["drawing-1"])

        XCTAssertNil(ledger.pushedAssetID(forDrawing: "d1"))
        XCTAssertEqual(result.map(\.id), ["photo-1", "photo-2"])
    }

    func testUploadsADrawingAsAnOrdinaryPhotoOnceHomieForgetsIt() {
        let result = PhotoUploader.uploadable(album, excluding: [])

        XCTAssertEqual(result.map(\.id), ["photo-1", "drawing-1", "photo-2"])
    }
}
