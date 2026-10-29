import SwiftUI

struct SidebarView: View {
    let selection: AppPane
    let isSignedIn: Bool
    let onSelect: (AppPane) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Color.clear.frame(height: 52)
            ForEach(AppPane.sections, id: \.heading) { section in
                if let heading = section.heading {
                    Text(heading)
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Color.homieSectionHeading)
                        .padding(.horizontal, 10)
                        .padding(.top, 16)
                        .padding(.bottom, 4)
                }
                ForEach(section.panes) { pane in
                    SidebarRow(
                        pane: pane,
                        isSelected: selection == pane,
                        isEnabled: isSignedIn || pane.isAvailableSignedOut,
                        onSelect: { onSelect(pane) }
                    )
                }
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 12)
        .padding(.bottom, 16)
        .frame(width: 220)
        .frame(maxHeight: .infinity)
        .overlay(alignment: .trailing) {
            Rectangle().fill(Color.white.opacity(0.08)).frame(width: 1)
        }
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
                        .overlay(Capsule().strokeBorder(Color.white.opacity(0.25), lineWidth: 1.5))
                }
            }
            .padding(.horizontal, 10)
            .frame(height: 36)
            .contentShape(Rectangle())
            .background(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .fill(Color.white.opacity(isSelected ? 0.12 : (isHovering && isEnabled ? 0.06 : 0)))
            )
        }
        .buttonStyle(.plain)
        .disabled(!isEnabled)
        .opacity(isEnabled ? 1 : 0.38)
        .onHover { isHovering = $0 }
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}
