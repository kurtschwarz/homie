import SwiftUI
import Combine

struct HomeView: View {
    @Environment(\.scenePhase) private var scenePhase
    @State private var homeManager = HomeManager()
    @State private var displayManager = DisplayManager()
    @State private var isAdding = false
    @State private var addDetent: PresentationDetent = .medium

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    AgentChatCard(weather: homeManager.weather, remainingToday: homeManager.remainingToday)

                    KitchenScreenSection(photoURL: displayManager.currentPhotoURL, photo: displayManager.currentPhoto, headline: homeManager.headline)

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
            .scrollBounceBehavior(.always)
            .homieRefreshable { await loadAll() }
            .background(Color(red: 244.0 / 255, green: 246.0 / 255, blue: 248.0 / 255))
            .navigationTitle("Homie")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    NavigationLink {
                        SettingsView()
                    } label: {
                        Image("phosphor-sliders-horizontal")
                    }
                    .accessibilityLabel("Settings")
                }
            }
            .task {
                await loadAll()
            }
            .task(id: displayManager.loadGeneration) {
                try? await Task.sleep(for: displayManager.refreshDelay)
                guard !Task.isCancelled else { return }
                await displayManager.load()
            }
            .sheet(isPresented: $isAdding, onDismiss: { addDetent = .medium }) {
                AddToHomieSheet(detent: $addDetent, onEventSaved: { Task { await homeManager.load() } })
                    .presentationDetents([.medium, .large], selection: $addDetent)
                    .presentationBackground(Color(uiColor: .systemGroupedBackground))
                    .presentationDragIndicator(.visible)
            }
            .onChange(of: scenePhase) { _, phase in
                guard phase == .active else { return }
                Task { await loadAll() }
            }
        }
    }

    private func loadAll() async {
        async let home: Void = homeManager.load()
        async let display: Void = displayManager.load()
        _ = await (home, display)
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
    let photoURL: URL?
    let photo: DisplayManager.Photo?
    let headline: HomeManager.Headline?

    @State private var now = Date()
    private let clockTimer = Timer.publish(every: 30, on: .main, in: .common).autoconnect()

    var body: some View {
        VStack(spacing: 0) {
            GeometryReader { geometry in
                let photoWidth = geometry.size.width * 0.49
                HStack(spacing: 0) {
                    ZStack(alignment: .topLeading) {
                        Color(uiColor: .tertiarySystemFill)
                        if let photoURL {
                            AsyncImage(url: photoURL) { image in
                                FocalPhotoCrop(image: image, photo: photo, frameSize: CGSize(width: photoWidth, height: geometry.size.height))
                                    .frame(width: photoWidth, height: geometry.size.height)
                                    .clipped()
                            } placeholder: {
                                Color(uiColor: .tertiarySystemFill)
                            }
                        }
                        HStack(spacing: 5) {
                            Circle()
                                .frame(width: 6, height: 6)
                            Text("Live")
                        }
                            .font(.caption.bold())
                            .foregroundStyle(.white)
                            .padding(.horizontal, 9)
                            .padding(.vertical, 5)
                            .background(.black.opacity(0.55), in: Capsule())
                            .padding(10)
                    }
                    .frame(width: photoWidth, height: geometry.size.height)
                    .clipped()

                    VStack(alignment: .leading, spacing: 0) {
                        Text(now.formatted(.dateTime.hour().minute()))
                            .font(.system(size: 18, weight: .semibold, design: .rounded))
                            .lineLimit(1)
                            .minimumScaleFactor(0.7)
                        Spacer(minLength: 14)
                        Text(headline?.label ?? "NEXT")
                            .font(.caption.weight(.bold))
                            .tracking(1.2)
                            .foregroundStyle(Color(red: 0.88, green: 0.36, blue: 0.18))
                            .lineLimit(1)
                        Text(headline?.title ?? "Nothing else today")
                            .font(.headline)
                            .lineLimit(2)
                            .minimumScaleFactor(0.75)
                            .padding(.top, 5)
                        Text(headline?.detail ?? Self.eveningPhrase(now))
                            .font(.body)
                            .foregroundStyle(.secondary)
                            .lineLimit(2)
                            .minimumScaleFactor(0.75)
                            .padding(.top, 4)
                        Spacer(minLength: 0)
                    }
                    .padding(16)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
                }
            }
            .frame(height: 160)

            Rectangle()
                .fill(Color(uiColor: .separator))
                .frame(height: 1)

            HStack(spacing: 0) {
                control(title: "Pause", icon: "phosphor-pause")
                control(title: "Next", icon: "phosphor-skip-forward")
                control(title: "Dim", icon: "phosphor-moon")
            }
            .padding(.vertical, 12)
        }
        .background(Color(uiColor: .secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 20))
        .onReceive(clockTimer) { now = $0 }
    }

    private func control(title: String, icon: String) -> some View {
        Button {} label: {
            HStack(spacing: 6) {
                Image(icon)
                    .resizable()
                    .renderingMode(.template)
                    .frame(width: 20, height: 20)
                Text(title)
                    .font(.subheadline.weight(.semibold))
            }
            .foregroundStyle(.primary)
            .frame(maxWidth: .infinity)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    private static func eveningPhrase(_ date: Date) -> String {
        switch Calendar.current.component(.hour, from: date) {
        case ..<12: "Enjoy the day"
        case ..<17: "Enjoy the afternoon"
        default: "Enjoy the evening"
        }
    }
}

private struct FocalPhotoCrop: View {
    let image: Image
    let photo: DisplayManager.Photo?
    let frameSize: CGSize

    var body: some View {
        let geometry = photoGeometry(photo: photo, frame: frameSize)
        image
            .resizable()
            .scaledToFill()
            .frame(width: geometry.size.width, height: geometry.size.height)
            .offset(x: geometry.offset.width, y: geometry.offset.height)
    }

    private func photoGeometry(photo: DisplayManager.Photo?, frame: CGSize) -> (size: CGSize, offset: CGSize) {
        guard let photo else { return (frame, .zero) }
        let aspectRatio = CGFloat(photo.width) / CGFloat(photo.height)
        var width = frame.height * aspectRatio
        var height = frame.height
        if width < frame.width {
            width = frame.width
            height = frame.width / aspectRatio
        }

        guard let focus = photo.focus else {
            return (CGSize(width: width, height: height), CGSize(width: (frame.width - width) / 2, height: (frame.height - height) / 2))
        }
        let focusX = CGFloat(focus.x + focus.width / 2)
        let focusY = CGFloat(focus.y + focus.height / 2)
        let x = min(0, max(frame.width - width, frame.width / 2 - focusX * width))
        let y = min(0, max(frame.height - height, frame.height / 2 - focusY * height))
        return (CGSize(width: width, height: height), CGSize(width: x, height: y))
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

#Preview {
    HomeView().environment(PeopleManager())
}
