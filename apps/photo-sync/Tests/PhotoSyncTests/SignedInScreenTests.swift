import Testing
@testable import PhotoSync

struct SignedInScreenTests {
    @Test func statusWithoutSelectionShowsEmptyState() {
        #expect(SignedInScreen.resolve(pane: .status, hasSelection: false) == .emptyStatus)
    }

    @Test func statusWithSelectionShowsSyncStatus() {
        #expect(SignedInScreen.resolve(pane: .status, hasSelection: true) == .status)
    }

    @Test(arguments: [false, true])
    func albumsAlwaysShowsPicker(hasSelection: Bool) {
        #expect(SignedInScreen.resolve(pane: .albums, hasSelection: hasSelection) == .picker)
    }

    @Test func otherPanesShowPlaceholder() {
        #expect(SignedInScreen.resolve(pane: .general, hasSelection: true) == .placeholder)
    }
}
