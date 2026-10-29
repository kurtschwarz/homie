import Photos

enum KeychainAccess: Equatable {
    case allowed
    case needsPrompt
    case denied
}

struct PermissionsGate: Equatable {
    var photos: PHAuthorizationStatus
    var keychain: KeychainAccess

    var photosAllowed: Bool { photos == .authorized || photos == .limited }
    var isSatisfied: Bool { photosAllowed && keychain == .allowed }

    var continueTitle: String {
        if !photosAllowed { return "Allow Photos to continue" }
        if keychain != .allowed { return "Allow Keychain to continue" }
        return "Continue"
    }

    nonisolated static func showsScreen(currentlyShowing: Bool, gate: PermissionsGate) -> Bool {
        currentlyShowing || !gate.isSatisfied
    }
}

enum StartupPhase: Equatable {
    case permissions
    case gettingReady
    case signedOut
    case signedIn

    nonisolated static func resolve(showsPermissions: Bool, isCheckingSession: Bool, isSignedIn: Bool) -> StartupPhase {
        if showsPermissions { return .permissions }
        if isCheckingSession { return .gettingReady }
        return isSignedIn ? .signedIn : .signedOut
    }
}
