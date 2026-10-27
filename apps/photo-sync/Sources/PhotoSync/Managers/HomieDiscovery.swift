import Combine
import Foundation

@MainActor
final class HomieDiscovery: NSObject, ObservableObject {
    @Published private(set) var hosts: [HomieHost] = []
    @Published private(set) var isSearching = false

    private let browser = NetServiceBrowser()
    private var services: [NetService] = []
    private var resolvedHosts: [String: HomieHost] = [:]
    private var started = false

    override init() {
        super.init()
        browser.delegate = self
    }

    func start() {
        guard !started else { return }
        started = true
        isSearching = true
        browser.searchForServices(ofType: "_homie._tcp.", inDomain: "local.")

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
}

extension HomieDiscovery: NetServiceBrowserDelegate, NetServiceDelegate {
    nonisolated func netServiceBrowser(_ browser: NetServiceBrowser, didFind service: NetService, moreComing: Bool) {
        Task { @MainActor in
            services.append(service)
            service.delegate = self
            service.resolve(withTimeout: 5)
        }
    }

    nonisolated func netServiceDidResolveAddress(_ sender: NetService) {
        Task { @MainActor in
            guard let hostname = sender.hostName else { return }
            var components = URLComponents()
            components.scheme = "http"
            components.host = hostname
            components.port = sender.port
            components.path = "/graphql"
            guard let endpoint = components.url else { return }
            let host = HomieHost(name: sender.name, graphQLURL: endpoint, source: .discovered)
            resolvedHosts[serviceKey(sender)] = host
            if let index = hosts.firstIndex(where: { $0.id == host.id }) {
                hosts[index] = host
            } else {
                hosts.append(host)
            }
            hosts.sort { $0.name.localizedStandardCompare($1.name) == .orderedAscending }
            isSearching = false
        }
    }

    nonisolated func netServiceBrowserDidStopSearch(_ browser: NetServiceBrowser) {
        Task { @MainActor in
            isSearching = false
            started = false
        }
    }

    nonisolated func netServiceBrowser(_ browser: NetServiceBrowser, didRemove service: NetService, moreComing: Bool) {
        Task { @MainActor in
            services.removeAll { $0 === service }
            if let host = resolvedHosts.removeValue(forKey: serviceKey(service)) {
                hosts.removeAll { $0.id == host.id }
            }
        }
    }

    nonisolated func netServiceBrowser(_ browser: NetServiceBrowser, didNotSearch errorDict: [String: NSNumber]) {
        Task { @MainActor in
            isSearching = false
            started = false
        }
    }

    nonisolated func netService(_ sender: NetService, didNotResolve errorDict: [String: NSNumber]) {
        Task { @MainActor in services.removeAll { $0 === sender } }
    }

    private func serviceKey(_ service: NetService) -> String {
        "\(service.name)|\(service.type)|\(service.domain)"
    }
}
