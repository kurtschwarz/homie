import SwiftUI

/// The light card shown under the no-signal glitch screen: explains that
/// events and reminders still work offline, with the add button nested inside.
public struct HomieOfflineModeCard<Icon: View>: View {
    public var title: String
    public var detail: String
    public var buttonTitle: String
    public var onAdd: () -> Void
    @ViewBuilder public var buttonIcon: () -> Icon

    public init(
        title: String = "Working offline",
        detail: String = "You can still add events and reminders. They’ll sync as soon as Homie’s back.",
        buttonTitle: String = "Add an event or reminder",
        onAdd: @escaping () -> Void,
        @ViewBuilder buttonIcon: @escaping () -> Icon = { EmptyView() }
    ) {
        self.title = title
        self.detail = detail
        self.buttonTitle = buttonTitle
        self.onAdd = onAdd
        self.buttonIcon = buttonIcon
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            HStack(alignment: .center, spacing: 16) {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(Color(uiColor: .tertiarySystemFill))
                    .frame(width: 40, height: 40)
                    .overlay {
                        Image("phosphor-plugs")
                            .resizable()
                            .renderingMode(.template)
                            .scaledToFit()
                            .frame(width: 18, height: 18)
                            .foregroundStyle(.primary)
                    }

                VStack(alignment: .leading, spacing: 4) {
                    Text(title)
                        .font(.system(size: 17, weight: .bold))
                    Text(detail)
                        .font(.system(size: 14))
                        .foregroundStyle(.secondary)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }

            HomieButton(title: buttonTitle, iconSize: 20, fillsWidth: true, action: onAdd, icon: buttonIcon)
        }
        .padding(20)
        .homieCard(cornerRadius: CardStyle.cornerRadius)
    }
}

#if DEBUG
#Preview {
    HomieOfflineModeCard {} buttonIcon: {
        Image(systemName: "plus")
            .resizable()
            .scaledToFit()
    }
    .padding(20)
    .background(Color(uiColor: .systemGroupedBackground))
}
#endif
