import SwiftUI

private struct ShimmerModifier: ViewModifier {
    @State private var isAnimating = false

    func body(content: Content) -> some View {
        content
            .opacity(isAnimating ? 0.4 : 1)
            .animation(.easeInOut(duration: 0.9).repeatForever(autoreverses: true), value: isAnimating)
            .onAppear { isAnimating = true }
    }
}

extension View {
    func shimmering() -> some View {
        modifier(ShimmerModifier())
    }
}

/// Applies the shimmer only while `isActive`, for views reused across loading
/// and loaded states (e.g. a placeholder color that is also the real fallback).
struct ShimmerIf: ViewModifier {
    let isActive: Bool

    init(_ isActive: Bool) {
        self.isActive = isActive
    }

    func body(content: Content) -> some View {
        if isActive {
            content.shimmering()
        } else {
            content
        }
    }
}

/// A placeholder bar for a line of loading data, sized by its frame.
struct SkeletonBlock: View {
    var cornerRadius: CGFloat = 6

    var body: some View {
        RoundedRectangle(cornerRadius: cornerRadius)
            .fill(Color(uiColor: .tertiarySystemFill))
            .shimmering()
    }
}
