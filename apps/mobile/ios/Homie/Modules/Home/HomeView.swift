import SwiftUI
import WebKit

struct HomeView: View {
    @Environment(\.scenePhase) private var scenePhase
    @State private var previewPage = WebPage()
    @State private var homeManager = HomeManager()
    @State private var isAdding = false
    @State private var isShowingSettings = false
    @State private var addDetent: PresentationDetent = .medium

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    AgentChatCard(weather: homeManager.weather, remainingToday: homeManager.remainingToday)

                    KitchenScreenSection(page: previewPage)

                    QuickActions()

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
            .background(Color(red: 244.0 / 255, green: 246.0 / 255, blue: 248.0 / 255))
            .navigationTitle("Homie")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        isShowingSettings = true
                    } label: {
                        Image("phosphor-sliders-horizontal")
                    }
                    .accessibilityLabel("Settings")
                }
            }
            .task {
                previewPage.load(URLRequest(url: HomieEndpoint.baseURL))
                await homeManager.load()
            }
            .sheet(isPresented: $isAdding, onDismiss: { addDetent = .medium }) {
                AddToHomieSheet(detent: $addDetent, onEventSaved: { Task { await homeManager.load() } })
                    .presentationDetents([.medium, .large], selection: $addDetent)
                    .presentationBackground(Color(uiColor: .systemGroupedBackground))
                    .presentationDragIndicator(.visible)
            }
            .sheet(isPresented: $isShowingSettings) {
                SettingsView()
                    .presentationDetents([.large])
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

private struct AgentChatCard: View {
    let weather: HomeManager.Weather?
    let remainingToday: Int

    private let now = Date()

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            HomieAvatar(showsGroundShadow: false)
                .frame(width: 64, height: 64)

            VStack(alignment: .leading, spacing: 6) {
                Text("\(Self.salutation(now))!")
                    .font(.headline)
                    .foregroundStyle(.primary)

                HStack(spacing: 7) {
                    Image(systemName: weather?.symbolName ?? "cloud.fill")
                        .foregroundStyle(.orange)
                    if let weather {
                        Text("\(weather.temperature)° · \(Self.weatherDescription(weather.condition))")
                    } else {
                        Text("Weather unavailable")
                    }
                }
                .font(.body)
                .foregroundStyle(.secondary)

                Text(Self.thingsPhrase(remainingToday))
                    .font(.body)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                Color(uiColor: .secondarySystemGroupedBackground),
                in: RoundedRectangle(cornerRadius: 24)
            )
        }
    }

    private static func weatherDescription(_ condition: String) -> String {
        switch condition {
        case "SUN": "sunny"
        case "CLOUD_SUN": "partly cloudy"
        case "CLOUD": "cloudy"
        case "RAIN": "rainy"
        case "MOON": "clear tonight"
        default: "cloudy"
        }
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
        case 0: "Nothing else is planned today"
        case 1: "You have 1 thing left today"
        default: "You have \(count) things left today"
        }
    }
}

private struct KitchenScreenSection: View {
    let page: WebPage

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Your home")
                .font(.title3.bold())
                .padding(.horizontal, 4)

            HStack(spacing: 14) {
                KioskPreview(page: page, compact: true)
                    .frame(width: 138)

                VStack(alignment: .leading, spacing: 5) {
                    Text("Kitchen screen")
                        .font(.headline)
                        .foregroundStyle(.primary)

                    HStack(spacing: 6) {
                        Circle()
                            .fill(.green)
                            .frame(width: 7, height: 7)
                        Text("Live · Kitchen display")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                            .lineLimit(1)
                    }
                }

                Spacer(minLength: 0)
            }
            .padding(12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                Color(uiColor: .secondarySystemGroupedBackground),
                in: RoundedRectangle(cornerRadius: CardStyle.cornerRadius)
            )
        }
    }
}

private struct QuickActions: View {
    private let columns = [GridItem(.flexible()), GridItem(.flexible())]
    @State private var isStartingTimer = false

    var body: some View {
        LazyVGrid(columns: columns, spacing: 12) {
            QuickAction(symbol: "timer", title: "Start a timer") {
                isStartingTimer = true
            }

            QuickAction(symbol: "bubble.left", title: "Show a message") {
                // Message actions are not built yet.
            }
        }
        .sheet(isPresented: $isStartingTimer) {
            StartTimerSheet()
                .presentationDetents([.medium, .large])
                .presentationBackground(Color(uiColor: .systemGroupedBackground))
                .presentationDragIndicator(.visible)
        }
    }
}

