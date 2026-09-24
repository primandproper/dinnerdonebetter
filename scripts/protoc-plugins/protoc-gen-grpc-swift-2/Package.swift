// swift-tools-version: 6.0
import PackageDescription

// protoc-gen-grpc-swift-2, pinned. See scripts/protoc-plugins/README.md for why the plugins
// are built here rather than installed with brew, and why this is its own package.
let package = Package(
  name: "grpc-swift-plugin",
  dependencies: [
    .package(url: "https://github.com/grpc/grpc-swift-protobuf.git", exact: "2.1.1"),
    // Transitive, and pinned anyway: the generator's output is produced by grpc-swift-2's
    // GRPCCodeGen, so letting it float would reintroduce the drift the pin above prevents.
    .package(url: "https://github.com/grpc/grpc-swift-2.git", exact: "2.4.3"),
    // Transitive too. grpc-swift-2 2.4.3 disables swift-protobuf's default traits, and
    // SwiftPM 6.2 refuses that against a swift-protobuf that declares none, which every
    // release before 1.36.0 does. This is the one plugin that cannot share 1.33.3.
    .package(url: "https://github.com/apple/swift-protobuf.git", exact: "1.36.1"),
  ]
)
