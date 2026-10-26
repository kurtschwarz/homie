import CoreLocation
import Foundation
import Photos

struct PhotoAlbum: Identifiable, Hashable {
    let id: String
    let title: String
    let count: Int
    let isShared: Bool
    let cover: PHAsset?
}

struct PhotoAsset: Identifiable {
    let id: String
    let asset: PHAsset
    let takenAt: Date?
    let version: String
    let pixelWidth: Int
    let pixelHeight: Int
    let coordinate: CLLocationCoordinate2D?
    let caption: String?

    init(_ asset: PHAsset) {
        self.id = asset.localIdentifier
        self.asset = asset
        self.takenAt = asset.creationDate
        self.version = String(Int((asset.modificationDate ?? asset.creationDate ?? .distantPast).timeIntervalSince1970))
        self.pixelWidth = asset.pixelWidth
        self.pixelHeight = asset.pixelHeight
        self.coordinate = asset.location?.coordinate
        self.caption = asset.value(forKey: "title") as? String
    }
}
