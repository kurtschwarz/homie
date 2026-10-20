//
//  HomieNoSignal.swift
//  Homie
//
//  The "can't reach Homie" glitch state from the design board, in SwiftUI:
//  Homie splits into red and cyan layers, slices of him jump sideways, everything flickers
//  like a bad TV, and "NO SIGNAL" glitches along with him. No image assets.
//
//  Three pieces, use whichever you need:
//
//      HomieNoSignalCard(detail: "Homie dropped off the Wi-Fi.\nTrying again in 8 s…") {
//          reconnect()
//      }
//
//      HomieGlitchAvatar()              // just the glitching ball
//          .frame(width: 132, height: 132)
//
//      GlitchText("NO SIGNAL", size: 30)
//
//  Timings and offsets match the CSS on the board (same keyframes, held between steps
//  like `steps(1, end)`), so it flickers in the same rhythm.
//  Requires iOS 15 / macOS 12 (Canvas + TimelineView).
//  Respects Reduce Motion: you get the split look standing still, with no flicker.
//

import SwiftUI

// MARK: - Card

public struct HomieNoSignalCard: View {
    public var title: String
    public var detail: String
    public var buttonTitle: String
    public var onRetry: () async -> Void

    @State private var isReconnecting = false

    public init(
        title: String = "NO SIGNAL",
        detail: String = "We can’t reach Homie right now, but we’ll keep trying. You can still add events and reminders in the meantime.",
        buttonTitle: String = "Try now",
        onRetry: @escaping () async -> Void
    ) {
        self.title = title
        self.detail = detail
        self.buttonTitle = buttonTitle
        self.onRetry = onRetry
    }

    public var body: some View {
        VStack(spacing: 8) {
            HomieGlitchAvatar()
                .frame(width: 132, height: 132)
                .padding(.bottom, -22)
            GlitchText(title, size: 30)
                .padding(.bottom, 24)
            Text(detail)
                .font(.system(size: 14))
                .foregroundColor(.white.opacity(0.7))
                .multilineTextAlignment(.center)
                .lineSpacing(3)
            HomieButton(title: isReconnecting ? "Reconnecting…" : buttonTitle) {
                guard !isReconnecting else { return }
                Task {
                    isReconnecting = true
                    let startedAt = Date()
                    await onRetry()
                    let remaining = 0.8 - Date().timeIntervalSince(startedAt)
                    if remaining > 0 {
                        try? await Task.sleep(for: .seconds(remaining))
                    }
                    isReconnecting = false
                }
            } icon: {
                if isReconnecting {
                    SpinningNotch()
                } else {
                    Image("phosphor-arrow-clockwise")
                        .resizable()
                        .renderingMode(.template)
                }
            }
            .disabled(isReconnecting)
            .padding(.top, 12)
        }
        .padding(.horizontal, 20)
        .padding(.top, 24)
        .padding(.bottom, 24)
        .frame(maxWidth: .infinity)
        .background(
            ZStack {
                RadialGradient(
                    colors: [Color(glitchHex: 0x1B1E2A), Color(glitchHex: 0x0B0C11)],
                    center: UnitPoint(x: 0.5, y: 0.3), startRadius: 0, endRadius: 320
                )
                Streaks()
            }
        )
        .overlay(Scanlines().allowsHitTesting(false))
        .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
        .shadow(color: .black.opacity(0.35), radius: 13, x: 0, y: 10)
        .accessibilityElement(children: .combine)
        .accessibilityLabel("No signal. \(detail)")
    }
}

/// The retry spinner. It starts turning when it appears, so the icon swap
/// lands first and the spin runs from the notch's resting position.
private struct SpinningNotch: View {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var isSpinning = false

    var body: some View {
        Image("phosphor-circle-notch")
            .resizable()
            .renderingMode(.template)
            .frame(width: 16, height: 16)
            .rotationEffect(.degrees(isSpinning ? 360 : 0))
            .onAppear {
                guard !reduceMotion else { return }
                withAnimation(.linear(duration: 0.8).repeatForever(autoreverses: false)) {
                    isSpinning = true
                }
            }
    }
}

// MARK: - Avatar

/// Homie's ball with the "lost" face, split into red and green-blue layers that drift apart,
/// two slices that jump sideways, plus flicker and jitter.
public struct HomieGlitchAvatar: View {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    public init() {}

