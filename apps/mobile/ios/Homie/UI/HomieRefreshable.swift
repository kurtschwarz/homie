import SwiftUI
import UIKit

private struct HomieRefreshableModifier: ViewModifier {
    let action: () async -> Void

    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var pull: CGFloat = 0
    @State private var isArmed = false
    @State private var isRefreshing = false
    @State private var isInteracting = false
    @State private var rotation = 0.0

    private let threshold: CGFloat = 80
    private let indicatorHeight: CGFloat = 112

    func body(content: Content) -> some View {
        content
            .overlay(alignment: .top) {
                HomieRefreshIndicator(
                    pull: pull,
                    threshold: threshold,
                    isArmed: isArmed,
                    isRefreshing: isRefreshing,
                    rotation: rotation
                )
                .frame(height: indicatorHeight)
                .offset(y: -indicatorHeight + (isRefreshing ? indicatorHeight : min(pull, indicatorHeight)))
                .clipped()
                .opacity(isRefreshing ? 0 : 1)
                .accessibilityHidden(!isRefreshing && pull == 0)
            }
            .safeAreaInset(edge: .top, spacing: 0) {
                if isRefreshing {
                    HomieRefreshIndicator(
                        pull: pull,
                        threshold: threshold,
                        isArmed: false,
                        isRefreshing: true,
                        rotation: rotation
                    )
                    .frame(height: indicatorHeight)
                    .transition(.move(edge: .top).combined(with: .opacity))
                }
            }
            .onScrollGeometryChange(for: CGFloat.self) { geometry in
                max(0, -(geometry.contentOffset.y + geometry.contentInsets.top))
            } action: { _, distance in
                guard !isRefreshing else { return }
                pull = distance
                if isInteracting {
                    let wasArmed = isArmed
                    isArmed = distance >= threshold
                    if isArmed && !wasArmed {
                        UIImpactFeedbackGenerator(style: .medium).impactOccurred()
                    }
                }
            }
            .onScrollPhaseChange { oldPhase, newPhase, _ in
                if newPhase == .interacting {
                    isInteracting = true
                    isArmed = false
                    pull = 0
                    return
                }
                isInteracting = false
                guard newPhase == .decelerating || newPhase == .idle else { return }
                guard oldPhase != newPhase else { return }
                guard isArmed, !isRefreshing else {
                    pull = 0
                    isArmed = false
                    return
                }
                Task { await refresh() }
            }
    }

    @MainActor
    private func refresh() async {
        isRefreshing = true
        pull = 0
        isArmed = false
        if !reduceMotion {
            rotation = 0
            withAnimation(.linear(duration: 0.9).repeatForever(autoreverses: false)) {
                rotation = 360
            }
        }
        let start = ContinuousClock.now
        await action()
        let elapsed = start.duration(to: .now)
        let minimum = Duration.milliseconds(600)
        if elapsed < minimum {
            try? await Task.sleep(for: minimum - elapsed)
        }
        withAnimation(.easeOut(duration: 0.2)) {
            isRefreshing = false
            pull = 0
        }
        rotation = 0
    }
}

private struct HomieRefreshIndicator: View {
    let pull: CGFloat
    let threshold: CGFloat
    let isArmed: Bool
    let isRefreshing: Bool
    let rotation: Double

    var body: some View {
        VStack(spacing: 12) {
            HomieAvatar(mood: isArmed && !isRefreshing ? .idle : .thinking, showsGroundShadow: false)
                .frame(width: 40, height: 40)
                .rotationEffect(.degrees(isRefreshing ? rotation : Double(pull / threshold) * 180))
            Text(isRefreshing ? "Syncing with Homie…" : isArmed ? "Let go to refresh" : "Pull to refresh")
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding(.top, 8)
        .accessibilityElement(children: .combine)
    }
}

extension View {
    func homieRefreshable(action: @escaping () async -> Void) -> some View {
        modifier(HomieRefreshableModifier(action: action))
    }
}
