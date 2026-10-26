import AppKit
import Photos
import SwiftUI

struct ContentView: View {
    @ObservedObject private var hostStore: HomieHostStore
    @StateObject private var auth: AuthModel
    @StateObject private var library: PhotoLibrary
    @StateObject private var uploader: PhotoUploader
    @State private var selectedPhoto: PhotoAsset?
    @State private var showsHostPicker = false

    init(hostStore: HomieHostStore) {
        self.hostStore = hostStore
        let endpoint = hostStore.selected.graphQLURL
        let api = HomieAPI(endpoint: endpoint, sessionToken: { HomieSession.token(for: endpoint) })
        let auth = AuthModel(api: api)
        let library = PhotoLibrary()
        _auth = StateObject(wrappedValue: auth)
        _library = StateObject(wrappedValue: library)
        _uploader = StateObject(wrappedValue: PhotoUploader(
            library: library,
            api: api,
            onAuthenticationRequired: { auth.sessionExpired() }
        ))
    }

    var body: some View {
        Group {
            switch auth.state {
            case .checking:
                ProgressView("Checking your Homie session…")
                    .frame(minWidth: 720, minHeight: 620)
            case .signedOut:
                SignInView(auth: auth, hostStore: hostStore)
            case .signedIn:
                photoLibraryContent
            }
        }
        .frame(minWidth: 900, minHeight: 620)
        .task {
            await auth.restoreSession()
        }
    }

    @ViewBuilder
    private var photoLibraryContent: some View {
        switch library.authorization {
        case .authorized, .limited:
            browser
                .onAppear {
                    if library.albums.isEmpty { library.loadAlbums() }
                }
        case .denied, .restricted:
            ContentUnavailableView("Photos Access Needed", systemImage: "photo.badge.exclamationmark", description: Text("Allow Photo Sync to access your Photos library in System Settings → Privacy & Security → Photos."))
        case .notDetermined:
            ContentUnavailableView {
                Label("Connect to Photos", systemImage: "photo.on.rectangle.angled")
            } description: {
                Text("Photo Sync reads your Homie album and displays photo metadata locally.")
            } actions: {
                Button("Allow Photo Access") { Task { await library.authorize() } }
                    .buttonStyle(.borderedProminent)
            }
        @unknown default:
            ContentUnavailableView("Photos Unavailable", systemImage: "photo")
        }
    }

    private var browser: some View {
        VStack(spacing: 0) {
            header
            Divider()
            HSplitView {
                photoGrid
                    .frame(minWidth: 430)
                detail
                    .frame(minWidth: 330, idealWidth: 390)
            }
        }
    }

    private var header: some View {
        HStack(spacing: 16) {
            Label("Photo Sync", systemImage: "photo.on.rectangle.angled")
                .font(.title2.weight(.semibold))
            Picker("Album", selection: Binding(
                get: { library.selectedAlbumID },
                set: { if let newValue = $0 { library.loadAlbum(id: newValue); selectedPhoto = nil } }
            )) {
                ForEach(library.albums) { album in
                    Text("\(album.title) (\(album.count))").tag(Optional(album.id))
                }
            }
            .frame(maxWidth: 300)
            Button { library.loadAlbums() } label: { Image(systemName: "arrow.clockwise") }
                .help("Refresh albums")
            Spacer()
            Button { showsHostPicker = true } label: {
                Label(hostStore.selected.name, systemImage: "display")
            }
            .disabled(uploader.isUploading)
            .sheet(isPresented: $showsHostPicker) {
                HomieHostPicker(store: hostStore, isPresented: $showsHostPicker)
            }
            if library.isLoading { ProgressView().controlSize(.small) }
            if uploader.isUploading {
                ProgressView(value: Double(uploader.completed), total: Double(max(uploader.total, 1)))
                    .frame(width: 90)
                Text("\(uploader.completed)/\(uploader.total)").font(.caption.monospacedDigit())
            }
            Button("Upload All") { uploader.uploadAll(library.photos) }
                .disabled(uploader.isUploading || library.photos.isEmpty)
            Button("Clear Local State", systemImage: "arrow.counterclockwise") {
                uploader.clearLocalState()
            }
            .disabled(uploader.isUploading)
            .help("Clear upload history and cached place lookups")
            if case .signedIn(let user) = auth.state {
                Menu(user.username) {
                    Button("Sign Out", role: .destructive) { auth.signOut() }
                }
            }
            Text("\(library.photos.count) photos · \(library.gpsCount) with GPS (\(library.gpsPercent)%) · \(library.enumerationDuration, specifier: "%.2f")s")
                .font(.callout.monospacedDigit())
                .foregroundStyle(.secondary)
        }
        .padding(16)
    }

