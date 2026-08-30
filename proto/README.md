# Proto

NOTE: this folder contains incomplete code that should not be utilized.

## filtering is not here

`primandproper/platform/filtering/v1/filtering.proto` — the `QueryFilter` a caller sends and the
`Pagination` they are answered with — is **not** in this directory and must not be copied into it.
platform-go ships that file inside the published module, so `backend/go.mod` already pins which
version of the schema this repo builds against, and there is nothing to keep in sync.

The root `Makefile` puts the module's proto directory on protoc's path:

```make
PLATFORM_PROTO_PATH := $(shell cd backend && go list -m -f '{{.Dir}}' github.com/primandproper/platform-go/v13)/filtering/proto
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
`platform-go/v13/filtering/filteringpb`, via the `-M` mapping in `proto_golang` — because the
page-size clamp, the default, and the cursor asymmetry are server-side rules, and a second copy of
one can be wrong in a way nothing reports. The conversions live in `platform-go/v13/filtering/grpc`;
this repo has no filter converters of its own.

Swift and TypeScript *do* generate it, from the same file, which is why `proto_swift` and
`proto_typescript` name it explicitly on protoc's command line. A generated `QueryFilter` in those
languages is a data class with eight fields and no rules to restate.
