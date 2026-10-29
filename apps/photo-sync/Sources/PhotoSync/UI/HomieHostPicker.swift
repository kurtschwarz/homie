import SwiftUI

struct HomieHostControl: View {
    @ObservedObject var store: HomieHostStore
    @ObservedObject private var discovery: HomieDiscovery
    @State private var showsPicker = false

    init(store: HomieHostStore) {
        self.store = store
        discovery = store.discovery
    }

    var body: some View {
        Button { showsPicker = true } label: {
            HStack(spacing: 5) {
                Text(store.selected.name)
                    .font(.system(size: 13))
                    .foregroundStyle(Color.homieSectionHeading)
                Image(systemName: "chevron.down")
                    .font(.system(size: 9, weight: .semibold))
                    .foregroundStyle(Color.homieSectionHeading)
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .help(statusText)
        .sheet(isPresented: $showsPicker) {
            HomieHostPicker(store: store, isPresented: $showsPicker)
        }
        .task {
            let initialHost = store.selected
            discovery.start()
            await store.validateSelected()
            if store.health == nil, store.selected.source == .fallback {
                do { try await Task.sleep(for: .seconds(2.5)) }
                catch { return }
                if !Task.isCancelled,
                   !store.isChecking,
                   store.selected.id == initialHost.id,
                   store.selected.source == .fallback,
                   let discovered = discovery.hosts.first {
                    try? await store.select(discovered)
                }
            }
        }
    }

    private var statusText: String {
        if store.isChecking { return "Checking \(store.selected.address)…" }
        if store.requiresSetup { return "Setup required · \(store.selected.address)" }
        if store.health != nil {
            return store.selected.source == .discovered
                ? "Found on your network · \(store.selected.address)"
                : "Connected · \(store.selected.address)"
        }
        return store.selected.address
    }
}

struct HomieHostPicker: View {
    @ObservedObject var store: HomieHostStore
    @ObservedObject private var discovery: HomieDiscovery
    @Binding var isPresented: Bool
    @State private var manualAddress = ""
    @State private var errorMessage: String?
    @State private var isConnecting = false

    init(store: HomieHostStore, isPresented: Binding<Bool>) {
        self.store = store
        discovery = store.discovery
        _isPresented = isPresented
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text("Choose your Homie").font(.title2.bold())
                    Text("Select one found nearby or enter its address.").foregroundStyle(.secondary)
                }
                Spacer()
                Button("Done") { isPresented = false }
            }

            GroupBox("On your network") {
                VStack(spacing: 0) {
                    if discovery.hosts.isEmpty {
                        HStack {
                            if discovery.isSearching { ProgressView().controlSize(.small) }
                            Text(discovery.isSearching ? "Looking for Homie…" : "No Homies found")
                                .foregroundStyle(.secondary)
                            Spacer()
                        }
                        .padding(12)
                    } else {
                        ForEach(discovery.hosts) { host in
                            Button { connect(to: host) } label: {
                                HStack {
                                    Image(systemName: "display").foregroundStyle(.green)
                                    VStack(alignment: .leading) {
                                        Text(host.name).foregroundStyle(.primary)
                                        Text(host.address).font(.caption).foregroundStyle(.secondary)
                                    }
                                    Spacer()
                                    if store.selected.id == host.id { Image(systemName: "checkmark").foregroundStyle(.green) }
                                }
                                .contentShape(Rectangle())
                                .padding(10)
                            }
                            .buttonStyle(.plain)
                        }
                    }
                }
            }

            GroupBox("Manual address") {
                HStack {
                    TextField("homie.local", text: $manualAddress)
                        .textFieldStyle(.roundedBorder)
                        .onSubmit(connectManually)
                    Button("Connect", action: connectManually)
                        .disabled(manualAddress.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || isConnecting)
                }
                .padding(10)
            }

            if let errorMessage {
                Label(errorMessage, systemImage: "exclamationmark.triangle.fill")
                    .font(.callout)
                    .foregroundStyle(.red)
            }
        }
        .padding(24)
        .frame(width: 520)
        .task { discovery.start() }
        .overlay {
            if isConnecting { ProgressView("Connecting…").padding(18).background(.regularMaterial, in: RoundedRectangle(cornerRadius: 12)) }
        }
    }

    private func connectManually() {
        guard let host = HomieHost.manual(manualAddress) else {
            errorMessage = "Enter a valid hostname or URL."
            return
        }
        connect(to: host)
    }

    private func connect(to host: HomieHost) {
        guard !isConnecting else { return }
        isConnecting = true
        errorMessage = nil
        Task {
            defer { isConnecting = false }
            do {
                try await store.select(host)
                isPresented = false
            } catch {
                errorMessage = "Could not connect to \(host.address): \(error.localizedDescription)"
            }
        }
    }
}
