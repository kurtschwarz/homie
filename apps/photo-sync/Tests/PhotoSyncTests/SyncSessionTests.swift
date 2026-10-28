import XCTest
@testable import PhotoSync

final class SyncSessionTests: XCTestCase {
    private func poll(
        signedIn: Bool = true,
        access: Bool = true,
        selection: Bool = true,
        keepSyncing: Bool = true,
        uploading: Bool = false
    ) -> Bool {
        SyncSession.shouldPoll(
            signedIn: signedIn,
            hasPhotosAccess: access,
            hasSelection: selection,
            keepSyncing: keepSyncing,
            isUploading: uploading
        )
    }

    func testPollsWhenReady() {
        XCTAssertTrue(poll())
    }

    func testNeverOverlapsARunningSync() {
        XCTAssertFalse(poll(uploading: true))
    }

    func testSkipsWhenSignedOut() {
        XCTAssertFalse(poll(signedIn: false))
    }

    func testSkipsWithoutPhotosAccessOrSelection() {
        XCTAssertFalse(poll(access: false))
        XCTAssertFalse(poll(selection: false))
    }

    func testSkipsWhenKeepSyncingIsOff() {
        XCTAssertFalse(poll(keepSyncing: false))
    }
}
