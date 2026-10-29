import Foundation

struct ArtworkAlbums {
    enum Status: Equatable {
        case off
        case needsChoice
        case unavailable
        case ready
    }

    let candidates: [PhotoAlbum]
    let current: PhotoAlbum?
    let status: Status

    init(albums: [PhotoAlbum], storedID: String?, enabled: Bool, search: String = "") {
        let writable = albums.filter(\.isArtworkCandidate)
        current = storedID.flatMap { id in writable.first { $0.id == id } }

        let query = search.trimmingCharacters(in: .whitespaces)
        let matching = query.isEmpty ? writable : writable.filter { $0.title.localizedCaseInsensitiveContains(query) }
        candidates = matching.sorted { $0.title.localizedCaseInsensitiveCompare($1.title) == .orderedAscending }

        if !enabled {
            status = .off
        } else if storedID == nil {
            status = .needsChoice
        } else {
            status = current == nil ? .unavailable : .ready
        }
    }
}
