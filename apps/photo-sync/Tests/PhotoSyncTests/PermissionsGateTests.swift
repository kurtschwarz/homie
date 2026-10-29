import Network
import Photos
import Testing
@testable import PhotoSync

struct PermissionsGateTests {
    private static func gate(
        _ photos: PHAuthorizationStatus,
        localNetwork: AccessState = .allowed,
        keychain: AccessState = .allowed
    ) -> PermissionsGate {
        PermissionsGate(photos: photos, localNetwork: localNetwork, keychain: keychain)
    }

    @Test(arguments: [PHAuthorizationStatus.authorized, .limited])
    func satisfiedWhenAllAllowed(photos: PHAuthorizationStatus) {
        #expect(Self.gate(photos).isSatisfied)
    }

    @Test(arguments: [
        (PHAuthorizationStatus.notDetermined, AccessState.allowed, AccessState.allowed),
        (.denied, .allowed, .allowed),
        (.authorized, .needsPrompt, .allowed),
        (.authorized, .denied, .allowed),
        (.authorized, .allowed, .needsPrompt),
        (.authorized, .allowed, .denied),
    ])
    func notSatisfiedWhenAnyMissing(photos: PHAuthorizationStatus, localNetwork: AccessState, keychain: AccessState) {
        #expect(!Self.gate(photos, localNetwork: localNetwork, keychain: keychain).isSatisfied)
    }

    @Test func screenNeverShownWhenSatisfied() {
        #expect(!PermissionsGate.showsScreen(currentlyShowing: false, gate: Self.gate(.authorized)))
    }

    @Test(arguments: [
        (PHAuthorizationStatus.denied, AccessState.allowed),
        (.authorized, .denied),
    ])
    func screenAppearsWhenAccessIsRevoked(photos: PHAuthorizationStatus, localNetwork: AccessState) {
        #expect(PermissionsGate.showsScreen(currentlyShowing: false, gate: Self.gate(photos, localNetwork: localNetwork)))
    }

    @Test func screenStaysUntilContinueEvenWhenSatisfied() {
        #expect(PermissionsGate.showsScreen(currentlyShowing: true, gate: Self.gate(.authorized)))
    }

    @Test(arguments: [
        (PHAuthorizationStatus.notDetermined, AccessState.needsPrompt, AccessState.needsPrompt, "Allow Photos to continue"),
        (.authorized, .needsPrompt, .needsPrompt, "Allow Local Network to continue"),
        (.authorized, .allowed, .needsPrompt, "Allow Keychain to continue"),
        (.authorized, .allowed, .allowed, "Continue"),
    ])
    func continueTitleNamesTheFirstMissingPermission(
        photos: PHAuthorizationStatus,
        localNetwork: AccessState,
        keychain: AccessState,
        title: String
    ) {
        #expect(Self.gate(photos, localNetwork: localNetwork, keychain: keychain).continueTitle == title)
    }

    @Test(arguments: [
        (StartupPhase.permissions, true),
        (.gettingReady, true),
        (.signedOut, false),
        (.signedIn, false),
    ])
    func sidebarIsSkeletonUntilPastTheStartupScreens(phase: StartupPhase, hidesSidebar: Bool) {
        #expect(phase.hidesSidebar == hidesSidebar)
    }

    @Test(arguments: [
        (true, true, false, StartupPhase.permissions),
        (false, true, false, .gettingReady),
        (false, false, false, .signedOut),
        (false, false, true, .signedIn),
    ])
    func startupOrder(showsPermissions: Bool, isCheckingSession: Bool, isSignedIn: Bool, expected: StartupPhase) {
        #expect(
            StartupPhase.resolve(
                showsPermissions: showsPermissions,
                isCheckingSession: isCheckingSession,
                isSignedIn: isSignedIn
            ) == expected
        )
    }
}

struct KeychainGrantTests {
    @Test func grantHoldsForTheSameCodeSignature() {
        #expect(CodeSignedGrant.isCurrent(recorded: "abc", identity: "abc"))
    }

    @Test func grantFromAnotherBuildIsStale() {
        #expect(!CodeSignedGrant.isCurrent(recorded: "abc", identity: "def"))
    }

    @Test(arguments: [(nil, "abc"), ("", "")] as [(String?, String)])
    func noRecordOrUnreadableSignatureIsNotAGrant(recorded: String?, identity: String) {
        #expect(!CodeSignedGrant.isCurrent(recorded: recorded, identity: identity))
    }
}

struct LocalNetworkProbeTests {
    @Test func readyBrowserSaysNothingBecauseItIsReadyWhetherOrNotAccessIsAllowed() {
        #expect(LocalNetworkProbe.access(for: .ready) == nil)
    }

    @Test func findingOurOwnAdvertisedServiceMeansAllowed() {
        #expect(LocalNetworkProbe.access(foundNames: ["other", "probe-1"], own: "probe-1") == .allowed)
    }

    @Test(arguments: [["other"], []])
    func notFindingOurOwnServiceSaysNothing(foundNames: [String]) {
        #expect(LocalNetworkProbe.access(foundNames: foundNames, own: "probe-1") == nil)
    }

    @Test func policyDenialMeansDenied() {
        let denied = NWBrowser.State.waiting(.dns(LocalNetworkProbe.policyDenied))
        #expect(LocalNetworkProbe.access(for: denied) == .denied)
    }

    @Test func otherWaitingReasonsSayNothingAboutPermission() {
        #expect(LocalNetworkProbe.access(for: .waiting(.dns(-65537))) == nil)
        #expect(LocalNetworkProbe.access(for: .setup) == nil)
    }
}
