import CoreLocation
import Combine
import Foundation

enum UploadState {
    case uploading
    case uploaded(String)
    case skipped
    case failed(String)
}

@MainActor
final class PhotoUploader: ObservableObject {
    @Published private(set) var states: [String: UploadState] = [:]
    @Published private(set) var completed = 0
    @Published private(set) var total = 0
    @Published private(set) var isUploading = false

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

    func clearLocalState() {
        guard !isUploading else { return }
        ledger.clear()
        states.removeAll()
        completed = 0
        total = 0
    }

    func upload(_ photo: PhotoAsset) {
        guard !isUploading else { return }
        Task { await uploadPhotos([photo], lookupPlaces: true) }
    }

    func uploadAll(_ photos: [PhotoAsset], lookupPlaces: Bool = true) {
        guard !isUploading else { return }
        Task { await uploadPhotos(photos, lookupPlaces: lookupPlaces) }
    }

    private func uploadPhotos(_ photos: [PhotoAsset], lookupPlaces: Bool) async {
        isUploading = true
        total = photos.count
        completed = 0
        defer { isUploading = false }

        for photo in photos {
            if ledger.isUploaded(photo) {
                states[photo.id] = .skipped
                completed += 1
                continue
            }
            guard let takenAt = photo.takenAt else {
                states[photo.id] = .failed("Skipped: capture date is unavailable.")
                completed += 1
                continue
            }

            states[photo.id] = .uploading
            do {
                let place = lookupPlaces ? await bestEffortPlace(for: photo.coordinate) : photo.coordinate.flatMap { ledger.place(for: $0) }
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
            } catch {
                states[photo.id] = .failed(error.localizedDescription)
                if case HomieAPIError.unauthorized = error {
                    onAuthenticationRequired()
                    completed += 1
                    break
                }
            }
            completed += 1
        }
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
