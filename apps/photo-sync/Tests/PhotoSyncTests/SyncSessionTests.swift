import Testing
@testable import PhotoSync

struct SyncSessionTests {
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

    @Test func pollsWhenReady() {
        #expect(poll())
    }

    @Test func neverOverlapsARunningSync() {
        #expect(!poll(uploading: true))
    }

    @Test func skipsWhenSignedOut() {
        #expect(!poll(signedIn: false))
    }

    @Test func skipsWithoutPhotosAccess() {
        #expect(!poll(access: false))
    }

    @Test func skipsWithoutSelection() {
        #expect(!poll(selection: false))
    }

    @Test func skipsWhenKeepSyncingIsOff() {
        #expect(!poll(keepSyncing: false))
    }

    private func needsWindow(
        permissions: Bool = false,
        signIn: Bool = false,
        selection: Bool = true
    ) -> Bool {
        SyncSession.needsWindow(showsPermissions: permissions, needsSignIn: signIn, hasSelection: selection)
    }

    @Test func staysInTheMenuBarOnceSetUp() {
        #expect(!needsWindow())
    }

    @Test func opensTheWindowWhenPermissionsAreMissing() {
        #expect(needsWindow(permissions: true))
    }

    @Test func opensTheWindowWhenSignInIsNeeded() {
        #expect(needsWindow(signIn: true))
    }

    @Test func opensTheWindowWithoutSelection() {
        #expect(needsWindow(selection: false))
    }
}
