import SwiftUI

struct SidebarView: View {
    let selection: AppPane
    let isSignedIn: Bool
    var isLoading = false
    let onSelect: (AppPane) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Color.clear.frame(height: GlassWindow.sidebarTopInset)
            ForEach(AppPane.sections) { section in
                if let heading = section.heading {
                    Text(heading)
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Color.homieSectionHeading)
                        .padding(.horizontal, 10)
                        .padding(.top, 16)
                        .padding(.bottom, 4)
                        .opacity(isLoading ? 0 : 1)
                        .overlay(alignment: .leading) {
                            if isLoading { SkeletonBar(width: 70, height: 10).padding(.leading, 10).padding(.top, 12) }
                        }
                }
                ForEach(section.panes) { pane in
                    if isLoading {
                        SkeletonRow(pane: pane)
                    } else {
                        SidebarRow(
                            pane: pane,
                            isSelected: selection == pane,
                            isEnabled: isSignedIn || pane.isAvailableSignedOut,
                            onSelect: { onSelect(pane) }
                        )
                    }
                }
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 12)
        .padding(.bottom, 16)
        .frame(width: 220)
        .frame(maxHeight: .infinity)
        .overlay(alignment: .trailing) {
            Rectangle().fill(Color.homieDivider).frame(width: 1)
        }
        .accessibilityElement(children: isLoading ? .ignore : .contain)
        .accessibilityLabel(isLoading ? "Sidebar, loading" : "")
    }
}

private struct SkeletonBar: View {
    let width: CGFloat
    let height: CGFloat
    @State private var dimmed = false

    var body: some View {
        Capsule()
            .fill(Color.white.opacity(dimmed ? 0.06 : 0.14))
            .frame(width: width, height: height)
            .animation(.easeInOut(duration: 0.9).repeatForever(autoreverses: true), value: dimmed)
            .onAppear { dimmed = true }
    }
}

private struct SkeletonRow: View {
    let pane: AppPane

    var body: some View {
        HStack(spacing: 10) {
            PaneIconTile(pane: pane)
                .grayscale(1)
                .opacity(0.4)
            SkeletonBar(width: CGFloat(60 + (pane.title.count * 7) % 50), height: 10)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 10)
        .frame(height: 36)
    }
}

private struct SidebarRow: View {
    let pane: AppPane
    let isSelected: Bool
    let isEnabled: Bool
    let onSelect: () -> Void
    @State private var isHovering = false

    var body: some View {
        Button(action: onSelect) {
            HStack(spacing: 10) {
                PaneIconTile(pane: pane)
                Text(pane.title)
                    .font(.system(size: 14))
                    .foregroundStyle(Color.homieText)
                Spacer(minLength: 0)
                if pane.isNew {
                    Text("New")
                        .font(.system(size: 11, weight: .bold))
                        .foregroundStyle(Color.homieBadgeText)
                        .padding(.horizontal, 8)
                        .frame(height: 20)
                        .overlay(Capsule().strokeBorder(Color.homieBadgeBorder, lineWidth: 1.5))
                }
            }
            .padding(.horizontal, 10)
            .frame(height: 36)
            .contentShape(Rectangle())
            .background(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .fill(isSelected ? Color.homieRowSelected : (isHovering && isEnabled ? Color.homieRowHover : .clear))
            )
        }
        .buttonStyle(.plain)
        .disabled(!isEnabled)
        .opacity(isEnabled ? 1 : 0.38)
        .onHover { isHovering = $0 }
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}
