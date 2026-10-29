// swift-tools-version: 6.2
import PackageDescription

let swiftSettings: [SwiftSetting] = [
    .swiftLanguageMode(.v5),
    .defaultIsolation(MainActor.self),
    .enableUpcomingFeature("DisableOutwardActorInference"),
    .enableUpcomingFeature("GlobalActorIsolatedTypesUsability"),
    .enableUpcomingFeature("InferIsolatedConformances"),
    .enableUpcomingFeature("InferSendableFromCaptures"),
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
