//
//  HomieAvatar.swift
//  Homie
//
//  Homie, the helper, drawn natively with SwiftUI Canvas.
//  No image assets or third-party libraries: it stays crisp at any size and animates on its own.
//
//  Usage:
//      HomieAvatar()                                   // idle, warm, with ground shadow
//      HomieAvatar(mood: .thinking)
//      HomieAvatar(mood: .sleepy, palette: .night)
//      HomieAvatar(mood: .idle, palette: .blue, wobbles: true)
//          .frame(width: 64, height: 64)
//
//  Sizing: the ball fills 80% of the view's width. The rest is room for the listening pulse,
//  the ground shadow and the sleepy z's, so a 60 pt frame gives you a 48 pt ball.
//
//  Requires iOS 15 / macOS 12 (Canvas + TimelineView). The #Preview at the bottom needs Xcode 15+.
//  Respects Reduce Motion: every animation stops and Homie holds its resting pose.
//

import SwiftUI

public struct HomieAvatar: View {

    public enum Mood: String, CaseIterable, Sendable {
        /// Blinks now and then, glances away and back. The shading rolls with the face.
        case idle
        /// Eyes wide, a little "o", and a soft pulse around the ball.
        case listening
        /// Looks up and away, wiggly mouth, three dots.
        case thinking
        /// Happy eyes, mouth flapping.
        case talking
        /// Night mode: closed eyes and drifting z's.
        case sleepy

        var accessibilityDescription: String {
            switch self {
            case .idle: return "Idle"
            case .listening: return "Listening"
            case .thinking: return "Thinking"
            case .talking: return "Talking"
            case .sleepy: return "Asleep"
            }
        }
    }

    public enum Palette: Sendable {
        /// Orange to pink. Closest to the original reference.
        case warm
        /// Blue to lilac.
        case blue
        /// Near-black.
        case ink
        /// Dimmed warm, for the kitchen screen at night.
        case night
    }

    public var mood: Mood
    public var palette: Palette
    public var showsGroundShadow: Bool
    public var wobbles: Bool
    /// Colour of the sleepy z's. Defaults to the system secondary label colour.
    public var zColor: Color

    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    public init(
        mood: Mood = .idle,
        palette: Palette = .warm,
        showsGroundShadow: Bool = true,
        wobbles: Bool = false,
        zColor: Color = .secondary
    ) {
        self.mood = mood
        self.palette = palette
        self.showsGroundShadow = showsGroundShadow
        self.wobbles = wobbles
        self.zColor = zColor
    }

    public var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 30.0, paused: reduceMotion)) { timeline in
            let renderer = HomieAvatarRenderer(
                mood: mood,
                palette: palette,
                ground: showsGroundShadow,
                wobble: wobbles,
                zColor: zColor,
                time: timeline.date.timeIntervalSinceReferenceDate,
                animated: !reduceMotion
            )
            Canvas { context, size in
                renderer.draw(in: &context, size: size)
            }
        }
        .aspectRatio(1, contentMode: .fit)
    }
}

// MARK: - Renderer

/// Draws Homie in a 100 × 100 design space (ball centred at 50,50 with radius 48), the same
/// coordinates as the SVG on the design board. The canvas shows -10…110 on both axes.
private struct HomieAvatarRenderer {
    let mood: HomieAvatar.Mood
    let palette: HomieAvatar.Palette
    let ground: Bool
    let wobble: Bool
    let zColor: Color
    let time: TimeInterval
    let animated: Bool

    static let viewMin: CGFloat = -10
    static let viewSpan: CGFloat = 120

    private let white = Color.white
    private let lineWidth: CGFloat = 4.6

    func draw(in ctx: inout GraphicsContext, size: CGSize) {
        let s = min(size.width, size.height) / Self.viewSpan
        ctx.translateBy(x: (size.width - Self.viewSpan * s) / 2, y: (size.height - Self.viewSpan * s) / 2)
        ctx.scaleBy(x: s, y: s)
        ctx.translateBy(x: -Self.viewMin, y: -Self.viewMin)

        if wobble && animated {
            let angle = value(period: 3.2, [(0, 0), (0.25, -4), (0.75, 3), (1, 0)])
            ctx.translateBy(x: 50, y: 50)
            ctx.rotate(by: .degrees(angle))
            ctx.translateBy(x: -50, y: -50)
        }

        if ground { drawGround(in: ctx) }
        if mood == .listening && animated { drawPulse(in: ctx) }
        drawBall(in: ctx)
        drawShade(in: ctx)

        switch mood {
        case .idle: drawIdle(in: ctx)
        case .listening: drawListening(in: ctx)
        case .thinking: drawThinking(in: ctx)
        case .talking: drawTalking(in: ctx)
        case .sleepy: drawSleepy(in: ctx)
        }
    }

