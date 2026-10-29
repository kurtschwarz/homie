import CoreLocation
import Foundation
import Testing
@testable import PhotoSync

@MainActor
struct UploadLedgerTests {
    private let defaults = TestDefaults.make()
    private let coordinate = CLLocationCoordinate2D(latitude: 40.7, longitude: -74)

    @Test func clearPreventsExistingLedgersFromWritingStaleState() {
        var staleLedger = UploadLedger(defaults: defaults)
        var resettingLedger = UploadLedger(defaults: defaults)

        resettingLedger.clear()
        staleLedger.savePlace("New York", for: coordinate)
        staleLedger.flush()

        #expect(UploadLedger(defaults: defaults).place(for: coordinate) == nil)
    }

    @Test func clearedLedgerCanStoreNewState() {
        var ledger = UploadLedger(defaults: defaults)

        ledger.clear()
        ledger.savePlace("New York", for: coordinate)

        #expect(ledger.place(for: coordinate) == "New York")
    }

    @Test func flushedStateSurvivesReload() {
        var ledger = UploadLedger(defaults: defaults)
        ledger.savePlace("New York", for: coordinate)
        ledger.savePushedDrawing("d1", assetID: "asset-1")
        ledger.savePushedDrawing("d2", assetID: "asset-2")
        ledger.forgetPushedDrawing("d2")
        ledger.flush()

        let reloaded = UploadLedger(defaults: defaults)

        #expect(reloaded.place(for: coordinate) == "New York")
        #expect(reloaded.pushedAssetID(forDrawing: "d1") == "asset-1")
        #expect(reloaded.pushedAssetID(forDrawing: "d2") == nil)
    }

    @Test func unflushedStateIsNotPersisted() {
        var ledger = UploadLedger(defaults: defaults)
        ledger.savePlace("New York", for: coordinate)

        #expect(UploadLedger(defaults: defaults).place(for: coordinate) == nil)
    }

    @Test func generationSurvivesReloadSoReloadedLedgerCanStillWrite() {
        var resettingLedger = UploadLedger(defaults: defaults)
        resettingLedger.clear()
        var reloaded = UploadLedger(defaults: defaults)

        reloaded.savePlace("Boston", for: coordinate)
        reloaded.flush()

        #expect(UploadLedger(defaults: defaults).place(for: coordinate) == "Boston")
    }
}
