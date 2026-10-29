import CoreLocation
import Foundation
import Testing
@testable import PhotoSync

@MainActor
struct UploadLedgerTests {
    private let defaults = TestDefaults.make()
    private let coordinate = CLLocationCoordinate2D(latitude: 40.7, longitude: -74)

    @Test func clearPreventsExistingLedgersFromWritingStaleState() {
        let staleLedger = UploadLedger(defaults: defaults)
        let resettingLedger = UploadLedger(defaults: defaults)

        resettingLedger.clear()
        staleLedger.savePlace("New York", for: coordinate)

        #expect(UploadLedger(defaults: defaults).place(for: coordinate) == nil)
    }

    @Test func clearedLedgerCanStoreNewState() {
        let ledger = UploadLedger(defaults: defaults)

        ledger.clear()
        ledger.savePlace("New York", for: coordinate)

        #expect(ledger.place(for: coordinate) == "New York")
    }
}
