import Foundation
import Observation

@Observable
final class AlbumSelection {
    private static let albumsKey = "photo-sync.selectedAlbums"
    private static let keepSyncingKey = "photo-sync.keepSyncing"
    private static let lastSyncedKey = "photo-sync.lastSyncedAt"
    private static let artworkAlbumKey = "photo-sync.artworkAlbum"

    private(set) var artworkAlbumID: String?
    private(set) var albumIDs: Set<String>
    private(set) var keepSyncing: Bool
    private(set) var lastSyncedAt: Date?

    private let defaults: UserDefaults

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
        albumIDs = Set(defaults.stringArray(forKey: Self.albumsKey) ?? [])
        keepSyncing = defaults.object(forKey: Self.keepSyncingKey) as? Bool ?? true
        lastSyncedAt = defaults.object(forKey: Self.lastSyncedKey) as? Date
        artworkAlbumID = defaults.string(forKey: Self.artworkAlbumKey)
    }

    func setArtworkAlbum(_ id: String?) {
        artworkAlbumID = id
        if let id { defaults.set(id, forKey: Self.artworkAlbumKey) } else { defaults.removeObject(forKey: Self.artworkAlbumKey) }
    }

    var hasSelection: Bool { !albumIDs.isEmpty }

    func contains(_ id: String) -> Bool { albumIDs.contains(id) }

    func toggle(_ id: String) {
        if albumIDs.contains(id) { albumIDs.remove(id) } else { albumIDs.insert(id) }
        defaults.set(albumIDs.sorted(), forKey: Self.albumsKey)
    }

    func setKeepSyncing(_ value: Bool) {
        keepSyncing = value
        defaults.set(value, forKey: Self.keepSyncingKey)
    }

    func markSynced(at date: Date = Date()) {
        lastSyncedAt = date
        defaults.set(date, forKey: Self.lastSyncedKey)
    }

    static func clear(defaults: UserDefaults) {
        defaults.removeObject(forKey: albumsKey)
        defaults.removeObject(forKey: keepSyncingKey)
        defaults.removeObject(forKey: lastSyncedKey)
        defaults.removeObject(forKey: artworkAlbumKey)
    }
}
