import SwiftUI

/// The sheet behind the home screen's add button: a chooser that swaps to the
/// new-event form in place. "Reminder" is not built yet.
///
/// The two screens cross-fade rather than being pushed onto a NavigationStack.
/// UIKit animates a detent change whether or not SwiftUI's half of the
/// transition is animated, so a push slid sideways while the sheet separately
/// changed height. A fade runs with the resize instead of against it.
struct AddToHomieSheet: View {
    private enum Mode {
        case chooser
        case event
    }

    @Environment(\.dismiss) private var dismiss

    /// Owned by the presenter, which is where presentationDetents is applied.
    @Binding var detent: PresentationDetent

    @State private var mode: Mode = .chooser

    /// Fades the outgoing screen out before the incoming one fades in. Fading
    /// both at once overlaps two title bars and two card stacks at half
    /// opacity, which reads as a double exposure rather than a change of screen.
    private static let fadeThrough: AnyTransition = .asymmetric(
        insertion: .opacity.animation(.easeIn(duration: 0.16).delay(0.09)),
        removal: .opacity.animation(.easeOut(duration: 0.12))
    )

    var body: some View {
        ZStack {
            switch mode {
            case .chooser:
                chooser.transition(Self.fadeThrough)
            case .event:
                NavigationStack {
                    NewEventView(close: { show(.chooser, at: .medium) })
                }
                .transition(Self.fadeThrough)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color(uiColor: .systemGroupedBackground))
    }

    /// Swaps screens and resizes the sheet together, so the fade and the height
    /// change read as one movement.
    private func show(_ newMode: Mode, at newDetent: PresentationDetent) {
        withAnimation(.easeInOut(duration: 0.27)) {
            mode = newMode
            detent = newDetent
        }
    }

    private var chooser: some View {
        // Wrapped so the title sits in a bar like the form's, rather than the
        // two screens disagreeing about where a title belongs.
        NavigationStack {
            chooserBody
                .navigationTitle("Add to Homie")
                .navigationBarTitleDisplayMode(.inline)
                .toolbar {
                    ToolbarItem(placement: .cancellationAction) {
                        // Titled so VoiceOver still says "Cancel".
                        Button("Cancel", systemImage: "xmark") { dismiss() }
                            .labelStyle(.iconOnly)
                            // The toolbar's default glyph fills the circle, so
                            // the mark is sized against it by hand.
                            .font(.system(size: 13, weight: .semibold))
                    }
                }
        }
    }

    private var chooserBody: some View {
        VStack(alignment: .leading, spacing: 20) {
            RoundedCard {
                ChoiceRow(
                    symbol: "calendar.badge.plus",
                    title: "Event",
                    subtitle: "Goes on the calendar, like a dentist visit or a game"
                ) {
                    show(.event, at: .large)
                }

                Divider().padding(.leading, 72)

                ChoiceRow(
                    symbol: "bell.badge",
                    title: "Reminder",
                    subtitle: "Shows on the screen during a time window, like “art bag” on Monday mornings"
                ) {
                    // Reminders are not built yet.
                }
            }

            Spacer(minLength: 0)
        }
        .padding(.horizontal, 16)
        .padding(.top, 20)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(Color(uiColor: .systemGroupedBackground))
    }
}

private struct ChoiceRow: View {
    let symbol: String
    let title: String
    let subtitle: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 14) {
                Image(systemName: symbol)
                    .font(.title3)
                    .foregroundStyle(.tint)
                    .frame(width: 44, height: 44)
                    .background(Color.accentColor.opacity(0.12), in: RoundedRectangle(cornerRadius: 12))

                VStack(alignment: .leading, spacing: 2) {
                    Text(title).font(.title3.bold()).foregroundStyle(.primary)
                    Text(subtitle).font(.subheadline).foregroundStyle(.secondary)
                        .fixedSize(horizontal: false, vertical: true)
                        .multilineTextAlignment(.leading)
                }

                Spacer(minLength: 0)

                Image(systemName: "chevron.right")
                    .font(.footnote.bold())
                    .foregroundStyle(.tertiary)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 14)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }
}

#Preview {
    @Previewable @State var detent: PresentationDetent = .medium

    Color.clear.sheet(isPresented: .constant(true)) {
        AddToHomieSheet(detent: $detent)
            .presentationDetents([.medium, .large], selection: $detent)
            .environment(PeopleManager())
    }
}
