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
