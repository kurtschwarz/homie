import CoreLocation
import Foundation
import Photos

struct PhotoAsset: Identifiable {
    let id: String
    let asset: PHAsset
    let takenAt: Date?
    let version: String
    let coordinate: CLLocationCoordinate2D?

    init(_ asset: PHAsset) {
        self.id = asset.localIdentifier
        self.asset = asset
        self.takenAt = asset.creationDate
        self.version = String(Int((asset.modificationDate ?? asset.creationDate ?? .distantPast).timeIntervalSince1970))
        self.coordinate = asset.location?.coordinate
    }
}
