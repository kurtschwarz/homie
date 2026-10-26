import CoreLocation
import XCTest
@testable import PhotoSync

final class UploadLedgerTests: XCTestCase {
    @MainActor
    func testClearPreventsExistingLedgersFromWritingStaleState() {
        let defaults = UserDefaults(suiteName: "UploadLedgerResetTests")!
        defaults.removePersistentDomain(forName: "UploadLedgerResetTests")
        let staleLedger = UploadLedger(defaults: defaults)
        let resettingLedger = UploadLedger(defaults: defaults)
        let coordinate = CLLocationCoordinate2D(latitude: 40.7, longitude: -74)

        resettingLedger.clear()
        staleLedger.savePlace("New York", for: coordinate)

        XCTAssertNil(UploadLedger(defaults: defaults).place(for: coordinate))
    }

    @MainActor
    func testClearedLedgerCanStoreNewState() {
        let defaults = UserDefaults(suiteName: "UploadLedgerReuseTests")!
        defaults.removePersistentDomain(forName: "UploadLedgerReuseTests")
        let ledger = UploadLedger(defaults: defaults)
        let coordinate = CLLocationCoordinate2D(latitude: 40.7, longitude: -74)

        ledger.clear()
        ledger.savePlace("New York", for: coordinate)

        XCTAssertEqual(ledger.place(for: coordinate), "New York")
    }
}
