import SwiftUI
import WebKit

struct HomeView: View {
    @Environment(\.scenePhase) private var scenePhase
    @State private var previewPage = WebPage()
    @State private var homeManager = HomeManager()
    @State private var isAdding = false
    @State private var addDetent: PresentationDetent = .medium

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    KioskPreview(page: previewPage)
                        .frame(maxWidth: .infinity)
                        .overlay(alignment: .topLeading) {
                            LivePill()
                                .padding(12)
                        }

                    Greeting(weather: homeManager.weather, remainingToday: homeManager.remainingToday)

                    Button {
                        // Dismissing from the form can leave the selection on
                        // .large, so the chooser's height is set on the way in.
                        addDetent = .medium
                        isAdding = true
                    } label: {
                        Label("Add an event or reminder", systemImage: "plus")
                            .font(.headline)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 10)
                    }
                    .buttonStyle(.borderedProminent)
                    // Matches RoundedCard, so the button squares up with the
                    // cards it sits above.
                    .buttonBorderShape(.roundedRectangle(radius: CardStyle.cornerRadius))
                }
                .padding(.horizontal, 16)
                .padding(.top, 8)
            }
            .background(Color(uiColor: .systemGroupedBackground))
            .toolbar(.hidden, for: .navigationBar)
            .task {
                previewPage.load(URLRequest(url: HomieEndpoint.baseURL))
                await homeManager.load()
            }
            .sheet(isPresented: $isAdding, onDismiss: { addDetent = .medium }) {
                AddToHomieSheet(detent: $addDetent)
                    .presentationDetents([.medium, .large], selection: $addDetent)
                    .presentationCornerRadius(CardStyle.cornerRadius)
                    .presentationDragIndicator(.visible)
            }
            .onChange(of: scenePhase) { _, phase in
                guard phase == .active else { return }
                previewPage.reload()
                Task { await homeManager.load() }
            }
        }
    }
}

private struct LivePill: View {
    var body: some View {
        HStack(spacing: 6) {
            Image(systemName: "circle.fill")
                .font(.system(size: 7))
                .foregroundStyle(.green)
                .symbolEffect(.pulse)

            Text("Live")
                .font(.caption.bold())
                .foregroundStyle(.white)
        }
        .padding(.horizontal, 9)
        .padding(.vertical, 5)
        .background(Capsule().fill(.black.opacity(0.6)))
    }
}

private struct Greeting: View {
    let weather: HomeManager.Weather?
    let remainingToday: Int

    private let now = Date()

    private let ballDiameter: CGFloat = 52
    /// HomieAvatar draws the ball at 80% of its frame.
    private var avatarFrame: CGFloat { ballDiameter / 0.8 }

    var body: some View {
        HStack(spacing: 14) {
            HomieAvatar(showsGroundShadow: false)
                .frame(width: avatarFrame, height: avatarFrame)
                // The frame carries transparent margin around the ball; trim it
                // so the layout box matches the ball you can actually see.
                .padding(-(avatarFrame - ballDiameter) / 2)

            text
        }
    }

    private var text: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(Self.dateLine(now))
                .font(.caption.bold())
                .tracking(1.5)
                .foregroundStyle(.secondary)

            Text(Self.salutation(now))
                .font(.system(size: 32, weight: .bold))

            if let weather {
                HStack(spacing: 8) {
                    Image(systemName: weather.symbolName)
                        .foregroundStyle(.orange)
                    Text("\(weather.temperature)° · \(Self.thingsPhrase(remainingToday))")
                }
                .font(.subheadline)
                .foregroundStyle(.secondary)
            }
        }
    }

    private static func dateLine(_ date: Date) -> String {
        date.formatted(.dateTime.weekday(.wide).month(.wide).day()).uppercased()
    }

    private static func salutation(_ date: Date) -> String {
        switch Calendar.current.component(.hour, from: date) {
        case ..<12: "Good morning"
        case ..<17: "Good afternoon"
        default: "Good evening"
        }
    }

    private static func thingsPhrase(_ count: Int) -> String {
        switch count {
        case 0: "nothing else today"
        case 1: "1 thing today"
        default: "\(count) things today"
        }
    }
}

/// A non-interactive miniature of the kitchen display.
///
/// The display SPA is laid out to a fixed 1920x1080 kiosk panel
/// (--hm-size-screen-w/h in @homie/ui) and does not reflow to phone widths, so
/// the web content renders at full kiosk size and is scaled down to fit.
private struct KioskPreview: View {
    let page: WebPage

    private let kioskWidth: CGFloat = 1920
    private let kioskHeight: CGFloat = 1080
    private let cornerRadius: CGFloat = 6

    var body: some View {
        GeometryReader { geometry in
            WebView(page)
                .frame(width: kioskWidth, height: kioskHeight)
                .scaleEffect(geometry.size.width / kioskWidth, anchor: .topLeading)
                .frame(width: geometry.size.width, height: geometry.size.height, alignment: .topLeading)
                .allowsHitTesting(false)
        }
        // allowsHitTesting blocks touches but still exposes the web content's
        // own buttons to VoiceOver, so hide the mirror from accessibility too.
        .accessibilityHidden(true)
        .aspectRatio(kioskWidth / kioskHeight, contentMode: .fit)
        .background(.white)
        .clipShape(RoundedRectangle(cornerRadius: cornerRadius))
        .overlay {
            RoundedRectangle(cornerRadius: cornerRadius)
                .strokeBorder(.black, lineWidth: 4)
        }
    }
}

#Preview {
    HomeView().environment(PeopleManager())
}
