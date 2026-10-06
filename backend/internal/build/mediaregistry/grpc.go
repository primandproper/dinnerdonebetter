/*
Package mediaregistry mounts platform-go's media registry surface over gRPC.

It replaces this application's own uploaded-media gRPC service, which was written over
platform's table and had the bugs a shared one does not: an upload's name was joined into its key
unchecked, so "../../<somebody>/x.png" planted an object under somebody else's part of the
bucket, and a registration accepted any unclaimed key as the caller's. Platform's surface refuses
both, and what this package supplies is the policy it asks for by value:

  - who is calling: the signed-in user, in the global directory every object is filed under —
    the same Caller the HTTP byte-serve is built with
  - where an upload goes: <user>/<object id>/<name>, which also makes <user>/ the only part of
    the bucket a caller may register keys under
  - which content types, and how large: uploadedmedia.SupportedMimeTypes and
    uploadedmedia.MaxUploadBytes
  - what an upload is metered against: the caller's active account

# Avatars

An avatar is an upload attached to its uploader, belongs_to {user, <their id>}, which is the one
attachment this surface authorizes without asking anybody. The newest such object is the
avatar; ListObjectsBySubject reads it back. Nothing else records which object is on show.

# Domain attachments

An upload attached to a recipe or a meal goes through mealplanning's own upload RPCs, which
authorize the subject they name. This surface refuses an attachment to anything but the caller.
*/
package mediaregistry

import (
	"context"
	"net/url"
	"path"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/platform-go/v15/mediaregistry"
	mediaregistrygrpc "github.com/primandproper/platform-go/v15/mediaregistry/grpc"
	mediaregistryhttp "github.com/primandproper/platform-go/v15/mediaregistry/http"
	"github.com/primandproper/platform-go/v15/metering"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/uploads"

	"github.com/samber/do/v2"
)

// Caller is the signed-in user, in the directory the upload registry files under. Both media
// surfaces are built with it, so the caller an upload is owned by and the caller a read is
// checked against are one answer.
func Caller(ctx context.Context) (mediaregistryhttp.Caller, error) {
	userID := sessions.FromContext(ctx).GetUserID()
	if userID == "" {
		return mediaregistryhttp.Caller{}, callers.ErrNoPrincipal
	}

	return mediaregistryhttp.Caller{PrincipalID: userID, Scope: tenancy.Global()}, nil
}

// ObjectKey lays an upload out as <user>/<object id>/<name>.
//
// The user first, so the part of the bucket that is one person's is a prefix: it is what
// mediaregistrygrpc.KeysUnderPrefix reads as the only keys a caller may register. The user is
// escaped into one segment, so an identifier with a slash in it cannot reach into somebody
// else's prefix; the name is already one segment, because the surface refuses any other.
//
// Every object is filed under the global scope, so the scope is left out of the layout.
func ObjectKey(caller mediaregistryhttp.Caller, objectID, name string) string {
	return path.Join(url.PathEscape(caller.PrincipalID), objectID, name)
}

// uploadMeter counts the bytes an upload added against the account the caller is acting in.
//
// A failure is returned to the surface, which records it on the upload's span and still tells
// the client their upload succeeded — which it did. The bytes and the row are committed by the
// time this runs, and nothing enforces this meter, so an uncounted record costs a gap in a
// dashboard rather than a wrong invoice; the metering package's dropped-record metric is what
// makes the gap alertable.
//
// The idempotency key is the registry row's ID rather than a request ID, because that is what
// is actually stable here. A client that retries a timed-out upload sends the bytes again, gets
// a new ID, and stores a second object — genuinely new usage that a request-scoped key would
// have deduped away into an object nobody is charged for.
func uploadMeter(db database.Client, recorder metering.Recorder) mediaregistrygrpc.AfterUpload {
	return func(ctx context.Context, _ mediaregistryhttp.Caller, object *mediaregistry.Object) error {
		return db.WithTransaction(ctx, func(tx database.Tx) error {
			return recorder.Record(ctx, tx, tenancy.Global(), metering.Usage{
				Subject:        sessions.FromContext(ctx).GetActiveAccountID(),
				Meter:          appmetering.UploadedMediaBytesMeter,
				Quantity:       object.Size,
				IdempotencyKey: object.ID,
				Dimensions: map[string]string{
					// Stored against the event for later analysis and deliberately not part of
					// the aggregate or of enforcement. The cardinality is bounded because the
					// surface refused everything outside SupportedMimeTypes before this ran.
					"mime_type": object.ContentType,
				},
			})
		})
	}
}

// NewServer builds platform's media registry surface with this application's policy.
func NewServer(
	store mediaregistry.Store,
	db database.Client,
	manager uploads.UploadManager,
	recorder metering.Recorder,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*mediaregistrygrpc.Server, error) {
	return mediaregistrygrpc.NewServer(store, db, manager,
		mediaregistrygrpc.WithCallerResolver(Caller),
		mediaregistrygrpc.WithKeyFunc(ObjectKey),
		mediaregistrygrpc.WithContentTypePolicy(mediaregistrygrpc.AllowContentTypes(uploadedmedia.SupportedMimeTypes()...)),
		mediaregistrygrpc.WithMaxBytes(uploadedmedia.MaxUploadBytes),
		mediaregistrygrpc.WithAfterUpload(uploadMeter(db, recorder)),
		mediaregistrygrpc.WithLogger(logger),
		mediaregistrygrpc.WithTracerProvider(tracerProvider),
		mediaregistrygrpc.WithMetricsProvider(metricsProvider),
	)
}

// RegisterMediaRegistryService registers platform's media registry surface with the injector.
func RegisterMediaRegistryService(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*mediaregistrygrpc.Server, error) {
		return NewServer(
			do.MustInvoke[mediaregistry.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[uploads.UploadManager](i),
			do.MustInvoke[metering.Recorder](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}

// Permissions is platform's map: every method a caller holding the media grants may make,
// against their own objects. The grants are a person's — see authorization.ServiceUserPermissions.
func Permissions() map[string][]authorization.Permission {
	return mediaregistrygrpc.Permissions()
}
