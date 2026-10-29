import SwiftUI

struct PaneHeader: View {
    let pane: AppPane
    var subtitle: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 14) {
                PaneIconTile(pane: pane, size: 40)
                Text(pane.title)
                    .font(.system(size: 22, weight: .bold))
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
