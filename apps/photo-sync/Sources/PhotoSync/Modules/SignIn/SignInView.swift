import SwiftUI

struct SignInView: View {
    let auth: AuthModel
    let hostStore: HomieHostStore
    @State private var username = ""
    @State private var password = ""
    @FocusState private var focusedField: Field?

    private enum Field { case username, password }

    private static let contentWidth: CGFloat = 340

    var body: some View {
        VStack(spacing: 0) {
            HomieMark(blinks: true)
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
            .frame(width: Self.contentWidth)
            .background(Color.homieDivider, in: RoundedRectangle(cornerRadius: 12))
            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.homieDivider, lineWidth: 1))

            if case .signedOut(let message) = auth.state, let message {
                Text(message)
                    .font(.system(size: 12))
                    .foregroundStyle(Color.homieError)
                    .multilineTextAlignment(.center)
                    .fixedSize(horizontal: false, vertical: true)
                    .frame(width: Self.contentWidth)
                    .padding(.top, 12)
            }

            Button(action: signIn) {
                HStack(spacing: 6) {
                    if auth.isSigningIn { ProgressView().controlSize(.small) }
                    Text(auth.isSigningIn ? "Signing in…" : "Sign in")
                }
            }
            .buttonStyle(HomieButtonStyle(prominent: true, fullWidth: true))
            .disabled(auth.isSigningIn)
            .frame(width: Self.contentWidth)
            .padding(.top, 14)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .onAppear { focusedField = .username }
    }

    private func signIn() {
        Task { await auth.signIn(username: username, password: password) }
    }
}
