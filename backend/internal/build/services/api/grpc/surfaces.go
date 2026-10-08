package grpcapi

import (
	"maps"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	issuereportsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/issuereports"
	mediaregistrybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/mediaregistry"
	oauth2clientsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/oauth2clients"
	passkeysbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passkeys"
	signinbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/signin"
	waitlistsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/waitlists"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	internalopssvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	internalopsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/grpc"

	auditpb "github.com/primandproper/platform-go/v15/audit/auditpb"
	auditgrpc "github.com/primandproper/platform-go/v15/audit/grpc"
	"github.com/primandproper/platform-go/v15/authentication/oauth2clients/oauth2clientspb"
	"github.com/primandproper/platform-go/v15/authentication/passkeys/passkeyspb"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset/passwordresetpb"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	billingpb "github.com/primandproper/platform-go/v15/billing/billingpb"
	paymentsgrpc "github.com/primandproper/platform-go/v15/billing/grpc"
	commentspb "github.com/primandproper/platform-go/v15/comments/commentspb"
	commentsgrpc "github.com/primandproper/platform-go/v15/comments/grpc"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	issuereportsgrpc "github.com/primandproper/platform-go/v15/issuereports/grpc"
	issuereportspb "github.com/primandproper/platform-go/v15/issuereports/issuereportspb"
	mediaregistrygrpc "github.com/primandproper/platform-go/v15/mediaregistry/grpc"
	notificationsgrpc "github.com/primandproper/platform-go/v15/notifications/grpc"
	notificationspb "github.com/primandproper/platform-go/v15/notifications/notificationspb"
	settingsgrpc "github.com/primandproper/platform-go/v15/settings/grpc"
	settingspb "github.com/primandproper/platform-go/v15/settings/settingspb"
	waitlistspb "github.com/primandproper/platform-go/v15/waitlists/waitlistspb"
	webhooksgrpc "github.com/primandproper/platform-go/v15/webhooks/grpc"
	webhookspb "github.com/primandproper/platform-go/v15/webhooks/webhookspb"
	platformgrpc "github.com/primandproper/primitives-go/v2/server/grpc"

	"github.com/samber/do/v2"
	"google.golang.org/grpc"
)

// surface is what one gRPC surface contributes to this server: the registration that mounts it,
// the permission table its methods ship with, and the amendments this deployment makes to that
// table.
//
// permissions is nil for a surface whose every method is anonymous — password reset's — since
// there is nothing for the enforcer to require. overrides is nil for every surface this
// deployment takes as it ships.
type surface struct {
	mount       func(i do.Injector) platformgrpc.RegistrationFunc
	permissions func() map[string][]authorization.Permission
	overrides   func() map[string][]authorization.Permission
}

// surfaces is every surface this server mounts, and the one list the registration funcs and
// the permission tables are both built from. Adding or removing a surface is one entry here.
func surfaces() []surface {
	return []surface{
		{mount: mountWithAdministration(auditpb.RegisterAuditServiceServer), permissions: auditgrpc.Permissions},
		{mount: mount(commentspb.RegisterCommentsServiceServer), permissions: commentsgrpc.Permissions},
		{mount: mount(identitypb.RegisterIdentityServiceServer), permissions: identitybuild.Permissions, overrides: identitybuild.PermissionOverrides},
		{mount: mount(internalopssvcpb.RegisterInternalOperationsServer), permissions: internalOpsPermissions},
		{mount: mount(issuereportspb.RegisterIssueReportsServiceServer), permissions: issuereportsgrpc.Permissions, overrides: issuereportsbuild.PermissionOverrides},
		// Domain: mealplanning
		{mount: mealplanningregistration.MountGRPC, permissions: mealplanningregistration.GRPCPermissions},
		{mount: mount(notificationspb.RegisterNotificationsServiceServer), permissions: notificationsgrpc.Permissions},
		{mount: mount(oauth2clientspb.RegisterOAuth2ClientsServiceServer), permissions: oauth2clientsbuild.Permissions},
		{mount: mount(billingpb.RegisterBillingServiceServer), permissions: paymentsgrpc.Permissions},
		{mount: mount(settingspb.RegisterSettingsServiceServer), permissions: settingsgrpc.Permissions},
		{mount: mountWithAdministration(signinpb.RegisterSignInServiceServer), permissions: signinbuild.Permissions},
		{mount: mount(passwordresetpb.RegisterPasswordResetServiceServer)},
		{mount: mount(passkeyspb.RegisterPasskeysServiceServer), permissions: passkeysbuild.Permissions},
		{mount: mountRegistrar[*mediaregistrygrpc.Server](), permissions: mediaregistrybuild.Permissions},
		{mount: mount(waitlistspb.RegisterWaitlistsServiceServer), permissions: waitlistsbuild.Permissions, overrides: waitlistsbuild.PermissionOverrides},
		{mount: mount(webhookspb.RegisterWebhooksServiceServer), permissions: webhooksgrpc.Permissions},
	}
}

