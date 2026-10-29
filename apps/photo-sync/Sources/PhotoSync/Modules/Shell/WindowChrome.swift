import AppKit
import SwiftUI

extension View {
    func glassWindowChrome() -> some View {
        modifier(GlassWindowChrome())
    }
}

private struct GlassWindowChrome: ViewModifier {
    private static let tint = Color.homieWindowTint

    func body(content: Content) -> some View {
        #if compiler(>=6.2)
        if #available(macOS 26, *) {
            content
                .background(GlassWindowConfigurator())
                .glassEffect(.regular.tint(Self.tint.opacity(0.55)), in: Rectangle())
                .ignoresSafeArea()
        } else {
            clipped(content)
        }
        #else
        clipped(content)
        #endif
    }

    private func clipped(_ content: Content) -> some View {
        content
            .background(GlassWindowConfigurator())
            .background(Self.tint.opacity(0.72))
            .background(VisualEffectBackground())
            .clipShape(RoundedRectangle(cornerRadius: 26, style: .continuous))
            .ignoresSafeArea()
    }
}

private struct VisualEffectBackground: NSViewRepresentable {
    func makeNSView(context: Context) -> NSVisualEffectView {
        let view = NSVisualEffectView()
        view.material = .hudWindow
        view.blendingMode = .behindWindow
        view.state = .active
        return view
    }

    func updateNSView(_ view: NSVisualEffectView, context: Context) {}
}

private struct GlassWindowConfigurator: NSViewRepresentable {
    func makeNSView(context: Context) -> NSView { GlassWindowConfiguringView() }
    func updateNSView(_ view: NSView, context: Context) {}
}

private final class GlassWindowConfiguringView: NSView {
    static let contentSize = NSSize(width: 680, height: 760)

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        guard let window else { return }
        window.setContentSize(Self.contentSize)
        window.contentMinSize = Self.contentSize
        window.contentMaxSize = Self.contentSize
        window.styleMask.remove(.resizable)
        window.styleMask.insert(.fullSizeContentView)
        window.titleVisibility = .hidden
        window.titlebarAppearsTransparent = true
        window.isMovableByWindowBackground = true
        window.isOpaque = false
        window.backgroundColor = .clear
        window.appearance = NSAppearance(named: .darkAqua)
        window.standardWindowButton(.zoomButton)?.isEnabled = false
        DispatchQueue.main.async { [weak window] in
            guard let window else { return }
            for type in [NSWindow.ButtonType.closeButton, .miniaturizeButton, .zoomButton] {
                guard let button = window.standardWindowButton(type) else { continue }
                button.setFrameOrigin(NSPoint(x: button.frame.origin.x + 12, y: button.frame.origin.y - 8))
            }
            window.invalidateShadow()
        }
    }
}
