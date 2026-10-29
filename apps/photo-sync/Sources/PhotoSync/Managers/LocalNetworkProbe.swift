import Foundation
import Network

/// macOS has no API that reports Local Network access, and browsing for a service is what raises the
/// prompt. A denial surfaces as a policy error on the browser, so that is the only signal there is.
enum LocalNetworkProbe {
    static let policyDenied: Int32 = -65570
    private static let grantedKey = "localNetworkAccessGranted"
    private static let serviceType = "_homie._tcp"

    static func access(for state: NWBrowser.State) -> AccessState? {
        switch state {
        case .ready:
            return .allowed
        case .waiting(.dns(policyDenied)):
            return .denied
        default:
            return nil
        }
    }

    static func recordedAccess(defaults: UserDefaults = .standard) -> AccessState {
        defaults.bool(forKey: grantedKey) ? .allowed : .needsPrompt
    }

    /// Browses until the system settles on an answer. With the prompt open that waits for the user, so
    /// `timeout` bounds it; nil means no answer arrived.
    static func request(timeout: Duration, defaults: UserDefaults = .standard) async -> AccessState? {
        let result = await browse(timeout: timeout)
        if let result { defaults.set(result == .allowed, forKey: grantedKey) }
        return result
    }

    private static func browse(timeout: Duration) async -> AccessState? {
        let browser = NWBrowser(for: .bonjour(type: serviceType, domain: nil), using: NWParameters())
        let settled = Settled()
        return await withCheckedContinuation { continuation in
            browser.stateUpdateHandler = { state in
                guard let answer = access(for: state), settled.claim() else { return }
                browser.cancel()
                continuation.resume(returning: answer)
            }
            browser.start(queue: .global(qos: .userInitiated))
            Task {
                try? await Task.sleep(for: timeout)
                guard settled.claim() else { return }
                browser.cancel()
                continuation.resume(returning: nil)
            }
        }
    }
}

private final class Settled: @unchecked Sendable {
    private let lock = NSLock()
    private var done = false

    func claim() -> Bool {
        lock.lock()
        defer { lock.unlock() }
        if done { return false }
        done = true
        return true
    }
}
