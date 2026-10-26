import Photos

struct PhotoAlbum: Identifiable, Hashable {
    let id: String
    let title: String
    let count: Int
    let isShared: Bool
    let cover: PHAsset?
}
