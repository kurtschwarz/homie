import SwiftUI

struct SignInView: View {
    @ObservedObject var auth: AuthModel
    @ObservedObject var hostStore: HomieHostStore
    @State private var username = ""
    @State private var password = ""
    @FocusState private var focusedField: Field?

    private enum Field { case username, password }

    var body: some View {
        VStack(spacing: 0) {
            Spacer()
            HomieMark()
                .frame(width: 154, height: 154)

            VStack(spacing: 14) {
                Text("Sign in to Homie")
                    .font(.system(size: 36, weight: .bold))
                    .tracking(-0.8)
                Text("Use a manager account to send iCloud Photos\nto your Homie.")
                    .font(.system(size: 20, weight: .regular))
                    .lineSpacing(5)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }
            .padding(.top, 12)
            .padding(.bottom, 28)

            VStack(alignment: .leading, spacing: 16) {
                HomieHostControl(store: hostStore)

                VStack(alignment: .leading, spacing: 7) {
                    Text("Username").font(.headline)
                    TextField("Your Homie username", text: $username)
                        .textFieldStyle(.roundedBorder)
                        .focused($focusedField, equals: .username)
                        .onSubmit { focusedField = .password }
                }
                VStack(alignment: .leading, spacing: 7) {
                    Text("Password").font(.headline)
                    SecureField("Your Homie password", text: $password)
                        .textFieldStyle(.roundedBorder)
                        .focused($focusedField, equals: .password)
                        .onSubmit(signIn)
                }
                Button(action: signIn) {
                    HStack {
                        if auth.isSigningIn { ProgressView().controlSize(.small) }
                        Text(auth.isSigningIn ? "Signing in…" : "Sign in")
                    }
                    .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.large)
                .disabled(auth.isSigningIn || username.isEmpty || password.isEmpty)

                if case .signedOut(let message) = auth.state, let message {
                    Label(message, systemImage: "exclamationmark.triangle.fill")
                        .font(.callout)
                        .foregroundStyle(.red)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            .frame(maxWidth: 540)

            Spacer()
            Text("Only admin accounts can sync photos. Your Apple ID stays on this Mac.")
                .font(.callout)
                .foregroundStyle(.tertiary)
                .padding(.bottom, 28)
        }
        .padding(.horizontal, 56)
        .onAppear { focusedField = .username }
    }

    private func signIn() {
        Task { await auth.signIn(username: username, password: password) }
    }
}
