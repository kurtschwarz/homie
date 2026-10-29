import Photos

struct PhotoAlbum: Identifiable, Hashable {
    let id: String
    let title: String
    let count: Int
    let isShared: Bool
    let canAddContent: Bool

    var isArtworkCandidate: Bool { !isShared && canAddContent }
    let recency: Date?
    let cover: PHAsset?
}
