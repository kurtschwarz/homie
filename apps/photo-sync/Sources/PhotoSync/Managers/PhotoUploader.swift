import Foundation
import Observation

struct AlbumProgress: Equatable {
    var completed = 0
    var total = 0
    var failed = 0
    var isDone: Bool { total > 0 && completed >= total }
}

protocol PhotoSource {
    func assets(inAlbum id: String) -> [PhotoAsset]
    func fullSizeJPEG(for photo: PhotoAsset) async throws -> PhotoJPEG
}

protocol PhotoUploadService {
    func upload(_ input: PhotoUploadInput, jpeg: Data, filename: String) async throws -> UploadedPhoto
}

extension PhotoLibrary: PhotoSource {}
extension HomieAPI: PhotoUploadService {}

@Observable
final class PhotoUploader {
    private(set) var completed = 0
    private(set) var total = 0
    private(set) var albumProgress: [String: AlbumProgress] = [:]

    var failed: Int { albumProgress.values.reduce(0) { $0 + $1.failed } }

    private static let ledgerFlushInterval = 50

    private let source: PhotoSource
    private let service: PhotoUploadService
    private let places: PlaceLookup
    private let defaults: UserDefaults
    private var ledger: UploadLedger

    init(
        source: PhotoSource,
        service: PhotoUploadService,
        places: PlaceLookup,
        defaults: UserDefaults = .standard
    ) {
        self.source = source
        self.service = service
        self.places = places
        self.defaults = defaults
        ledger = UploadLedger(defaults: defaults)
    }

    nonisolated static func uploadable<Asset: Identifiable>(
        _ assets: [Asset],
        excluding drawingAssetIDs: Set<String>
    ) -> [Asset] where Asset.ID == String {
        assets.filter { !drawingAssetIDs.contains($0.id) }
    }

    func upload(albumIDs: [String], excluding drawingAssetIDs: Set<String>, run: SyncRunLog) async throws {
        ledger = UploadLedger(defaults: defaults)
        places.reload()
        let batches = albumIDs.map { id in
            (id, Self.uploadable(source.assets(inAlbum: id), excluding: drawingAssetIDs))
        }
        albumProgress = Dictionary(uniqueKeysWithValues: batches.map { ($0.0, AlbumProgress(total: $0.1.count)) })
        total = batches.reduce(0) { $0 + $1.1.count }
        completed = 0
        run.info("Photos to check", detail: "albums=\(batches.count) photos=\(total)")

        for (albumID, photos) in batches {
            try await uploadPhotos(photos, albumID: albumID, run: run)
        }
    }

    private func finish(albumID: String, failed: Bool = false) {
        completed += 1
        albumProgress[albumID]?.completed += 1
        if failed { albumProgress[albumID]?.failed += 1 }
    }

    private func uploadPhotos(_ photos: [PhotoAsset], albumID: String, run: SyncRunLog) async throws {
        defer {
            ledger.flush()
            places.flush()
        }
        for photo in photos {
            if ledger.isUploaded(photo) {
                finish(albumID: albumID)
                continue
            }
            guard let takenAt = photo.takenAt else {
                run.error("Photo skipped: no capture date", detail: "asset=\(photo.id)")
                finish(albumID: albumID, failed: true)
                continue
            }

            do {
                let place = try await places.place(for: photo.coordinate)
                let jpeg = try await source.fullSizeJPEG(for: photo)
                let coordinate = photo.coordinate
                let input = PhotoUploadInput(
                    assetID: photo.id,
                    version: photo.version,
                    takenAt: Int(takenAt.timeIntervalSince1970),
                    width: jpeg.width,
                    height: jpeg.height,
                    latitude: coordinate?.latitude,
                    longitude: coordinate?.longitude,
                    place: place,
                    focus: Saliency.focalRegion(for: jpeg.cgImage)
                )
                _ = try await service.upload(input, jpeg: jpeg.data, filename: "\(photo.id).jpg")
                ledger.markUploaded(photo)
                finish(albumID: albumID)
                if completed.isMultiple(of: Self.ledgerFlushInterval) {
                    ledger.flush()
                    places.flush()
                }
            } catch {
                if error is CancellationError {
                    run.info("Sync cancelled")
                    throw error
                }
                run.error("Upload failed", detail: "asset=\(photo.id) error=\(error.localizedDescription)")
                finish(albumID: albumID, failed: true)
                if case HomieAPIError.unauthorized = error {
                    run.error("Session expired; stopping")
                    throw SessionExpired()
                }
            }
        }
    }
}
