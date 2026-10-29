import Photos

nonisolated enum AccessState: Equatable {
    case allowed
    case needsPrompt
    case denied
}

nonisolated struct PermissionsGate: Equatable {
    var photos: PHAuthorizationStatus
    var localNetwork: AccessState
    var keychain: AccessState

    var photosAllowed: Bool { photos == .authorized || photos == .limited }
    var isSatisfied: Bool { photosAllowed && localNetwork == .allowed && keychain == .allowed }

    static let count = 3
    var allowedCount: Int {
        [photosAllowed, localNetwork == .allowed, keychain == .allowed].filter { $0 }.count
    }

    var continueTitle: String {
        if !photosAllowed { return "Allow Photos to continue" }
        if localNetwork != .allowed { return "Allow Local Network to continue" }
        if keychain != .allowed { return "Allow Keychain to continue" }
        return "Continue"
    }

    static func showsScreen(currentlyShowing: Bool, gate: PermissionsGate) -> Bool {
        currentlyShowing || !gate.isSatisfied
    }
}

enum StartupPhase: Equatable {
    case permissions
    case gettingReady
    case signedOut
    case signedIn

    /// The sidebar has nothing to point at until the app is past its startup screens.
    var hidesSidebar: Bool { self == .permissions || self == .gettingReady }

    nonisolated static func resolve(showsPermissions: Bool, isCheckingSession: Bool, isSignedIn: Bool) -> StartupPhase {
        if showsPermissions { return .permissions }
        if isCheckingSession { return .gettingReady }
        return isSignedIn ? .signedIn : .signedOut
    }
}