    // MARK: Body

    private var ball: Path { Path(ellipseIn: CGRect(x: 2, y: 2, width: 96, height: 96)) }

    private func drawGround(in ctx: GraphicsContext) {
        var g = ctx
        g.translateBy(x: 50, y: 103)
        g.scaleBy(x: 1, y: 4.5 / 34)
        g.fill(
            Path(ellipseIn: CGRect(x: -34, y: -34, width: 68, height: 68)),
            with: .radialGradient(
                Gradient(colors: [Color.black.opacity(0.18), Color.black.opacity(0)]),
                center: .zero, startRadius: 0, endRadius: 34
            )
        )
    }

    private func drawPulse(in ctx: GraphicsContext) {
        let raw = (time.truncatingRemainder(dividingBy: 1.6)) / 1.6
        let p = 1 - pow(1 - raw, 2)            // ease-out
        let r = 48 + 18 * p
        var ring = Path(ellipseIn: CGRect(x: 50 - r, y: 50 - r, width: r * 2, height: r * 2))
        ring.addPath(ball)
        ctx.fill(ring, with: .color(Color(hex: 0xE06C3F).opacity(0.45 * (1 - raw))), style: FillStyle(eoFill: true))
    }

    private func drawBall(in ctx: GraphicsContext) {
        // SVG gradient x1 .15 y1 .08 → x2 .9 y2 .95 over the ball's bounding box (2…98).
        ctx.fill(
            ball,
            with: .linearGradient(
                Gradient(stops: palette.stops),
                startPoint: CGPoint(x: 2 + 0.15 * 96, y: 2 + 0.08 * 96),
                endPoint: CGPoint(x: 2 + 0.9 * 96, y: 2 + 0.95 * 96)
            )
        )
    }

    /// Inner shadow: a soft glow centred on the face and a dark band at the far edge.
    /// It sits where the face is looking and, in idle, rolls a bit further than the face moves.
    private func drawShade(in ctx: GraphicsContext) {
        var base: CGPoint
        switch mood {
        case .idle: base = CGPoint(x: 11, y: -5)
        case .listening: base = CGPoint(x: 10, y: -4)
        case .thinking: base = CGPoint(x: 2, y: -11)
        case .talking: base = CGPoint(x: 12, y: -2)
        case .sleepy: base = CGPoint(x: 11, y: 1)
        }
        if mood == .idle && animated {
            base.x += value(period: 9, [(0, 0), (0.38, 0), (0.44, -13), (0.58, -13), (0.64, 3), (0.72, 3), (1, 0)])
            base.y += value(period: 9, [(0, 0), (0.38, 0), (0.44, 4), (0.58, 4), (0.64, -3), (0.72, -3), (1, 0)])
        }
        let edge = Color(hex: 0x1F0900)
        var c = ctx
        c.clip(to: ball)
        c.fill(
            Path(CGRect(x: -40, y: -40, width: 180, height: 180)),
            with: .radialGradient(
                Gradient(stops: [
                    .init(color: Color.white.opacity(0.16), location: 0),
                    .init(color: Color.white.opacity(0), location: 0.42),
                    .init(color: edge.opacity(0), location: 0.8),
                    .init(color: edge.opacity(0.34), location: 0.93),
                    .init(color: edge.opacity(0.62), location: 1)
                ]),
                center: CGPoint(x: 50 + base.x, y: 50 + base.y),
                startRadius: 0, endRadius: 60
            )
        )
    }

    // MARK: Faces

