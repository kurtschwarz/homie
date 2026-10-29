import Foundation
import CoreLocation

struct UploadLedger {
    private let defaults: UserDefaults
    private let versionsKey = "photo-sync.uploadedVersions"
    private let placesKey = "photo-sync.geocodedPlaces"
    private let pushedDrawingsKey = "photo-sync.pushedDrawings"
    private let generationKey = "photo-sync.ledgerGeneration"
    private var generation: Int

    private var versions: [String: String]
    private var pushedDrawings: [String: String]
    private var places: [String: String]

    private var versionsDirty = false
    private var pushedDrawingsDirty = false
    private var placesDirty = false

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
        generation = defaults.integer(forKey: generationKey)
        versions = defaults.dictionary(forKey: versionsKey) as? [String: String] ?? [:]
        pushedDrawings = defaults.dictionary(forKey: pushedDrawingsKey) as? [String: String] ?? [:]
        places = defaults.dictionary(forKey: placesKey) as? [String: String] ?? [:]
    }

    func isUploaded(_ photo: PhotoAsset) -> Bool { versions[photo.id] == photo.version }

    mutating func markUploaded(_ photo: PhotoAsset) {
        versions[photo.id] = photo.version
        versionsDirty = true
    }

    // Remembers drawings already added to Photos so a failed markDrawingPushed
    // is retried without creating a second asset.
    func pushedAssetID(forDrawing id: String) -> String? { pushedDrawings[id] }

    mutating func savePushedDrawing(_ id: String, assetID: String) {
        pushedDrawings[id] = assetID
        pushedDrawingsDirty = true
    }

    mutating func forgetPushedDrawing(_ id: String) {
        pushedDrawings.removeValue(forKey: id)
        pushedDrawingsDirty = true
    }

    func place(for coordinate: CLLocationCoordinate2D) -> String? { places[key(for: coordinate)] }

    mutating func savePlace(_ place: String, for coordinate: CLLocationCoordinate2D) {
        places[key(for: coordinate)] = place
        placesDirty = true
    }

    // Writes only collections changed since the last flush, and nothing once another ledger has cleared.
    mutating func flush() {
        guard isCurrent else {
            versionsDirty = false
            pushedDrawingsDirty = false
            placesDirty = false
            return
        }
        if versionsDirty {
            defaults.set(versions, forKey: versionsKey)
            versionsDirty = false
        }
        if pushedDrawingsDirty {
            defaults.set(pushedDrawings, forKey: pushedDrawingsKey)
            pushedDrawingsDirty = false
        }
        if placesDirty {
            defaults.set(places, forKey: placesKey)
            placesDirty = false
        }
    }

    mutating func clear() {
        generation += 1
        defaults.set(generation, forKey: generationKey)
        defaults.removeObject(forKey: versionsKey)
        defaults.removeObject(forKey: placesKey)
        defaults.removeObject(forKey: pushedDrawingsKey)
        versions = [:]
        pushedDrawings = [:]
        places = [:]
        versionsDirty = false
        pushedDrawingsDirty = false
        placesDirty = false
    }

    private var isCurrent: Bool {
        generation == defaults.integer(forKey: generationKey)
    }

    private func key(for coordinate: CLLocationCoordinate2D) -> String {
        String(format: "%.3f,%.3f", coordinate.latitude, coordinate.longitude)
    }
}
