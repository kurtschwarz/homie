import Foundation

nonisolated struct HomieHost: Identifiable, Hashable {
    enum Source: Hashable {
        case discovered
        case manual
        case fallback
    }

    let name: String
    let graphQLURL: URL
    let source: Source

    var id: String {
        let scheme = graphQLURL.scheme?.lowercased() ?? "http"
        let port = graphQLURL.port ?? (scheme == "https" ? 443 : 80)
        return "\(scheme)://\(Self.normalizedHostname(for: graphQLURL)):\(port)\(graphQLURL.path())"
    }

    var address: String {
        guard let host = graphQLURL.host() else { return graphQLURL.absoluteString }
        let cleanHost = host.hasSuffix(".") ? String(host.dropLast()) : host
        guard let port = graphQLURL.port, port != 80, port != 443 else { return cleanHost }
        return "\(cleanHost):\(port)"
    }

    static func configured(defaults: UserDefaults = .standard) -> HomieHost {
        let endpoint = HomieAPI.configuredEndpoint(defaults: defaults)
        let isOverride = defaults.string(forKey: "HomieAPIURL") != nil
        let wasDiscovered = defaults.bool(forKey: "HomieAPIDiscovered")
        return HomieHost(
            name: defaults.string(forKey: "HomieAPIName") ?? (endpoint.host() == "homie" ? "Homie" : endpoint.host() ?? "Homie"),
            graphQLURL: endpoint,
            source: wasDiscovered ? .discovered : (isOverride ? .manual : .fallback)
        )
    }

    static func manual(_ input: String) -> HomieHost? {
        var value = input.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !value.isEmpty else { return nil }
        if !value.contains("://") { value = "http://\(value)" }
        guard let url = URL(string: value), url.host() != nil else { return nil }
        let endpoint = url.path().isEmpty || url.path() == "/" ? url.appending(path: "graphql") : url
        return HomieHost(name: url.host() ?? "Homie", graphQLURL: endpoint, source: .manual)
    }

    static func normalizedHostname(for endpoint: URL) -> String {
        var host = endpoint.host()?.lowercased() ?? endpoint.absoluteString
        if host.hasSuffix(".") { host.removeLast() }
        if host.hasSuffix(".local") { host.removeLast(".local".count) }
        return host
    }
}
