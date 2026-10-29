import Foundation
import os

enum Log {
    static let subsystem = "ar.schw.homie.photo-sync"

    static let sync = Logger(subsystem: subsystem, category: "sync")
    static let api = Logger(subsystem: subsystem, category: "api")
    static let permissions = Logger(subsystem: subsystem, category: "permissions")
    static let keychain = Logger(subsystem: subsystem, category: "keychain")
    static let discovery = Logger(subsystem: subsystem, category: "discovery")
}

/// Every line carries the run ID, so one sync can be filtered out of a log archive with
/// `log show --predicate 'subsystem == "ar.schw.homie.photo-sync" AND eventMessage CONTAINS "run 1a2b3c4d"'`.
/// `message` is logged in the clear, so keep personal data (hosts, asset IDs, error text) in `detail`.
struct SyncRunLog: Sendable {
    let id: String

    init(id: String = Self.makeID()) {
        self.id = id
    }

    static func makeID() -> String {
        String(UUID().uuidString.prefix(8)).lowercased()
    }

    func info(_ message: String, detail: String = "") {
        Log.sync.info("[run \(id, privacy: .public)] \(message, privacy: .public) \(detail, privacy: .private)")
    }

    func error(_ message: String, detail: String = "") {
        Log.sync.error("[run \(id, privacy: .public)] \(message, privacy: .public) \(detail, privacy: .private)")
    }
}
