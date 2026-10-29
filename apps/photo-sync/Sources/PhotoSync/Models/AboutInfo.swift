import Foundation

nonisolated struct AboutInfo {
    static let name = "Homie Photos"
    static let summary = "Sends iCloud Photos to your Homie, and brings paintings made on Homie back to your Mac."

    let versionLine: String?
    let commit: String?

    init(info: [String: Any]?) {
        let version = Self.nonEmpty(info?["CFBundleShortVersionString"])
        let build = Self.nonEmpty(info?["CFBundleVersion"])
        commit = Self.nonEmpty(info?["HomieGitCommit"])
        guard let version else {
            versionLine = nil
            return
        }
        versionLine = build.map { "Version \(version) (\($0))" } ?? "Version \(version)"
    }

    static var current: AboutInfo { AboutInfo(info: Bundle.main.infoDictionary) }

    private static func nonEmpty(_ value: Any?) -> String? {
        guard let string = value as? String, !string.isEmpty else { return nil }
        return string
    }
}
