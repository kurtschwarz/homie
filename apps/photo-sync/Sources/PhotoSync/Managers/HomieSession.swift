import Foundation
import Security

enum HomieSession {
    private static let service = "ar.schw.homie.photo-sync"
    private static let legacyAccount = "session-token"

    static func token(for endpoint: URL) -> String? {
        if let token = read(account: sessionAccount(for: endpoint)) { return token }
        guard let legacyToken = read(account: legacyAccount) else { return nil }
        do {
            try save(legacyToken, for: endpoint)
            SecItemDelete(itemQuery(account: legacyAccount) as CFDictionary)
        } catch {}
        return legacyToken
    }

    private static let grantedKey = "keychainGrantedForCode"

    /// Reading an item's data is the only way to learn whether the keychain would prompt, and asking is
    /// what triggers the prompt. So launch only looks for the item's attributes, which never prompt, and
    /// trusts what the last Allow told us, for the same code signature only: re-signing the app
    /// invalidates the keychain's grant, so a record from an older build proves nothing.
    static func probeAccess(for endpoint: URL, defaults: UserDefaults = .standard) -> AccessState {
        guard savedAccounts(for: endpoint).contains(where: itemExists) else { return .allowed }
        return grantIsCurrent(recorded: defaults.string(forKey: grantedKey), identity: codeIdentity) ? .allowed : .needsPrompt
    }

    nonisolated static func grantIsCurrent(recorded: String?, identity: String) -> Bool {
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

    static func requestAccess(for endpoint: URL, defaults: UserDefaults = .standard) -> AccessState {
        for account in savedAccounts(for: endpoint) {
            var query = itemQuery(account: account)
            query[kSecReturnData as String] = true
            query[kSecMatchLimit as String] = kSecMatchLimitOne
            var result: CFTypeRef?
            switch SecItemCopyMatching(query as CFDictionary, &result) {
            case errSecSuccess, errSecItemNotFound:
                continue
            default:
                defaults.removeObject(forKey: grantedKey)
                return .denied
            }
        }
        defaults.set(codeIdentity, forKey: grantedKey)
        return .allowed
    }

    private static func savedAccounts(for endpoint: URL) -> [String] {
        [sessionAccount(for: endpoint), legacyAccount]
    }

    private static func itemExists(account: String) -> Bool {
        var query = itemQuery(account: account)
        query[kSecReturnAttributes as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        var result: CFTypeRef?
        return SecItemCopyMatching(query as CFDictionary, &result) == errSecSuccess
    }

    private static func read(account: String) -> String? {
        var query = itemQuery(account: account)
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        var result: CFTypeRef?
        guard SecItemCopyMatching(query as CFDictionary, &result) == errSecSuccess,
              let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    static func save(_ token: String, for endpoint: URL) throws {
        clear(for: endpoint)
        var item = itemQuery(account: sessionAccount(for: endpoint))
        item[kSecValueData as String] = Data(token.utf8)
        item[kSecAttrAccessible as String] = kSecAttrAccessibleWhenUnlocked
        let status = SecItemAdd(item as CFDictionary, nil)
        guard status == errSecSuccess else { throw HomieSessionError.couldNotSave(status) }
    }

    static func clear(for endpoint: URL) {
        SecItemDelete(itemQuery(account: sessionAccount(for: endpoint)) as CFDictionary)
    }

    static func clearAll() {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
        ]
        SecItemDelete(query as CFDictionary)
    }

    private static func itemQuery(account: String) -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
    }

    static func sessionAccount(for endpoint: URL) -> String {
        let scheme = endpoint.scheme?.lowercased() ?? "http"
        let port = endpoint.port ?? (scheme == "https" ? 443 : 80)
        return "\(legacyAccount)@\(scheme)://\(HomieHost.normalizedHostname(for: endpoint)):\(port)"
    }
}

private enum HomieSessionError: LocalizedError {
    case couldNotSave(OSStatus)

    var errorDescription: String? {
        switch self {
        case .couldNotSave(let status):
            "Could not save your Homie session in the Keychain (error \(status))."
        }
    }
}
