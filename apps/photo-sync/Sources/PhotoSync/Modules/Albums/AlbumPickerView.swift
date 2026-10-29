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

    private var artworkCandidates: [PhotoAlbum] { library.albums.filter(\.isArtworkCandidate) }

    private var artworkAlbum: PhotoAlbum? {
        selection.artworkAlbumID.flatMap { id in library.albums.first { $0.id == id } }
    }

    private var artworkAlbumUnavailable: Bool {
        guard let id = selection.artworkAlbumID else { return false }
        return !(library.albums.first { $0.id == id }?.isArtworkCandidate ?? false)
    }

    @ViewBuilder
    private var artworkAlbumSection: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 14) {
                RoundedRectangle(cornerRadius: 10)
                    .fill(Color.homieGreen.opacity(0.18))
                    .frame(width: 34, height: 34)
                    .overlay {
                        Image(systemName: "paintbrush.pointed.fill")
                            .font(.system(size: 14))
                            .foregroundStyle(Color.homieGreenDark)
                    }
                VStack(alignment: .leading, spacing: 2) {
                    Text("Save paintings to").font(.system(size: 14, weight: .semibold))
                    Text("Pictures painted on Homie are added to this album.")
                        .font(.system(size: 12))
                        .foregroundStyle(.secondary)
                }
                Spacer()
                if !artworkCandidates.isEmpty {
                    Menu {
                        Button("None") { selection.setArtworkAlbum(nil) }
                        Divider()
                        ForEach(artworkCandidates) { album in
                            Button(album.title) { selection.setArtworkAlbum(album.id) }
                        }
                    } label: {
                        HStack {
                            Text(artworkAlbum?.title ?? "Choose album")
                                .font(.system(size: 14))
                                .lineLimit(1)
                            Spacer(minLength: 8)
                            Image(systemName: "chevron.down")
                                .font(.system(size: 10, weight: .semibold))
                                .foregroundStyle(.secondary)
                        }
                        .padding(.horizontal, 12)
                        .frame(width: 220, height: 34)
                        .background(RoundedRectangle(cornerRadius: 8).fill(Color(nsColor: .controlBackgroundColor)))
                        .overlay(RoundedRectangle(cornerRadius: 8).stroke(Color.secondary.opacity(0.3), lineWidth: 1))
                        .contentShape(Rectangle())
                    }
                    .menuStyle(.borderlessButton)
                    .menuIndicator(.hidden)
                    .fixedSize()
                }
            }
            if artworkCandidates.isEmpty {
                Text("No albums can be used for artwork yet. Create an album in Photos, then reopen Photo Sync.")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
            if artworkAlbumUnavailable {
                Text("The selected artwork album is no longer available or can't be written to. Choose another album.")
                    .font(.system(size: 12))
                    .foregroundStyle(.red)
            }
        }
    }

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    Text("Pick albums for Homie")
                        .font(.system(size: 24, weight: .bold))
                    Text("Photos from these albums show in the slideshow on \(hostName).")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                        .padding(.top, 4)

                    HStack {
                        Text("ICLOUD ALBUMS")
                            .font(.system(size: 12, weight: .semibold))
                            .tracking(0.6)
                            .foregroundStyle(.secondary)
                        Spacer()
                        Button(allSelected ? "Select none" : "Select all") {
                            selection.select(allSelected ? [] : library.albums.map(\.id))
                        }
                        .buttonStyle(.plain)
                        .font(.system(size: 13, weight: .semibold))
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
            VStack(spacing: 22) {
                artworkAlbumSection
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Keep syncing new photos").font(.system(size: 14, weight: .semibold))
                        Text("New photos in these albums go to Homie automatically.")
                            .font(.system(size: 12))
                            .foregroundStyle(.secondary)
                    }
                    Spacer()
                    Toggle("", isOn: Binding(get: { selection.keepSyncing }, set: { selection.setKeepSyncing($0) }))
                        .toggleStyle(.switch)
                        .labelsHidden()
                        .tint(.homieGreen)
                }
                HStack {
                    (Text("\(selectedAlbums.count) \(selectedAlbums.count == 1 ? "album" : "albums")").fontWeight(.semibold)
                        + Text(" · \(photoCount.formatted()) photos · about \(sizeText)"))
                        .font(.system(size: 13))
                    Spacer()
                    Button("Start sync", action: onStart)
                        .buttonStyle(HomieButtonStyle(prominent: true))
                        .disabled(selectedAlbums.isEmpty)
                }
                HStack {
                    Text("Signed in as \(username)")
                        .font(.system(size: 12))
                        .foregroundStyle(.secondary)
                    Spacer()
                    Button("Sign out", action: onSignOut)
                        .buttonStyle(.plain)
                        .font(.system(size: 12, weight: .semibold))
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
                Text(album.title).font(.system(size: 14, weight: .semibold))
                Text("\(album.count.formatted()) photos").font(.system(size: 12)).foregroundStyle(.secondary)
            }
            Spacer()
            if album.isShared {
                Text("Shared")
                    .font(.system(size: 12, weight: .medium))
                    .padding(.horizontal, 12)
                    .padding(.vertical, 5)
                    .background(Color.secondary.opacity(0.12), in: Capsule())
            }
        }
        .padding(.vertical, 8)
    }
}
