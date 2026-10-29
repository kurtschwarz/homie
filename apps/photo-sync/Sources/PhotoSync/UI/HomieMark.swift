import SwiftUI

struct HomieMark: View {
    var body: some View {
        Canvas { context, size in
            let scale = min(size.width, size.height) / 220
            context.scaleBy(x: scale, y: scale)
            let ball = Path(ellipseIn: CGRect(x: 24, y: 4, width: 172, height: 172))

            context.drawLayer { shadow in
                shadow.addFilter(.blur(radius: 8))
                shadow.fill(
                    Path(ellipseIn: CGRect(x: 52, y: 193, width: 116, height: 10)),
                    with: .color(.black.opacity(0.2))
                )
            }

            context.fill(
                ball,
                with: .linearGradient(
                    Gradient(colors: [
                        Color(red: 0.22, green: 0.73, blue: 0.53),
                        Color(red: 0.47, green: 0.84, blue: 0.68),
                        Color(red: 0.77, green: 0.94, blue: 0.84),
                    ]),
                    startPoint: CGPoint(x: 48, y: 18),
                    endPoint: CGPoint(x: 182, y: 168)
                )
            )

            var shade = context
            shade.clip(to: ball)
            shade.fill(
                Path(CGRect(x: 0, y: -20, width: 220, height: 220)),
                with: .radialGradient(
                    Gradient(stops: [
                        .init(color: .white.opacity(0.12), location: 0),
                        .init(color: .white.opacity(0), location: 0.5),
                        .init(color: Color(red: 0.02, green: 0.28, blue: 0.19).opacity(0), location: 0.72),
                        .init(color: Color(red: 0.02, green: 0.28, blue: 0.19).opacity(0.65), location: 1),
                    ]),
                    center: CGPoint(x: 148, y: 48),
                    startRadius: 0,
                    endRadius: 132
                )
            )

            let face = Color(red: 0.02, green: 0.31, blue: 0.22)
            context.fill(oval(center: CGPoint(x: 113, y: 73), radii: CGSize(width: 8, height: 14), rotation: -10), with: .color(face))
            context.fill(oval(center: CGPoint(x: 147, y: 55), radii: CGSize(width: 8, height: 13), rotation: -20), with: .color(face))

            var smile = Path()
            smile.move(to: CGPoint(x: 94, y: 106))
            smile.addQuadCurve(to: CGPoint(x: 158, y: 103), control: CGPoint(x: 126, y: 137))
            context.stroke(smile, with: .color(face), style: StrokeStyle(lineWidth: 8, lineCap: .round))
        }
    }

    private func oval(center: CGPoint, radii: CGSize, rotation: CGFloat) -> Path {
        Path(ellipseIn: CGRect(
            x: -radii.width,
            y: -radii.height,
            width: radii.width * 2,
            height: radii.height * 2
        ))
        .applying(CGAffineTransform(rotationAngle: rotation * .pi / 180))
        .applying(CGAffineTransform(translationX: center.x, y: center.y))
    }
}
