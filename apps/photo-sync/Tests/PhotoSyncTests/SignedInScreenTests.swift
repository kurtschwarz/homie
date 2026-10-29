import XCTest
@testable import PhotoSync

final class SignedInScreenTests: XCTestCase {
    func testStatusWithoutSelectionShowsEmptyState() {
        XCTAssertEqual(SignedInScreen.resolve(pane: .status, hasSelection: false), .emptyStatus)
    }

    func testStatusWithSelectionShowsSyncStatus() {
        XCTAssertEqual(SignedInScreen.resolve(pane: .status, hasSelection: true), .status)
    }

    func testAlbumsAlwaysShowsPicker() {
        XCTAssertEqual(SignedInScreen.resolve(pane: .albums, hasSelection: false), .picker)
        XCTAssertEqual(SignedInScreen.resolve(pane: .albums, hasSelection: true), .picker)
    }

    func testOtherPanesShowPlaceholder() {
        XCTAssertEqual(SignedInScreen.resolve(pane: .general, hasSelection: true), .placeholder)
    }
}
