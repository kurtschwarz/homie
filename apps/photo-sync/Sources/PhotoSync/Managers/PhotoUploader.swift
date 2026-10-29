import CoreLocation
import Combine
import Foundation

struct AlbumProgress: Equatable {
    var completed = 0
    var total = 0
    var failed = 0
    var isDone: Bool { total > 0 && completed >= total }
}

@MainActor
final class PhotoUploader: ObservableObject {
    @Published private(set) var completed = 0
    @Published private(set) var total = 0
    @Published private(set) var isUploading = false
    @Published private(set) var albumProgress: [String: AlbumProgress] = [:]
    @Published private(set) var drawingProgress: AlbumProgress?
    @Published private(set) var drawingMessage: String?

    private let library: PhotoLibrary
    private let api: HomieAPI
    private let ledger = UploadLedger()
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
        Task {
            await syncAll(albumIDs, artworkAlbumID: artworkAlbumID)
            isUploading = false
            onFinished()
        }
    }

    private func syncAll(_ albumIDs: [String], artworkAlbumID: String?) async {
        drawingProgress = nil
        drawingMessage = nil
        let shouldContinue = await pushDrawings(toAlbum: artworkAlbumID)
        guard shouldContinue else { return }
        let drawingAssetIDs: Set<String>
        do {
            drawingAssetIDs = try await api.pushedDrawingAssetIDs()
        } catch {
            if case HomieAPIError.unauthorized = error { onAuthenticationRequired() }
            return
        }
        await syncAlbums(albumIDs, excluding: drawingAssetIDs)
    }

    private func pushDrawings(toAlbum albumID: String?) async -> Bool {
        guard let albumID, library.canAddDrawings(toAlbum: albumID) else {
            drawingMessage = "Choose an artwork album to save Homie drawings to Photos."
            return true
        }
        let drawings: [HomieDrawing]
        do {
            drawings = try await api.unpushedDrawings()
        } catch {
            drawingMessage = error.localizedDescription
            if case HomieAPIError.unauthorized = error {
                onAuthenticationRequired()
                return false
            }
            return true
        }
        guard !drawings.isEmpty else { return true }

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

    private func syncAlbums(_ albumIDs: [String], excluding drawingAssetIDs: Set<String>) async {
        let batches = albumIDs.map { id in
            (id, Self.uploadable(library.assets(inAlbum: id), excluding: drawingAssetIDs))
        }
        albumProgress = Dictionary(uniqueKeysWithValues: batches.map { ($0.0, AlbumProgress(total: $0.1.count)) })
        total = batches.reduce(0) { $0 + $1.1.count }
        completed = 0

        for (albumID, photos) in batches {
            let shouldContinue = await uploadPhotos(photos, albumID: albumID)
            if !shouldContinue { break }
        }
    }

    private func finish(albumID: String, failed: Bool = false) {
        completed += 1
        albumProgress[albumID]?.completed += 1
        if failed { albumProgress[albumID]?.failed += 1 }
    }

    private func uploadPhotos(_ photos: [PhotoAsset], albumID: String) async -> Bool {
        for photo in photos {
            if ledger.isUploaded(photo) {
                finish(albumID: albumID)
                continue
            }
            guard let takenAt = photo.takenAt else {
                finish(albumID: albumID, failed: true)
                continue
            }

            do {
                let place = await bestEffortPlace(for: photo.coordinate)
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
            } catch {
                finish(albumID: albumID, failed: true)
                if case HomieAPIError.unauthorized = error {
                    onAuthenticationRequired()
                    return false
                }
            }
        }
        return true
    }

    private func bestEffortPlace(for coordinate: CLLocationCoordinate2D?) async -> String? {
        guard let coordinate else { return nil }
        if let cached = ledger.place(for: coordinate) { return cached }
        do {
            // Keep requests serial and leave time for PhotoKit/iCloud between lookups.
            try await Task.sleep(for: .seconds(1))
            let placemarks = try await geocoderWithTimeout(for: coordinate)
            guard let placemark = placemarks.first else { return nil }
            let place = [placemark.locality, placemark.administrativeArea, placemark.country]
                .compactMap { $0 }
                .reduce(into: [String]()) { result, item in if result.last != item { result.append(item) } }
                .joined(separator: ", ")
            guard !place.isEmpty else { return nil }
            ledger.savePlace(place, for: coordinate)
            return place
        } catch {
            return nil
        }
    }

    private func geocoderWithTimeout(for coordinate: CLLocationCoordinate2D) async throws -> [CLPlacemark] {
        try await withCheckedThrowingContinuation { continuation in
            let lock = NSLock()
            var finished = false
            func finish(_ result: Result<[CLPlacemark], Error>) {
                lock.lock()
                defer { lock.unlock() }
                guard !finished else { return }
                finished = true
                continuation.resume(with: result)
            }

            geocoder.reverseGeocodeLocation(CLLocation(latitude: coordinate.latitude, longitude: coordinate.longitude)) { placemarks, error in
                if let error { finish(.failure(error)) }
                else { finish(.success(placemarks ?? [])) }
            }
            DispatchQueue.main.asyncAfter(deadline: .now() + 5) {
                self.geocoder.cancelGeocode()
                finish(.failure(PhotoUploaderError.geocodingTimedOut))
            }
        }
    }
}

private enum PhotoUploaderError: LocalizedError {
    case geocodingTimedOut
    var errorDescription: String? { "Place lookup timed out." }
}
