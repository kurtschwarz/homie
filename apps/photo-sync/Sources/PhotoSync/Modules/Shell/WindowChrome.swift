import AppKit
import SwiftUI

enum GlassWindow {
    static let contentSize = NSSize(width: 680, height: 760)
    static let cornerRadius: CGFloat = 26
    static let trafficLightOffset = CGSize(width: 12, height: -8)
    // Clears the traffic lights, which sit inside the sidebar's top edge.
    static let sidebarTopInset: CGFloat = 52
}

extension View {
    func glassWindowChrome() -> some View {
        modifier(GlassWindowChrome())
    }
}

private struct GlassWindowChrome: ViewModifier {
    private static let tint = Color.homieWindowTint
    // Tuned by eye per OS: Liquid Glass and the visual-effect material composite a tint differently.
    private static let glassTintOpacity = 0.55
    private static let fallbackTintOpacity = 0.72

    func body(content: Content) -> some View {
        #if compiler(>=6.2)
        if #available(macOS 26, *) {
            content
                .background(GlassWindowConfigurator())
                .glassEffect(.regular.tint(Self.tint.opacity(Self.glassTintOpacity)), in: Rectangle())
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
            .background(Self.tint.opacity(Self.fallbackTintOpacity))
            .background(VisualEffectBackground())
            .clipShape(RoundedRectangle(cornerRadius: GlassWindow.cornerRadius, style: .continuous))
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

private var trafficLightsOffsetKey: UInt8 = 0

private final class GlassWindowConfiguringView: NSView {
    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        guard let window else { return }
        window.setContentSize(GlassWindow.contentSize)
        window.contentMinSize = GlassWindow.contentSize
        window.contentMaxSize = GlassWindow.contentSize
        window.styleMask.remove(.resizable)
        window.styleMask.insert(.fullSizeContentView)
        window.titlebarAppearsTransparent = true
        window.isMovableByWindowBackground = true
        window.isOpaque = false
        window.backgroundColor = .clear
        window.appearance = NSAppearance(named: .darkAqua)
        window.standardWindowButton(.zoomButton)?.isEnabled = false
        offsetTrafficLightsOnce(in: window)
    }

    // The host view is rebuilt on a host change but the window is not, and the
    // offset is relative, so apply it once per window.
    private func offsetTrafficLightsOnce(in window: NSWindow) {
        guard objc_getAssociatedObject(window, &trafficLightsOffsetKey) == nil else { return }
        objc_setAssociatedObject(window, &trafficLightsOffsetKey, true, .OBJC_ASSOCIATION_RETAIN_NONATOMIC)
        // SwiftUI lays out the titlebar after the view attaches, so move the buttons a turn later.
        DispatchQueue.main.async { [weak window] in
            guard let window else { return }
            let offset = GlassWindow.trafficLightOffset
            for type in [NSWindow.ButtonType.closeButton, .miniaturizeButton, .zoomButton] {
                guard let button = window.standardWindowButton(type) else { continue }
                button.setFrameOrigin(NSPoint(x: button.frame.origin.x + offset.width, y: button.frame.origin.y + offset.height))
            }
            window.invalidateShadow()
        }
    }
}
