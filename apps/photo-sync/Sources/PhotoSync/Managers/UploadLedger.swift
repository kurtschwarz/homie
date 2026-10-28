import Foundation
import CoreLocation

@MainActor
final class UploadLedger {
    private let defaults: UserDefaults
    private let versionsKey = "photo-sync.uploadedVersions"
    private let placesKey = "photo-sync.geocodedPlaces"
    private let pushedDrawingsKey = "photo-sync.pushedDrawings"
    private let generationKey = "photo-sync.ledgerGeneration"
    private var generation: Int

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
        generation = defaults.integer(forKey: generationKey)
    }

    private var versions: [String: String] {
        get { defaults.dictionary(forKey: versionsKey) as? [String: String] ?? [:] }
        set { defaults.set(newValue, forKey: versionsKey) }
    }

    private var pushedDrawings: [String: String] {
        get { defaults.dictionary(forKey: pushedDrawingsKey) as? [String: String] ?? [:] }
        set { defaults.set(newValue, forKey: pushedDrawingsKey) }
    }

    private var places: [String: String] {
        get { defaults.dictionary(forKey: placesKey) as? [String: String] ?? [:] }
        set { defaults.set(newValue, forKey: placesKey) }
    }

    func isUploaded(_ photo: PhotoAsset) -> Bool { versions[photo.id] == photo.version }

    func markUploaded(_ photo: PhotoAsset) {
        guard isCurrent else { return }
        var value = versions
        value[photo.id] = photo.version
        versions = value
    }

    // Remembers drawings already added to Photos so a failed markDrawingPushed
    // is retried without creating a second asset.
    func pushedAssetID(forDrawing id: String) -> String? { pushedDrawings[id] }

    func savePushedDrawing(_ id: String, assetID: String) {
        guard isCurrent else { return }
        var value = pushedDrawings
        value[id] = assetID
        pushedDrawings = value
    }

    func forgetPushedDrawing(_ id: String) {
        guard isCurrent else { return }
        var value = pushedDrawings
        value.removeValue(forKey: id)
        pushedDrawings = value
    }

    func place(for coordinate: CLLocationCoordinate2D) -> String? { places[key(for: coordinate)] }

    func savePlace(_ place: String, for coordinate: CLLocationCoordinate2D) {
        guard isCurrent else { return }
        var value = places
        value[key(for: coordinate)] = place
        places = value
    }

    func clear() {
        generation += 1
        defaults.set(generation, forKey: generationKey)
        defaults.removeObject(forKey: versionsKey)
        defaults.removeObject(forKey: placesKey)
        defaults.removeObject(forKey: pushedDrawingsKey)
    }

    private var isCurrent: Bool {
        generation == defaults.integer(forKey: generationKey)
    }

    private func key(for coordinate: CLLocationCoordinate2D) -> String {
        String(format: "%.3f,%.3f", coordinate.latitude, coordinate.longitude)
    }
}
