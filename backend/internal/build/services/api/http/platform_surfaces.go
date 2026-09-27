package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	"github.com/primandproper/platform-go/v14/callers"
	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	dataprivacyhttp "github.com/primandproper/platform-go/v14/dataprivacy/http"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/mediaregistry"
	mediaregistryhttp "github.com/primandproper/platform-go/v14/mediaregistry/http"
	"github.com/primandproper/platform-go/v14/operations"
	operationshttp "github.com/primandproper/platform-go/v14/operations/http"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/routing"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/uploads"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

// ArtifactSuffix is where a completed export's artifact is served, under its request.
const ArtifactSuffix = "/artifact"

// PlatformSurfaces are the surfaces platform-go serves over HTTP rather than gRPC, as this
// server mounts them on its router.
//
// Privacy requests, the operations that fulfill them, and the uploaded-object read. They are
// on HTTP because their flows are: a confirmation arrives as a link, progress is an event
// stream, and an object is bytes under their own content type. What this application adds is
// who is asking, resolved from the same session every gRPC surface reads — see
// sessionFromAuthorization.
type PlatformSurfaces struct {
	sessions func(ctx context.Context, header string) (*sessions.ContextData, error)
	// requiresChange is whether a user has been told to change their password.
	requiresChange func(ctx context.Context, userID string) (bool, error)
	privacy        *dataprivacyhttp.Handlers
	operations     *operationshttp.Handlers
	objects        *mediaregistryhttp.Handler
	requests       platformdataprivacy.Service
	logger         logging.Logger
	tracer         tracing.Tracer
}

// RegisterPlatformSurfaces registers the HTTP surfaces with the injector.
func RegisterPlatformSurfaces(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*PlatformSurfaces, error) {
		logger := do.MustInvoke[logging.Logger](i)
		tracerProvider := do.MustInvoke[tracing.Provider](i)

		requests := do.MustInvoke[platformdataprivacy.Service](i)

		privacy, err := dataprivacyhttp.New(requests,
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
			// The default entitlement, OwnerOnly, is the rule UploadedMediaService's reads
			// already apply: an object is its uploader's.
			mediaregistryhttp.WithCallerResolver(objectCaller),
			mediaregistryhttp.WithLogger(logger),
			mediaregistryhttp.WithTracerProvider(tracerProvider),
		)
		if err != nil {
			return nil, err
		}

		return &PlatformSurfaces{
			sessions:       do.MustInvoke[*interceptors.AuthInterceptor](i).SessionFromAuthorization,
			requiresChange: do.MustInvoke[*interceptors.AuthInterceptor](i).RequiresPasswordChange,
			privacy:        privacy,
			operations:     ops,
			objects:        objects,
			requests:       requests,
			logger:         logging.NewNamedLogger(logger, "platform_http_surfaces"),
			tracer:         tracing.NewNamedTracer(tracerProvider, "platform_http_surfaces"),
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

	router.Use(p.sessionFromAuthorization)

	p.privacy.Mount(router)
	p.operations.Mount(router)
	p.objects.Mount(router)

	router.Handle(http.MethodGet,
		path.Join(dataprivacyhttp.BasePath, "/{requestID}", ArtifactSuffix),
		http.HandlerFunc(p.serveArtifact))
}

// sessionFromAuthorization attaches the caller a bearer token names.
//
// It is global middleware, so it reaches the OAuth2 and payment routes too, and it must not
// change what they answer: a request with no bearer token, or one that does not work, goes on
// with no session attached. The surfaces that need a caller refuse its absence themselves, as
// callers.ErrNoPrincipal, which the mapper answers 401.
//
// Two genuine tokens are refused here, because the gRPC interceptor refuses them and an HTTP
// door that did not would be the way around it: a user whose account status does not admit
// sign-in is 403, as it is PermissionDenied there, and so is a user who has been told to change
// their password, since no HTTP route is the change.
func (p *PlatformSurfaces) sessionFromAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		header := req.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			next.ServeHTTP(res, req)
			return
		}

		data, err := p.sessions(req.Context(), header)
		if errors.Is(err, platformidentity.ErrSignInNotAdmitted) {
			http.Error(res, "this account may not sign in", http.StatusForbidden)
			return
		}

		if err != nil || data == nil {
			next.ServeHTTP(res, req)
			return
		}

		requiresChange, err := p.requiresChange(req.Context(), data.GetUserID())
		if err != nil {
			http.Error(res, "checking password change requirement", http.StatusInternalServerError)
			return
		}

		if requiresChange {
			http.Error(res, "password change required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(res, req.WithContext(sessions.AttachToContext(req.Context(), data)))
	})
}

// serveArtifact hands a completed export to its subject.
//
// This application's own route, beside platform's five, and platform's documentation says why
// there is one: serving the bytes is a different kind of endpoint from a resource surface, and
// the artifact is encrypted at rest, so Download's signed URL would be ciphertext and Open —
// which decrypts — is the only way a subject reads it. The request is checked against the
// caller before a byte is read, and a request that is not theirs is absent rather than
// forbidden, as it is on platform's read.
func (p *PlatformSurfaces) serveArtifact(res http.ResponseWriter, req *http.Request) {
	ctx, span := p.tracer.StartSpan(req.Context())
	defer span.End()

	subject, err := privacySubject(ctx)
	if err != nil {
		http.Error(res, "authentication required", http.StatusUnauthorized)
		return
	}

	requestID := chi.URLParam(req, "requestID")

	stored, err := p.requests.Get(ctx, nil, requestID)
	if err != nil || stored == nil || stored.Subject.ID != subject.ID {
		http.Error(res, "not found", http.StatusNotFound)
		return
	}

	artifact, err := p.requests.Open(ctx, nil, stored.ID)
	if err != nil {
		if errors.Is(err, platformdataprivacy.ErrArtifactUnavailable) {
			http.Error(res, "the artifact is not available", http.StatusConflict)
			return
		}

		p.logger.WithSpan(span).Error("opening a privacy request artifact", err)
		http.Error(res, "opening the artifact", http.StatusInternalServerError)

		return
	}

	defer func() {
		if closeErr := artifact.Close(); closeErr != nil {
			p.logger.WithSpan(span).Error("closing a privacy request artifact", closeErr)
		}
	}()

	res.Header().Set("Content-Type", "application/octet-stream")
	res.Header().Set("Cache-Control", "no-store")

	if _, err = io.Copy(res, artifact); err != nil {
		p.logger.WithSpan(span).Error("writing a privacy request artifact", err)
	}
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

// objectCaller is the signed-in user, in the directory the upload registry files under.
func objectCaller(ctx context.Context) (mediaregistryhttp.Caller, error) {
	userID := sessions.FromContext(ctx).GetUserID()
	if userID == "" {
		return mediaregistryhttp.Caller{}, callers.ErrNoPrincipal
	}

	return mediaregistryhttp.Caller{PrincipalID: userID, Scope: ddbuploadedmedia.Scope()}, nil
}
