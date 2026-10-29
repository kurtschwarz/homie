import SwiftUI

enum AppPane: String, CaseIterable, Identifiable {
    case status, albums, artwork, sync, connection, account, general, about

    var id: String { rawValue }

    var title: String { rawValue.capitalized }

    var symbol: String? {
        switch self {
        case .status: nil
        case .albums: "photo"
        case .artwork: "paintbrush.pointed.fill"
        case .sync: "arrow.triangle.2.circlepath"
        case .connection: "wifi"
        case .account: "person.fill"
        case .general: "gearshape.fill"
        case .about: "info"
        }
    }

    var gradient: [Color] {
        switch self {
        case .status: [Color.white.opacity(0.1), Color.white.opacity(0.1)]
        case .albums: [Color(hex: 0x6FB7FF), Color(hex: 0x2F7BEA)]
        case .artwork: [Color(hex: 0xFF8FC2), Color(hex: 0xE8458E)]
        case .sync: [Color(hex: 0x6FE0A9), Color(hex: 0x2FB277)]
        case .connection: [Color(hex: 0x6FD9E8), Color(hex: 0x1FA5BE)]
        case .account: [Color(hex: 0xB9A0FF), Color(hex: 0x7A5AF0)]
        case .general, .about: [Color(hex: 0xA9ADB5), Color(hex: 0x6E737C)]
        }
    }

    var isNew: Bool { self == .artwork }

    var isAvailableSignedOut: Bool { self == .about }

    static let sections: [(heading: String?, panes: [AppPane])] = [
        (nil, [.status]),
        ("Photos", [.albums, .artwork, .sync]),
        ("Homie", [.connection, .account]),
        ("Homie Photos", [.general, .about])
    ]
}
