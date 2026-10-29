import Network
import Photos
import XCTest
@testable import PhotoSync

final class PermissionsGateTests: XCTestCase {
    private func gate(
        _ photos: PHAuthorizationStatus,
        localNetwork: AccessState = .allowed,
        keychain: AccessState = .allowed
    ) -> PermissionsGate {
        PermissionsGate(photos: photos, localNetwork: localNetwork, keychain: keychain)
    }

    func testSatisfiedWhenAllAllowed() {
        XCTAssertTrue(gate(.authorized).isSatisfied)
        XCTAssertTrue(gate(.limited).isSatisfied)
    }

    func testNotSatisfiedWhenAnyMissing() {
        XCTAssertFalse(gate(.notDetermined).isSatisfied)
        XCTAssertFalse(gate(.denied).isSatisfied)
        XCTAssertFalse(gate(.authorized, localNetwork: .needsPrompt).isSatisfied)
        XCTAssertFalse(gate(.authorized, localNetwork: .denied).isSatisfied)
        XCTAssertFalse(gate(.authorized, keychain: .needsPrompt).isSatisfied)
        XCTAssertFalse(gate(.authorized, keychain: .denied).isSatisfied)
    }

    func testScreenNeverShownWhenSatisfied() {
        XCTAssertFalse(PermissionsGate.showsScreen(currentlyShowing: false, gate: gate(.authorized)))
    }

    func testScreenAppearsWhenAccessIsRevoked() {
        XCTAssertTrue(PermissionsGate.showsScreen(currentlyShowing: false, gate: gate(.denied)))
        XCTAssertTrue(PermissionsGate.showsScreen(currentlyShowing: false, gate: gate(.authorized, localNetwork: .denied)))
    }

    func testScreenStaysUntilContinueEvenWhenSatisfied() {
        XCTAssertTrue(PermissionsGate.showsScreen(currentlyShowing: true, gate: gate(.authorized)))
    }

    func testContinueTitleNamesTheFirstMissingPermission() {
        XCTAssertEqual(gate(.notDetermined, localNetwork: .needsPrompt, keychain: .needsPrompt).continueTitle, "Allow Photos to continue")
        XCTAssertEqual(gate(.authorized, localNetwork: .needsPrompt, keychain: .needsPrompt).continueTitle, "Allow Local Network to continue")
        XCTAssertEqual(gate(.authorized, keychain: .needsPrompt).continueTitle, "Allow Keychain to continue")
        XCTAssertEqual(gate(.authorized).continueTitle, "Continue")
    }

    func testSidebarIsSkeletonUntilPastTheStartupScreens() {
        XCTAssertTrue(StartupPhase.permissions.hidesSidebar)
        XCTAssertTrue(StartupPhase.gettingReady.hidesSidebar)
        XCTAssertFalse(StartupPhase.signedOut.hidesSidebar)
        XCTAssertFalse(StartupPhase.signedIn.hidesSidebar)
    }

    func testStartupOrder() {
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: true, isCheckingSession: true, isSignedIn: false), .permissions)
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: false, isCheckingSession: true, isSignedIn: false), .gettingReady)
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: false, isCheckingSession: false, isSignedIn: false), .signedOut)
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: false, isCheckingSession: false, isSignedIn: true), .signedIn)
    }
}

final class KeychainGrantTests: XCTestCase {
    func testGrantHoldsForTheSameCodeSignature() {
        XCTAssertTrue(CodeSignedGrant.isCurrent(recorded: "abc", identity: "abc"))
    }

    func testGrantFromAnotherBuildIsStale() {
        XCTAssertFalse(CodeSignedGrant.isCurrent(recorded: "abc", identity: "def"))
    }

    func testNoRecordOrUnreadableSignatureIsNotAGrant() {
        XCTAssertFalse(CodeSignedGrant.isCurrent(recorded: nil, identity: "abc"))
        XCTAssertFalse(CodeSignedGrant.isCurrent(recorded: "", identity: ""))
    }
}

final class LocalNetworkProbeTests: XCTestCase {
    func testReadyBrowserSaysNothingBecauseItIsReadyWhetherOrNotAccessIsAllowed() {
        XCTAssertNil(LocalNetworkProbe.access(for: .ready))
    }

    func testFindingOurOwnAdvertisedServiceMeansAllowed() {
        XCTAssertEqual(LocalNetworkProbe.access(foundNames: ["other", "probe-1"], own: "probe-1"), .allowed)
    }

    func testNotFindingOurOwnServiceSaysNothing() {
        XCTAssertNil(LocalNetworkProbe.access(foundNames: ["other"], own: "probe-1"))
        XCTAssertNil(LocalNetworkProbe.access(foundNames: [], own: "probe-1"))
    }

    func testPolicyDenialMeansDenied() {
        let denied = NWBrowser.State.waiting(.dns(LocalNetworkProbe.policyDenied))
        XCTAssertEqual(LocalNetworkProbe.access(for: denied), .denied)
    }

    func testOtherWaitingReasonsSayNothingAboutPermission() {
        XCTAssertNil(LocalNetworkProbe.access(for: .waiting(.dns(-65537))))
        XCTAssertNil(LocalNetworkProbe.access(for: .setup))
    }
}
