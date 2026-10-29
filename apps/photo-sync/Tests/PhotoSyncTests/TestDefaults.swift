import Foundation

enum TestDefaults {
    static func make() -> UserDefaults {
        UserDefaults(suiteName: "PhotoSyncTests.\(UUID().uuidString)")!
    }
}
