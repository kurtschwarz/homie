import SwiftUI

/// The new-event form reached from ``AddToHomieSheet``.
struct NewEventView: View {
    @Environment(PeopleManager.self) private var peopleManager
    @Environment(\.dismiss) private var dismiss
    /// Returns to the chooser. Supplied by the sheet, which resizes as it swaps.
    let close: () -> Void
    /// Called after a successful save, so the presenter can refresh its own
    /// event count without this view knowing about HomeManager.
    var onSaved: () -> Void = {}

    @State private var title = ""
    @State private var location = ""
    @State private var personID: String?
    @State private var day = Calendar.current.startOfDay(for: Date())
    @State private var time = Self.nextHalfHour()
    @State private var isAllDay = false
    @State private var isSaving = false
    @State private var errorMessage: String?

    /// How far ahead the day strip runs.
    private let dayCount = 14

    private var trimmedTitle: String {
        title.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private var selectedPerson: Person? {
        peopleManager.people.first { $0.id == personID }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                section("What") {
                    RoundedCard {
                        TextField("Dentist", text: $title)
                            .font(.title3)
                            .padding(16)
                    }
                }

                section("Where") {
                    RoundedCard {
                        TextField("Location (optional)", text: $location)
                            .padding(16)
                    }
                }

                section("For") {
                    if peopleManager.people.isEmpty {
                        if peopleManager.isLoading {
                            ProgressView("Loading people")
                        } else {
                            Text(peopleManager.errorMessage ?? "Add a person in Settings before creating an event.")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    } else {
                        peoplePicker
                    }
                }

                section("When") {
                    VStack(alignment: .leading, spacing: 12) {
                        dayStrip
                        timeRow
                    }
                }

                if let errorMessage {
                    Text(errorMessage)
                        .font(.subheadline)
                        .foregroundStyle(.red)
                }

                kitchenPreview
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 20)
        }
        .background(Color(uiColor: .systemGroupedBackground))
        .navigationTitle("New event")
        .navigationBarTitleDisplayMode(.inline)
        .navigationBarBackButtonHidden()
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                // Titled so VoiceOver still says "Cancel". Sized to match the
                // chooser's close button, which the toolbar's default fills.
                Button("Cancel", systemImage: "chevron.left") { close() }
                    .labelStyle(.iconOnly)
                    .font(.system(size: 13, weight: .semibold))
                    .disabled(isSaving)
            }
            ToolbarItem(placement: .confirmationAction) {
                if isSaving {
                    ProgressView()
                } else {
                    Button("Add") { Task { await save() } }
                        .disabled(trimmedTitle.isEmpty || personID == nil)
                }
            }
        }
        .task {
            await peopleManager.loadPeople()
            if personID == nil { personID = peopleManager.people.first?.id }
        }
    }

    /// Start of the event: `day`'s date combined with `time`'s time of day,
    /// or `day` itself at midnight when all-day.
    private var start: Date {
        guard !isAllDay else { return day }
        let calendar = Calendar.current
        let timeParts = calendar.dateComponents([.hour, .minute], from: time)
        return calendar.date(bySettingHour: timeParts.hour ?? 0, minute: timeParts.minute ?? 0, second: 0, of: day) ?? day
    }

    /// End of the event: one day after `start` when all-day (DTEND is
    /// exclusive), otherwise an hour after `start`.
    private var end: Date {
        let calendar = Calendar.current
        if isAllDay {
            return calendar.date(byAdding: .day, value: 1, to: start) ?? start
        }
        return calendar.date(byAdding: .hour, value: 1, to: start) ?? start
    }

    private func save() async {
        isSaving = true
        errorMessage = nil
        defer { isSaving = false }
        do {
            var input: [String: Any] = [
                "title": trimmedTitle,
                "start": Int(start.timeIntervalSince1970),
                "end": Int(end.timeIntervalSince1970),
                "allDay": isAllDay,
                "personIds": personID.map { [$0] } ?? [],
            ]
            let trimmedLocation = location.trimmingCharacters(in: .whitespacesAndNewlines)
            if !trimmedLocation.isEmpty { input["location"] = trimmedLocation }
            struct Payload: Decodable { let createEvent: CreatedEvent }
            struct CreatedEvent: Decodable { let id: String }
            let _: Payload = try await HomieAPI.perform(
                """
                mutation($input: EventInput!) {
                    createEvent(input: $input) { id }
                }
                """,
                variables: ["input": input]
            )
            onSaved()
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private var peoplePicker: some View {
        FlowLayout(spacing: 10) {
            ForEach(peopleManager.people) { person in
                let isSelected = person.id == personID
                Button { personID = person.id } label: {
                    HStack(spacing: 8) {
                        PersonBadge(person: person, diameter: 32)
                        Text(person.name).font(.headline).foregroundStyle(.primary)
                    }
                    .padding(.leading, 6)
                    .padding(.trailing, 14)
                    .padding(.vertical, 6)
                    .background(Color(uiColor: .secondarySystemGroupedBackground), in: Capsule())
                    .overlay {
                        Capsule().strokeBorder(isSelected ? person.color : .clear, lineWidth: 2)
                    }
                }
                .buttonStyle(.plain)
            }
        }
    }

    private var dayStrip: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                ForEach(days, id: \.self) { date in
                    let isSelected = Calendar.current.isDate(date, inSameDayAs: day)
                    Button { day = date } label: {
                        VStack(spacing: 2) {
                            Text(Self.dayLabel(date)).font(.subheadline.weight(.medium))
                            Text(date.formatted(.dateTime.month(.abbreviated).day())).font(.headline)
                        }
                        .frame(minWidth: 72)
                        .padding(.vertical, 10)
                        .padding(.horizontal, 8)
                        .foregroundStyle(isSelected ? Color.white : .primary)
                        .background(
                            isSelected ? AnyShapeStyle(Color.accentColor)
                                       : AnyShapeStyle(Color(uiColor: .secondarySystemGroupedBackground)),
                            in: RoundedRectangle(cornerRadius: 12)
                        )
                    }
                    .buttonStyle(.plain)
                }
            }
        }
        // Let the cards bleed to the screen edge while the form stays inset.
        .padding(.horizontal, -16)
        .contentMargins(.horizontal, 16, for: .scrollContent)
    }

    private var timeRow: some View {
        HStack(spacing: 10) {
            RoundedCard {
                DatePicker("Time", selection: $time, displayedComponents: .hourAndMinute)
                    .labelsHidden()
                    .padding(12)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            .opacity(isAllDay ? 0.4 : 1)
            .disabled(isAllDay)

            Button { isAllDay.toggle() } label: {
                Text("All day")
                    .font(.headline)
                    .foregroundStyle(isAllDay ? Color.white : .primary)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 14)
                    .background(
                        isAllDay ? AnyShapeStyle(Color.accentColor)
                                 : AnyShapeStyle(Color(uiColor: .secondarySystemGroupedBackground)),
                        in: RoundedRectangle(cornerRadius: CardStyle.cornerRadius)
                    )
            }
            .buttonStyle(.plain)
        }
    }

    private var kitchenPreview: some View {
        RoundedCard {
            VStack(alignment: .leading, spacing: 12) {
                Text("ON THE KITCHEN SCREEN")
                    .font(.caption.bold()).tracking(1.5).foregroundStyle(.secondary)

                HStack(spacing: 14) {
                    Text(isAllDay ? "All day" : time.formatted(.dateTime.hour().minute()))
                        .foregroundStyle(.secondary)
                    if let selectedPerson {
                        PersonBadge(person: selectedPerson, diameter: 32)
                    }
                    Text(trimmedTitle.isEmpty ? "Your event" : trimmedTitle)
                        .font(.body.weight(.semibold))
                    Spacer(minLength: 0)
                }

                Text(summary)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .padding(16)
        }
    }

    private var summary: String {
        let when = Self.dayLabel(day).lowercased()
        let at = isAllDay ? "all day" : "at \(time.formatted(.dateTime.hour().minute()))"
        guard let selectedPerson else { return "Shows on the screen \(when) \(at)." }
        return "Shows with \(selectedPerson.name)’s \(selectedPerson.badgeText) \(when) \(at)."
    }

    private var days: [Date] {
        let calendar = Calendar.current
        let today = calendar.startOfDay(for: Date())
        return (0..<dayCount).compactMap { calendar.date(byAdding: .day, value: $0, to: today) }
    }

    private func section<Content: View>(_ title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title.uppercased()).font(.caption.bold()).tracking(1.5).foregroundStyle(.secondary)
            content()
        }
    }

    private static func dayLabel(_ date: Date) -> String {
        let calendar = Calendar.current
        if calendar.isDateInToday(date) { return "Today" }
        if calendar.isDateInTomorrow(date) { return "Tmrw" }
        return date.formatted(.dateTime.weekday(.abbreviated))
    }

    /// Events are rarely scheduled to the minute, so the time starts on the
    /// next half hour rather than whatever minute the form was opened.
    private static func nextHalfHour() -> Date {
        let calendar = Calendar.current
        let now = Date()
        var parts = calendar.dateComponents([.year, .month, .day, .hour, .minute], from: now)
        let minute = parts.minute ?? 0
        // Rebuild from the top of the hour so the seconds are dropped too.
        parts.minute = 0
        guard let hourStart = calendar.date(from: parts) else { return now }
        return calendar.date(byAdding: .minute, value: minute < 30 ? 30 : 60, to: hourStart) ?? now
    }
}

#Preview {
    NavigationStack {
        NewEventView(close: {}).environment(PeopleManager())
    }
}
