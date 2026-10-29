@preconcurrency import CoreLocation
import Foundation
import Observation

struct AlbumProgress: Equatable {
    var completed = 0
    var total = 0
    var failed = 0
    var isDone: Bool { total > 0 && completed >= total }
}

@Observable
final class PhotoUploader {
    private(set) var completed = 0
    private(set) var total = 0
    private(set) var isUploading = false
    private(set) var albumProgress: [String: AlbumProgress] = [:]
    private(set) var drawingProgress: AlbumProgress?
    private(set) var drawingMessage: String?

    private static let ledgerFlushInterval = 50

    private let library: PhotoLibrary
    private let api: HomieAPI
    private var ledger = UploadLedger()
    private let geocoder = CLGeocoder()
    private let onAuthenticationRequired: () -> Void

    init(library: PhotoLibrary, api: HomieAPI, onAuthenticationRequired: @escaping () -> Void) {
        self.library = library
        self.api = api
        self.onAuthenticationRequired = onAuthenticationRequired
    }

    func sync(albumIDs: [String], artworkAlbumID: String? = nil, onFinished: @escaping () -> Void = {}) {
        guard !isUploading else { return }
        isUploading = true
        let run = SyncRunLog()
        Task {
            run.info("Sync started", detail: "albums=\(albumIDs.count)")
            await syncAll(albumIDs, artworkAlbumID: artworkAlbumID, run: run)
            run.info(
                "Sync finished",
                detail: "completed=\(completed) total=\(total) failed=\(albumProgress.values.reduce(0) { $0 + $1.failed })"
            )
            isUploading = false
            onFinished()
        }
    }

    private func syncAll(_ albumIDs: [String], artworkAlbumID: String?, run: SyncRunLog) async {
        ledger = UploadLedger()
        drawingProgress = nil
        drawingMessage = nil
        let shouldContinue = await pushDrawings(toAlbum: artworkAlbumID, run: run)
        guard shouldContinue else { return }
        let drawingAssetIDs: Set<String>
        do {
            drawingAssetIDs = try await api.pushedDrawingAssetIDs()
        } catch {
            run.error("Could not load pushed drawings; aborting", detail: error.localizedDescription)
            if case HomieAPIError.unauthorized = error { onAuthenticationRequired() }
            return
        }
        await syncAlbums(albumIDs, excluding: drawingAssetIDs, run: run)
    }

    private func pushDrawings(toAlbum albumID: String?, run: SyncRunLog) async -> Bool {
        guard let albumID, library.canAddDrawings(toAlbum: albumID) else {
            drawingMessage = "Choose an artwork album to save Homie drawings to Photos."
            run.info("Skipping drawings: no artwork album")
            return true
        }
        let drawings: [HomieDrawing]
        do {
            drawings = try await api.unpushedDrawings()
            run.info("Drawings to push", detail: "count=\(drawings.count)")
        } catch {
            run.error("Could not list drawings", detail: error.localizedDescription)
            drawingMessage = error.localizedDescription
            if case HomieAPIError.unauthorized = error {
                onAuthenticationRequired()
                return false
            }
            return true
        }
        guard !drawings.isEmpty else { return true }
        defer { ledger.flush() }

        var progress = AlbumProgress(total: drawings.count)
        drawingProgress = progress
        for drawing in drawings {
            do {
                let assetID: String
                if let known = ledger.pushedAssetID(forDrawing: drawing.id) {
                    assetID = known
                } else {
                    let jpeg = try await api.downloadDrawing(drawing)
                    assetID = try await library.addDrawing(
                        jpeg,
                        createdAt: Date(timeIntervalSince1970: TimeInterval(drawing.takenAt)),
                        toAlbum: albumID
                    )
                    ledger.savePushedDrawing(drawing.id, assetID: assetID)
                }
                try await api.markDrawingPushed(id: drawing.id, assetID: assetID)
                ledger.forgetPushedDrawing(drawing.id)
                progress.completed += 1
            } catch {
                progress.completed += 1
                progress.failed += 1
                run.error("Drawing failed", detail: "id=\(drawing.id) error=\(error.localizedDescription)")
                drawingMessage = "Drawing didn't sync: \(error.localizedDescription)"
                if case HomieAPIError.unauthorized = error {
                    drawingProgress = progress
                    onAuthenticationRequired()
                    return false
                }
            }
            drawingProgress = progress
        }
        return true
    }

