import Foundation
import Security

enum HomieSession {
    private static let service = "ar.schw.homie.photo-sync"
    private static let account = "session-token"

    static var token: String? {
        var query = itemQuery
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        var result: CFTypeRef?
        guard SecItemCopyMatching(query as CFDictionary, &result) == errSecSuccess,
              let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    static func save(_ token: String) throws {
        clear()
        var item = itemQuery
        item[kSecValueData as String] = Data(token.utf8)
        item[kSecAttrAccessible as String] = kSecAttrAccessibleWhenUnlocked
        let status = SecItemAdd(item as CFDictionary, nil)
        guard status == errSecSuccess else { throw HomieSessionError.couldNotSave(status) }
    }

    static func clear() {
        SecItemDelete(itemQuery as CFDictionary)
    }

    private static var itemQuery: [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
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
