import AppKit
import SwiftUI

struct AboutPaneView: View {
    let info: AboutInfo

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                PaneHeader(pane: .about)
                card
            }
            .padding(.horizontal, 32)
            .padding(.top, 34)
            .padding(.bottom, 24)
        }
    }

    private var card: some View {
        VStack(spacing: 10) {
            Image(nsImage: NSApp.applicationIconImage)
                .resizable()
                .frame(width: 96, height: 96)
                .accessibilityHidden(true)
            Text(AboutInfo.name)
                .font(.system(size: 19, weight: .bold))
                .foregroundStyle(Color.homieText)
            if let versionLine = info.versionLine {
                Text(versionLine)
                    .font(.system(size: 13))
                    .foregroundStyle(Color.homieSectionHeading)
                    .textSelection(.enabled)
                    .help(info.commit ?? "")
            }
            Text(AboutInfo.summary)
                .font(.system(size: 13))
                .foregroundStyle(Color.homieText)
                .multilineTextAlignment(.center)
                .fixedSize(horizontal: false, vertical: true)
                .padding(.top, 6)
        }
        .frame(maxWidth: .infinity)
        .padding(.horizontal, 24)
        .padding(.vertical, 28)
        .background(Color.white.opacity(0.05), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
        .accessibilityElement(children: .combine)
    }
}