    private func drawIdle(in ctx: GraphicsContext) {
        var look = CGAffineTransform.identity
        var blink: CGFloat = 1
        if animated {
            let lx = value(period: 9, [(0, 0), (0.38, 0), (0.44, -9), (0.58, -9), (0.64, 2), (0.72, 2), (1, 0)])
            let ly = value(period: 9, [(0, 0), (0.38, 0), (0.44, 3), (0.58, 3), (0.64, -2), (0.72, -2), (1, 0)])
            look = CGAffineTransform(translationX: lx, y: ly)
            blink = value(period: 5, [(0, 1), (0.88, 1), (0.92, 0.1), (1, 1)])
        }
        // Blink squashes both eyes toward a line 60% down their combined bounds (y ≈ 39.6).
        let squash = CGAffineTransform(translationX: 0, y: 39.6).scaledBy(x: 1, y: blink).translatedBy(x: 0, y: -39.6)
        var eyes = Path()
        eyes.addPath(ellipse(53, 41, 4.6, 7.2, deg: -14))
        eyes.addPath(ellipse(72, 33, 4.0, 6.4, deg: -22))
        ctx.fill(eyes.applying(squash).applying(look), with: .color(white))

        var mouth = Path()
        mouth.move(to: CGPoint(x: 44, y: 57))
        mouth.addQuadCurve(to: CGPoint(x: 77, y: 56), control: CGPoint(x: 60, y: 71))
        stroke(mouth.applying(look), in: ctx)
    }

    private func drawListening(in ctx: GraphicsContext) {
        var shapes = Path()
        shapes.addPath(ellipse(51, 40, 6.0, 9.4, deg: -10))
        shapes.addPath(ellipse(71, 34, 4.6, 7.4, deg: -18))
        shapes.addPath(ellipse(62, 63, 4.2, 5.0, deg: 0))
        ctx.fill(shapes, with: .color(white))
    }

    private func drawThinking(in ctx: GraphicsContext) {
        var eyes = Path()
        eyes.addPath(ellipse(44, 31, 4.0, 6.6, deg: 8))
        eyes.addPath(ellipse(61, 27, 3.6, 6.0, deg: 4))
        ctx.fill(eyes, with: .color(white))

        // "M40 56 q5 -4 10 0 t10 0 t10 0"
        var mouth = Path()
        mouth.move(to: CGPoint(x: 40, y: 56))
        mouth.addQuadCurve(to: CGPoint(x: 50, y: 56), control: CGPoint(x: 45, y: 52))
        mouth.addQuadCurve(to: CGPoint(x: 60, y: 56), control: CGPoint(x: 55, y: 60))
        mouth.addQuadCurve(to: CGPoint(x: 70, y: 56), control: CGPoint(x: 65, y: 52))
        stroke(mouth, in: ctx)

        let dots: [CGPoint] = [CGPoint(x: 66, y: 74), CGPoint(x: 74, y: 72), CGPoint(x: 82, y: 68)]
        for (i, d) in dots.enumerated() {
            let o = animated ? value(period: 1.2, delay: Double(i) * 0.2, [(0, 0.25), (0.5, 1), (1, 0.25)]) : 1
            ctx.fill(Path(ellipseIn: CGRect(x: d.x - 2.6, y: d.y - 2.6, width: 5.2, height: 5.2)), with: .color(white.opacity(o)))
        }
    }

    private func drawTalking(in ctx: GraphicsContext) {
        var eyes = Path()
        eyes.move(to: CGPoint(x: 47, y: 42))
        eyes.addQuadCurve(to: CGPoint(x: 57, y: 42), control: CGPoint(x: 52, y: 34))
        eyes.move(to: CGPoint(x: 66, y: 35))
        eyes.addQuadCurve(to: CGPoint(x: 75, y: 35), control: CGPoint(x: 70.5, y: 28))
        stroke(eyes, in: ctx)

        let open: CGFloat = animated ? value(period: 0.45, [(0, 0.45), (0.5, 1), (1, 0.45)]) : 1
        let flap = CGAffineTransform(translationX: 0, y: 60).scaledBy(x: 1, y: open).translatedBy(x: 0, y: -60)
        ctx.fill(ellipse(62, 60, 9, 6.5, deg: 0).applying(flap), with: .color(white))
    }

    private func drawSleepy(in ctx: GraphicsContext) {
        var lines = Path()
        lines.move(to: CGPoint(x: 48, y: 44))
        lines.addQuadCurve(to: CGPoint(x: 58, y: 44), control: CGPoint(x: 53, y: 48))
        lines.move(to: CGPoint(x: 66, y: 37))
        lines.addQuadCurve(to: CGPoint(x: 75, y: 37), control: CGPoint(x: 70.5, y: 40.6))
        lines.move(to: CGPoint(x: 56, y: 62))
        lines.addQuadCurve(to: CGPoint(x: 66, y: 62), control: CGPoint(x: 61, y: 65))
        stroke(lines, in: ctx)

        drawZ(in: ctx, at: CGPoint(x: 82, y: 18), size: 18, delay: 0)
        drawZ(in: ctx, at: CGPoint(x: 94, y: 2), size: 13, delay: 1.3)
    }

