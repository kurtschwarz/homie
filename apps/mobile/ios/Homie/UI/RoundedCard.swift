import SwiftUI

/// The corner radius cards share with anything that has to line up with them,
/// like the add button and the add sheet.
enum CardStyle {
    static let cornerRadius: CGFloat = 20
}

struct RoundedCard<Content: View>: View {
    private let content: Content

    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    var body: some View {
        VStack(spacing: 0) {
            content
        }
        .background(Color(uiColor: .secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: CardStyle.cornerRadius))
    }
}