private struct StartTimerSheet: View {
    @Environment(\.dismiss) private var dismiss
    @State private var title = "Pasta"
    @State private var emoji = "🍝"
    @State private var minutes = 5
    @State private var isSaving = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Form {
                Section("Timer") {
                    TextField("Name", text: $title)
                    TextField("Emoji", text: $emoji)
                        .textInputAutocapitalization(.never)
                    Stepper("Duration: \(minutes) min", value: $minutes, in: 1...120)
                }

                if let errorMessage {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                }
            }
            .navigationTitle("Start a timer")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                        .disabled(isSaving)
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(isSaving ? "Starting…" : "Start") {
                        Task { await startTimer() }
                    }
                    .disabled(isSaving || title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                }
            }
        }
    }

    private func startTimer() async {
        isSaving = true
        errorMessage = nil
        defer { isSaving = false }

        do {
            let _: StartTimerPayload = try await HomieAPI.perform(
                """
                mutation StartTimer($title: String!, $emoji: String!, $durationSeconds: Int!) {
                    startTimer(title: $title, emoji: $emoji, durationSeconds: $durationSeconds) { title }
                }
                """,
                variables: [
                    "title": title.trimmingCharacters(in: .whitespacesAndNewlines),
                    "emoji": emoji,
                    "durationSeconds": minutes * 60,
                ]
            )
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

private struct StartTimerPayload: Decodable {
    let startTimer: StartedTimer
}

private struct StartedTimer: Decodable {
    let title: String
}

private struct QuickAction: View {
    let symbol: String
    let title: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: 0) {
                Image(systemName: symbol)
                    .font(.title2)
                    .foregroundStyle(.tint)

                Spacer(minLength: 12)

                Text(title)
                    .font(.headline)
                    .foregroundStyle(.primary)
                    .multilineTextAlignment(.leading)
            }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .frame(height: 104)
            .contentShape(Rectangle())
            .background {
                RoundedRectangle(cornerRadius: CardStyle.cornerRadius)
                    .fill(Color(uiColor: .secondarySystemGroupedBackground))
            }
        }
        .buttonStyle(.plain)
    }
}

/// A non-interactive miniature of the kitchen display.
///
/// The display SPA is laid out to a fixed 1920x1080 kiosk panel
/// (--hm-size-screen-w/h in @homie/ui) and does not reflow to phone widths, so
/// the web content renders at full kiosk size and is scaled down to fit.
private struct KioskPreview: View {
    let page: WebPage
    let compact: Bool

    private let kioskWidth: CGFloat = 1920
    private let kioskHeight: CGFloat = 1080
    private var bezelWidth: CGFloat { compact ? 4 : 12 }
    private var screenCornerRadius: CGFloat { compact ? 3 : 6 }
    private var frameCornerRadius: CGFloat { compact ? 8 : 18 }

    init(page: WebPage, compact: Bool = false) {
        self.page = page
        self.compact = compact
    }

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
        .clipShape(RoundedRectangle(cornerRadius: screenCornerRadius))
        .padding(bezelWidth)
        .background {
            RoundedRectangle(cornerRadius: frameCornerRadius)
                .fill(
                    LinearGradient(
                        colors: [Color(white: 0.30), Color(white: 0.06), Color(white: 0.16)],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
        }
        .shadow(color: .black.opacity(0.30), radius: compact ? 4 : 14, x: 0, y: compact ? 2 : 12)
        .overlay {
            RoundedRectangle(cornerRadius: frameCornerRadius)
                .strokeBorder(
                    LinearGradient(
                        colors: [Color(white: 0.54), Color(white: 0.17), Color(white: 0.025), Color(white: 0.22)],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    ),
                    lineWidth: 1.5
                )
        }
        .overlay {
            RoundedRectangle(cornerRadius: screenCornerRadius)
                .strokeBorder(.white.opacity(0.26), lineWidth: 1)
                .padding(bezelWidth)
        }
    }
}

#Preview {
    HomeView().environment(PeopleManager())
}
