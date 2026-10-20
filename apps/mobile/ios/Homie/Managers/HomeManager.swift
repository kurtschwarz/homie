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

    struct Headline {
        let label: String
        let title: String
        let detail: String?
    }

    private(set) var weather: Weather?
    /// Events still to come today. The API's events window starts at "now", so
    /// anything that already finished today is not counted.
    private(set) var remainingToday = 0
    private(set) var headline: Headline?
    private(set) var errorMessage: String?
    private(set) var hasLoaded = false

    private struct Payload: Decodable {
        struct WeatherDTO: Decodable {
            let condition: String
            let temperature: Int
            let stale: Bool
        }
        struct EventDTO: Decodable {
            let id: String
            let title: String
            let location: String?
            let start: Int
            let end: Int
            let allDay: Bool
        }
        let weather: WeatherDTO
        let events: [EventDTO]
    }

    func load() async {
        defer { hasLoaded = true }
        do {
            let payload: Payload = try await HomieAPI.perform(
                "{ weather { condition temperature stale } events(days: 21) { id title location start end allDay } }"
            )
            weather = Weather(
                condition: payload.weather.condition,
                temperature: payload.weather.temperature,
                stale: payload.weather.stale
            )
            remainingToday = Self.countToday(payload.events)
            headline = Self.headline(payload.events)
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

    private static func headline(_ events: [Payload.EventDTO], now: Date = Date()) -> Headline? {
        let upcoming = events
            .filter { !$0.allDay && Date(timeIntervalSince1970: TimeInterval($0.end)) > now }
            .sorted { $0.start < $1.start }
        guard let event = upcoming.first else { return nil }

        let start = Date(timeIntervalSince1970: TimeInterval(event.start))
        let end = Date(timeIntervalSince1970: TimeInterval(event.end))
        let isLive = start <= now
        let label = isLive ? "NOW · until \(Self.time(end))" : "NEXT"
        let timeRange = "\(Self.time(start))–\(Self.time(end))"
        let detail = [timeRange, event.location].compactMap { $0 }.joined(separator: " · ")
        return Headline(label: label, title: event.title, detail: detail)
    }

    private static func time(_ date: Date) -> String {
        date.formatted(.dateTime.hour().minute())
    }
}
