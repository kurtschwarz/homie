import Testing

@testable import PhotoSync

struct SyncRunLogTests {
    @Test func runIDsAreShortAndDistinct() {
        let ids = (0..<50).map { _ in SyncRunLog().id }
        #expect(Set(ids).count == ids.count)
        #expect(ids.allSatisfy { $0.count == 8 })
    }

    @Test func subsystemIsStable() {
        #expect(Log.subsystem == "ar.schw.homie.photo-sync")
    }
}
