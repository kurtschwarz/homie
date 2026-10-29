import Photos
import XCTest
@testable import PhotoSync

final class PermissionsGateTests: XCTestCase {
    private func gate(_ photos: PHAuthorizationStatus, _ keychain: KeychainAccess) -> PermissionsGate {
        PermissionsGate(photos: photos, keychain: keychain)
    }

    func testSatisfiedWhenBothAllowed() {
        XCTAssertTrue(gate(.authorized, .allowed).isSatisfied)
        XCTAssertTrue(gate(.limited, .allowed).isSatisfied)
    }

    func testNotSatisfiedWhenEitherMissing() {
        XCTAssertFalse(gate(.notDetermined, .allowed).isSatisfied)
        XCTAssertFalse(gate(.denied, .allowed).isSatisfied)
        XCTAssertFalse(gate(.authorized, .needsPrompt).isSatisfied)
        XCTAssertFalse(gate(.authorized, .denied).isSatisfied)
    }

    func testScreenNeverShownWhenSatisfied() {
        XCTAssertFalse(PermissionsGate.showsScreen(currentlyShowing: false, gate: gate(.authorized, .allowed)))
    }

    func testScreenAppearsWhenAccessIsRevoked() {
        XCTAssertTrue(PermissionsGate.showsScreen(currentlyShowing: false, gate: gate(.denied, .allowed)))
    }

    func testScreenStaysUntilContinueEvenWhenSatisfied() {
        XCTAssertTrue(PermissionsGate.showsScreen(currentlyShowing: true, gate: gate(.authorized, .allowed)))
    }

    func testContinueTitleExplainsWhatIsMissing() {
        XCTAssertEqual(gate(.notDetermined, .allowed).continueTitle, "Allow Photos to continue")
        XCTAssertEqual(gate(.authorized, .needsPrompt).continueTitle, "Allow Keychain to continue")
        XCTAssertEqual(gate(.authorized, .allowed).continueTitle, "Continue")
    }

    func testStartupOrder() {
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: true, isCheckingSession: true, isSignedIn: false), .permissions)
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: false, isCheckingSession: true, isSignedIn: false), .gettingReady)
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: false, isCheckingSession: false, isSignedIn: false), .signedOut)
        XCTAssertEqual(StartupPhase.resolve(showsPermissions: false, isCheckingSession: false, isSignedIn: true), .signedIn)
    }
}
