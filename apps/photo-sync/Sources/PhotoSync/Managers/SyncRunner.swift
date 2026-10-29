import Foundation
import Observation

@Observable
final class SyncRunner {
    private(set) var isRunning = false
    let drawings: DrawingPusher
    let photos: PhotoUploader

    private let onAuthenticationRequired: () -> Void

    init(drawings: DrawingPusher, photos: PhotoUploader, onAuthenticationRequired: @escaping () -> Void) {
        self.drawings = drawings
        self.photos = photos
        self.onAuthenticationRequired = onAuthenticationRequired
    }

    func sync(albumIDs: [String], artworkAlbumID: String? = nil, onFinished: @escaping () -> Void = {}) {
        guard !isRunning else { return }
        isRunning = true
        Task {
            await run(albumIDs: albumIDs, artworkAlbumID: artworkAlbumID)
            isRunning = false
            onFinished()
        }
    }

    func run(albumIDs: [String], artworkAlbumID: String?) async {
        let log = SyncRunLog()
        log.info("Sync started", detail: "albums=\(albumIDs.count)")
        do {
            try await drawings.push(toAlbum: artworkAlbumID, run: log)
            let drawingAssetIDs: Set<String>
            do {
                drawingAssetIDs = try await drawings.pushedAssetIDs()
            } catch {
                log.error("Could not load pushed drawings; aborting", detail: error.localizedDescription)
                try throwIfSessionExpired(error)
                return
            }
            try await photos.upload(albumIDs: albumIDs, excluding: drawingAssetIDs, run: log)
        } catch is SessionExpired {
            onAuthenticationRequired()
        } catch {
            // Cancellation: the uploader has already logged it.
        }
        log.info("Sync finished", detail: "completed=\(photos.completed) total=\(photos.total) failed=\(photos.failed)")
    }
}