    public var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 30.0, paused: reduceMotion)) { timeline in
            let t = reduceMotion ? 0 : timeline.date.timeIntervalSinceReferenceDate
            Canvas { ctx, size in
                GlitchRenderer(time: t, size: size).draw(in: &ctx)
            }
            .opacity(Glitch.flicker(t))
        }
        .aspectRatio(1, contentMode: .fit)
        .accessibilityHidden(true)
    }
}

private struct GlitchRenderer {
    let time: TimeInterval
    let size: CGSize

    func draw(in ctx: inout GraphicsContext) {
        let side = min(size.width, size.height)
        let px = side / 132            // the board's offsets are in px at 132 pt
        let jitter = Glitch.jitter(time)
        let scale = side / 156          // leave room for the displaced glitch slices
        ctx.translateBy(
            x: (size.width - side) / 2 + side / 2 - 50 * scale + jitter.width * px,
            y: (size.height - side) / 2 + side / 2 - 50 * scale + jitter.height * px
        )
        ctx.scaleBy(x: scale, y: scale)
        let unit = px / scale           // preserve board-pixel offsets at the smaller drawing scale

        // Red-only and green/blue-only copies, added together with screen blending.
        // Where they line up you get Homie's normal colours; where they drift, red and cyan fringes.
        let r = Glitch.redOffset(time)
        var red = ctx
        red.translateBy(x: r.width * unit, y: r.height * unit)
        HomieFace.draw(in: red, tint: .redOnly)

        var cyan = ctx
        cyan.blendMode = .screen
        cyan.translateBy(x: -r.width * unit, y: -r.height * unit)
        HomieFace.draw(in: cyan, tint: .greenBlueOnly)

        // Slices: horizontal bands of Homie, shifted sideways.
        for (slice, tint) in [(Glitch.sliceA(time), HomieFace.Tint.teal), (Glitch.sliceB(time), HomieFace.Tint.bright)] {
            var band = ctx
            let top = slice.top, bottom = slice.bottom
            band.clip(to: Path(CGRect(x: -50, y: top, width: 200, height: max(0, 100 - top - bottom))))
            band.translateBy(x: slice.x * unit, y: 0)
            HomieFace.draw(in: band, tint: tint)
        }
    }
}

/// Homie in a 100 × 100 design space with the "lost" face: small dot eyes and a little frown.
private enum HomieFace {
    enum Tint { case normal, redOnly, greenBlueOnly, teal, bright }

    struct RGB { var r, g, b: Double }

    static func color(_ c: RGB, _ tint: Tint, opacity: Double = 1) -> Color {
        let m: RGB
        switch tint {
        case .normal: m = c
        case .redOnly: m = RGB(r: c.r, g: 0, b: 0)
        case .greenBlueOnly: m = RGB(r: 0, g: min(1, c.g + 0.04), b: min(1, c.b + 0.12))
        case .teal: m = RGB(r: c.b * 0.5, g: min(1, c.r * 0.9), b: min(1, c.r))   // roughly a 160° hue turn
        case .bright: m = RGB(r: min(1, c.r * 1.4), g: min(1, c.g * 1.4), b: min(1, c.b * 1.4))
        }
        return Color(.sRGB, red: m.r, green: m.g, blue: m.b, opacity: opacity)
    }

    static let warm = [RGB(r: 0.831, g: 0.373, b: 0.157), RGB(r: 0.878, g: 0.424, b: 0.247), RGB(r: 0.933, g: 0.541, b: 0.639)]
    static let white = RGB(r: 1, g: 1, b: 1)
    static let edge = RGB(r: 0.122, g: 0.035, b: 0)

