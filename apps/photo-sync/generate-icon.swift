import AppKit
import Foundation

guard CommandLine.arguments.count == 2 else {
    fputs("usage: generate-icon.swift OUTPUT.icns\n", stderr)
    exit(2)
}

let output = URL(fileURLWithPath: CommandLine.arguments[1])
let iconset = output.deletingLastPathComponent().appending(path: "AppIcon.iconset")
let files: [(String, Int)] = [
    ("icon_16x16.png", 16),
    ("icon_16x16@2x.png", 32),
    ("icon_32x32.png", 32),
    ("icon_32x32@2x.png", 64),
    ("icon_128x128.png", 128),
    ("icon_128x128@2x.png", 256),
    ("icon_256x256.png", 256),
    ("icon_256x256@2x.png", 512),
    ("icon_512x512.png", 512),
    ("icon_512x512@2x.png", 1024),
]

try? FileManager.default.removeItem(at: iconset)
try FileManager.default.createDirectory(at: iconset, withIntermediateDirectories: true)

for (name, pixels) in files {
    guard let bitmap = NSBitmapImageRep(
        bitmapDataPlanes: nil,
        pixelsWide: pixels,
        pixelsHigh: pixels,
        bitsPerSample: 8,
        samplesPerPixel: 4,
        hasAlpha: true,
        isPlanar: false,
        colorSpaceName: .deviceRGB,
        bytesPerRow: 0,
        bitsPerPixel: 0
    ) else { fatalError("Could not create \(pixels)-pixel icon") }

    bitmap.size = NSSize(width: pixels, height: pixels)
    NSGraphicsContext.saveGraphicsState()
    let graphics = NSGraphicsContext(bitmapImageRep: bitmap)!
    NSGraphicsContext.current = graphics
    graphics.cgContext.scaleBy(x: CGFloat(pixels) / 1024, y: CGFloat(pixels) / 1024)

    NSColor.clear.setFill()
    NSRect(x: 0, y: 0, width: 1024, height: 1024).fill()

    let shadow = NSShadow()
    shadow.shadowColor = NSColor.black.withAlphaComponent(0.22)
    shadow.shadowBlurRadius = 36
    shadow.shadowOffset = NSSize(width: 0, height: -16)
    shadow.set()
    NSColor.black.withAlphaComponent(0.12).setFill()
    NSBezierPath(ovalIn: NSRect(x: 245, y: 110, width: 534, height: 76)).fill()
    NSGraphicsContext.restoreGraphicsState()
    NSGraphicsContext.saveGraphicsState()
    NSGraphicsContext.current = graphics
    graphics.cgContext.scaleBy(x: CGFloat(pixels) / 1024, y: CGFloat(pixels) / 1024)

    let ball = NSBezierPath(ovalIn: NSRect(x: 128, y: 152, width: 768, height: 768))
    NSGradient(colors: [
        NSColor(red: 0.05, green: 0.55, blue: 0.37, alpha: 1),
        NSColor(red: 0.29, green: 0.78, blue: 0.59, alpha: 1),
        NSColor(red: 0.72, green: 0.93, blue: 0.81, alpha: 1),
    ])!.draw(in: ball, angle: 32)

    let face = NSColor(red: 0.01, green: 0.29, blue: 0.20, alpha: 1)
    face.setFill()
    rotatedOval(center: NSPoint(x: 515, y: 635), size: NSSize(width: 82, height: 142), degrees: 10).fill()
    rotatedOval(center: NSPoint(x: 665, y: 705), size: NSSize(width: 76, height: 130), degrees: 20).fill()

    let smile = NSBezierPath()
    smile.move(to: NSPoint(x: 420, y: 525))
    smile.curve(
        to: NSPoint(x: 710, y: 535),
        controlPoint1: NSPoint(x: 515, y: 405),
        controlPoint2: NSPoint(x: 625, y: 405)
    )
    smile.lineWidth = 48
    smile.lineCapStyle = .round
    face.setStroke()
    smile.stroke()

    NSGraphicsContext.restoreGraphicsState()
    guard let png = bitmap.representation(using: .png, properties: [:]) else {
        fatalError("Could not encode \(pixels)-pixel icon")
    }
    try png.write(to: iconset.appending(path: name))
}

let process = Process()
process.executableURL = URL(fileURLWithPath: "/usr/bin/iconutil")
process.arguments = ["--convert", "icns", "--output", output.path, iconset.path]
try process.run()
process.waitUntilExit()
guard process.terminationStatus == 0 else { exit(process.terminationStatus) }
try FileManager.default.removeItem(at: iconset)

func rotatedOval(center: NSPoint, size: NSSize, degrees: CGFloat) -> NSBezierPath {
    let oval = NSBezierPath(ovalIn: NSRect(
        x: center.x - size.width / 2,
        y: center.y - size.height / 2,
        width: size.width,
        height: size.height
    ))
    var transform = AffineTransform(
        translationByX: center.x,
        byY: center.y
    )
    transform.rotate(byDegrees: degrees)
    transform.translate(x: -center.x, y: -center.y)
    oval.transform(using: transform)
    return oval
}
