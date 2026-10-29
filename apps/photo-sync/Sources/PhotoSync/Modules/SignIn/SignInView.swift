import SwiftUI

struct SignInView: View {
    @ObservedObject var auth: AuthModel
    @ObservedObject var hostStore: HomieHostStore
    @State private var username = ""
    @State private var password = ""
    @State private var isBlinking = false
    @FocusState private var focusedField: Field?

    private enum Field { case username, password }

    var body: some View {
        VStack(spacing: 0) {
            HomieMark(eyesClosed: isBlinking)
                .frame(width: 88, height: 88)

            Text("Sign in to Homie")
                .font(.system(size: 24, weight: .bold))
                .foregroundStyle(.white)
                .padding(.top, 14)

            HomieHostControl(store: hostStore)
                .padding(.top, 6)
                .padding(.bottom, 22)

            VStack(spacing: 0) {
                TextField("Username", text: $username)
                    .textFieldStyle(.plain)
                    .focused($focusedField, equals: .username)
                    .onSubmit { focusedField = .password }
                    .padding(.horizontal, 14)
                    .frame(height: 44)
                Rectangle().fill(Color.homieDivider).frame(height: 1)
                SecureField("Password", text: $password)
                    .textFieldStyle(.plain)
                    .focused($focusedField, equals: .password)
                    .onSubmit(signIn)
                    .padding(.horizontal, 14)
                    .frame(height: 44)
            }
            .frame(width: 340)
            .background(Color.homieDivider, in: RoundedRectangle(cornerRadius: 12))
            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.homieDivider, lineWidth: 1))

            if case .signedOut(let message) = auth.state, let message {
                Text(message)
                    .font(.system(size: 12))
                    .foregroundStyle(Color.homieError)
                    .multilineTextAlignment(.center)
                    .fixedSize(horizontal: false, vertical: true)
                    .frame(width: 340)
                    .padding(.top, 12)
            }

            Button(action: signIn) {
                HStack(spacing: 6) {
                    if auth.isSigningIn { ProgressView().controlSize(.small) }
                    Text(auth.isSigningIn ? "Signing in…" : "Sign in")
                }
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(Color.homieButtonText)
                .frame(width: 340, height: 40)
                .background(
                    LinearGradient(
                        colors: [Color.homieGreenLight, Color.homieGreen],
                        startPoint: .top,
                        endPoint: .bottom
                    ),
                    in: RoundedRectangle(cornerRadius: 12)
                )
            }
            .buttonStyle(.plain)
            .disabled(auth.isSigningIn)
            .opacity(auth.isSigningIn ? 0.7 : 1)
            .padding(.top, 14)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .onAppear { focusedField = .username }
        .task {
            while !Task.isCancelled {
                do { try await Task.sleep(for: .seconds(3.5)) } catch { return }
                isBlinking = true
                do { try await Task.sleep(for: .milliseconds(150)) } catch { return }
                isBlinking = false
            }
        }
    }

    private func signIn() {
        Task { await auth.signIn(username: username, password: password) }
    }
}
