import SwiftUI

struct PaneHeader: View {
    let pane: AppPane
    var subtitle: String?

    var body: some View {
        HStack(spacing: 14) {
            PaneIconTile(pane: pane, size: 40)
            VStack(alignment: .leading, spacing: 2) {
                Text(pane.title)
                    .font(.system(size: 22, weight: .bold))
                    .foregroundStyle(Color.homieText)
                if let subtitle {
                    Text(subtitle)
                        .font(.system(size: 13))
                        .foregroundStyle(Color.homieSectionHeading)
                }
            }
            Spacer(minLength: 0)
        }
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(.isHeader)
    }
}
