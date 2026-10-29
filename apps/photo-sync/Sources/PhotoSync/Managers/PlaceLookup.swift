@preconcurrency import CoreLocation
import Foundation

final class PlaceLookup {
    typealias Resolver = @Sendable (CLLocationCoordinate2D) async throws -> [String]

    private let defaults: UserDefaults
    private let pause: Duration
    private let timeout: Duration
    private let resolver: Resolver
    private var ledger: UploadLedger

    init(
        defaults: UserDefaults = .standard,
        pause: Duration = .seconds(1),
        timeout: Duration = .seconds(5),
        resolver: @escaping Resolver = PlaceLookup.geocode
    ) {
        self.defaults = defaults
        self.pause = pause
        self.timeout = timeout
        self.resolver = resolver
        ledger = UploadLedger(defaults: defaults)
    }

    func reload() { ledger = UploadLedger(defaults: defaults) }

    func flush() { ledger.flush() }

    func place(for coordinate: CLLocationCoordinate2D?) async throws -> String? {
        guard let coordinate else { return nil }
        if let cached = ledger.place(for: coordinate) { return cached }
        // Keep requests serial and leave time for PhotoKit/iCloud between lookups.
        try await Task.sleep(for: pause)
        let components: [String]?
        do {
            components = try await withTimeout(timeout) { [resolver] in try await resolver(coordinate) }
        } catch {
            try Task.checkCancellation()
            Log.sync.error("Place lookup failed: \(error.localizedDescription, privacy: .private)")
            return nil
        }
        guard let place = Self.name(from: components ?? []) else { return nil }
        ledger.savePlace(place, for: coordinate)
        return place
    }

    nonisolated static func name(from components: [String]) -> String? {
        let place = components
            .reduce(into: [String]()) { result, item in if result.last != item { result.append(item) } }
            .joined(separator: ", ")
        return place.isEmpty ? nil : place
    }

    nonisolated static func geocode(_ coordinate: CLLocationCoordinate2D) async throws -> [String] {
        let geocoder = CLGeocoder()
        let location = CLLocation(latitude: coordinate.latitude, longitude: coordinate.longitude)
        let placemarks = try await withTaskCancellationHandler {
            try await geocoder.reverseGeocodeLocation(location)
        } onCancel: {
            geocoder.cancelGeocode()
        }
        guard let placemark = placemarks.first else { return [] }
        return [placemark.locality, placemark.administrativeArea, placemark.country].compactMap { $0 }
    }
}
