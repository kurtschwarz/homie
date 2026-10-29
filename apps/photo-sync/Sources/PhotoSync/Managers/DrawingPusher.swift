import Foundation
import Observation

protocol DrawingService {
    func unpushedDrawings() async throws -> [HomieDrawing]
    func pushedDrawingAssetIDs() async throws -> Set<String>
    func downloadDrawing(_ drawing: HomieDrawing) async throws -> Data
    func markDrawingPushed(id: String, assetID: String) async throws
}

protocol DrawingAlbum {
    func canAddDrawings(toAlbum id: String) -> Bool
    func addDrawing(_ jpeg: Data, createdAt: Date, toAlbum albumID: String) async throws -> String
}

extension HomieAPI: DrawingService {}
extension PhotoLibrary: DrawingAlbum {}

@Observable
final class DrawingPusher {
    private(set) var progress: AlbumProgress?
    private(set) var message: String?

    private let service: DrawingService
    private let album: DrawingAlbum
    private let defaults: UserDefaults

    init(service: DrawingService, album: DrawingAlbum, defaults: UserDefaults = .standard) {
        self.service = service
        self.album = album
        self.defaults = defaults
    }

    func pushedAssetIDs() async throws -> Set<String> {
        try await service.pushedDrawingAssetIDs()
    }

    func push(toAlbum albumID: String?, enabled: Bool = true, run: SyncRunLog) async throws {
        progress = nil
        message = nil
        var ledger = UploadLedger(defaults: defaults)

        guard enabled else {
            run.info("Skipping drawings: saving paintings is off")
            return
        }
        guard let albumID else {
            message = "Choose an artwork album to save Homie drawings to Photos."
            run.info("Skipping drawings: no artwork album")
            return
        }
        guard album.canAddDrawings(toAlbum: albumID) else {
            message = "The artwork album is missing or can't be added to. Choose another in Artwork."
            run.info("Skipping drawings: artwork album unusable")
            return
        }
        let drawings: [HomieDrawing]
        do {
            drawings = try await service.unpushedDrawings()
            run.info("Drawings to push", detail: "count=\(drawings.count)")
        } catch {
            run.error("Could not list drawings", detail: error.localizedDescription)
            message = error.localizedDescription
            try throwIfSessionExpired(error)
            return
        }
        guard !drawings.isEmpty else { return }
        defer { ledger.flush() }

        var current = AlbumProgress(total: drawings.count)
        progress = current
        for drawing in drawings {
            do {
                let assetID: String
                if let known = ledger.pushedAssetID(forDrawing: drawing.id) {
                    assetID = known
                } else {
                    let jpeg = try await service.downloadDrawing(drawing)
                    assetID = try await album.addDrawing(
                        jpeg,
                        createdAt: Date(timeIntervalSince1970: TimeInterval(drawing.takenAt)),
                        toAlbum: albumID
                    )
                    ledger.savePushedDrawing(drawing.id, assetID: assetID)
                }
                try await service.markDrawingPushed(id: drawing.id, assetID: assetID)
                ledger.forgetPushedDrawing(drawing.id)
                current.completed += 1
            } catch {
                current.completed += 1
                current.failed += 1
                run.error("Drawing failed", detail: "id=\(drawing.id) error=\(error.localizedDescription)")
                message = "Drawing didn't sync: \(error.localizedDescription)"
                if case HomieAPIError.unauthorized = error {
                    progress = current
                    throw SessionExpired()
                }
            }
            progress = current
        }
    }
}
