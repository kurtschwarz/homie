import Foundation
import Testing
@testable import PhotoSync

@MainActor
struct DrawingPusherTests {
    private final class FakeService: DrawingService {
        var drawings: [HomieDrawing] = []
        var listError: Error?
        var downloadError: Error?
        var marked: [String] = []

        func unpushedDrawings() async throws -> [HomieDrawing] {
            if let listError { throw listError }
            return drawings
        }
        func pushedDrawingAssetIDs() async throws -> Set<String> { [] }
        func downloadDrawing(_ drawing: HomieDrawing) async throws -> Data {
            if let downloadError { throw downloadError }
            return Data()
        }
        func markDrawingPushed(id: String, assetID: String) async throws { marked.append(id) }
    }

    private struct FakeAlbum: DrawingAlbum {
        var canAdd = true
        func canAddDrawings(toAlbum id: String) -> Bool { canAdd }
        func addDrawing(_ jpeg: Data, createdAt: Date, toAlbum albumID: String) async throws -> String { "asset" }
    }

    private let defaults = TestDefaults.make()
    private let service = FakeService()

    private func drawing(_ id: String) -> HomieDrawing {
        HomieDrawing(id: id, url: "/d/\(id)", takenAt: 0, coordinate: nil)
    }

    private func pusher(canAdd: Bool = true) -> DrawingPusher {
        DrawingPusher(service: service, album: FakeAlbum(canAdd: canAdd), defaults: defaults)
    }

    @Test func explainsWhenNoArtworkAlbumIsChosen() async throws {
        let pusher = pusher(canAdd: false)

        try await pusher.push(toAlbum: nil, run: SyncRunLog())

        #expect(pusher.message == "Choose an artwork album to save Homie drawings to Photos.")
    }

    @Test func skipsSilentlyWhenSavingIsSwitchedOff() async throws {
        service.drawings = [drawing("a")]
        let pusher = pusher()

        try await pusher.push(toAlbum: "album", enabled: false, run: SyncRunLog())

        #expect(pusher.message == nil)
        #expect(service.marked.isEmpty)
    }

    @Test func warnsWhenTheChosenAlbumIsUnusable() async throws {
        service.drawings = [drawing("a")]
        let pusher = pusher(canAdd: false)

        try await pusher.push(toAlbum: "album", run: SyncRunLog())

        #expect(pusher.message?.contains("missing or can't be added to") == true)
        #expect(service.marked.isEmpty)
    }

    @Test func pushesEachDrawingAndMarksItPushed() async throws {
        service.drawings = [drawing("a"), drawing("b")]
        let pusher = pusher()

        try await pusher.push(toAlbum: "album", run: SyncRunLog())

        #expect(service.marked == ["a", "b"])
        #expect(pusher.progress == AlbumProgress(completed: 2, total: 2, failed: 0))
    }

    @Test func countsFailuresAndKeepsGoing() async throws {
        service.drawings = [drawing("a")]
        service.downloadError = HomieAPIError.http(500)
        let pusher = pusher()

        try await pusher.push(toAlbum: "album", run: SyncRunLog())

        #expect(pusher.progress?.failed == 1)
    }

    @Test func anExpiredSessionEndsThePush() async {
        service.drawings = [drawing("a"), drawing("b")]
        service.downloadError = HomieAPIError.unauthorized
        let pusher = pusher()

        await #expect(throws: SessionExpired.self) {
            try await pusher.push(toAlbum: "album", run: SyncRunLog())
        }
        #expect(pusher.progress?.completed == 1)
    }

    @Test func anExpiredSessionWhileListingEndsThePush() async {
        service.listError = HomieAPIError.unauthorized
        let pusher = pusher()

        await #expect(throws: SessionExpired.self) {
            try await pusher.push(toAlbum: "album", run: SyncRunLog())
        }
    }
}
