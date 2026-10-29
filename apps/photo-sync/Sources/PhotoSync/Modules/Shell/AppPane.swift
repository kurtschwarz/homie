import SwiftUI

enum PaneIcon {
    case homie
    case symbol(String, gradient: [Color])
}

struct AppPaneSection: Identifiable {
    let heading: String?
    let panes: [AppPane]

    var id: String { heading ?? "" }
}

enum SignedInScreen: Equatable {
    case emptyStatus
    case status
    case picker
    case placeholder

    nonisolated static func resolve(pane: AppPane, hasSelection: Bool) -> SignedInScreen {
        switch pane {
        case .status: hasSelection ? .status : .emptyStatus
        case .albums: .picker
        default: .placeholder
        }
    }
}

enum AppPane: String, CaseIterable, Identifiable {
    case status, albums, artwork, sync, connection, account, general, about

    var id: String { rawValue }

    var title: String { rawValue.capitalized }

    var icon: PaneIcon {
        switch self {
        case .status: .homie
        case .albums: .symbol("photo", gradient: [Color(hex: 0x6FB7FF), Color(hex: 0x2F7BEA)])
        case .artwork: .symbol("paintbrush.pointed.fill", gradient: [Color(hex: 0xFF8FC2), Color(hex: 0xE8458E)])
        case .sync: .symbol("arrow.triangle.2.circlepath", gradient: [Color(hex: 0x6FE0A9), Color(hex: 0x2FB277)])
        case .connection: .symbol("wifi", gradient: [Color(hex: 0x6FD9E8), Color(hex: 0x1FA5BE)])
        case .account: .symbol("person.fill", gradient: [Color(hex: 0xB9A0FF), Color(hex: 0x7A5AF0)])
        case .general: .symbol("gearshape.fill", gradient: [Color(hex: 0xA9ADB5), Color(hex: 0x6E737C)])
        case .about: .symbol("info", gradient: [Color(hex: 0xA9ADB5), Color(hex: 0x6E737C)])
        }
    }

    var isNew: Bool { self == .artwork }

    var isAvailableSignedOut: Bool { self == .about }

    static let sections: [AppPaneSection] = [
        AppPaneSection(heading: nil, panes: [.status]),
        AppPaneSection(heading: "Photos", panes: [.albums, .artwork, .sync]),
        AppPaneSection(heading: "Homie", panes: [.connection, .account]),
        AppPaneSection(heading: "Homie Photos", panes: [.general, .about])
    ]
}