    static func draw(in ctx: GraphicsContext, tint: Tint) {
        let ball = Path(ellipseIn: CGRect(x: 2, y: 2, width: 96, height: 96))
        ctx.fill(ball, with: .linearGradient(
            Gradient(stops: [
                .init(color: color(warm[0], tint), location: 0),
                .init(color: color(warm[1], tint), location: 0.55),
                .init(color: color(warm[2], tint), location: 1)
            ]),
            startPoint: CGPoint(x: 2 + 0.15 * 96, y: 2 + 0.08 * 96),
            endPoint: CGPoint(x: 2 + 0.9 * 96, y: 2 + 0.95 * 96)
        ))

        var shade = ctx
        shade.clip(to: ball)
        shade.fill(Path(CGRect(x: -40, y: -40, width: 180, height: 180)), with: .radialGradient(
            Gradient(stops: [
                .init(color: color(white, tint, opacity: 0.16), location: 0),
                .init(color: color(white, tint, opacity: 0), location: 0.42),
                .init(color: color(edge, tint, opacity: 0), location: 0.8),
                .init(color: color(edge, tint, opacity: 0.34), location: 0.93),
                .init(color: color(edge, tint, opacity: 0.62), location: 1)
            ]),
            center: CGPoint(x: 61, y: 45), startRadius: 0, endRadius: 60
        ))

        let ink = color(white, tint)
        var eyes = Path()
        eyes.addEllipse(in: CGRect(x: 50 - 3.6, y: 40 - 5.4, width: 7.2, height: 10.8))
        eyes.addEllipse(in: CGRect(x: 68 - 3.2, y: 34 - 4.8, width: 6.4, height: 9.6))
        ctx.fill(eyes, with: .color(ink))

        var mouth = Path()
        mouth.move(to: CGPoint(x: 48, y: 62))
        mouth.addQuadCurve(to: CGPoint(x: 72, y: 62), control: CGPoint(x: 60, y: 55))
        ctx.stroke(mouth, with: .color(ink), style: StrokeStyle(lineWidth: 4.6, lineCap: .round, lineJoin: .round))
    }
}

// MARK: - Glitch text

/// Text with cyan and pink copies offset either side, a jitter, and a slice that jumps.
public struct GlitchText: View {
    public var text: String
    public var size: CGFloat
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    public init(_ text: String, size: CGFloat = 30) {
        self.text = text
        self.size = size
    }

    public var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 30.0, paused: reduceMotion)) { timeline in
            let t = reduceMotion ? 0 : timeline.date.timeIntervalSinceReferenceDate
            let r = Glitch.redOffset(t), j = Glitch.jitter(t), slice = Glitch.sliceA(t)
            ZStack {
                label.foregroundColor(Color(glitchHex: 0x00F0FF)).offset(x: -r.width, y: -r.height)
                label.foregroundColor(Color(glitchHex: 0xFF2D6F)).offset(x: r.width, y: r.height).blendMode(.screen)
                label.foregroundColor(.white)
                label.foregroundColor(.white)
                    .offset(x: slice.x)
                    // The mask sits on the text's layout frame, so the band stays put while the slice slides.
                    .mask(
                        GeometryReader { geo in
                            Rectangle()
                                .frame(height: max(0, geo.size.height * (100 - slice.top - slice.bottom) / 100))
                                .offset(y: geo.size.height * slice.top / 100)
                        }
                    )
            }
            .compositingGroup()
            .offset(x: j.width, y: j.height)
        }
        .accessibilityLabel(text)
    }

    private var label: some View {
        Text(text)
            .font(.system(size: size, weight: .bold))
            .kerning(size * 0.06)
            .lineLimit(1)
            .fixedSize()
    }
}

// MARK: - Background pieces

private struct Scanlines: View {
    var body: some View {
        Canvas { ctx, size in
            var y: CGFloat = 0
            while y < size.height {
                ctx.fill(Path(CGRect(x: 0, y: y, width: size.width, height: 1)), with: .color(.white.opacity(0.05)))
                y += 3
            }
        }
    }
}

private struct Streaks: View {
    // (left, width as a fraction of the card, top in pt, height, colour)
    private let bars: [(CGFloat, CGFloat, CGFloat, CGFloat, Color)] = [
        (0.04, 0.38, 46, 3, Color(glitchHex: 0x00F0FF).opacity(0.5)),
        (0.58, 0.36, 112, 2, Color(glitchHex: 0xFF2D6F).opacity(0.55)),
        (0.10, 0.30, 168, 4, Color.white.opacity(0.18)),
        (0.62, 0.30, 196, 2, Color(glitchHex: 0x00F0FF).opacity(0.4))
    ]

