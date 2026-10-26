import SwiftUI

struct SyncStatusView: View {
    @ObservedObject var library: PhotoLibrary
    @ObservedObject var selection: AlbumSelection
    @ObservedObject var uploader: PhotoUploader
    let onSyncNow: () -> Void
    let onChangeAlbums: () -> Void

    private var albums: [PhotoAlbum] { library.albums.filter { selection.contains($0.id) } }
    private var failedCount: Int { uploader.albumProgress.values.reduce(0) { $0 + $1.failed } }
    private var photoTotal: Int { uploader.isUploading || uploader.total > 0 ? uploader.total : albums.reduce(0) { $0 + $1.count } }
    private var photoDone: Int { uploader.total > 0 ? uploader.completed : (selection.lastSyncedAt == nil ? 0 : photoTotal) }
    private var fraction: Double { photoTotal == 0 ? 0 : Double(photoDone) / Double(photoTotal) }

    private enum Phase { case ready, syncing, failed, upToDate }

    private var phase: Phase {
        if uploader.isUploading { return .syncing }
        if failedCount > 0 { return .failed }
        if selection.lastSyncedAt != nil { return .upToDate }
        return .ready
    }

    private var title: String {
        switch phase {
        case .ready: "Ready to sync"
        case .syncing: "Syncing to Homie"
        case .failed: "Some photos didn't sync"
        case .upToDate: "Homie is up to date"
        }
    }

    private var subtitle: String {
        let albumsText = "\(albums.count) \(albums.count == 1 ? "album" : "albums")"
        return "\(photoTotal.formatted()) photos from \(albumsText) \(phase == .upToDate ? "are in the slideshow." : "will be in the slideshow.")"
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            ScrollView {
                VStack(spacing: 20) {
                    ForEach(albums) { album in
                        SyncAlbumRow(album: album, library: library, progress: uploader.albumProgress[album.id], isUploading: uploader.isUploading)
                    }
                }
                .padding(.horizontal, 32)
                .padding(.vertical, 24)
            }
            Divider()
            footer
        }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack(spacing: 18) {
                ZStack {
                    Circle().fill(iconColor.opacity(0.15))
                    if phase == .syncing {
                        ProgressView().controlSize(.small)
                    } else {
                        Image(systemName: iconName).font(.system(size: 22, weight: .semibold)).foregroundStyle(iconColor)
                    }
                }
                .frame(width: 76, height: 76)
                VStack(alignment: .leading, spacing: 4) {
                    Text(title).font(.system(size: 24, weight: .bold))
                    Text(subtitle).font(.system(size: 13)).foregroundStyle(.secondary)
                }
            }
            ProgressView(value: fraction).tint(.homieGreen)
            HStack {
                Text("\(photoDone.formatted()) of \(photoTotal.formatted()) photos").font(.system(size: 13, weight: .semibold))
                Spacer()
                if let last = selection.lastSyncedAt, phase != .syncing {
                    Text("Last synced \(last.formatted(.relative(presentation: .named)))")
                        .font(.system(size: 13)).foregroundStyle(.secondary)
                }
            }
        }
        .padding(.horizontal, 32)
        .padding(.vertical, 24)
    }

    private var footer: some View {
        VStack(alignment: .leading, spacing: 14) {
            if selection.keepSyncing {
                HStack(spacing: 10) {
                    Circle().fill(Color.homieGreen).frame(width: 10, height: 10)
                    Text("New photos will sync automatically").font(.system(size: 13)).foregroundStyle(.secondary)
                }
            }
            HStack {
                Spacer()
                Button("Change albums", action: onChangeAlbums)
                    .buttonStyle(HomieButtonStyle())
                    .disabled(uploader.isUploading)
                Button("Sync now", action: onSyncNow)
                    .buttonStyle(HomieButtonStyle(prominent: true))
                    .disabled(uploader.isUploading || albums.isEmpty)
            }
        }
        .padding(.horizontal, 32)
        .padding(.vertical, 18)
        .background(Color(nsColor: .windowBackgroundColor))
    }

    private var iconName: String {
        switch phase {
        case .failed: "exclamationmark.triangle.fill"
        case .upToDate: "checkmark"
        default: "arrow.up.circle"
        }
    }

    private var iconColor: Color {
        switch phase {
        case .failed: .orange
        case .ready: .secondary
        default: .homieGreenDark
        }
    }
}

private struct SyncAlbumRow: View {
    let album: PhotoAlbum
    @ObservedObject var library: PhotoLibrary
    let progress: AlbumProgress?
    let isUploading: Bool

    private var status: String {
        guard let progress else { return "Waiting" }
        if progress.failed > 0 { return "\(progress.failed) failed" }
        if progress.isDone { return "Done" }
        return "\(progress.completed.formatted()) of \(progress.total.formatted())"
    }

    private var fraction: Double {
        guard let progress, progress.total > 0 else { return 0 }
        return Double(progress.completed) / Double(progress.total)
    }

    var body: some View {
        HStack(spacing: 18) {
            AlbumCover(album: album, library: library, size: 70)
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Text(album.title).font(.system(size: 14, weight: .semibold))
                    Spacer()
                    Text(status).font(.system(size: 12, weight: .semibold)).foregroundStyle(progress?.failed ?? 0 > 0 ? Color.orange : Color.homieGreenDark)
                }
                ProgressView(value: progress == nil ? 1 : fraction).tint(.homieGreen)
            }
        }
    }
}
