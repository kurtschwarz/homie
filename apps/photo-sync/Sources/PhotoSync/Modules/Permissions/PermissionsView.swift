import AppKit
import Photos
import SwiftUI

struct PermissionsView: View {
    @ObservedObject var permissions: PermissionsModel
    @ObservedObject var library: PhotoLibrary
    @Environment(\.openWindow) private var openWindow

    private static let contentWidth: CGFloat = 340

    var body: some View {
        let gate = permissions.gate
        VStack(spacing: 0) {
            HomieMark()
                .frame(width: 88, height: 88)

            Text("Three quick permissions")
                .font(.system(size: 24, weight: .bold))
                .foregroundStyle(.white)
                .padding(.top, 14)
                .padding(.bottom, 22)

            VStack(spacing: 0) {
                PermissionRow(
                    title: "Photos",
                    detail: photosDetail(gate),
                    isDone: gate.photosAllowed,
                    actionTitle: gate.photos == .denied || gate.photos == .restricted ? "Open System Settings" : "Allow",
                    isBusy: permissions.isRequestingPhotos
                ) {
                    if gate.photos == .denied || gate.photos == .restricted {
                        permissions.openPhotosSettings()
                    } else {
                        Task {
                            async let request: Void = permissions.allowPhotos()
                            // PhotoKit answers about two seconds late, so keep asking rather than wait.
                            let nudge = Task {
                                while !Task.isCancelled {
                                    try? await Task.sleep(for: .milliseconds(50))
                                    comeBack()
                                }
                            }
                            await request
                            nudge.cancel()
                            comeBack()
                        }
                    }
                }
                Rectangle().fill(Color.homieDivider).frame(height: 1)
                PermissionRow(
                    title: "Local Network",
                    detail: gate.localNetwork == .denied ? "Not allowed" : "Lets Homie Photos find Homie on your Wi-Fi",
                    isDone: gate.localNetwork == .allowed,
                    actionTitle: gate.localNetwork == .denied ? "Open System Settings" : "Allow",
                    isBusy: permissions.isRequestingLocalNetwork
                ) {
                    if gate.localNetwork == .denied {
                        permissions.openLocalNetworkSettings()
                    } else {
                        Task {
                            await permissions.allowLocalNetwork()
                            comeBack()
                        }
                    }
                }
                Rectangle().fill(Color.homieDivider).frame(height: 1)
                PermissionRow(
                    title: "Keychain",
                    detail: gate.keychain == .denied ? "Not allowed" : "Keeps you signed in to Homie",
                    isDone: gate.keychain == .allowed,
                    actionTitle: gate.keychain == .denied ? "Try again" : "Allow",
                    isBusy: permissions.isRequestingKeychain
                ) {
                    Task {
                        await permissions.allowKeychain()
                        comeBack()
                    }
                }
            }
            .frame(width: Self.contentWidth)
            .background(Color.homieDivider, in: RoundedRectangle(cornerRadius: 12))
            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.homieDivider, lineWidth: 1))

            Button("Reset saved sign-in") { permissions.resetSavedSignIn() }
                .buttonStyle(.plain)
                .font(.system(size: 12))
                .foregroundStyle(Color.homieText.opacity(0.7))
                .padding(.top, 12)

            Button(gate.continueTitle) { permissions.continueIfSatisfied() }
                .buttonStyle(HomieButtonStyle(prominent: true, fullWidth: true))
                .disabled(!gate.isSatisfied)
                .frame(width: Self.contentWidth)
                .padding(.top, 14)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    /// A system prompt drops the window behind whatever else is on screen, and both openWindow and
    /// activate are no-ops while another app owns the front. orderFrontRegardless is not.
    private func comeBack() {
        openWindow(id: "main")
        NSApp.activate(ignoringOtherApps: true)
        NSApp.windows.first { $0.identifier?.rawValue == "main" }?.orderFrontRegardless()
    }

    private func photosDetail(_ gate: PermissionsGate) -> String {
        if permissions.isRequestingPhotos { return "Waiting for Photos…" }
        switch gate.photos {
        case .denied, .restricted: return "Not allowed"
        default: return "Reads your albums to sync them"
        }
    }
}

private struct PermissionRow: View {
    let title: String
    let detail: String
    let isDone: Bool
    let actionTitle: String
    let isBusy: Bool
    let action: () -> Void

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: isDone ? "checkmark.circle.fill" : "circle")
                .font(.system(size: 16))
                .foregroundStyle(isDone ? Color.homieGreen : Color.homieText.opacity(0.4))
                .frame(width: 18, height: 18)
            VStack(alignment: .leading, spacing: 1) {
                Text(title)
                    .font(.system(size: 14))
                    .foregroundStyle(Color.homieText)
                Text(detail)
                    .font(.system(size: 11))
                    .foregroundStyle(Color.homieText.opacity(0.6))
            }
            Spacer(minLength: 0)
            if !isDone {
                if isBusy {
                    ProgressView().controlSize(.small)
                } else {
                    Button(actionTitle, action: action)
                        .buttonStyle(.link)
                        .font(.system(size: 12, weight: .semibold))
                }
            }
        }
        .padding(.horizontal, 14)
        .frame(height: 52)
        .accessibilityElement(children: .combine)
        .accessibilityValue(isDone ? "Allowed" : "Not allowed")
    }
}
