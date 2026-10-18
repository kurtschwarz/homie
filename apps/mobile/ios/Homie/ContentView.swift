//
//  ContentView.swift
//  Homie
//
//  Created by Kurt Schwarz on 2026-10-18.
//

import Foundation
import SwiftUI

private struct Health: Decodable {
    let status: String
    let version: String
    let uptimeSeconds: Int
}

private struct HealthResponse: Decodable {
    struct Data: Decodable {
        let health: Health?
    }

    struct ErrorItem: Decodable {
        let message: String
    }

    let data: Data?
    let errors: [ErrorItem]?
}

struct ContentView: View {
    @State private var health: Health?
    @State private var errorMessage: String?
    @State private var isLoading = false

    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "house.fill")
                .font(.system(size: 44))
                .foregroundStyle(.tint)

            Text("Homie")
                .font(.largeTitle.bold())

            if let health {
                VStack(spacing: 8) {
                    Label(health.status.capitalized, systemImage: "checkmark.circle.fill")
                        .foregroundStyle(.green)
                    Text("Version \(health.version)")
                        .font(.title2.weight(.medium))
                        .accessibilityIdentifier("apiVersion")
                    Text("API uptime: \(health.uptimeSeconds) seconds")
                        .foregroundStyle(.secondary)
                }
            } else if let errorMessage {
                Text("Could not connect to the Homie API")
                    .font(.headline)
                Text(errorMessage)
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            } else {
                ProgressView("Connecting to the Homie API…")
            }

            Button("Refresh", systemImage: "arrow.clockwise") {
                Task { await refresh() }
            }
            .disabled(isLoading)
        }
        .padding(32)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .task { await refresh() }
    }

    private func refresh() async {
        guard !isLoading else { return }
        isLoading = true
        defer { isLoading = false }

        do {
            var request = URLRequest(url: Self.apiEndpoint)
            request.httpMethod = "POST"
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.setValue("application/json", forHTTPHeaderField: "Accept")
            request.httpBody = try JSONSerialization.data(withJSONObject: [
                "query": "{ health { status version uptimeSeconds } }"
            ])

            let (data, response) = try await URLSession.shared.data(for: request)
            guard let response = response as? HTTPURLResponse else {
                throw APIError.invalidResponse
            }
            guard (200..<300).contains(response.statusCode) else {
                throw APIError.http(response.statusCode)
            }

            let result = try JSONDecoder().decode(HealthResponse.self, from: data)
            if let message = result.errors?.first?.message {
                throw APIError.graphQL(message)
            }
            guard let health = result.data?.health else {
                throw APIError.missingHealth
            }

            self.health = health
            errorMessage = nil
        } catch {
            health = nil
            errorMessage = error.localizedDescription
        }
    }

    private static var apiEndpoint: URL {
        let configuredURL = UserDefaults.standard.string(forKey: "HomieAPIURL")
            ?? ProcessInfo.processInfo.environment["HOMIE_API_URL"]

        if let configuredURL,
           let url = URL(string: configuredURL),
           url.host() != nil {
            return url.path().isEmpty || url.path() == "/"
                ? url.appending(path: "graphql")
                : url
        }

        return URL(string: "http://homie:3000/graphql")!
    }
}

private enum APIError: LocalizedError {
    case invalidResponse
    case http(Int)
    case graphQL(String)
    case missingHealth

    var errorDescription: String? {
        switch self {
        case .invalidResponse:
            "The API returned an invalid response."
        case .http(let status):
            "The API returned HTTP \(status)."
        case .graphQL(let message):
            message
        case .missingHealth:
            "The API response did not include its health status."
        }
    }
}

#Preview {
    ContentView()
}
