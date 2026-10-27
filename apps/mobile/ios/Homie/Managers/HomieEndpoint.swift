import Foundation

enum HomieEndpoint {
    static var baseURL: URL {
        let configuredURL = UserDefaults.standard.string(forKey: "HomieAPIURL")
            ?? ProcessInfo.processInfo.environment["HOMIE_API_URL"]

        guard let configuredURL,
              let url = URL(string: configuredURL),
              url.host() != nil else {
            return URL(string: "http://homie")!
        }

        if url.path().isEmpty || url.path() == "/" {
            return url
        }

        var components = URLComponents(url: url, resolvingAgainstBaseURL: false)!
        if components.path.hasSuffix("/graphql") {
            components.path = String(components.path.dropLast("/graphql".count))
        }
        return components.url ?? url
    }

    static var graphQLURL: URL {
        let configuredURL = UserDefaults.standard.string(forKey: "HomieAPIURL")
            ?? ProcessInfo.processInfo.environment["HOMIE_API_URL"]

        if let configuredURL,
           let url = URL(string: configuredURL),
           url.host() != nil,
           !url.path().isEmpty,
           url.path() != "/" {
            return url
        }

        return baseURL.appending(path: "graphql")
    }
}
