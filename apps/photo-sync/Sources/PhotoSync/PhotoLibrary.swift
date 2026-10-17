import AppKit
import Combine
import CoreLocation
import Photos
import Vision

struct PhotoJPEG {
    let data: Data
    let width: Int
    let height: Int
    let cgImage: CGImage
}

@MainActor
final class PhotoLibrary: ObservableObject {
    @Published private(set) var authorization: PHAuthorizationStatus = PHPhotoLibrary.authorizationStatus(for: .readWrite)
    @Published private(set) var albums: [PhotoAlbum] = []
    @Published private(set) var photos: [PhotoAsset] = []
    @Published private(set) var selectedAlbumID: String?
    @Published private(set) var isLoading = false
    @Published private(set) var loadError: String?
    @Published private(set) var enumerationDuration: TimeInterval = 0
    @Published private(set) var gpsCount = 0

    private let imageManager = PHCachingImageManager()
    private let geocoder = CLGeocoder()

    var gpsPercent: Int { photos.isEmpty ? 0 : Int((Double(gpsCount) / Double(photos.count) * 100).rounded()) }

    func authorize() async {
        authorization = await PHPhotoLibrary.requestAuthorization(for: .readWrite)
        if authorization == .authorized || authorization == .limited {
            loadAlbums()
        }
    }

    func loadAlbums() {
        let collections = PHAssetCollection.fetchAssetCollections(with: .album, subtype: .any, options: nil)
        var found: [PhotoAlbum] = []
        collections.enumerateObjects { collection, _, _ in
            guard let title = collection.localizedTitle else { return }
            let options = PHFetchOptions()
            options.predicate = NSPredicate(format: "mediaType == %d AND isHidden == NO", PHAssetMediaType.image.rawValue)
            found.append(PhotoAlbum(id: collection.localIdentifier, title: title, count: PHAsset.fetchAssets(in: collection, options: options).count))
        }
        albums = found.sorted { $0.title.localizedCaseInsensitiveCompare($1.title) == .orderedAscending }
        if let homie = albums.first(where: { $0.title == "Homie" }) {
            // Keep the picker selection in sync with the default album.
            loadAlbum(id: homie.id)
        } else if let first = albums.first {
            loadAlbum(id: first.id)
        }
    }

    func loadAlbum(id: String) {
        guard let collection = PHAssetCollection.fetchAssetCollections(withLocalIdentifiers: [id], options: nil).firstObject else {
            loadError = "The selected album is no longer available."
            return
        }
        isLoading = true
        selectedAlbumID = id
        loadError = nil
        let started = Date()
        let options = PHFetchOptions()
        options.predicate = NSPredicate(format: "mediaType == %d AND isHidden == NO", PHAssetMediaType.image.rawValue)
        options.sortDescriptors = [NSSortDescriptor(key: "creationDate", ascending: false)]
        let result = PHAsset.fetchAssets(in: collection, options: options)
        var loaded: [PhotoAsset] = []
        result.enumerateObjects { asset, _, _ in loaded.append(PhotoAsset(asset)) }
        photos = loaded
        gpsCount = loaded.filter { $0.coordinate != nil }.count
        enumerationDuration = Date().timeIntervalSince(started)
        isLoading = false
    }

    func requestImage(for photo: PhotoAsset, size: CGSize, completion: @escaping (NSImage?, Bool) -> Void) {
        let options = PHImageRequestOptions()
        options.isSynchronous = false
        options.deliveryMode = .opportunistic
        options.resizeMode = .exact
        options.isNetworkAccessAllowed = true
        imageManager.requestImage(for: photo.asset, targetSize: size, contentMode: .aspectFit, options: options) { image, info in
            let degraded = (info?[PHImageResultIsDegradedKey] as? Bool) ?? false
            DispatchQueue.main.async { completion(image, degraded) }
        }
    }

    func jpegPreview(for photo: PhotoAsset) async throws -> (Data, NSImage) {
        let image = try await requestImage(for: photo, targetSize: CGSize(width: 1400, height: 1400))
        let jpeg = try encodeJPEG(image)
        return (jpeg.data, image)
    }

    func fullSizeJPEG(for photo: PhotoAsset) async throws -> PhotoJPEG {
        let image = try await requestImage(for: photo, targetSize: PHImageManagerMaximumSize)
        var jpeg = try encodeJPEG(image)
        if jpeg.data.count > 18 * 1024 * 1024 {
            jpeg = try encodeJPEG(image, quality: 0.65)
        }
        guard jpeg.data.count <= 20 * 1024 * 1024 else { throw PhotoLibraryError.jpegTooLarge }
        return jpeg
    }

    private func requestImage(for photo: PhotoAsset, targetSize: CGSize) async throws -> NSImage {
        try await withCheckedThrowingContinuation { continuation in
            let options = PHImageRequestOptions()
            options.isSynchronous = false
            options.deliveryMode = .highQualityFormat
            options.resizeMode = .none
            options.isNetworkAccessAllowed = true
            imageManager.requestImage(for: photo.asset, targetSize: targetSize, contentMode: .aspectFit, options: options) { image, info in
                if (info?[PHImageCancelledKey] as? Bool) == true {
                    continuation.resume(throwing: PhotoLibraryError.imageUnavailable)
                } else if let error = info?[PHImageErrorKey] as? Error {
                    continuation.resume(throwing: error)
                } else if let image, (info?[PHImageResultIsDegradedKey] as? Bool) != true {
                    continuation.resume(returning: image)
                } else if (info?[PHImageResultIsInCloudKey] as? Bool) == true {
                    continuation.resume(throwing: PhotoLibraryError.imageUnavailable)
                }
            }
        }
    }

    private func encodeJPEG(_ image: NSImage, quality: Double = 0.85) throws -> PhotoJPEG {
        guard let tiff = image.tiffRepresentation,
              let bitmap = NSBitmapImageRep(data: tiff),
              let jpeg = bitmap.representation(using: .jpeg, properties: [.compressionFactor: quality]) else {
            throw PhotoLibraryError.jpegEncodingFailed
        }
        guard let cgImage = bitmap.cgImage else { throw PhotoLibraryError.jpegEncodingFailed }
        return PhotoJPEG(data: jpeg, width: bitmap.pixelsWide, height: bitmap.pixelsHigh, cgImage: cgImage)
    }

    func place(for photo: PhotoAsset) async throws -> String? {
        guard let coordinate = photo.coordinate else { return nil }
        let placemarks = try await geocoder.reverseGeocodeLocation(CLLocation(latitude: coordinate.latitude, longitude: coordinate.longitude))
        guard let place = placemarks.first else { return nil }
        return [place.locality, place.administrativeArea, place.country]
            .compactMap { $0 }
            .reduce(into: [String]()) { result, item in if result.last != item { result.append(item) } }
            .joined(separator: ", ")
    }
}

enum PhotoLibraryError: LocalizedError {
    case imageUnavailable
    case jpegEncodingFailed
    case jpegTooLarge

    var errorDescription: String? {
        switch self {
        case .imageUnavailable: "PhotoKit did not return an image."
        case .jpegEncodingFailed: "Could not encode the preview as JPEG."
        case .jpegTooLarge: "The JPEG is larger than the server's 20 MB upload limit."
        }
    }
}
