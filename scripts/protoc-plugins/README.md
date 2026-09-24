# protoc plugins

The Swift codegen plugins are built from the manifests in this directory, not installed with brew.
`make proto_swift` builds them (a no-op once built) and passes them to protoc by path.

## Why not brew

`brew install` means "whatever is current on the machine that ran it", and these generators do
not produce the same output across versions:

- swift-protobuf began emitting `nonisolated extension` after 1.33.3.
- brew's build of `protoc-gen-grpc-swift` 2.1.1 is not the same generator as the git tag 2.1.1.
  The tag emits `Sendable` conformances on the service enums and passes the RPC type to each
  `MethodDescriptor`; brew's build does neither.

So identical `.proto` files produced different code on different machines, and a regeneration with
no schema change came out as a diff made entirely of toolchain noise.

## The versions

| plugin | built from | pinned |
| --- | --- | --- |
| `protoc-gen-swift` | `apple/swift-protobuf` | 1.33.3 |
| `protoc-gen-grpc-swift-2` | `grpc/grpc-swift-protobuf` | 2.1.1 |
| (its code generator) | `grpc/grpc-swift-2` | 2.4.3 |

These match [platform-client-swift](https://github.com/primandproper/platform-client-swift)'s
`scripts/plugins/Package.swift`. The two generate byte-identical code from the same schema, so its
client stays a drop-in for this app's. Bump them together.

The root `Makefile` asserts each plugin's `--version` before generating (`PROTOC_GEN_SWIFT_VERSION`,
`PROTOC_GEN_GRPC_SWIFT_VERSION`), so bump those with the manifests.

## Why two packages

platform-client-swift builds both plugins from one package. That package does not resolve on
SwiftPM 6.2 or later: grpc-swift-2 2.4.3 disables swift-protobuf's default traits, and SwiftPM 6.2
refuses that against any swift-protobuf before 1.36.0, which declares no traits. So
`protoc-gen-swift` is built alone at 1.33.3. The grpc plugin gets swift-protobuf 1.36.1, which only
parses descriptors for it and does not change what it writes.

## The runtime has to keep up

The generated code calls the grpc-swift-2 runtime the app links against, not the one the generator
was built with. The 2.4.3 generator emits `MethodDescriptor(service:method:type:)`, which the
runtime has had since 2.3.0. The app's `Package.resolved` pins grpc-swift-2 2.4.1. That is the
newest release with the initializer that does not also disable swift-protobuf's traits, which
would fail to resolve against the app's swift-protobuf 1.33.3.
