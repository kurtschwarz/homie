import SwiftUI

struct SignInView: View {
    @ObservedObject var auth: AuthModel
    @State private var username = ""
    @State private var password = ""
    @FocusState private var focusedField: Field?

    private enum Field { case username, password }

    var body: some View {
        VStack(spacing: 26) {
            Spacer()
            Image(systemName: "photo.on.rectangle.angled")
                .font(.system(size: 50, weight: .medium))
                .foregroundStyle(.white)
                .frame(width: 112, height: 112)
                .background(Color.accentColor.gradient, in: RoundedRectangle(cornerRadius: 28))
                .shadow(color: .black.opacity(0.16), radius: 16, y: 8)

            VStack(spacing: 8) {
                Text("Sign in to Homie")
                    .font(.system(size: 38, weight: .bold))
                Text("Use an admin account to send iCloud Photos to your kitchen screen.")
                    .font(.title3)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }

            VStack(alignment: .leading, spacing: 16) {
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
        .frame(minWidth: 720, minHeight: 620)
        .onAppear { focusedField = .username }
    }

    private func signIn() {
        Task { await auth.signIn(username: username, password: password) }
    }
}
