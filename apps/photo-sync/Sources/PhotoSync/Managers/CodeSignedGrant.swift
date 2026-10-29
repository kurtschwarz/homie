import Foundation
import Security

/// A record of the user's last Allow, tied to the app's code signature: re-signing the app invalidates
/// the system's grant, so a record from an older build proves nothing.
nonisolated struct CodeSignedGrant {
    let key: String

    func isCurrent(defaults: UserDefaults = .standard) -> Bool {
        Self.isCurrent(recorded: defaults.string(forKey: key), identity: Self.codeIdentity)
    }

    func record(defaults: UserDefaults = .standard) {
        defaults.set(Self.codeIdentity, forKey: key)
    }

    func revoke(defaults: UserDefaults = .standard) {
        defaults.removeObject(forKey: key)
    }

    static func isCurrent(recorded: String?, identity: String) -> Bool {
        !identity.isEmpty && recorded == identity
    }

    static var codeIdentity: String {
        var code: SecCode?
        var staticCode: SecStaticCode?
        var info: CFDictionary?
        guard SecCodeCopySelf([], &code) == errSecSuccess, let code,
              SecCodeCopyStaticCode(code, [], &staticCode) == errSecSuccess, let staticCode,
              SecCodeCopySigningInformation(staticCode, SecCSFlags(rawValue: kSecCSSigningInformation), &info) == errSecSuccess,
              let unique = (info as? [String: Any])?[kSecCodeInfoUnique as String] as? Data else { return "" }
        return unique.map { String(format: "%02x", $0) }.joined()
    }
}
