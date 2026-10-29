import SwiftUI

struct GettingReadyView: View {
    let hostName: String

    private static let contentWidth: CGFloat = 340

    var body: some View {
        VStack(spacing: 0) {
            HomieMark()
                .frame(width: 88, height: 88)
                .padding(14)
                .overlay { SpinnerRing() }

            Text("Getting ready")
                .font(.system(size: 24, weight: .bold))
                .foregroundStyle(.white)
                .padding(.top, 14)
                .padding(.bottom, 22)

            VStack(spacing: 0) {
                ChecklistRow(title: "Started", isDone: true)
                Rectangle().fill(Color.homieDivider).frame(height: 1)
                ChecklistRow(title: "Contacting \(hostName)", isDone: false)
            }
            .frame(width: Self.contentWidth)
            .background(Color.homieDivider, in: RoundedRectangle(cornerRadius: 12))
            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.homieDivider, lineWidth: 1))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Getting ready. Contacting \(hostName).")
    }
}

private struct SpinnerRing: View {
    @State private var spinning = false

    var body: some View {
        Circle()
            .trim(from: 0, to: 0.28)
            .stroke(Color.homieGreen, style: StrokeStyle(lineWidth: 3, lineCap: .round))
            .rotationEffect(.degrees(spinning ? 360 : 0))
            .animation(.linear(duration: 1).repeatForever(autoreverses: false), value: spinning)
            .onAppear { spinning = true }
    }
}

private struct ChecklistRow: View {
    let title: String
    let isDone: Bool

    var body: some View {
        HStack(spacing: 10) {
            ZStack {
                if isDone {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 16))
                        .foregroundStyle(Color.homieGreen)
                } else {
                    ProgressView().controlSize(.small)
                }
            }
            .frame(width: 18, height: 18)
            Text(title)
                .font(.system(size: 14))
                .foregroundStyle(Color.homieText)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 14)
        .frame(height: 44)
    }
}