    private var photoGrid: some View {
        Group {
            if let error = library.loadError {
                ContentUnavailableView("Could Not Load Album", systemImage: "exclamationmark.triangle", description: Text(error))
            } else if library.photos.isEmpty {
                ContentUnavailableView("No Photos", systemImage: "photo", description: Text("The selected album has no visible images."))
            } else {
                ScrollView {
                    LazyVGrid(columns: [GridItem(.adaptive(minimum: 130), spacing: 10)], spacing: 10) {
                        ForEach(library.photos) { photo in
                            PhotoThumbnail(photo: photo, library: library, uploader: uploader, isSelected: selectedPhoto?.id == photo.id)
                                .onTapGesture { selectedPhoto = photo }
                        }
                    }
                    .padding(12)
                }
            }
        }
    }

    private var detail: some View {
        Group {
            if let selectedPhoto {
                PhotoDetail(photo: selectedPhoto, library: library, uploader: uploader)
            } else {
                ContentUnavailableView("Select a Photo", systemImage: "viewfinder", description: Text("Choose a photo to inspect its full-size JPEG preview and metadata."))
            }
        }
        .background(Color(nsColor: .windowBackgroundColor))
    }
}

private struct PhotoThumbnail: View {
    let photo: PhotoAsset
    @ObservedObject var library: PhotoLibrary
    @ObservedObject var uploader: PhotoUploader
    let isSelected: Bool
    @State private var image: NSImage?

    var body: some View {
        VStack(alignment: .leading, spacing: 5) {
            ZStack(alignment: .topTrailing) {
                Group {
                    if let image {
                        Image(nsImage: image).resizable().scaledToFill()
                    } else {
                        Rectangle().fill(.quaternary).overlay { ProgressView().controlSize(.small) }
                    }
                }
                .frame(height: 125)
                .frame(maxWidth: .infinity)
                .clipped()
                if photo.coordinate != nil {
                    Image(systemName: "location.fill")
                        .font(.caption2)
                        .padding(5)
                        .background(.regularMaterial, in: Circle())
                        .padding(6)
                }
                if let state = uploader.states[photo.id] {
                    uploadBadge(state)
                        .padding(6)
                        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottomTrailing)
                }
            }
            Text(photo.takenAt?.formatted(date: .abbreviated, time: .shortened) ?? "Date unknown")
                .font(.caption)
                .lineLimit(1)
                .foregroundStyle(.secondary)
        }
        .padding(5)
        .background(isSelected ? Color.accentColor.opacity(0.15) : Color.clear, in: RoundedRectangle(cornerRadius: 8))
        .overlay(RoundedRectangle(cornerRadius: 8).stroke(isSelected ? Color.accentColor : .clear, lineWidth: 2))
        .task(id: photo.id) {
            guard image == nil else { return }
            library.requestImage(for: photo, size: CGSize(width: 400, height: 400)) { result, degraded in
                if !degraded, let result { image = result }
                else if image == nil, let result { image = result }
            }
        }
    }

    @ViewBuilder
    private func uploadBadge(_ state: UploadState) -> some View {
        switch state {
        case .uploading:
            ProgressView().controlSize(.small).padding(5).background(.regularMaterial, in: Circle())
        case .uploaded:
            Image(systemName: "checkmark.circle.fill").foregroundStyle(.green).padding(5).background(.regularMaterial, in: Circle())
        case .skipped:
            Image(systemName: "checkmark.circle").foregroundStyle(.secondary).padding(5).background(.regularMaterial, in: Circle())
        case .failed:
            Image(systemName: "exclamationmark.circle.fill").foregroundStyle(.red).padding(5).background(.regularMaterial, in: Circle())
        }
    }
}

