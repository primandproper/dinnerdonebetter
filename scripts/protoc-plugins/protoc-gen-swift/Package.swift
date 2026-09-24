// swift-tools-version: 6.0
import PackageDescription

// protoc-gen-swift, pinned. See scripts/protoc-plugins/README.md for why the plugins are
// built here rather than installed with brew, and why this is its own package.
let package = Package(
  name: "swift-protobuf-plugin",
  dependencies: [
    .package(url: "https://github.com/apple/swift-protobuf.git", exact: "1.33.3"),
  ]
)
