import Foundation
import Testing
@testable import PhotoSync

struct AlbumSortPreferenceTests {
    private let defaults = TestDefaults.make()

    @Test func defaultsToRecent() {
        #expect(AlbumSelection(defaults: defaults).sort == .recent)
    }

    @Test func persistsAcrossLaunches() {
        AlbumSelection(defaults: defaults).setSort(.mostPhotos)

        #expect(AlbumSelection(defaults: defaults).sort == .mostPhotos)
    }

    @Test func onboardingResetClearsIt() {
        AlbumSelection(defaults: defaults).setSort(.name)

        AlbumSelection.clear(defaults: defaults)

        #expect(AlbumSelection(defaults: defaults).sort == .recent)
    }
}
