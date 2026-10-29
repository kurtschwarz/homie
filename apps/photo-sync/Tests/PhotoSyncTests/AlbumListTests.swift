import Foundation
import Testing
@testable import PhotoSync

struct AlbumListTests {
    private static func album(
        _ id: String,
        count: Int = 10,
        shared: Bool = false,
        recency: Date? = nil
    ) -> PhotoAlbum {
        PhotoAlbum(id: id, title: id, count: count, isShared: shared, canAddContent: true, recency: recency, cover: nil)
    }

    private let library = [
        album("Family"),
        album("Family Trip", shared: true),
        album("Holidays"),
        album("Pets", shared: true)
    ]

    @Test func searchNarrowsByTitleIgnoringCaseAndChipsCountWhatSearchLeft() {
        let list = AlbumList(albums: library, selectedIDs: ["Family"], search: "fam")

        #expect(list.groups.flatMap(\.albums).map(\.id) == ["Family", "Family Trip"])
        #expect(list.counts == AlbumList.Counts(all: 2, onHomie: 1, notOnHomie: 1))
        #expect(list.libraryTotal == 4)
    }

    @Test func groupsWithNoMatchesDisappear() {
        let list = AlbumList(albums: library, selectedIDs: [], search: "pets")

        #expect(list.groups.map(\.title) == ["Shared albums"])
        #expect(list.groups.first?.albums.map(\.id) == ["Pets"])
    }

    @Test func nothingMatchingYieldsAnEmptyListButKeepsTheLibraryTotal() {
        let list = AlbumList(albums: library, selectedIDs: [], search: "zzz")

        #expect(list.isEmpty)
        #expect(list.libraryTotal == 4)
    }

    @Test func sortsAreAppliedWithinEachGroup() {
        let old = Date(timeIntervalSince1970: 1_000)
        let new = Date(timeIntervalSince1970: 2_000)
        let albums = [
            Self.album("B", count: 5, recency: old),
            Self.album("A", count: 1, recency: nil),
            Self.album("C", count: 9, recency: new),
            Self.album("S2", count: 3, shared: true),
            Self.album("S1", count: 7, shared: true)
        ]

        func order(_ sort: AlbumList.Sort) -> [String] {
            AlbumList(albums: albums, selectedIDs: [], sort: sort).groups.flatMap(\.albums).map(\.id)
        }

        #expect(order(.name) == ["A", "B", "C", "S1", "S2"])
        #expect(order(.mostPhotos) == ["C", "B", "A", "S1", "S2"])
        #expect(order(.recent) == ["C", "B", "A", "S1", "S2"])
    }

    @Test func chipsPartitionTheListBySwitchState() {
        let on = AlbumList(albums: library, selectedIDs: ["Family", "Pets"], filter: .onHomie)
        let off = AlbumList(albums: library, selectedIDs: ["Family", "Pets"], filter: .notOnHomie)

        #expect(on.groups.flatMap(\.albums).map(\.id).sorted() == ["Family", "Pets"])
        #expect(off.groups.flatMap(\.albums).map(\.id).sorted() == ["Family Trip", "Holidays"])
    }
}
