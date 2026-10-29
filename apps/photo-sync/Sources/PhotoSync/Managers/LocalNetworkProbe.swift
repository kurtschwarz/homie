import Foundation
import Network

/// macOS has no API that reports Local Network access, and a browser reaches `ready` whether or not the
/// user has answered. So the app advertises a service of its own and browses for it: seeing it means
/// access is on, and a policy error means it is off.
enum LocalNetworkProbe {
    static let policyDenied: Int32 = -65570
    static let serviceType = "_homie-photos-probe._tcp"
    private static let grantedKey = "photo-sync.localNetworkGrantedForCode"

    static func access(for state: NWBrowser.State) -> AccessState? {
        if case .waiting(.dns(policyDenied)) = state { return .denied }
        return nil
    }

    static func access(foundNames: [String], own: String) -> AccessState? {
        foundNames.contains(own) ? .allowed : nil
    }

    static func recordedAccess(defaults: UserDefaults = .standard) -> AccessState {
        grantIsCurrent(recorded: defaults.string(forKey: grantedKey), identity: HomieSession.codeIdentity)
            ? .allowed : .needsPrompt
    }

    /// Waits for the system to settle on an answer. With the prompt open that waits for the user, so
    /// `timeout` bounds it; nil means no answer arrived.
    static func request(timeout: Duration, defaults: UserDefaults = .standard) async -> AccessState? {
        let result = await probe(timeout: timeout)
        switch result {
        case .allowed: defaults.set(HomieSession.codeIdentity, forKey: grantedKey)
        case .denied: defaults.removeObject(forKey: grantedKey)
        default: break
        }
        return result
    }

    nonisolated static func grantIsCurrent(recorded: String?, identity: String) -> Bool {
        HomieSession.grantIsCurrent(recorded: recorded, identity: identity)
    }

    private static func probe(timeout: Duration) async -> AccessState? {
        let name = "probe-\(UUID().uuidString.prefix(8))"
        guard let listener = try? NWListener(using: .tcp) else { return nil }
        listener.service = NWListener.Service(name: name, type: serviceType)
        listener.newConnectionHandler = { $0.cancel() }

        let browser = NWBrowser(for: .bonjour(type: serviceType, domain: nil), using: NWParameters())
        let settled = Settled()

        return await withCheckedContinuation { continuation in
            @Sendable func finish(_ answer: AccessState?) {
                guard settled.claim() else { return }
                browser.cancel()
                listener.cancel()
                continuation.resume(returning: answer)
            }
            browser.stateUpdateHandler = { state in
                if let answer = access(for: state) { finish(answer) }
            }
            browser.browseResultsChangedHandler = { results, _ in
                let names = results.compactMap { result -> String? in
                    if case .service(let found, _, _, _) = result.endpoint { return found }
                    return nil
                }
                if let answer = access(foundNames: names, own: name) { finish(answer) }
            }
            listener.start(queue: .global(qos: .userInitiated))
            browser.start(queue: .global(qos: .userInitiated))
            Task {
                try? await Task.sleep(for: timeout)
                finish(nil)
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