    var body: some View {
        GeometryReader { geo in
            ForEach(0..<bars.count, id: \.self) { i in
                let b = bars[i]
                LinearGradient(stops: [
                    .init(color: .clear, location: 0), .init(color: b.4, location: 0.3),
                    .init(color: b.4, location: 0.7), .init(color: .clear, location: 1)
                ], startPoint: .leading, endPoint: .trailing)
                .frame(width: geo.size.width * b.1, height: b.3)
                .position(x: geo.size.width * (b.0 + b.1 / 2), y: b.2 + b.3 / 2)
            }
        }
        .allowsHitTesting(false)
    }
}

// MARK: - Timing

/// The board's CSS keyframes, held between steps like `steps(1, end)`.
private enum Glitch {
    static func step<T>(_ time: TimeInterval, period: Double, _ keys: [(Double, T)]) -> T {
        let phase = (time / period).truncatingRemainder(dividingBy: 1)
        var value = keys[0].1
        for (at, v) in keys where phase >= at { value = v }
        return value
    }

    static func flicker(_ t: TimeInterval) -> Double {
        step(t, period: 2.4, [(0, 1), (0.06, 0.55), (0.07, 1), (0.33, 1), (0.34, 0.25), (0.35, 0.9), (0.36, 1), (0.72, 0.7), (0.73, 1)])
    }

    static func jitter(_ t: TimeInterval) -> CGSize {
        step(t, period: 3.1, [
            (0, CGSize(width: 0, height: 0)), (0.08, CGSize(width: -3, height: 1)), (0.09, CGSize(width: 2, height: -1)),
            (0.10, CGSize(width: 0, height: 0)), (0.59, CGSize(width: 4, height: 0)), (0.60, CGSize(width: -2, height: 1)),
            (0.61, CGSize(width: 0, height: 0))
        ])
    }

    /// How far the red copy sits to the right; the cyan copy mirrors it to the left.
    static func redOffset(_ t: TimeInterval) -> CGSize {
        step(t, period: 2.2, [
            (0, CGSize(width: 5, height: 0)), (0.30, CGSize(width: 7, height: -1)), (0.31, CGSize(width: 3, height: 1)),
            (0.64, CGSize(width: 9, height: 0)), (0.65, CGSize(width: 5, height: 0))
        ])
    }

    struct Slice { var top: CGFloat; var bottom: CGFloat; var x: CGFloat }   // top/bottom insets in %, x in px

    static func sliceA(_ t: TimeInterval) -> Slice {
        step(t, period: 1.7, [
            (0, Slice(top: 28, bottom: 62, x: 0)), (0.12, Slice(top: 28, bottom: 62, x: -14)), (0.13, Slice(top: 58, bottom: 30, x: 16)),
            (0.40, Slice(top: 58, bottom: 30, x: 0)), (0.41, Slice(top: 12, bottom: 80, x: -10)), (0.70, Slice(top: 70, bottom: 20, x: 20)),
            (0.71, Slice(top: 28, bottom: 62, x: 0))
        ])
    }

    static func sliceB(_ t: TimeInterval) -> Slice {
        step(t, period: 2.9, [
            (0, Slice(top: 45, bottom: 47, x: 8)), (0.25, Slice(top: 80, bottom: 10, x: -18)),
            (0.26, Slice(top: 40, bottom: 52, x: 12)), (0.55, Slice(top: 18, bottom: 74, x: -6))
        ])
    }
}

private extension Color {
    init(glitchHex hex: UInt32) {
        self.init(.sRGB, red: Double((hex >> 16) & 0xFF) / 255, green: Double((hex >> 8) & 0xFF) / 255, blue: Double(hex & 0xFF) / 255, opacity: 1)
    }
}

// MARK: - Preview

#if DEBUG
struct HomieNoSignal_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: 24) {
            HomieNoSignalCard(detail: "Homie dropped off the Wi-Fi.\nTrying again in 8 s…") {}
            HStack(spacing: 24) {
                HomieGlitchAvatar().frame(width: 64, height: 64)
                GlitchText("NO SIGNAL", size: 22)
            }
            .padding()
            .background(Color.black)
            .clipShape(RoundedRectangle(cornerRadius: 16))
        }
        .padding(20)
        .background(Color(glitchHex: 0xF4F6F8))
    }
}
#endif
