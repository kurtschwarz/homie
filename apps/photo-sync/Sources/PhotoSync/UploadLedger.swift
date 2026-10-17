import Foundation
import CoreLocation

@MainActor
final class UploadLedger {
    private let defaults = UserDefaults.standard
    private let versionsKey = "photo-sync.uploadedVersions"
    private let placesKey = "photo-sync.geocodedPlaces"

    private var versions: [String: String] {
        get { defaults.dictionary(forKey: versionsKey) as? [String: String] ?? [:] }
        set { defaults.set(newValue, forKey: versionsKey) }
    }

    private var places: [String: String] {
        get { defaults.dictionary(forKey: placesKey) as? [String: String] ?? [:] }
        set { defaults.set(newValue, forKey: placesKey) }
    }

    func isUploaded(_ photo: PhotoAsset) -> Bool { versions[photo.id] == photo.version }

    func markUploaded(_ photo: PhotoAsset) {
        var value = versions
        value[photo.id] = photo.version
        versions = value
    }

    func place(for coordinate: CLLocationCoordinate2D) -> String? { places[key(for: coordinate)] }

    func savePlace(_ place: String, for coordinate: CLLocationCoordinate2D) {
        var value = places
        value[key(for: coordinate)] = place
        places = value
    }

    func clear() {
        defaults.removeObject(forKey: versionsKey)
        defaults.removeObject(forKey: placesKey)
    }

    private func key(for coordinate: CLLocationCoordinate2D) -> String {
        String(format: "%.3f,%.3f", coordinate.latitude, coordinate.longitude)
    }
}
