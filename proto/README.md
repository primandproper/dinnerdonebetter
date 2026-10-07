# Proto

The gRPC schema for this repository's own services: `mealplanning` and `internal_ops`, plus
`common.proto`, which both import. Every other service the apps call is platform's, and its
schema lives in platform-go, not here. `make proto` from the repository root regenerates the Go,
Swift and TypeScript code from these files; CI's `proto` workflow fails when what is committed is
not what the pinned toolchain produces.

## filtering is not here

`primandproper/platform/filtering/v1/filtering.proto` — the `QueryFilter` a caller sends and the
`Pagination` they are answered with — is **not** in this directory and must not be copied into it.
primitives-go ships that file inside the published module, so `backend/go.mod` already pins which
version of the schema this repo builds against, and there is nothing to keep in sync.

The root `Makefile` puts the module's proto directory on protoc's path:

```make
PLATFORM_PROTO_PATH := $(shell cd backend && go list -m -f '{{.Dir}}' github.com/primandproper/primitives-go/v2)/filtering/proto
```

A file here imports it by its canonical name, exactly as it already imports
`google/protobuf/timestamp.proto`:

```proto
import "primandproper/platform/filtering/v1/filtering.proto";

message GetThingsRequest {
  primandproper.platform.filtering.v1.QueryFilter filter = 1;
}
```

Go does not generate that file. It links against the bindings platform already generated, in
`primitives-go/v2/filtering/filteringpb`, via the `-M` mapping in `proto_golang` — because the
page-size clamp, the default, and the cursor asymmetry are server-side rules, and a second copy of
one can be wrong in a way nothing reports. The conversions live in `primitives-go/v2/filtering/grpc`;
this repo has no filter converters of its own.

Swift and TypeScript do not generate it either. Their client libraries, platform-client-swift
and `@primandproper/platform-client`, already ship it generated from the same schema, so each
generator is told to import it from there: Swift through `swift_module_mappings.asciipb`, and
ts-proto through the `M` options in the root `Makefile`'s `PROTO_TS_IMPORT_MAPPINGS`, which also
stops ts-proto writing a copy. The apps import `QueryFilter` from the library, the same type this
repository's generated stubs use.

`primandproper/platform/mediaregistry/v1/mediaregistry.proto` arrives the same way, from
platform-go, and is mapped the same way in all three languages. A platform proto newly imported
here needs a mapping in each, or Swift and TypeScript generate a second copy of it.
