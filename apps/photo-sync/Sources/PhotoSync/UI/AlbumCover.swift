import AppKit
import SwiftUI

struct AlbumCover: View {
    let album: PhotoAlbum
    let library: PhotoLibrary
    var size: CGFloat = 70
    @State private var image: NSImage?

    var body: some View {
        Group {
            if let image {
                Image(nsImage: image).resizable().scaledToFill()
            } else {
                Rectangle().fill(.quaternary)
            }
        }
        .frame(width: size, height: size)
        .clipShape(RoundedRectangle(cornerRadius: size * 0.2))
        .task(id: album.id) {
            guard let cover = album.cover else { return }
            library.requestThumbnail(for: cover, size: CGSize(width: size * 3, height: size * 3)) { result in
                if let result { image = result }
            }
        }
    }
}