private struct PhotoDetail: View {
    let photo: PhotoAsset
    @ObservedObject var library: PhotoLibrary
    @ObservedObject var uploader: PhotoUploader
    @State private var image: NSImage?
    @State private var jpegBytes: Int?
    @State private var place: String?
    @State private var placeError: String?
    @State private var isLoading = true
    @State private var isGeocoding = false

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            ZStack {
                Color.black.opacity(0.05)
                if let image {
                    Image(nsImage: image).resizable().scaledToFit()
                        .padding(12)
                } else if isLoading {
                    ProgressView("Loading full-size preview…")
                } else {
                    ContentUnavailableView("Preview Unavailable", systemImage: "photo.badge.exclamationmark")
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .clipShape(RoundedRectangle(cornerRadius: 10))

            VStack(alignment: .leading, spacing: 7) {
                Label(photo.takenAt?.formatted(date: .complete, time: .shortened) ?? "Date unknown", systemImage: "calendar")
                Label("\(photo.pixelWidth) × \(photo.pixelHeight) source", systemImage: "aspectratio")
                if let jpegBytes, let image {
                    let rep = image.representations.first
                    Label("Preview JPEG · \(rep?.pixelsWide ?? 0) × \(rep?.pixelsHigh ?? 0) · \(jpegBytes / 1024) KB", systemImage: "arrow.down.right.and.arrow.up.left")
                }
                if let coordinate = photo.coordinate {
                    Label(String(format: "GPS · %.5f, %.5f", coordinate.latitude, coordinate.longitude), systemImage: "location")
                    if let place {
                        Label(place, systemImage: "mappin.and.ellipse")
                    } else if isGeocoding {
                        Label("Looking up place…", systemImage: "ellipsis")
                    } else if let placeError {
                        Text(placeError).font(.caption).foregroundStyle(.secondary)
                    } else {
                        Button("Reverse-geocode this photo") {
                            isGeocoding = true
                            Task {
                                defer { isGeocoding = false }
                                do { place = try await library.place(for: photo) ?? "No place name returned" }
                                catch { placeError = error.localizedDescription }
                            }
                        }
                    }
                } else {
                    Label("No GPS location", systemImage: "location.slash")
                }
                if let caption = photo.caption, !caption.isEmpty {
                    Label(caption, systemImage: "text.quote").lineLimit(2)
                }
                Button("Upload Photo") { uploader.upload(photo) }
                    .buttonStyle(.borderedProminent)
                    .disabled(uploader.isUploading || photo.takenAt == nil)
                if photo.takenAt == nil {
                    Text("Can't upload: capture date is unavailable.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                if let state = uploader.states[photo.id] {
                    switch state {
                    case .uploading: Label("Uploading…", systemImage: "arrow.up.circle")
                    case .uploaded(let url): Label("Uploaded · \(url)", systemImage: "checkmark.circle").lineLimit(2)
                    case .skipped: Label("Already uploaded", systemImage: "checkmark.circle")
                    case .failed(let message): Label(message, systemImage: "exclamationmark.triangle").lineLimit(3)
                    }
                }
            }
            .font(.callout)
            .foregroundStyle(.secondary)
        }
        .padding(16)
        .task(id: photo.id) {
            image = nil
            jpegBytes = nil
            place = nil
            placeError = nil
            isLoading = true
            do {
                let (data, preview) = try await library.jpegPreview(for: photo)
                image = preview
                jpegBytes = data.count
            } catch {
                placeError = error.localizedDescription
            }
            isLoading = false
        }
    }
}
