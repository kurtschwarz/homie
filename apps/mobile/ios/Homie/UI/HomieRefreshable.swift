import SwiftUI
import UIKit

/// The row held open at the top of the scroll content, as an animatable height.
///
/// `contentMargins` takes a plain value and does not interpolate towards a new
/// one, so animating the state behind it does nothing. Declaring the height as
/// `animatableData` has SwiftUI re-run this body on every frame of the animation
/// instead, applying a fresh flat margin each time.
private struct ScrollRowMargin: ViewModifier, Animatable {
    var height: CGFloat

    var animatableData: CGFloat {
        get { height }
        set { height = newValue }
    }

    func body(content: Content) -> some View {
        content.contentMargins(.top, height, for: .scrollContent)
    }
}

private struct HomieRefreshableModifier: ViewModifier {
    let action: () async -> Void

    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var chatAvatarFrame: CGRect = .zero
    @State private var cardFrame: CGRect = .zero
    @State private var headerFrame: CGRect = .zero
    @State private var measuredAnchors: Set<HomieRefreshAnchor> = []
    @State private var pull: CGFloat = 0
    @State private var isArmed = false
    @State private var isRefreshing = false
    @State private var isInteracting = false
    @State private var isReturning = false
    /// Set when he sets off home and held for a beat after he lands.
    @State private var isDizzy = false
    /// Height of the row held open for the indicator, as a scroll content margin.
    /// Opened flat and animated shut through `ScrollRowMargin`.
    @State private var rowHeight: CGFloat = 0
    /// How far the indicator is pulled into view, in points. Separate from
    /// `rowHeight` because the drag reveals the indicator over content the row
    /// is not holding open yet.
    @State private var indicatorReveal: CGFloat = 0
    @State private var travelProgress: CGFloat = 0
    /// Pull beyond the threshold, where he has nowhere left to travel.
    @State private var overscroll: CGFloat = 0
    @State private var rotation = 0.0

    private let threshold: CGFloat = 80
    /// Travel is spread over more pull than arming needs, so the journey reads at
    /// a watchable pace rather than finishing in the first inch of the drag.
    /// Whatever is left when the finger lifts is animated out in `refresh()`.
    private let travelDistance: CGFloat = 140
    private let indicatorHeight: CGFloat = 112