    private func drawZ(in ctx: GraphicsContext, at point: CGPoint, size: CGFloat, delay: Double) {
        var c = ctx
        var p = point
        if animated {
            c.opacity = value(period: 2.6, delay: delay, [(0, 0), (0.3, 1), (0.7, 1), (1, 0)])
            p.x += value(period: 2.6, delay: delay, [(0, 0), (1, 4)])
            p.y += value(period: 2.6, delay: delay, [(0, 4), (1, -6)])
        }
        c.draw(
            Text("z").font(.system(size: size, weight: .semibold, design: .rounded)).foregroundColor(zColor),
            at: p, anchor: .bottomLeading
        )
    }

    // MARK: Helpers

    private func stroke(_ path: Path, in ctx: GraphicsContext) {
        ctx.stroke(path, with: .color(white), style: StrokeStyle(lineWidth: lineWidth, lineCap: .round, lineJoin: .round))
    }

    private func ellipse(_ cx: CGFloat, _ cy: CGFloat, _ rx: CGFloat, _ ry: CGFloat, deg: Double) -> Path {
        Path(ellipseIn: CGRect(x: -rx, y: -ry, width: rx * 2, height: ry * 2))
            .applying(CGAffineTransform(rotationAngle: CGFloat(deg * .pi / 180)))
            .applying(CGAffineTransform(translationX: cx, y: cy))
    }

    /// Keyframed value, eased in and out between frames like CSS `ease-in-out`.
    /// Frames are (progress 0…1, value) pairs in order.
    private func value(period: Double, delay: Double = 0, _ frames: [(Double, Double)]) -> CGFloat {
        var p = (time - delay).truncatingRemainder(dividingBy: period) / period
        if p < 0 { p += 1 }
        for i in 1..<frames.count where p <= frames[i].0 {
            let (p0, v0) = frames[i - 1]
            let (p1, v1) = frames[i]
            let u = p1 > p0 ? (p - p0) / (p1 - p0) : 1
            let eased = u * u * (3 - 2 * u)
            return CGFloat(v0 + (v1 - v0) * eased)
        }
        return CGFloat(frames[frames.count - 1].1)
    }
}

// MARK: - Palette

private extension HomieAvatar.Palette {
    var stops: [Gradient.Stop] {
        switch self {
        case .warm:
            return [.init(color: Color(hex: 0xD45F28), location: 0),
                    .init(color: Color(hex: 0xE06C3F), location: 0.55),
                    .init(color: Color(hex: 0xEE8AA3), location: 1)]
        case .blue:
            return [.init(color: Color(hex: 0x2550C7), location: 0),
                    .init(color: Color(hex: 0x4E6FE0), location: 0.6),
                    .init(color: Color(hex: 0xA88BEA), location: 1)]
        case .ink:
            return [.init(color: Color(hex: 0x15161A), location: 0),
                    .init(color: Color(hex: 0x3A3D46), location: 1)]
        case .night:
            return [.init(color: Color(hex: 0x8A4A26), location: 0),
                    .init(color: Color(hex: 0x94606A), location: 1)]
        }
    }
}

private extension Color {
    init(hex: UInt32) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255,
            opacity: 1
        )
    }
}

// MARK: - Preview

#Preview("Homie moods") {
    VStack(spacing: 24) {
        HomieAvatar(wobbles: true)
            .frame(width: 220, height: 220)
        HStack(spacing: 12) {
            ForEach(HomieAvatar.Mood.allCases, id: \.self) { mood in
                VStack(spacing: 6) {
                    HomieAvatar(mood: mood, palette: mood == .sleepy ? .night : .warm)
                        .frame(width: 64, height: 64)
                    Text(mood.rawValue.capitalized).font(.caption)
                }
            }
        }
        HStack(spacing: 16) {
            HomieAvatar(palette: .warm).frame(width: 56, height: 56)
            HomieAvatar(palette: .blue).frame(width: 56, height: 56)
            HomieAvatar(palette: .ink).frame(width: 56, height: 56)
        }
    }
    .padding()
}
