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

    private func needsWindow(
        permissions: Bool = false,
        signIn: Bool = false,
        selection: Bool = true
    ) -> Bool {
        SyncSession.needsWindow(showsPermissions: permissions, needsSignIn: signIn, hasSelection: selection)
    }

    func testStaysInTheMenuBarOnceSetUp() {
        XCTAssertFalse(needsWindow())
    }

    func testOpensTheWindowWhenSetupIsUnfinished() {
        XCTAssertTrue(needsWindow(permissions: true))
        XCTAssertTrue(needsWindow(signIn: true))
        XCTAssertTrue(needsWindow(selection: false))
    }
}