    func body(content: Content) -> some View {
        ZStack(alignment: .top) {
            content
                // Reserves the row while loading. This has to be a content margin
                // rather than a safe area inset: a margin moves the scroll view's
                // own resting position, so releasing an overscroll decelerates
                // into the open row in one motion. An inset only pushes the
                // content, leaving the scroll view to snap back to the top first
                // and the row to shove it down again afterwards.
                .modifier(ScrollRowMargin(height: rowHeight))
                .onScrollGeometryChange(for: CGFloat.self) { geometry in
                    max(0, -(geometry.contentOffset.y + geometry.contentInsets.top))
                } action: { _, distance in
                    // Travel is driven by the same value as the indicator, including
                    // through the release bounce. Letting these two diverge puts the
                    // travelling Homie and the indicator's slot out of step.
                    guard !isRefreshing, !isReturning else { return }
                    pull = distance
                    travelProgress = min(1, max(0, distance / travelDistance))
                    // Rolling on the spot only starts once he has actually
                    // arrived, so it never doubles up with the journey.
                    overscroll = max(0, distance - travelDistance)
                    guard isInteracting else { return }
                    let wasArmed = isArmed
                    isArmed = distance >= threshold
                    if isArmed && !wasArmed {
                        UIImpactFeedbackGenerator(style: .medium).impactOccurred()
                    }
                }
                .onScrollPhaseChange { oldPhase, newPhase, _ in
                    if newPhase == .interacting {
                        isInteracting = true
                        isArmed = false
                        return
                    }
                    isInteracting = false
                    guard newPhase == .decelerating || newPhase == .idle else { return }
                    guard oldPhase != newPhase else { return }
                    // Not armed: the bounce carries `pull` back to zero, and travel
                    // follows it, so Homie rides home with the content.
                    guard isArmed, !isRefreshing else {
                        isArmed = false
                        return
                    }
                    Task { await refresh() }
                }

            // Clipped on its own so the indicator stays hidden while parked above
            // the top edge, without that clip also cutting the travelling Homie.
            Color.clear
                .overlay(alignment: .top) {
                    HomieRefreshIndicator(
                        isArmed: isArmed,
                        isRefreshing: isRefreshing,
                        rotation: rotation,
                        hidesAvatar: shouldShowTravelingAvatar
                    )
                    .frame(height: indicatorHeight)
                    .offset(y: -indicatorHeight + revealedHeight)
                }
                .clipped()
                .accessibilityHidden(!isRefreshing && pull == 0)

            if shouldShowTravelingAvatar {
                travelingAvatar
            }
        }
        // This stack owns the travel coordinate space and is the direct parent of
        // everything measured in it, so `.position` lands where the anchors were
        // read. Declaring the space elsewhere let the scroll content and the
        // indicator resolve against different origins.
        .coordinateSpace(.named(HomieRefresh.coordinateSpace))
        .onPreferenceChange(HomieRefreshAnchorPreferenceKey.self) { anchors in
            // Measurements, not motion. While the row closes, the card's avatar
            // climbs with the content and these are re-read every frame; letting
            // them inherit that animation would ease the parked Homie towards
            // each reading and leave him trailing the card he is sitting on.
            var measure = Transaction()
            measure.disablesAnimations = true
            withTransaction(measure) {
                if let frame = anchors[.chatAvatar], !frame.isEmpty {
                    chatAvatarFrame = frame
                    measuredAnchors.insert(.chatAvatar)
                }
                if let frame = anchors[.chatCard], !frame.isEmpty {
                    cardFrame = frame
                    measuredAnchors.insert(.chatCard)
                }
                if let frame = anchors[.header], !frame.isEmpty {
                    headerFrame = frame
                    measuredAnchors.insert(.header)
                }
            }
        }
        .environment(\.homieRefreshSlot, slot)
    }

    @MainActor
    private func refresh() async {
        pull = 0
        isArmed = false
        isReturning = false
        // Set without animation: the margin is the scroll view's new resting
        // position, and it has to be in place before the release bounce resolves
        // or the content settles at the old top and jumps afterwards.
        isRefreshing = true
        rowHeight = indicatorHeight
        indicatorReveal = indicatorHeight
        // Arming happens before the journey is finished, so he covers the rest
        // of the way himself instead of jumping to the slot.
        if reduceMotion {
            travelProgress = 1
        } else {
            withAnimation(.easeOut(duration: 0.4)) {
                travelProgress = 1
            }
        }
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
        // The spin leaves him unbothered; it is the roll home that makes him
        // dizzy, and it lingers after he lands so it reads as an after-effect.
        //
        // `overscroll` unwinds with the journey home; resetting it afterwards
        // would snap away however far the extra pull had turned him.
        isReturning = true
        isDizzy = true
        // He sets off home and the row collapses under him at the same time, on
        // one curve: waiting for him to land first left an empty row sitting
        // open for the whole trip.
        //
        // The row shrinking is what carries the content up: the scroll view
        // rests against the margin, so every point the margin gives back is a
        // point the content climbs. The indicator retracts by the same amount,
        // so it rides out inside its own row rather than over the cards. Both
        // ends of his journey are measured live, so his path rises with them and
        // he still lands on the card's slot.
        let home = reduceMotion ? 0.22 : 0.65
        withAnimation(.easeInOut(duration: home)) {
            travelProgress = 0
            overscroll = 0
            rowHeight = 0
            indicatorReveal = 0
        }
        try? await Task.sleep(for: .milliseconds(Int(home * 1000) + 60))
        pull = 0
        // Cleared only now: it keeps the label reading "Syncing…" until the row
        // is gone, rather than flipping back to "Pull to refresh" on its way out.
        isRefreshing = false
        isReturning = false
        rotation = 0
        try? await Task.sleep(for: .seconds(2))
        isDizzy = false
    }

