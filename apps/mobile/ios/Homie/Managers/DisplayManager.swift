import Foundation
import Observation

@MainActor
@Observable
final class DisplayManager {
    struct Photo: Decodable {
        let url: String
        let width: Int
        let height: Int
        let place: String?
        let focus: Focus?
    }

    struct Focus: Decodable {
        let x: Double
        let y: Double
        let width: Double
        let height: Double
    }

    private(set) var currentPhoto: Photo?
    private(set) var currentPhotoURL: URL?
    private(set) var endsAt: Date?
    private(set) var loadGeneration = 0
    private(set) var hasLoaded = false

    private struct Payload: Decodable {
        struct Showing: Decodable {
            let endsAt: Int
            let photo: Photo
        }
        struct State: Decodable {
            let current: Showing?
        }
        let photoShowing: State
    }

    func load() async {
        defer { loadGeneration += 1; hasLoaded = true }
        do {
            let payload: Payload = try await HomieAPI.perform(
                "{ photoShowing { current { endsAt photo { url width height place focus { x y width height } } } } }"
            )
            currentPhoto = payload.photoShowing.current?.photo
            currentPhotoURL = payload.photoShowing.current.flatMap { URL(string: $0.photo.url, relativeTo: HomieEndpoint.baseURL)?.absoluteURL }
            endsAt = payload.photoShowing.current.map { Date(timeIntervalSince1970: TimeInterval($0.endsAt)) }
        } catch {
            // Keep the last known photo while the API is temporarily unavailable.
        }
    }

    var refreshDelay: Duration {
        guard let endsAt else { return .seconds(60) }
        let seconds = max(15, endsAt.timeIntervalSinceNow + 1)
        return .seconds(seconds)
    }
}
