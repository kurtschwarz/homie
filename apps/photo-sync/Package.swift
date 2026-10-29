// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "PhotoSync",
    platforms: [.macOS(.v15)],
    targets: [
        .executableTarget(
            name: "PhotoSync",
            path: "Sources/PhotoSync",
            swiftSettings: [.swiftLanguageMode(.v5)]
        ),
        .testTarget(
            name: "PhotoSyncTests",
            dependencies: ["PhotoSync"],
            path: "Tests/PhotoSyncTests",
            swiftSettings: [.swiftLanguageMode(.v5)]
        )
    ]
)