    /// Driven by the drag while pulling, then by `indicatorReveal`, so the
    /// indicator and the content it displaces always move as one.
    private var revealedHeight: CGFloat {
        max(min(pull, indicatorHeight), indicatorReveal)
    }

    private var slot: HomieRefreshSlot {
        guard shouldShowTravelingAvatar else { return .occupied }
        // The trip home ends when `isReturning` clears; `isDizzy` outlives it by
        // the beat he spends parked on the slot, so the two together are the only
        // stretch where he is back but still drawn as the travelling copy.
        return isDizzy && !isReturning ? .covered : .vacant
    }

    private var shouldShowTravelingAvatar: Bool {
        // `isDizzy` keeps him on screen after he lands. At zero progress he sits
        // exactly over the card's slot at its size, so holding him there is
        // indistinguishable from the card's own avatar, which stays hidden.
        !reduceMotion && (travelProgress > 0 || isReturning || isRefreshing || isDizzy)
            && measuredAnchors.contains(.chatAvatar)
            && measuredAnchors.contains(.header)
    }

    private var travelingAvatar: some View {
        let progress = min(1, max(0, travelProgress))
        // Both anchors are measured in the same space, so these interpolate
        // directly. `shouldShowTravelingAvatar` guarantees both are non-empty.
        let origin = CGPoint(x: chatAvatarFrame.midX, y: chatAvatarFrame.midY)
        let target = CGPoint(x: headerFrame.midX, y: headerFrame.midY)
        // Ends at exactly the slot's size, so the handoff back to the indicator's
        // own avatar is invisible.
        let size = chatAvatarFrame.width + (headerFrame.width - chatAvatarFrame.width) * progress
        let pose = self.pose(from: origin, to: target, size: size, progress: progress)
        // He only spins once he is rolling right; the climb leaves him untouched.
        // A whole turn both matches the roll over that distance and leaves him
        // upright at the slot rather than parked on his side.
        //
        // Once he is parked, further pull has nowhere to carry him, so it rolls
        // him on the spot at the rate a ball his size would turn over it.
        let inPlace = size > 0 ? Double(overscroll / (.pi * size)) * 360 : 0
        // The spin belongs to the wait, not the trip home, where the unwinding
        // roll takes over.
        let spin = isRefreshing && !isReturning ? rotation : 0
        let degrees = Double(pose.roll) * 360 + inPlace + spin

        // He carries his resting face out, across, and through the spin; the
        // dizziness only catches up with him on the way home.
        return HomieAvatar(mood: isDizzy ? .dizzy : .idle, showsGroundShadow: false)
            .frame(width: size, height: size)
            .rotationEffect(.degrees(degrees))
            .position(pose.position)
            .allowsHitTesting(false)
            .accessibilityHidden(true)
    }

    private struct TravelPose {
        let position: CGPoint
        /// Zero for the whole climb, then 0...1 once he turns and rolls right.
        /// Rotation and mood hang off this, so he rides up unchanged.
        let roll: CGFloat
    }

    /// Homie climbs the card's left edge until his bottom clears the card's top,
    /// then turns and rolls right to the indicator. The split is weighted by each
    /// leg's length so he keeps one speed through the corner.
    private func pose(from origin: CGPoint, to target: CGPoint, size: CGFloat, progress: CGFloat) -> TravelPose {
        // The turn happens where his bottom edge is level with the card's top.
        // Never above the target, or the second leg would double back down.
        let corner = cardFrame.isEmpty
            ? CGPoint(x: origin.x, y: target.y)
            : CGPoint(x: origin.x, y: max(target.y, cardFrame.minY - size / 2))
        let rise = abs(corner.y - origin.y)
        let run = hypot(target.x - corner.x, target.y - corner.y)
        guard rise + run > 0 else { return TravelPose(position: target, roll: 1) }
        let travelled = (rise + run) * progress
        if travelled <= rise {
            let leg = rise == 0 ? 1 : travelled / rise
            return TravelPose(
                position: CGPoint(x: origin.x, y: origin.y + (corner.y - origin.y) * leg),
                roll: 0
            )
        }
        let leg = run == 0 ? 1 : (travelled - rise) / run
        return TravelPose(
            position: CGPoint(
                x: corner.x + (target.x - corner.x) * leg,
                y: corner.y + (target.y - corner.y) * leg
            ),
            roll: leg
        )
    }

}

