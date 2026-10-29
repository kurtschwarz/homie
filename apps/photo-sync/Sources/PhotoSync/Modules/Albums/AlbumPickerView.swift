import SwiftUI

struct AlbumPickerView: View {
    static let estimatedBytesPerPhoto = 2.4 * 1_000_000

    let library: PhotoLibrary
    let selection: AlbumSelection
    let hostName: String
    let username: String
    let onSignOut: () -> Void

    @State private var search = ""
    @State private var filter: AlbumList.Filter = .all
    @FocusState private var searchFocused: Bool

    private var selectedAlbums: [PhotoAlbum] { library.albums.filter { selection.contains($0.id) } }
    private var photoCount: Int { selectedAlbums.reduce(0) { $0 + $1.count } }

    private var list: AlbumList {
        AlbumList(albums: library.albums, selectedIDs: selection.albumIDs, search: search, filter: filter, sort: selection.sort)
    }

    private var sizeText: String {
        ByteCountFormatter.string(fromByteCount: Int64(Double(photoCount) * Self.estimatedBytesPerPhoto), countStyle: .file)
    }

    var body: some View {
        let list = list
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                if list.isEmpty {
                    emptyState
                }
                ForEach(list.groups, id: \.title) { group in
                    AlbumGroup(title: group.title, albums: group.albums, library: library, selection: selection)
                }
            }
            .padding(.horizontal, 32)
            .padding(.vertical, 16)
            .animation(.default, value: list.groups.flatMap(\.albums).map(\.id))
        }
        .softScrollEdges()
        .safeAreaInset(edge: .top, spacing: 0) { header(list) }
        .safeAreaInset(edge: .bottom, spacing: 0) { footer }
        .background {
            Button("") { searchFocused = true }
                .keyboardShortcut("f", modifiers: .command)
                .opacity(0)
        }
    }

    private func header(_ list: AlbumList) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            PaneHeader(pane: .albums, subtitle: "Photos from these albums show in the slideshow on \(hostName).")
            HStack(spacing: 10) {
                searchField(libraryTotal: list.libraryTotal)
                sortMenu
            }
            HStack(spacing: 8) {
                chip("All", count: list.counts.all, value: .all)
                chip("On Homie", count: list.counts.onHomie, value: .onHomie)
                chip("Not on Homie", count: list.counts.notOnHomie, value: .notOnHomie)
                Spacer()
            }
        }
        .padding(.horizontal, 32)
        .padding(.top, 34)
        .padding(.bottom, 14)
        .fadingMaterial(solidEdge: .top)
        .zIndex(1)
    }

    private func searchField(libraryTotal: Int) -> some View {
        HStack(spacing: 6) {
            Image(systemName: "magnifyingglass").foregroundStyle(.secondary)
            TextField("Search \(libraryTotal.formatted()) albums", text: $search)
                .textFieldStyle(.plain)
                .focused($searchFocused)
                .onKeyPress(.escape) {
                    if !search.isEmpty {
                        search = ""
                    } else {
                        searchFocused = false
                    }
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
    }

    private var sortMenu: some View {
        Picker("Sort", selection: Binding(get: { selection.sort }, set: { selection.setSort($0) })) {
            ForEach(AlbumList.Sort.allCases, id: \.self) { Text($0.title).tag($0) }
        }
        .pickerStyle(.menu)
        .fixedSize()
    }

    private func chip(_ title: String, count: Int, value: AlbumList.Filter) -> some View {
        let active = filter == value
        return Button {
            filter = value
        } label: {
            Text("\(title) \(count.formatted())")
                .font(.system(size: 12, weight: .semibold))
                .padding(.horizontal, 10)
                .padding(.vertical, 4)
                .foregroundStyle(active ? Color.white : Color.homieSectionHeading)
                .background(active ? Color.homieGreen : Color.white.opacity(0.07), in: Capsule())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(active ? .isSelected : [])
    }

    private var emptyState: some View {
        let libraryIsEmpty = library.albums.isEmpty
        return VStack(spacing: 8) {
            Text(libraryIsEmpty ? "No albums in your Photos library." : "No albums match.")
                .foregroundStyle(Color.homieSectionHeading)
            if !libraryIsEmpty {
                Button("Clear search") {
                    search = ""
                    filter = .all
                }
                .buttonStyle(.plain)
                .fontWeight(.semibold)
                .foregroundStyle(Color.homieGreenDark)
            }
        }
        .font(.system(size: 13))
        .frame(maxWidth: .infinity)
        .padding(.top, 40)
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
        .fadingMaterial(solidEdge: .bottom)
        .overlay(alignment: .top) { Color.homieDivider.frame(height: 1) }
    }
}

private extension View {
    // Solid behind the controls, fading over a strip just outside the bar so rows dissolve into it
    // rather than meeting a hard edge.
    func fadingMaterial(solidEdge: VerticalEdge) -> some View {
        let isTop = solidEdge == .top
        let fade: CGFloat = 28
        return background(alignment: isTop ? .bottom : .top) {
            Rectangle()
                .fill(.ultraThinMaterial)
                .frame(height: fade)
                .mask(
                    LinearGradient(
                        colors: [.black, .clear],
                        startPoint: isTop ? .top : .bottom,
                        endPoint: isTop ? .bottom : .top
                    )
                )
                .offset(y: isTop ? fade : -fade)
        }
        .background(.ultraThinMaterial)
    }

    @ViewBuilder
    func softScrollEdges() -> some View {
        #if compiler(>=6.2)
        if #available(macOS 26, *) { scrollEdgeEffectStyle(.soft, for: .vertical) } else { self }
        #else
        self
        #endif
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
