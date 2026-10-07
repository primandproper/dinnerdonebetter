package api

import (
	"context"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	mediaregistrybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/mediaregistry"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	"github.com/primandproper/platform-go/v15/callers"
	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	dataprivacyhttp "github.com/primandproper/platform-go/v15/dataprivacy/http"
	"github.com/primandproper/platform-go/v15/mediaregistry"
	mediaregistryhttp "github.com/primandproper/platform-go/v15/mediaregistry/http"
	"github.com/primandproper/platform-go/v15/operations"
	operationshttp "github.com/primandproper/platform-go/v15/operations/http"
	authzhttp "github.com/primandproper/primitives-go/v2/authorization/http"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/routing"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/uploads"

	"github.com/samber/do/v2"
)

// PlatformSurfaces are the surfaces platform-go serves over HTTP rather than gRPC, as this
// server mounts them on its router.
//
// Privacy requests, the operations that fulfill them, and the uploaded-object read. They are
// on HTTP because their flows are: a confirmation arrives as a link, progress is an event
// stream, and an object is bytes under their own content type. What this application adds is
// who is asking, resolved from the same session every gRPC surface reads — see
// AuthInterceptor.HTTPMiddleware.
type PlatformSurfaces struct {
	// sessions resolves a request's caller and attaches their session.
	sessions   func(http.Handler) http.Handler
	privacy    *dataprivacyhttp.Handlers
	operations *operationshttp.Handlers
	objects    *mediaregistryhttp.Handler
}

// RegisterPlatformSurfaces registers the HTTP surfaces with the injector.
func RegisterPlatformSurfaces(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*PlatformSurfaces, error) {
		logger := do.MustInvoke[logging.Logger](i)
		tracerProvider := do.MustInvoke[tracing.Provider](i)

		// Each route's permission is checked against the grants the session carries, the
		// same grants the gRPC enforcer reads. The permissions are a person's, held through
		// service_user — see authorization.ServiceUserPermissions.
		enforcer, err := authzhttp.NewEnforcer(sessions.GrantsFromContext,
			authzhttp.WithLogger(logger),
			authzhttp.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
		if err != nil {
			return nil, err
		}

		privacy, err := dataprivacyhttp.New(do.MustInvoke[platformdataprivacy.Service](i),
			dataprivacyhttp.WithEnforcer(enforcer),
			dataprivacyhttp.WithSubjectResolver(privacySubject),
			// No scope resolver: dataprivacyhttp.UnconfinedRequests is the default, and it is
			// this application's answer. A request here is about a person rather than an
			// account, and every route is already narrowed to the resolved subject.
			dataprivacyhttp.WithLogger(logger),
			dataprivacyhttp.WithTracerProvider(tracerProvider),
		)
		if err != nil {
			return nil, err
		}

		ops, err := operationshttp.New(do.MustInvoke[operations.Service](i),
			operationshttp.WithEnforcer(enforcer),
			// dataprivacy starts its operations owned by the subject, and no other operation
			// this application starts names an owner, so a caller's operations are their own.
			operationshttp.WithOwnerResolver(operationOwner),
			// The event stream, which a privacy request's receipt names beside the poll path.
			operationshttp.WithWatcher(do.MustInvoke[*operations.Watcher](i)),
			operationshttp.WithLogger(logger),
			operationshttp.WithTracerProvider(tracerProvider),
		)
		if err != nil {
			return nil, err
		}

		objects, err := mediaregistryhttp.New(
			do.MustInvoke[mediaregistry.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[uploads.UploadManager](i),
			// The default entitlement, OwnerOnly, is the rule the gRPC surface's reads apply
			// too: an object is its uploader's. The caller is the one that surface is built
			// with, so who owns an upload and who may read its bytes are one answer.
			mediaregistryhttp.WithCallerResolver(mediaregistrybuild.Caller),
			mediaregistryhttp.WithEnforcer(enforcer),
			mediaregistryhttp.WithLogger(logger),
			mediaregistryhttp.WithTracerProvider(tracerProvider),
		)
		if err != nil {
			return nil, err
		}

		return &PlatformSurfaces{
			sessions:   do.MustInvoke[*interceptors.AuthInterceptor](i).HTTPMiddleware,
			privacy:    privacy,
			operations: ops,
			objects:    objects,
		}, nil
	})
}

// Mount installs the session middleware and registers every route. It must be called before
// any other route is registered, because the middleware is global. A nil PlatformSurfaces
// mounts nothing, for a router built in a test that is about the other routes.
func (p *PlatformSurfaces) Mount(router *routing.Router) {
	if p == nil {
		return
	}

	// Global, so it reaches the OAuth2 and payment routes too, and it must not change what they
	// answer: a request with no bearer token, or one that does not work, goes on with no session
	// attached. The surfaces that need a caller refuse its absence themselves, as
	// callers.ErrNoPrincipal, which the mapper answers 401. A banned user and one who owes a
	// forced password change are refused here, because the gRPC interceptor refuses them and an
	// HTTP door that did not would be the way around it.
	router.Use(p.sessions)

	// The privacy surface's Mount includes the artifact download, which serves a subject
	// their own export and nobody else's.
	p.privacy.Mount(router)
	p.operations.Mount(router)
	p.objects.Mount(router)
}

// privacySubject is the signed-in user, as the subject of their own privacy requests.
func privacySubject(ctx context.Context) (platformdataprivacy.Subject, error) {
	userID := sessions.FromContext(ctx).GetUserID()
	if userID == "" {
		return platformdataprivacy.Subject{}, callers.ErrNoPrincipal
	}

	return platformdataprivacy.Subject{ID: userID, Type: platformdataprivacy.SubjectUser}, nil
}

// operationOwner is the owner dataprivacy files a person's operations under: the person.
func operationOwner(ctx context.Context) (tenancy.Scope, error) {
	userID := sessions.FromContext(ctx).GetUserID()
	if userID == "" {
		return tenancy.Scope{}, callers.ErrNoPrincipal
	}

	return tenancy.Of(userID), nil
}