struct HomieRefreshAnchorPreferenceKey: PreferenceKey {
    static let defaultValue: [HomieRefreshAnchor: CGRect] = [:]

    static func reduce(value: inout [HomieRefreshAnchor: CGRect], nextValue: () -> [HomieRefreshAnchor: CGRect]) {
        value.merge(nextValue(), uniquingKeysWith: { _, new in new })
    }
}

/// What the slot Homie normally sits in should draw. The avatar itself stays
/// hidden for as long as the travelling copy is on screen, but the outline only
/// stands in while he is actually away from it.
enum HomieRefreshSlot {
    /// He is in it.
    case occupied
    /// He has left: mark the spot.
    case vacant
    /// He has landed back on it, still drawn as the travelling copy. Nothing to
    /// mark, and an outline here would show through him.
    case covered
}

private struct HomieRefreshSlotKey: EnvironmentKey {
    static let defaultValue = HomieRefreshSlot.occupied
}

extension EnvironmentValues {
    var homieRefreshSlot: HomieRefreshSlot {
        get { self[HomieRefreshSlotKey.self] }
        set { self[HomieRefreshSlotKey.self] = newValue }
    }
}

private struct HomieRefreshVisibilityModifier: ViewModifier {
    @Environment(\.homieRefreshSlot) private var slot

    func body(content: Content) -> some View {
        content
            .opacity(slot == .occupied ? 1 : 0)
            // An overlay, so the placeholder inherits the avatar's exact frame and
            // nothing in the layout moves on the swap.
            //
            // Swapped outright rather than faded: at both ends of his journey the
            // travelling Homie is sitting exactly on this slot, so the cut happens
            // underneath him and is never seen. Fading instead shows the two of
            // them at once while he sets off, which reads as a ghost.
            .overlay { if slot == .vacant { HomieAvatarPlaceholder() } }
    }
}

/// The one coordinate space every travel anchor is measured in. It is declared by
/// `homieRefreshable` on the scroll view; nothing else may declare it, or anchors
/// inside the scroll content end up measured against a different origin.
enum HomieRefresh {
    static let coordinateSpace = "HomieRefreshTravel"
}

extension View {
    func homieRefreshAnchor(_ anchor: HomieRefreshAnchor) -> some View {
        background {
            GeometryReader { proxy in
                Color.clear.preference(
                    key: HomieRefreshAnchorPreferenceKey.self,
                    value: [anchor: proxy.frame(in: .named(HomieRefresh.coordinateSpace))]
                )
            }
        }
    }

    func hidesWhileHomieTravels() -> some View {
        modifier(HomieRefreshVisibilityModifier())
    }
}

private struct HomieRefreshIndicator: View {
    let isArmed: Bool
    let isRefreshing: Bool
    let rotation: Double
    /// True while the travelling Homie is on screen, so only one is ever drawn.
    let hidesAvatar: Bool

    var body: some View {
        VStack(spacing: 12) {
            // The slot keeps its size either way, so it is the anchor the
            // travelling Homie aims at and the layout never shifts on handoff.
            Color.clear
                .frame(width: 40, height: 40)
                .overlay {
                    if !hidesAvatar {
                        HomieAvatar(mood: .idle, showsGroundShadow: false)
                            .rotationEffect(.degrees(isRefreshing ? rotation : 0))
                    }
                }
                .homieRefreshAnchor(.header)
            Text(isRefreshing ? "Syncing with Homie…" : isArmed ? "Let go to refresh" : "Pull to refresh")
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding(.top, 8)
        .accessibilityElement(children: .combine)
    }
}

enum HomieRefreshAnchor: Hashable {
    case chatAvatar
    case chatCard
    case header
}

extension View {
    func homieRefreshable(action: @escaping () async -> Void) -> some View {
        modifier(HomieRefreshableModifier(action: action))
    }
}
