import SwiftUI

struct AlbumPickerView: View {
    static let estimatedBytesPerPhoto = 2.4 * 1_000_000

    let library: PhotoLibrary
    let selection: AlbumSelection
    let hostName: String
    let username: String
    let onSignOut: () -> Void

    private var selectedAlbums: [PhotoAlbum] { library.albums.filter { selection.contains($0.id) } }
    private var photoCount: Int { selectedAlbums.reduce(0) { $0 + $1.count } }
    private var iCloudAlbums: [PhotoAlbum] { library.albums.filter { !$0.isShared } }
    private var sharedAlbums: [PhotoAlbum] { library.albums.filter(\.isShared) }

    private var sizeText: String {
        ByteCountFormatter.string(fromByteCount: Int64(Double(photoCount) * Self.estimatedBytesPerPhoto), countStyle: .file)
    }

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: 24) {
                    PaneHeader(pane: .albums, subtitle: "Photos from these albums show in the slideshow on \(hostName).")
                    AlbumGroup(title: "iCloud Photos", albums: iCloudAlbums, library: library, selection: selection)
                    AlbumGroup(title: "Shared albums", albums: sharedAlbums, library: library, selection: selection)
                }
                .padding(.horizontal, 32)
                .padding(.vertical, 28)
            }

            footer
        }
    }

    private var footer: some View {
        VStack(spacing: 10) {
            HStack {
                if selectedAlbums.isEmpty {
                    Text("Turn on an album to start sending photos to Homie.")
                        .foregroundStyle(.secondary)
                } else {
                    (Text("\(selectedAlbums.count) \(selectedAlbums.count == 1 ? "album" : "albums")").fontWeight(.semibold)
                        + Text(" · \(photoCount.formatted()) photos · about \(sizeText)")
                        + Text(" · changes go to Homie on the next sync").foregroundStyle(.secondary))
                }
                Spacer()
            }
            .font(.system(size: 13))

            HStack(spacing: 12) {
                Toggle("Keep syncing", isOn: Binding(get: { selection.keepSyncing }, set: { selection.setKeepSyncing($0) }))
                    .toggleStyle(.switch)
                    .controlSize(.small)
                    .tint(.homieGreen)
                Spacer()
                Text("Signed in as \(username)")
                    .foregroundStyle(.secondary)
                Button("Sign out", action: onSignOut)
                    .buttonStyle(.plain)
                    .fontWeight(.semibold)
                    .foregroundStyle(Color.homieGreenDark)
            }
            .font(.system(size: 12))
        }
        .padding(.horizontal, 32)
        .padding(.vertical, 14)
        .background(.thinMaterial)
        .overlay(alignment: .top) { Color.homieDivider.frame(height: 1) }
    }
}

private struct AlbumGroup: View {
    let title: String
    let albums: [PhotoAlbum]
    let library: PhotoLibrary
    let selection: AlbumSelection

    var body: some View {
        if !albums.isEmpty {
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Text(title)
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Color.homieSectionHeading)
                    Spacer()
                    Text(albums.count.formatted())
                        .font(.system(size: 12))
                        .foregroundStyle(Color.homieSectionHeading)
                }
                .accessibilityElement(children: .combine)
                .accessibilityAddTraits(.isHeader)

                VStack(spacing: 0) {
                    ForEach(albums) { album in
                        AlbumRow(album: album, library: library, selection: selection)
                        if album.id != albums.last?.id {
                            Color.homieDivider.frame(height: 1).padding(.leading, 66)
                        }
                    }
                }
                .background(Color.white.opacity(0.05), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }
        }
    }
}

private struct AlbumRow: View {
    let album: PhotoAlbum
    let library: PhotoLibrary
    let selection: AlbumSelection
    @State private var isHovering = false

    private var isSelected: Bool { selection.contains(album.id) }
    private var isEmpty: Bool { album.count == 0 }

    private var binding: Binding<Bool> {
        Binding(get: { isSelected }, set: { _ in selection.toggle(album.id) })
    }

    var body: some View {
        HStack(spacing: 12) {
            AlbumCover(album: album, library: library, size: 40)
            VStack(alignment: .leading, spacing: 2) {
                Text(album.title)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(Color.homieText)
                    .lineLimit(1)
                HStack(spacing: 6) {
                    Text("\(album.count.formatted()) \(album.count == 1 ? "photo" : "photos")")
                    if isSelected {
                        Text("· on Homie").foregroundStyle(Color.homieGreenLight)
                    }
                }
                .font(.system(size: 12))
                .foregroundStyle(Color.homieSectionHeading)
            }
            Spacer()
            Toggle("", isOn: binding)
                .toggleStyle(.switch)
                .labelsHidden()
                .tint(.homieGreen)
                .allowsHitTesting(false)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(isHovering && !isEmpty ? Color.homieRowHover : Color.clear)
        .contentShape(Rectangle())
        .onTapGesture { if !isEmpty { selection.toggle(album.id) } }
        .onHover { isHovering = $0 }
        .opacity(isEmpty ? 0.45 : 1)
        .disabled(isEmpty)
        .accessibilityRepresentation {
            Toggle(isOn: binding) {
                Text("\(album.title), \(album.count.formatted()) \(album.count == 1 ? "photo" : "photos")")
            }
            .disabled(isEmpty)
        }
    }
}
