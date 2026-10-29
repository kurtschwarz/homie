import Foundation
import CoreLocation

// Each entry lives under its own defaults key ("<key>.<entry>"), so a flush
// writes only the entries that changed. A legacy whole-dictionary value under
// "<key>" is migrated to per-entry keys on load.
private struct PersistedMap {
    private let key: String
    private(set) var values: [String: String]
    private var changed: Set<String> = []

    private var prefix: String { key + "." }

    init(key: String, defaults: UserDefaults) {
        self.key = key
        let prefix = key + "."
        var loaded = defaults.dictionary(forKey: key) as? [String: String] ?? [:]
        for (storedKey, value) in defaults.dictionaryRepresentation() where storedKey.hasPrefix(prefix) {
            if let value = value as? String { loaded[String(storedKey.dropFirst(prefix.count))] = value }
        }
        values = loaded

        if defaults.object(forKey: key) != nil {
            for (entry, value) in loaded { defaults.set(value, forKey: prefix + entry) }
            defaults.removeObject(forKey: key)
        }
    }

    subscript(entry: String) -> String? {
        get { values[entry] }
        set {
            values[entry] = newValue
            changed.insert(entry)
        }
    }

    mutating func flush(to defaults: UserDefaults) {
        for entry in changed {
            if let value = values[entry] {
                defaults.set(value, forKey: prefix + entry)
            } else {
                defaults.removeObject(forKey: prefix + entry)
            }
        }
        changed.removeAll()
    }

    mutating func discardChanges() { changed.removeAll() }

    static func removeAll(key: String, from defaults: UserDefaults) {
        defaults.removeObject(forKey: key)
        let prefix = key + "."
        for storedKey in defaults.dictionaryRepresentation().keys where storedKey.hasPrefix(prefix) {
            defaults.removeObject(forKey: storedKey)
        }
    }
}

struct UploadLedger {
    private static let versionsKey = "photo-sync.uploadedVersions"
    private static let placesKey = "photo-sync.geocodedPlaces"
    private static let pushedDrawingsKey = "photo-sync.pushedDrawings"
    private static let generationKey = "photo-sync.ledgerGeneration"

    private let defaults: UserDefaults
    private var generation: Int

    private var uploadedVersions: PersistedMap
    private var pushedDrawings: PersistedMap
    private var geocodedPlaces: PersistedMap

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
        generation = defaults.integer(forKey: Self.generationKey)
        uploadedVersions = PersistedMap(key: Self.versionsKey, defaults: defaults)
        pushedDrawings = PersistedMap(key: Self.pushedDrawingsKey, defaults: defaults)
        geocodedPlaces = PersistedMap(key: Self.placesKey, defaults: defaults)
    }

    func isUploaded(_ photo: PhotoAsset) -> Bool { uploadedVersions[photo.id] == photo.version }

    mutating func markUploaded(_ photo: PhotoAsset) {
        uploadedVersions[photo.id] = photo.version
    }

    // Remembers drawings already added to Photos so a failed markDrawingPushed
    // is retried without creating a second asset.
    func pushedAssetID(forDrawing id: String) -> String? { pushedDrawings[id] }

    mutating func savePushedDrawing(_ id: String, assetID: String) {
        pushedDrawings[id] = assetID
    }

    mutating func forgetPushedDrawing(_ id: String) {
        pushedDrawings[id] = nil
    }

    func place(for coordinate: CLLocationCoordinate2D) -> String? { geocodedPlaces[key(for: coordinate)] }

    mutating func savePlace(_ place: String, for coordinate: CLLocationCoordinate2D) {
        geocodedPlaces[key(for: coordinate)] = place
    }

    // Writes only entries changed since the last flush, and nothing once another ledger has cleared.
    mutating func flush() {
        guard isCurrent else {
            uploadedVersions.discardChanges()
            pushedDrawings.discardChanges()
            geocodedPlaces.discardChanges()
            return
        }
        uploadedVersions.flush(to: defaults)
        pushedDrawings.flush(to: defaults)
        geocodedPlaces.flush(to: defaults)
    }

    static func clear(defaults: UserDefaults = .standard) {
        var ledger = UploadLedger(defaults: defaults)
        ledger.clear()
    }

    mutating func clear() {
        generation += 1
        defaults.set(generation, forKey: Self.generationKey)
        PersistedMap.removeAll(key: Self.versionsKey, from: defaults)
        PersistedMap.removeAll(key: Self.placesKey, from: defaults)
        PersistedMap.removeAll(key: Self.pushedDrawingsKey, from: defaults)
        uploadedVersions = PersistedMap(key: Self.versionsKey, defaults: defaults)
        pushedDrawings = PersistedMap(key: Self.pushedDrawingsKey, defaults: defaults)
        geocodedPlaces = PersistedMap(key: Self.placesKey, defaults: defaults)
    }

    private var isCurrent: Bool {
        generation == defaults.integer(forKey: Self.generationKey)
    }

    private func key(for coordinate: CLLocationCoordinate2D) -> String {
        String(format: "%.3f,%.3f", coordinate.latitude, coordinate.longitude)
    }
}