// internalOpsPermissions is the internal operations fragment as the plain map the table is
// assembled from.
func internalOpsPermissions() map[string][]authorization.Permission {
	return internalopsgrpc.ProvideMethodPermissions()
}

// mount registers a surface through its generated registration, over the implementation the
// container holds under the generated server interface.
func mount[S any](register func(grpc.ServiceRegistrar, S)) func(do.Injector) platformgrpc.RegistrationFunc {
	return func(i do.Injector) platformgrpc.RegistrationFunc {
		impl := do.MustInvoke[S](i)

		return func(server *grpc.Server) {
			register(server, impl)
		}
	}
}

// registrar is a server that mounts itself.
type registrar interface {
	RegisterOn(*grpc.Server)
}

// mountRegistrar mounts a surface through its own RegisterOn, over the concrete server the
// container holds.
func mountRegistrar[S registrar]() func(do.Injector) platformgrpc.RegistrationFunc {
	return func(i do.Injector) platformgrpc.RegistrationFunc {
		return do.MustInvoke[S](i).RegisterOn
	}
}

// mountWithAdministration mounts a platform surface through its own RegisterOn where it has
// one, which registers the surface's operator half beside it — AuditAdministrationService beside
// AuditService, SignInAdministrationService beside SignInService. Mounting the half grants
// nothing: each of its methods requires a permission only an operator holds. A server without
// RegisterOn, a test double say, is registered as the plain service.
func mountWithAdministration[S any](register func(grpc.ServiceRegistrar, S)) func(do.Injector) platformgrpc.RegistrationFunc {
	return func(i do.Injector) platformgrpc.RegistrationFunc {
		impl := do.MustInvoke[S](i)

		if r, ok := any(impl).(registrar); ok {
			return r.RegisterOn
		}

		return func(server *grpc.Server) {
			register(server, impl)
		}
	}
}

// BuildRegistrationFuncs resolves every surface's implementation from i and returns the
// registrations that mount them, in the order surfaces lists them.
func BuildRegistrationFuncs(i do.Injector) []platformgrpc.RegistrationFunc {
	all := surfaces()
	funcs := make([]platformgrpc.RegistrationFunc, 0, len(all))

	for _, s := range all {
		funcs = append(funcs, s.mount(i))
	}

	return funcs
}

// MethodPermissions is the table the server's authentication interceptor enforces: every method
// this deployment serves, and the permissions a caller must hold to make it. It is each surface's
// fragment with this deployment's overrides applied.
//
// It is assembled here rather than in the injector so that what reads it outside the server —
// the enforcer equivalence proof, and the conformance harness deriving which calls this
// deployment reserves to an operator — reads the table the server enforces rather than a copy.
//
// The authorization enforcer does not read it. It is built from the two halves separately —
// MethodPermissionFragments declared, MethodPermissionOverrides applied through
// RequirementsBuilder.Override — so the equivalence proof compares two derivations of the table
// rather than one table with itself.
func MethodPermissions() interceptors.MethodPermissionsMap {
	out := MethodPermissionFragments()
	maps.Copy(out, MethodPermissionOverrides())

	return out
}

// MethodPermissionFragments is every surface's table as it ships, before this deployment amends
// any of it.
func MethodPermissionFragments() interceptors.MethodPermissionsMap {
	out := make(interceptors.MethodPermissionsMap)

	for _, s := range surfaces() {
		if s.permissions != nil {
			maps.Copy(out, s.permissions())
		}
	}

	return out
}

// MethodPermissionOverrides is where this deployment amends a platform surface's fragment: each
// entry replaces what a method the fragment already declares demands.
//
// They are kept apart from the fragments rather than written over a copy of them because the
// requirements builder checks an override against what was declared: an override naming a
// method no fragment declares — a typo, or an RPC the surface renamed — fails the build, where a
// copy edited in place would declare it quietly and leave the real method on its default.
func MethodPermissionOverrides() map[string][]authorization.Permission {
	out := map[string][]authorization.Permission{}

	for _, s := range surfaces() {
		if s.overrides != nil {
			maps.Copy(out, s.overrides())
		}
	}

	return out
}
