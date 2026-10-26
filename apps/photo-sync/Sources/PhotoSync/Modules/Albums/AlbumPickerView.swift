import SwiftUI

struct AlbumPickerView: View {
    static let estimatedBytesPerPhoto = 2.4 * 1_000_000

    @ObservedObject var library: PhotoLibrary
    @ObservedObject var selection: AlbumSelection
    let hostName: String
    let username: String
    let onStart: () -> Void
    let onSignOut: () -> Void

    private var selectedAlbums: [PhotoAlbum] { library.albums.filter { selection.contains($0.id) } }
    private var photoCount: Int { selectedAlbums.reduce(0) { $0 + $1.count } }
    private var allSelected: Bool { !library.albums.isEmpty && selectedAlbums.count == library.albums.count }

    private var sizeText: String {
        ByteCountFormatter.string(fromByteCount: Int64(Double(photoCount) * Self.estimatedBytesPerPhoto), countStyle: .file)
    }

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    Text("Pick albums for Homie")
                        .font(.system(size: 28, weight: .bold))
                    Text("Photos from these albums show in the slideshow on \(hostName).")
                        .font(.system(size: 15))
                        .foregroundStyle(.secondary)
                        .padding(.top, 4)

                    HStack {
                        Text("ICLOUD PHOTOS")
                            .font(.system(size: 13, weight: .semibold))
                            .tracking(0.6)
                            .foregroundStyle(.secondary)
                        Spacer()
                        Button(allSelected ? "Select none" : "Select all") {
                            selection.select(allSelected ? [] : library.albums.map(\.id))
                        }
                        .buttonStyle(.plain)
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.homieGreenDark)
                    }
                    .padding(.top, 24)
                    .padding(.bottom, 10)

                    ForEach(library.albums) { album in
                        AlbumRow(album: album, library: library, isSelected: selection.contains(album.id))
                            .contentShape(Rectangle())
                            .onTapGesture { selection.toggle(album.id) }
                    }
                }
                .padding(.horizontal, 32)
                .padding(.vertical, 28)
            }

            Divider()
            VStack(spacing: 14) {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Keep syncing new photos").font(.system(size: 16, weight: .semibold))
                        Text("New photos in these albums go to Homie automatically.")
                            .font(.system(size: 13))
                            .foregroundStyle(.secondary)
                    }
                    Spacer()
                    Toggle("", isOn: Binding(get: { selection.keepSyncing }, set: selection.setKeepSyncing))
                        .toggleStyle(.switch)
                        .labelsHidden()
                        .tint(.homieGreen)
                }
                HStack {
                    (Text("\(selectedAlbums.count) \(selectedAlbums.count == 1 ? "album" : "albums")").fontWeight(.semibold)
                        + Text(" · \(photoCount.formatted()) photos · about \(sizeText)"))
                        .font(.system(size: 15))
                    Spacer()
                    Button("Start sync", action: onStart)
                        .buttonStyle(HomieButtonStyle(prominent: true))
                        .disabled(selectedAlbums.isEmpty)
                }
                HStack {
                    Text("Signed in as \(username)")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                    Spacer()
                    Button("Sign out", action: onSignOut)
                        .buttonStyle(.plain)
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Color.homieGreenDark)
                }
            }
            .padding(.horizontal, 32)
            .padding(.vertical, 18)
            .background(Color(nsColor: .windowBackgroundColor))
        }
    }
}

private struct AlbumRow: View {
    let album: PhotoAlbum
    @ObservedObject var library: PhotoLibrary
    let isSelected: Bool

    var body: some View {
        HStack(spacing: 14) {
            RoundedRectangle(cornerRadius: 4)
                .fill(isSelected ? Color.homieGreenDark : Color.clear)
                .overlay(RoundedRectangle(cornerRadius: 4).stroke(isSelected ? Color.clear : Color.secondary.opacity(0.4), lineWidth: 1.5))
                .overlay { if isSelected { Image(systemName: "checkmark").font(.system(size: 9, weight: .bold)).foregroundStyle(.white) } }
                .frame(width: 16, height: 16)
            AlbumCover(album: album, library: library)
            VStack(alignment: .leading, spacing: 2) {
                Text(album.title).font(.system(size: 17, weight: .semibold))
                Text("\(album.count.formatted()) photos").font(.system(size: 14)).foregroundStyle(.secondary)
            }
            Spacer()
            if album.isShared {
                Text("Shared")
                    .font(.system(size: 14, weight: .medium))
                    .padding(.horizontal, 12)
                    .padding(.vertical, 5)
                    .background(Color.secondary.opacity(0.12), in: Capsule())
            }
        }
        .padding(.vertical, 8)
    }
}
