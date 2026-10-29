import Foundation
import Network
import Observation

@Observable
final class HomieDiscovery {
    private(set) var hosts: [HomieHost] = []
    private(set) var isSearching = false

    @ObservationIgnored private var browser: NWBrowser?
    @ObservationIgnored private var resolutions: [String: Task<Void, Never>] = [:]
    @ObservationIgnored private var resolvedHosts: [String: HomieHost] = [:]

    func start() {
        guard browser == nil else { return }
        isSearching = true

        let browser = NWBrowser(for: .bonjour(type: "_homie._tcp", domain: "local."), using: NWParameters())
        browser.stateUpdateHandler = { [weak self] state in
            MainActor.assumeIsolated { self?.apply(state) }
        }
        browser.browseResultsChangedHandler = { [weak self] _, changes in
            MainActor.assumeIsolated { self?.apply(changes) }
        }
        self.browser = browser
        browser.start(queue: .main)

        Task {
            try? await Task.sleep(for: .seconds(2))
            guard hosts.isEmpty else { return }
            let fallback = HomieHost(
                name: "Homie",
                graphQLURL: URL(string: "http://homie.local/graphql")!,
                source: .fallback
            )
            if (try? await HomieAPI(endpoint: fallback.graphQLURL).health()) != nil,
               !hosts.contains(where: { $0.id == fallback.id }) {
                hosts.append(fallback)
            }
        }
    }

    private func apply(_ state: NWBrowser.State) {
        switch state {
        case .failed, .cancelled:
            browser = nil
            isSearching = false
        default:
            break
        }
    }

    private func apply(_ changes: Set<NWBrowser.Result.Change>) {
        for change in changes {
            switch change {
            case .added(let result), .changed(old: _, new: let result, flags: _):
                if case .service(let name, let type, let domain, _) = result.endpoint {
                    resolve(result.endpoint, key: Self.serviceKey(name: name, type: type, domain: domain), name: name)
                }
            case .removed(let result):
                if case .service(let name, let type, let domain, _) = result.endpoint {
                    remove(key: Self.serviceKey(name: name, type: type, domain: domain))
                }
            case .identical:
                break
            @unknown default:
                break
            }
        }
    }

    private func resolve(_ endpoint: NWEndpoint, key: String, name: String) {
        resolutions[key]?.cancel()
        resolutions[key] = Task {
            let target = await Self.resolveTarget(of: endpoint)
            guard !Task.isCancelled, let target,
                  let url = Self.graphQLURL(serviceName: name, target: target) else { return }
            resolutions[key] = nil
            let host = HomieHost(name: name, graphQLURL: url, source: .discovered)
            if let previous = resolvedHosts[key] { hosts.removeAll { $0.id == previous.id } }
            resolvedHosts[key] = host
            if let index = hosts.firstIndex(where: { $0.id == host.id }) {
                hosts[index] = host
            } else {
                hosts.append(host)
            }
            hosts.sort { $0.name.localizedStandardCompare($1.name) == .orderedAscending }
            isSearching = false
        }
    }

    private func remove(key: String) {
        resolutions.removeValue(forKey: key)?.cancel()
        if let host = resolvedHosts.removeValue(forKey: key) {
            hosts.removeAll { $0.id == host.id }
        }
    }

    private static func serviceKey(name: String, type: String, domain: String) -> String {
        "\(name)|\(type)|\(domain)"
    }

    /// Avahi advertises the service as "Homie · <hostname>", and the host answers mDNS as `<hostname>.local`.
    nonisolated static func hostname(fromServiceName name: String) -> String? {
        let prefix = "Homie · "
        guard name.hasPrefix(prefix) else { return nil }
        let hostname = name.dropFirst(prefix.count).trimmingCharacters(in: .whitespaces)
        return hostname.isEmpty ? nil : "\(hostname).local"
    }

    private nonisolated static func graphQLURL(serviceName: String, target: NWEndpoint) -> URL? {
        guard case .hostPort(let host, let port) = target else { return nil }
        var components = URLComponents()
        components.scheme = "http"
        components.host = hostname(fromServiceName: serviceName) ?? literal(for: host)
        components.port = Int(port.rawValue)
        components.path = "/graphql"
        return components.url
    }

    private nonisolated static func literal(for host: NWEndpoint.Host) -> String {
        let text = "\(host)".split(separator: "%").first.map(String.init) ?? "\(host)"
        return text.contains(":") ? "[\(text)]" : text
    }

    private nonisolated static func resolveTarget(of endpoint: NWEndpoint) async -> NWEndpoint? {
        let connection = NWConnection(to: endpoint, using: .tcp)
        let timeout = Task {
            try? await Task.sleep(for: .seconds(5))
            connection.cancel()
        }
        defer { timeout.cancel() }

        return await withTaskCancellationHandler {
            await withCheckedContinuation { continuation in
                let settled = Settled()
                connection.stateUpdateHandler = { state in
                    switch state {
                    case .ready:
                        if settled.claim() {
                            continuation.resume(returning: connection.currentPath?.remoteEndpoint)
                            connection.cancel()
                        }
                    case .failed, .cancelled:
                        if settled.claim() { continuation.resume(returning: nil) }
                    default:
                        break
                    }
                }
                connection.start(queue: .global(qos: .userInitiated))
            }
        } onCancel: {
            connection.cancel()
        }
    }
}
