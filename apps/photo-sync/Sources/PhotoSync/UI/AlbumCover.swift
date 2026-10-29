import AppKit
import SwiftUI

struct AlbumCover: View {
    let album: PhotoAlbum
    let library: PhotoLibrary
    var size: CGFloat = 70
    @State private var image: NSImage?

    private var placeholder: LinearGradient {
        let seed = album.id.unicodeScalars.reduce(0) { ($0 &* 31 &+ Int($1.value)) % 360 }
        let hue = Double(seed) / 360
        return LinearGradient(
            colors: [Color(hue: hue, saturation: 0.45, brightness: 0.85), Color(hue: (hue + 0.08).truncatingRemainder(dividingBy: 1), saturation: 0.6, brightness: 0.55)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    var body: some View {
        Group {
            if let image {
                Image(nsImage: image).resizable().scaledToFill()
            } else {
                Rectangle().fill(placeholder)
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
