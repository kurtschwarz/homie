import Foundation

struct AlbumList {
    enum Filter: CaseIterable {
        case all, onHomie, notOnHomie
    }

    enum Sort: String, CaseIterable {
        case recent, name, mostPhotos
    }

    struct Counts: Equatable {
        let all: Int
        let onHomie: Int
        let notOnHomie: Int
    }

    struct Group {
        let title: String
        let albums: [PhotoAlbum]
    }

    let libraryTotal: Int
    let counts: Counts
    let groups: [Group]

    var isEmpty: Bool { groups.isEmpty }

    init(
        albums: [PhotoAlbum],
        selectedIDs: Set<String>,
        search: String = "",
        filter: Filter = .all,
        sort: Sort = .recent
    ) {
        libraryTotal = albums.count

        let query = search.trimmingCharacters(in: .whitespaces)
        let searched = query.isEmpty ? albums : albums.filter { $0.title.localizedCaseInsensitiveContains(query) }
        let onHomie = searched.filter { selectedIDs.contains($0.id) }
        counts = Counts(all: searched.count, onHomie: onHomie.count, notOnHomie: searched.count - onHomie.count)

        let visible = switch filter {
        case .all: searched
        case .onHomie: onHomie
        case .notOnHomie: searched.filter { !selectedIDs.contains($0.id) }
        }

        groups = [
            Group(title: "iCloud Photos", albums: visible.filter { !$0.isShared }.sorted(by: sort.areInIncreasingOrder)),
            Group(title: "Shared albums", albums: visible.filter(\.isShared).sorted(by: sort.areInIncreasingOrder))
        ].filter { !$0.albums.isEmpty }
    }
}

extension AlbumList.Sort {
    var title: String {
        switch self {
        case .recent: "Recent"
        case .name: "Name"
        case .mostPhotos: "Most photos"
        }
    }

    fileprivate func areInIncreasingOrder(_ a: PhotoAlbum, _ b: PhotoAlbum) -> Bool {
        func byName() -> Bool { a.title.localizedCaseInsensitiveCompare(b.title) == .orderedAscending }
        switch self {
        case .name:
            return byName()
        case .mostPhotos:
            return a.count != b.count ? a.count > b.count : byName()
        case .recent:
            switch (a.recency, b.recency) {
            case let (x?, y?) where x != y: return x > y
            case (_?, nil): return true
            case (nil, _?): return false
            default: return byName()
            }
        }
    }
}