    nonisolated static func uploadable<Asset: Identifiable>(
        _ assets: [Asset],
        excluding drawingAssetIDs: Set<String>
    ) -> [Asset] where Asset.ID == String {
        assets.filter { !drawingAssetIDs.contains($0.id) }
    }

    private func syncAlbums(_ albumIDs: [String], excluding drawingAssetIDs: Set<String>, run: SyncRunLog) async {
        let batches = albumIDs.map { id in
            (id, Self.uploadable(library.assets(inAlbum: id), excluding: drawingAssetIDs))
        }
        albumProgress = Dictionary(uniqueKeysWithValues: batches.map { ($0.0, AlbumProgress(total: $0.1.count)) })
        total = batches.reduce(0) { $0 + $1.1.count }
        completed = 0
        run.info("Photos to check", detail: "albums=\(batches.count) photos=\(total)")

        for (albumID, photos) in batches {
            let shouldContinue = await uploadPhotos(photos, albumID: albumID, run: run)
            if !shouldContinue { break }
        }
    }

    private func finish(albumID: String, failed: Bool = false) {
        completed += 1
        albumProgress[albumID]?.completed += 1
        if failed { albumProgress[albumID]?.failed += 1 }
    }

    private func uploadPhotos(_ photos: [PhotoAsset], albumID: String, run: SyncRunLog) async -> Bool {
        defer { ledger.flush() }
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
                let place = try await bestEffortPlace(for: photo.coordinate)
                let jpeg = try await library.fullSizeJPEG(for: photo)
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
                _ = try await api.upload(input, jpeg: jpeg.data, filename: "\(photo.id).jpg")
                ledger.markUploaded(photo)
                finish(albumID: albumID)
                if completed.isMultiple(of: Self.ledgerFlushInterval) { ledger.flush() }
            } catch {
                if error is CancellationError {
                    run.info("Sync cancelled")
                    return false
                }
                run.error("Upload failed", detail: "asset=\(photo.id) error=\(error.localizedDescription)")
                finish(albumID: albumID, failed: true)
                if case HomieAPIError.unauthorized = error {
                    run.error("Session expired; stopping")
                    onAuthenticationRequired()
                    return false
                }
            }
        }
        return true
    }

    private func bestEffortPlace(for coordinate: CLLocationCoordinate2D?) async throws -> String? {
        guard let coordinate else { return nil }
        if let cached = ledger.place(for: coordinate) { return cached }
        // Keep requests serial and leave time for PhotoKit/iCloud between lookups.
        try await Task.sleep(for: .seconds(1))
        let placemarks: [CLPlacemark]?
        do {
            placemarks = try await withTimeout(.seconds(5)) { [geocoder] in
                let location = CLLocation(latitude: coordinate.latitude, longitude: coordinate.longitude)
                return try await withTaskCancellationHandler {
                    try await geocoder.reverseGeocodeLocation(location)
                } onCancel: {
                    geocoder.cancelGeocode()
                }
            }
        } catch {
            try Task.checkCancellation()
            Log.sync.error("Place lookup failed: \(error.localizedDescription, privacy: .private)")
            return nil
        }
        guard let placemark = placemarks?.first else { return nil }
        let place = [placemark.locality, placemark.administrativeArea, placemark.country]
            .compactMap { $0 }
            .reduce(into: [String]()) { result, item in if result.last != item { result.append(item) } }
            .joined(separator: ", ")
        guard !place.isEmpty else { return nil }
        ledger.savePlace(place, for: coordinate)
        return place
    }
}
