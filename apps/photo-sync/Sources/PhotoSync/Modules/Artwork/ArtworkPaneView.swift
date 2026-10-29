import SwiftUI

struct ArtworkPaneView: View {
    let library: PhotoLibrary
    let selection: AlbumSelection

    @State private var isExpanded = false
    @State private var search = ""
    @FocusState private var searchFocused: Bool

    private var artwork: ArtworkAlbums {
        ArtworkAlbums(
            albums: library.albums,
            storedID: selection.artworkAlbumID,
            enabled: selection.artworkEnabled,
            search: search
        )
    }

    var body: some View {
        let artwork = artwork
        ZStack {
            Color.clear
                .contentShape(Rectangle())
                .onTapGesture { collapse() }
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    PaneHeader(
                        pane: .artwork,
                        subtitle: "Pictures painted on Homie come back to your Mac and into iCloud Photos."
                    )
                    card(artwork)
                }
                .padding(.horizontal, 32)
                .padding(.top, 34)
                .padding(.bottom, 24)
            }
        }
        .onChange(of: selection.artworkEnabled) { _, enabled in
            if !enabled { collapse() }
        }
    }

    private func card(_ artwork: ArtworkAlbums) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Save paintings to")
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Color.homieSectionHeading)

            VStack(spacing: 0) {
                HStack {
                    Text("Save to iCloud Photos")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.homieText)
                    Spacer()
                    Toggle("Save to iCloud Photos", isOn: Binding(get: { selection.artworkEnabled }, set: { selection.setArtworkEnabled($0) }))
                        .toggleStyle(.switch)
                        .labelsHidden()
                        .tint(.homieGreen)
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 10)

                Color.homieDivider.frame(height: 1)

                albumRow(artwork)
                    .opacity(selection.artworkEnabled ? 1 : 0.45)
                    .disabled(!selection.artworkEnabled)

                if isExpanded && selection.artworkEnabled {
                    Color.homieDivider.frame(height: 1)
                    picker(artwork)
                }
            }
            .background(Color.white.opacity(0.05), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
            .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
        }
    }

    private func albumRow(_ artwork: ArtworkAlbums) -> some View {
        Button {
            isExpanded ? collapse() : expand()
        } label: {
            HStack(spacing: 12) {
                if let current = artwork.current {
                    AlbumCover(album: current, library: library, size: 40)
                } else {
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Color.white.opacity(0.08))
                        .frame(width: 40, height: 40)
                        .overlay(Image(systemName: "photo.on.rectangle").foregroundStyle(.secondary))
                }
                VStack(alignment: .leading, spacing: 2) {
                    Text("Album")
                        .font(.system(size: 12))
                        .foregroundStyle(Color.homieSectionHeading)
                    Text(artwork.current?.title ?? "Choose an album")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.homieText)
                        .lineLimit(1)
                    if let warning = warning(for: artwork.status) {
                        Label(warning, systemImage: "exclamationmark.triangle.fill")
                            .font(.system(size: 12))
                            .foregroundStyle(.orange)
                    }
                }
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(.secondary)
                    .rotationEffect(.degrees(isExpanded ? 90 : 0))
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    private func warning(for status: ArtworkAlbums.Status) -> String? {
        switch status {
        case .unavailable: "This album is missing or can't be added to. Choose another."
        case .needsChoice: "Choose an album to save paintings to."
        case .off, .ready: nil
        }
    }

    private func picker(_ artwork: ArtworkAlbums) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 6) {
                Image(systemName: "magnifyingglass").foregroundStyle(.secondary)
                TextField("Search albums", text: $search)
                    .textFieldStyle(.plain)
                    .focused($searchFocused)
                    .onKeyPress(.escape) {
                        if search.isEmpty { collapse() } else { search = "" }
                        return .handled
                    }
                if !search.isEmpty {
                    Button { search = "" } label: { Image(systemName: "xmark.circle.fill") }
                        .buttonStyle(.plain)
                        .foregroundStyle(.secondary)
                        .accessibilityLabel("Clear search")
                }
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 6)
            .background(Color.white.opacity(0.07), in: RoundedRectangle(cornerRadius: 8, style: .continuous))

            ScrollView {
                LazyVStack(spacing: 0) {
                    if artwork.candidates.isEmpty {
                        Text(search.isEmpty ? "No albums you can add to." : "No albums match.")
                            .font(.system(size: 13))
                            .foregroundStyle(Color.homieSectionHeading)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 24)
                    }
                    ForEach(artwork.candidates) { album in
                        CandidateRow(album: album, library: library, isCurrent: album.id == selection.artworkAlbumID) {
                            selection.setArtworkAlbum(album.id)
                            collapse()
                        }
                    }
                }
            }
            .frame(height: 240)

            Text("Shared albums aren't listed because paintings can only go into albums you own. Make a new album in Photos and it will appear here.")
                .font(.system(size: 12))
                .foregroundStyle(Color.homieSectionHeading)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(12)
    }

    private func expand() {
        guard selection.artworkEnabled else { return }
        isExpanded = true
        searchFocused = true
    }

    private func collapse() {
        guard isExpanded else { return }
        isExpanded = false
        search = ""
        searchFocused = false
    }
}

private struct CandidateRow: View {
    let album: PhotoAlbum
    let library: PhotoLibrary
    let isCurrent: Bool
    let onChoose: () -> Void
    @State private var isHovering = false

    var body: some View {
        Button(action: onChoose) {
            HStack(spacing: 12) {
                AlbumCover(album: album, library: library, size: 36)
                VStack(alignment: .leading, spacing: 2) {
                    Text(album.title)
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.homieText)
                        .lineLimit(1)
                    Text("\(album.count.formatted()) \(album.count == 1 ? "photo" : "photos")")
                        .font(.system(size: 12))
                        .foregroundStyle(Color.homieSectionHeading)
                }
                Spacer()
                if isCurrent {
                    Image(systemName: "checkmark").foregroundStyle(Color.homieGreenLight)
                }
            }
            .padding(.horizontal, 8)
            .padding(.vertical, 6)
            .background(isHovering ? Color.homieRowHover : Color.clear, in: RoundedRectangle(cornerRadius: 8))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .onHover { isHovering = $0 }
        .accessibilityAddTraits(isCurrent ? .isSelected : [])
    }
}
