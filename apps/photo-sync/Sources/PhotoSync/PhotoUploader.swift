import CoreLocation
import Combine
import Foundation

enum UploadState {
    case uploading
    case uploaded(String)
    case skipped
    case failed(String)
}

struct AlbumProgress: Equatable {
    var completed = 0
    var total = 0
    var failed = 0
    var isDone: Bool { total > 0 && completed >= total }
}

@MainActor
final class PhotoUploader: ObservableObject {
    @Published private(set) var states: [String: UploadState] = [:]
    @Published private(set) var completed = 0
    @Published private(set) var total = 0
    @Published private(set) var isUploading = false
    @Published private(set) var albumProgress: [String: AlbumProgress] = [:]

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

    func sync(albumIDs: [String], onFinished: @escaping () -> Void = {}) {
        guard !isUploading else { return }
        Task {
            await syncAlbums(albumIDs)
            onFinished()
        }
    }

    private func syncAlbums(_ albumIDs: [String]) async {
        let batches = albumIDs.map { ($0, library.assets(inAlbum: $0)) }
        isUploading = true
        defer { isUploading = false }
        states.removeAll()
        albumProgress = Dictionary(uniqueKeysWithValues: batches.map { ($0.0, AlbumProgress(total: $0.1.count)) })
        total = batches.reduce(0) { $0 + $1.1.count }
        completed = 0

        for (albumID, photos) in batches {
            let shouldContinue = await uploadPhotos(photos, albumID: albumID)
            if !shouldContinue { break }
        }
    }

    private func finish(_ photo: PhotoAsset, albumID: String, failed: Bool = false) {
        completed += 1
        albumProgress[albumID]?.completed += 1
        if failed { albumProgress[albumID]?.failed += 1 }
    }

    private func uploadPhotos(_ photos: [PhotoAsset], albumID: String) async -> Bool {
        for photo in photos {
            if ledger.isUploaded(photo) {
                states[photo.id] = .skipped
                finish(photo, albumID: albumID)
                continue
            }
            guard let takenAt = photo.takenAt else {
                states[photo.id] = .failed("Skipped: capture date is unavailable.")
                finish(photo, albumID: albumID, failed: true)
                continue
            }

            states[photo.id] = .uploading
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
                let uploaded = try await api.upload(input, jpeg: jpeg.data, filename: "\(photo.id).jpg")
                ledger.markUploaded(photo)
                states[photo.id] = .uploaded(uploaded.url)
                finish(photo, albumID: albumID)
            } catch {
                states[photo.id] = .failed(error.localizedDescription)
                finish(photo, albumID: albumID, failed: true)
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
