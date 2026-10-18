import Foundation
import Observation

@MainActor
@Observable
final class HomeManager {
    struct Weather {
        let condition: String
        let temperature: Int
        let stale: Bool

        /// Mirrors the WeatherCondition enum in weather.graphqls.
        var symbolName: String {
            switch condition {
            case "SUN": "sun.max.fill"
            case "CLOUD_SUN": "cloud.sun.fill"
            case "CLOUD": "cloud.fill"
            case "RAIN": "cloud.rain.fill"
            case "MOON": "moon.fill"
            default: "cloud.fill"
            }
        }
    }

    private(set) var weather: Weather?
    /// Events still to come today. The API's events window starts at "now", so
    /// anything that already finished today is not counted.
    private(set) var remainingToday = 0
    private(set) var errorMessage: String?

    private struct Payload: Decodable {
        struct WeatherDTO: Decodable {
            let condition: String
            let temperature: Int
            let stale: Bool
        }
        struct EventDTO: Decodable {
            let start: Int
        }
        let weather: WeatherDTO
        let events: [EventDTO]
    }

    func load() async {
        do {
            let payload: Payload = try await HomieAPI.perform(
                "{ weather { condition temperature stale } events(days: 2) { start } }"
            )
            weather = Weather(
                condition: payload.weather.condition,
                temperature: payload.weather.temperature,
                stale: payload.weather.stale
            )
            remainingToday = Self.countToday(payload.events)
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    /// Counts events starting before midnight tonight. A two-day window is
    /// requested because "today" is a local-calendar question the API does not
    /// answer, so tomorrow's events have to be filtered out here.
    private static func countToday(_ events: [Payload.EventDTO]) -> Int {
        let calendar = Calendar.current
        guard let endOfToday = calendar.dateInterval(of: .day, for: Date())?.end else { return 0 }
        return events.filter { Date(timeIntervalSince1970: TimeInterval($0.start)) < endOfToday }.count
    }
}
