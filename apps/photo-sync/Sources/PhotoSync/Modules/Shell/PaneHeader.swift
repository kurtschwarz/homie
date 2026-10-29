import SwiftUI

struct PaneHeader: View {
    let pane: AppPane
    var subtitle: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 12) {
                PaneIconTile(pane: pane, size: 32)
                Text(pane.title)
                    .font(.system(size: 19, weight: .bold))
                    .foregroundStyle(Color.homieText)
                Spacer(minLength: 0)
            }
            if let subtitle {
                Text(subtitle)
                    .font(.system(size: 13))
                    .foregroundStyle(Color.homieSectionHeading)
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(.isHeader)
    }
}
