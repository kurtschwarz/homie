// swift-tools-version: 6.2
import PackageDescription

let swiftSettings: [SwiftSetting] = [
    .swiftLanguageMode(.v6),
    .defaultIsolation(MainActor.self),
    .enableUpcomingFeature("InferIsolatedConformances"),
    .enableUpcomingFeature("NonisolatedNonsendingByDefault")
]

let package = Package(
    name: "PhotoSync",
    platforms: [.macOS(.v15)],
    targets: [
        .executableTarget(
            name: "PhotoSync",
            path: "Sources/PhotoSync",
            swiftSettings: swiftSettings
        ),
        .testTarget(
            name: "PhotoSyncTests",
            dependencies: ["PhotoSync"],
            path: "Tests/PhotoSyncTests",
            swiftSettings: swiftSettings
        )
    ]
)
