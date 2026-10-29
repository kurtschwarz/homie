import AppKit
import Observation
import Photos
import Vision

struct PhotoJPEG {
    let data: Data
    let width: Int
    let height: Int
    let cgImage: CGImage
}

@Observable
final class PhotoLibrary {
    private(set) var authorization: PHAuthorizationStatus = PHPhotoLibrary.authorizationStatus(for: .readWrite)
    private(set) var albums: [PhotoAlbum] = []

    private let imageManager = PHCachingImageManager()
    private var changeObserver: LibraryChangeObserver?

    func refreshAuthorization() {
        authorization = PHPhotoLibrary.authorizationStatus(for: .readWrite)
    }

    func authorize() async {
        // Off the main actor: PhotoKit keeps it busy for a second or two, which stalls any UI work
        // waiting to run while the prompt is up.
        authorization = await Self.requestAuthorization()
    }

    @concurrent
    private static func requestAuthorization() async -> PHAuthorizationStatus {
        await PHPhotoLibrary.requestAuthorization(for: .readWrite)
    }

    func loadAlbums() {
        if changeObserver == nil {
            let observer = LibraryChangeObserver { [weak self] in self?.loadAlbums() }
            PHPhotoLibrary.shared().register(observer)
            changeObserver = observer
        }
        let collections = PHAssetCollection.fetchAssetCollections(with: .album, subtype: .any, options: nil)
        var found: [PhotoAlbum] = []
        collections.enumerateObjects { collection, _, _ in
            guard let title = collection.localizedTitle else { return }
            let options = PHFetchOptions()
            options.predicate = NSPredicate(format: "mediaType == %d AND isHidden == NO", PHAssetMediaType.image.rawValue)
            let count = PHAsset.fetchAssets(in: collection, options: options).count
            options.sortDescriptors = [NSSortDescriptor(key: "creationDate", ascending: false)]
            options.fetchLimit = 1
            let cover = PHAsset.fetchAssets(in: collection, options: options).firstObject
            found.append(PhotoAlbum(
                id: collection.localIdentifier,
                title: title,
                count: count,
                isShared: collection.assetCollectionSubtype == .albumCloudShared,
                canAddContent: collection.canPerform(.addContent),
                recency: cover?.creationDate ?? collection.endDate ?? collection.startDate,
                cover: cover
            ))
        }
        albums = found.sorted {
            if $0.count != $1.count { return $0.count > $1.count }
            return $0.title.localizedCaseInsensitiveCompare($1.title) == .orderedAscending
        }
    }

    func assets(inAlbum id: String) -> [PhotoAsset] {
        guard let collection = PHAssetCollection.fetchAssetCollections(withLocalIdentifiers: [id], options: nil).firstObject else {
            return []
        }
        let options = PHFetchOptions()
        options.predicate = NSPredicate(format: "mediaType == %d AND isHidden == NO", PHAssetMediaType.image.rawValue)
        options.sortDescriptors = [NSSortDescriptor(key: "creationDate", ascending: false)]
        var loaded: [PhotoAsset] = []
        PHAsset.fetchAssets(in: collection, options: options).enumerateObjects { asset, _, _ in
            loaded.append(PhotoAsset(asset))
        }
        return loaded
    }

    func requestThumbnail(for asset: PHAsset, size: CGSize, completion: @escaping (NSImage?) -> Void) {
        let options = PHImageRequestOptions()
        options.isSynchronous = false
        options.deliveryMode = .opportunistic
        options.resizeMode = .fast
        options.isNetworkAccessAllowed = true
        imageManager.requestImage(for: asset, targetSize: size, contentMode: .aspectFill, options: options) { image, _ in
            DispatchQueue.main.async { completion(image) }
        }
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

    func canAddDrawings(toAlbum id: String) -> Bool {
        guard let collection = PHAssetCollection.fetchAssetCollections(withLocalIdentifiers: [id], options: nil).firstObject else {
            return false
        }
        return collection.canPerform(.addContent)
    }

    func addDrawing(_ jpeg: Data, createdAt: Date, toAlbum albumID: String) async throws -> String {
        guard let collection = PHAssetCollection.fetchAssetCollections(withLocalIdentifiers: [albumID], options: nil).firstObject else {
            throw PhotoLibraryError.albumUnavailable
        }
        let file = FileManager.default.temporaryDirectory.appendingPathComponent("\(UUID().uuidString).jpg")
        try jpeg.write(to: file)
        defer { try? FileManager.default.removeItem(at: file) }

        var identifier: String?
        try await PHPhotoLibrary.shared().performChanges {
            let request = PHAssetCreationRequest.forAsset()
            let options = PHAssetResourceCreationOptions()
            options.shouldMoveFile = false
            request.addResource(with: .photo, fileURL: file, options: options)
            request.creationDate = createdAt
            guard let placeholder = request.placeholderForCreatedAsset else { return }
            identifier = placeholder.localIdentifier
            PHAssetCollectionChangeRequest(for: collection)?.addAssets([placeholder] as NSArray)
        }
        guard let identifier else { throw PhotoLibraryError.assetNotCreated }
        return identifier
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
}

private final class LibraryChangeObserver: NSObject, PHPhotoLibraryChangeObserver {
    private let onChange: @MainActor () -> Void

    init(onChange: @escaping @MainActor () -> Void) {
        self.onChange = onChange
    }

    deinit {
        PHPhotoLibrary.shared().unregisterChangeObserver(self)
    }

    nonisolated func photoLibraryDidChange(_ changeInstance: PHChange) {
        Task { @MainActor in onChange() }
    }
}

enum PhotoLibraryError: LocalizedError {
    case imageUnavailable
    case jpegEncodingFailed
    case jpegTooLarge
    case albumUnavailable
    case assetNotCreated

    var errorDescription: String? {
        switch self {
        case .albumUnavailable: "The artwork album no longer exists in Photos."
        case .assetNotCreated: "Photos did not create the drawing."
        case .imageUnavailable: "PhotoKit did not return an image."
        case .jpegEncodingFailed: "Could not encode the preview as JPEG."
        case .jpegTooLarge: "The JPEG is larger than the server's 20 MB upload limit."
        }
    }
}
