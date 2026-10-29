import CoreLocation
import Foundation
import Testing
@testable import PhotoSync

struct PlaceLookupTests {
    private let defaults = TestDefaults.make()
    private let coordinate = CLLocationCoordinate2D(latitude: 43.65, longitude: -79.38)

    private func lookup(_ resolver: @escaping PlaceLookup.Resolver) -> PlaceLookup {
        PlaceLookup(defaults: defaults, pause: .zero, timeout: .seconds(1), resolver: resolver)
    }

    @Test func joinsComponentsAndDropsRepeats() {
        #expect(PlaceLookup.name(from: ["Toronto", "Ontario", "Canada"]) == "Toronto, Ontario, Canada")
        #expect(PlaceLookup.name(from: ["Singapore", "Singapore"]) == "Singapore")
        #expect(PlaceLookup.name(from: []) == nil)
    }

    @Test func returnsNothingWithoutACoordinate() async throws {
        let place = try await lookup { _ in ["x"] }.place(for: nil)

        #expect(place == nil)
    }

    @Test func cachesResolvedPlaces() async throws {
        let places = lookup { _ in ["Toronto", "Canada"] }
        _ = try await places.place(for: coordinate)

        places.flush()
        let reloaded = try await lookup { _ in Issue.record("resolver should not run"); return [] }.place(for: coordinate)
        #expect(reloaded == "Toronto, Canada")
    }

    @Test func aFailedLookupYieldsNoPlace() async throws {
        struct Failure: Error {}

        let place = try await lookup { _ in throw Failure() }.place(for: coordinate)

        #expect(place == nil)
    }
}
