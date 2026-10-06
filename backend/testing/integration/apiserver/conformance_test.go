package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	grpcapi "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api/grpc"
	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	mealplanninggenerated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/audit/auditpb"
	auditclient "github.com/primandproper/platform-go/v15/audit/grpc/client"
	oauth2clientsclient "github.com/primandproper/platform-go/v15/authentication/oauth2clients/grpc/client"
	"github.com/primandproper/platform-go/v15/authentication/passkeys/passkeyspb"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset/passwordresetpb"
	signinclient "github.com/primandproper/platform-go/v15/authentication/signin/grpc/client"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	"github.com/primandproper/platform-go/v15/billing"
	billingclient "github.com/primandproper/platform-go/v15/billing/grpc/client"
	commentsclient "github.com/primandproper/platform-go/v15/comments/grpc/client"
	"github.com/primandproper/platform-go/v15/conformance"
	conformanceall "github.com/primandproper/platform-go/v15/conformance/all"
	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	dataprivacyhttp "github.com/primandproper/platform-go/v15/dataprivacy/http"
	"github.com/primandproper/platform-go/v15/identity"
	identityclient "github.com/primandproper/platform-go/v15/identity/grpc/client"
	issuereportsclient "github.com/primandproper/platform-go/v15/issuereports/grpc/client"
	mediaregistryclient "github.com/primandproper/platform-go/v15/mediaregistry/grpc/client"
	mediaregistryhttp "github.com/primandproper/platform-go/v15/mediaregistry/http"
	"github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb"
	notificationsclient "github.com/primandproper/platform-go/v15/notifications/grpc/client"
	"github.com/primandproper/platform-go/v15/operations"
	operationshttp "github.com/primandproper/platform-go/v15/operations/http"
	settingsclient "github.com/primandproper/platform-go/v15/settings/grpc/client"
	waitlistsclient "github.com/primandproper/platform-go/v15/waitlists/grpc/client"
	webhooksclient "github.com/primandproper/platform-go/v15/webhooks/grpc/client"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/capitalism"
	"github.com/primandproper/primitives-go/v2/clock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestPlatformConformance runs platform-go's own promises against this deployment.
//
// Every surface platform serves here is asserted by platform's suite rather than by a copy of it
// kept in this package: the suite is written once, upstream, against the generated clients, and
// what this file supplies is how this deployment mints a caller and brings about the states no
// client can. What it proves is this application's wiring — the interceptors, the scope each
// surface resolves, the hooks — on this application's dialect.
//
// This application's surfaces do not agree on one tenancy (see sessions.AccountScopedPrincipal):
// issue reports, webhooks and the audit read are per account, and the rest are global. A
// subject says so the way the suites ask — Scope is the global directory, and Scopes names the
// account on each surface that confines to one.
func TestPlatformConformance(T *testing.T) {
	T.Parallel()

	// The suites run as parallel subtests, so what they skipped is only known once they have
	// all finished.
	T.Cleanup(func() { assertConformanceSkips(T) })

	conformanceall.Run(T, conformance.Seams{
		NewSubject: newConformanceSubject,
		Anonymous: func(context.Context) (grpc.ClientConnInterface, error) {
			return dialConformance()
		},
		AnonymousHTTP: func(context.Context) (*http.Client, error) {
			return &http.Client{}, nil
		},
		// A token platform's sign-in issued is carried the way every client here carries one.
		SignedIn: func(_ context.Context, issued *signinpb.IssuedToken) (grpc.ClientConnInterface, error) {
			return dialConformance(client.WithBearerTokenCredentials(issued.GetToken()))
		},
		Actions: conformance.Actions{
			Auditable:          conformanceAuditable,
			Authorized:         conformanceAuthorized,
			PasswordResetToken: conformancePasswordResetToken,
			Credentialed:       conformanceCredentialed,
			InvitationToken:    conformanceInvitationToken,
			EmailVerified:      conformanceEmailVerified,
			Subscribed:         conformanceSubscribed,
			Notified:           conformanceNotified,
			CommentTarget:      conformanceCommentTarget,
			VerificationToken:  conformanceVerificationToken,
			WaitlistLinks:      conformanceWaitlistLinks,
			Registered:         conformanceRegistered,
			HandleReminder:     conformanceHandleReminder,
			Operated:           conformanceOperated,
			ArtifactExpired:    conformanceArtifactExpired,
			// MagicLinkToken is left nil on purpose: this deployment names no magic link store,
			// so platform refuses RequestMagicLink and RedeemMagicLink outright and no link is
			// ever mailed — see internal/build/signin. signin/magic_links skips for it.
		},
		CommentTargetType: string(mealplanning.CommentTargetTypeRecipes),
		OperatorMethods:   conformanceOperatorMethods(),
		OperatorRoutes:    conformanceOperatorRoutes(),
		// The roles are foreign-keyed, so the suites grant from this deployment's own.
		Roles: conformance.Roles{
			Owner:   authorization.AccountAdminRoleName,
			Service: authorization.ServiceAdminRoleName,
			// The role sign-in's administrative door admits — see signin.WithAdminServiceRoles.
			Administrator: authorization.ServiceAdminRoleName,
			Membership:    [2]string{authorization.AccountAdminRoleName, authorization.AccountMemberRoleName},
		},
		Dialect: dialect.Postgres,
		// waitlists resolves a visitor to the directory's scope, which is global.
		VisitorScope: new(tenancy.Global()),
		// Settings here have one subject type, the user — see internal/domain/settings — so no
		// caller resolves an account's setting, their own included.
		AccountSettingsUnresolved: true,
		// The interceptor checks a platform sign-in token's login on every request.
		ImmediateRevocation: true,
		// GetPrincipal is built with identitygrpc.WithPermissionResolver.
		PrincipalPermissions: true,
		// authentication.RegistrationPolicy registers somebody in good standing and issues them a
		// second factor, unproven until they answer it.
		// InvitationTokenReturned is left false on purpose: the identity server is built without
		// identitygrpc.WithInvitationTokenReturned, so an invitation's token reaches only the
		// invitee's inbox and a sender has no link to copy. The copied-link assertion skips.
		InvitationTokenReturned:        false,
		RegistrantsAdmittedUnverified:  true,
		RegistrationIssuesSecondFactor: true,
		// And refuses a registrant who names no password: this application has no passwordless
		// arrival yet.
		PasswordlessRegistrationRefused: true,
		// The relying party the testing config renders, which is what PasskeysService verifies a
		// ceremony against.
		WebAuthn: &conformance.WebAuthnDeployment{
			RPID:   apiServiceConfig.Auth.Passkey.RelyingParty.RPID,
			Origin: apiServiceConfig.Auth.Passkey.RelyingParty.RPOrigins[0],
		},
	})
}

// assertConformanceSkips fails the run on any skip but the ones this deployment expects, and on
// an expected one that no longer happens, so a new wiring gap shows up as a diff here rather than
// as a quietly longer list of skips.
func assertConformanceSkips(t *testing.T) {
	t.Helper()

	skips := conformance.Skips(t)

	actual := make([]string, 0, len(skips))
	for i := range skips {
		actual = append(actual, skips[i].Test)
	}

	expected := conformanceExpectedSkips()

	for _, name := range actual {
		if !slices.Contains(expected, name) {
			t.Errorf("unexpected conformance skip: %s", name)
		}
	}

	for _, name := range expected {
		if !slices.Contains(actual, name) {
			t.Errorf("expected conformance skip did not happen, so it now runs and belongs off the list: %s", name)
		}
	}
}

// conformanceExpectedSkips is every assertion the suites skip against this deployment, grouped
// by why. Every one is either the shape of this deployment or a choice it made.
func conformanceExpectedSkips() []string {
	expected := []string{
		// Global tenancy. Every surface but audit, issue reports and webhooks serves one directory,
		// so there is no neighboring tenant for a cross-tenant assertion to stand in.
		"billing/products/a_catalog_listing_pages_the_caller's_tenant_only",
		"billing/products/a_product_is_stocked_in_the_caller's_catalog_and_nobody_else's",
		"billing/products/a_revision_of_a_neighboring_tenant's_product_is_absent_and_changes_nothing",
		"billing/products/archiving_a_neighboring_tenant's_product_is_absent_and_changes_nothing",
		"billing/subscriptions/a_neighboring_tenant's_subscription_is_absent",
		"billing/subscriptions/archiving_a_neighboring_tenant's_subscription_is_absent_and_changes_nothing",
		"billing/subscriptions/the_scope-wide_listing_is_the_tenant's,_and_asks_about_no_account",
		"comments/confinement/a_discussion's_listings_reach_the_caller's_tenant_only",
		"comments/confinement/a_neighbor's_comment_is_absent_to_every_call_that_names_it",
		"dataprivacy/a_request_is_absent_to_a_caller_in_another_tenant",
		"identity/a_listing_pages_the_caller's_directory_only",
		"identity/a_read_by_id_is_scoped_to_the_caller's_directory",
		"identity/accounts/a_transfer_to_somebody_in_another_directory_is_refused",
		"identity/accounts/an_account_listing_pages_the_caller's_directory_only",
		"identity/accounts/an_account_read_is_confined_to_the_caller's_directory",
		"identity/invitations/an_invitation_read_is_confined_to_the_sender's_directory",
		"identity/memberships/a_user's_memberships_are_refused_to_a_caller_from_another_directory",
		"identity/users/a_search_by_username_prefix_is_confined_to_the_caller's_directory",
		"mediaregistry/the_resource_surface/another_tenant's_upload_is_absent_from_every_read,_exactly_as_an_unknown_one_is",
		"mediaregistry/the_serve_route/another_tenant's_object_is_absent,_exactly_as_an_unknown_one_is",
		"notifications/devices/a_device_listing_holds_nothing_from_another_tenant",
		"notifications/the_inbox/an_inbox_holds_nothing_from_another_tenant",
		"oauth2clients/reach/a_listing_pages_the_caller's_whole_registry_and_nobody_else's",
		"oauth2clients/reach/a_registration_in_another_registry_is_absent_to_a_caller_naming_it",
		"oauth2clients/reach/another_registry's_registration_is_answered_exactly_as_one_never_minted",
		"operations/tenant-owned",
		"settings/confinement/a_catalog_listing_pages_the_caller's_tenant_only",
		"settings/confinement/a_neighbor_resolving_the_caller's_setting_finds_no_such_setting",
		"settings/confinement/a_neighbor's_definition_is_absent_to_every_catalog_call_that_names_it",
		"settings/confinement/an_account_in_a_neighboring_directory_is_not_the_caller's_to_resolve",
		"waitlists/erasure/an_erasure_reaches_the_caller's_tenant_only",
		"waitlists/lists/a_list_read_by_id_is_scoped_to_the_caller's_tenant",
		"waitlists/lists/an_update_will_not_reach_another_tenant's_list",
		"waitlists/lists/the_console's_catalog_is_the_caller's,_open_and_closed_alike",
		"waitlists/lists/the_open_catalog_places_a_signed-in_caller_by_their_principal",
		"waitlists/signups/a_list's_signups_are_its_own_and_its_tenant's",
		"waitlists/signups/a_signup_read_by_id_is_scoped_to_the_caller's_tenant",
		"waitlists/signups/a_signup_read_by_its_address_is_scoped_to_the_caller's_tenant",
		"waitlists/signups/an_invitation_will_not_reach_another_tenant's_signup",
		"waitlists/the_signup_page/a_signed-in_caller_cannot_join_another_tenant's_list",
		"waitlists/the_signup_page/a_visitor_cannot_be_joined_to_a_list_outside_the_tenant_they_land_in",
		"waitlists/the_signup_page/a_visitor_sees_the_open_catalog_of_the_tenant_they_land_in,_and_only_that",

		// Settings have one subject type here, the user (Seams.AccountSettingsUnresolved).
		"settings/confinement/an_account_the_caller_holds_no_membership_in_is_not_the_caller's_to_resolve",

		// Two listings answer the suite's request with a refusal rather than a page —
		// FailedPrecondition for the invitations, NotFound for the setting values — so there is
		// no pagination to read.
		"pagination/identity_ListInvitationsForEmailAddress/a_cursor_is_echoed_as_the_previous_one",
		"pagination/identity_ListInvitationsForEmailAddress/a_page_size_too_large_to_narrow_is_clamped_rather_than_wrapped",
		"pagination/identity_ListInvitationsForEmailAddress/a_sort_direction_is_reported_normalized",
		"pagination/identity_ListInvitationsForEmailAddress/an_absent_page_size_is_reported_as_the_default",
		"pagination/settings_ListValuesForDefinition/a_cursor_is_echoed_as_the_previous_one",
		"pagination/settings_ListValuesForDefinition/a_page_size_too_large_to_narrow_is_clamped_rather_than_wrapped",
		"pagination/settings_ListValuesForDefinition/a_sort_direction_is_reported_normalized",
		"pagination/settings_ListValuesForDefinition/an_absent_page_size_is_reported_as_the_default",

		// An address claimed through UpdateProfile is refused, so nobody holds an unverified one.
		"identity/invitations/a_caller_who_has_not_verified_their_address_is_refused_its_invitations",

		// Registration: a registrant must name a password (Seams.PasswordlessRegistrationRefused),
		// is issued a second factor (Seams.RegistrationIssuesSecondFactor), and the sign-up door
		// is open (Seams.RegistrationClosed).
		"passkeys/last_passkey/the_last_passkey_of_somebody_with_no_password_stays",
		"signin/registration/a_closed_sign-up_door_is_refused_by_name",
		"signin/registration/a_mailed_link_cannot_replace_a_password_somebody_already_holds",
		"signin/registration/a_registrant_with_no_password_attaches_one_through_the_mailed_link",
		"signin/self/proving_a_second_factor_nobody_issued_is_refused_as_a_precondition",

		// No magic link store, so Actions.MagicLinkToken is nil — see the Seams above.
		"signin/magic_links",

		// No invitation token comes back to its sender (Seams.InvitationTokenReturned).
		"signin/registration/a_copied_invitation_link_registers_the_addressed_person_into_the_inviting_account",
	}

	// Every operator-only method of this application's own services. The reservation suite
	// checks platform's methods; that a member is refused one of these is this application's
	// own test.
	for _, method := range conformanceOperatorMethods() {
		if !strings.HasPrefix(method, "/primandproper.platform.") {
			expected = append(expected, "reservations/"+method)
		}
	}

	return expected
}

// dialConformance connects to this suite's server, carrying whatever credentials opts add.
func dialConformance(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.NewClient(fmt.Sprintf("127.0.0.1:%d", apiServiceConfig.GRPCServer.Port),
		append([]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			identityclient.DefaultInterceptors(),
		}, opts...)...,
	)
}

func newConformanceSubject(ctx context.Context, opts ...conformance.SubjectOption) (*conformance.Subject, error) {
	req := conformance.NewSubjectRequest(opts...)

	if req.Admin {
		return conformanceAdmin(ctx, req)
	}

	user, err := createServiceUser(ctx, true, nil)
	if err != nil {
		return nil, err
	}

	// A tenant named on a surface that confines to an account is a colleague in that account.
	// Anything else — no tenant, or the global directory every other surface serves — is a
	// caller of their own, in the account registration made them.
	desiredAccount := ""
	if req.Scope != nil && slices.Contains(accountScopedSurfaces, req.Surface) {
		desiredAccount = req.Scope.Owner()
		if err = joinAccount(ctx, user, desiredAccount); err != nil {
			return nil, fmt.Errorf("joining account %s: %w", desiredAccount, err)
		}
	}

	return conformanceSubjectFor(ctx, user, desiredAccount)
}

// accountScopedSurfaces are the suites whose surface is built on
// sessions.AccountScopedPrincipal, and so confines a caller to their active account.
var accountScopedSurfaces = []string{"audit", "issuereports", "webhooks"}

// conformanceAdmin is a service administrator of the suite's own: somebody registered the way
// everybody is, in an account of their own, and then granted the service role. One per request,
// because the suites ask for two operators and tell them apart by tenant — an operator reading
// another operator's audit chain is a read that has to be observably somebody else's.
//
// The audit surface is the exception to "in an account of their own". Its reads confine to the
// caller's active account and its request names none, so an operator verifying somebody's chain
// has to be acting in their account — which this deployment's operators do by impersonating its
// owner, through InternalOperations.ImpersonateUser. The request is then the owner's, in their
// account, and names the operator as who made it.
func conformanceAdmin(ctx context.Context, req *conformance.SubjectRequest) (*conformance.Subject, error) {
	user, err := createServiceUser(ctx, true, nil)
	if err != nil {
		return nil, err
	}

	directory, store, err := conformanceDirectory()
	if err != nil {
		return nil, err
	}

	if _, err = directory.SetUserServiceRoles(ctx, tenancy.Global(), user.ID,
		[]string{authorization.ServiceAdminRoleName}); err != nil {
		return nil, fmt.Errorf("granting the service administrator role: %w", err)
	}

	token, err := loginForConformance(ctx, user, "")
	if err != nil {
		return nil, err
	}

	conformanceTokens.Store(user.ID, token)

	connToken := token
	if req.Scope != nil && req.Surface == "audit" {
		account, accountErr := store.GetAccount(ctx, databaseClient.Reader(), tenancy.Global(), req.Scope.Owner())
		if accountErr != nil {
			return nil, fmt.Errorf("reading the account an operator is asked to act in: %w", accountErr)
		}

		if connToken, err = impersonateForConformance(ctx, token, account.OwnerUserID, account.ID); err != nil {
			return nil, err
		}
	}

	subject, err := conformanceSubjectWithToken(ctx, user.ID, connToken)
	if err != nil {
		return nil, err
	}

	if req.Scope != nil && slices.Contains(accountScopedSurfaces, req.Surface) {
		subject.Scopes[req.Surface] = *req.Scope
	}

	return subject, nil
}

// impersonateForConformance mints the token an operator, signed in as operatorToken, acts as
// subjectID in accountID through.
func impersonateForConformance(ctx context.Context, operatorToken, subjectID, accountID string) (string, error) {
	operator, err := buildAuthedGRPCClientWithBearerToken(operatorToken)
	if err != nil {
		return "", err
	}
	defer func() { _ = operator.Close() }()

	res, err := operator.ImpersonateUser(ctx, &internalopssvc.ImpersonateUserRequest{
		SubjectId: subjectID,
		AccountId: accountID,
	})
	if err != nil {
		return "", fmt.Errorf("impersonating the owner of account %s: %w", accountID, err)
	}

	return res.GetToken(), nil
}

// conformanceOperatorMethods are the calls this deployment reserves to an operator: every method
// the server's permission table gates behind a permission an ordinary caller does not hold.
//
// Derived from the table rather than listed, so that the list the suites route by is the one the
// interceptor enforces. An ordinary caller is what newConformanceSubject mints — a service user,
// and an account admin in the account they act on.
func conformanceOperatorMethods() []string {
	expanded, err := platformauthz.ExpandInheritance(authorization.PlatformPolicy()...)
	if err != nil {
		panic(err)
	}

	service := authorization.NewServiceRolePermissionCheckerFromSet(nil, expanded[authorization.ServiceUserRoleName])
	account := authorization.NewAccountRolePermissionCheckerFromSet(nil, expanded[authorization.AccountAdminRoleName])

	var reserved []string
	for method, required := range grpcapi.MethodPermissions() {
		for _, p := range required {
			if !service.HasPermission(p) && !account.HasPermission(p) {
				reserved = append(reserved, method)
				break
			}
		}
	}

	slices.Sort(reserved)

	return reserved
}

// conformanceOperatorRoutes are the routes on platform's HTTP surfaces this deployment reserves
// to an operator: every route whose permission an ordinary caller does not hold, derived from the
// surfaces' own tables for conformanceOperatorMethods' reason.
func conformanceOperatorRoutes() []string {
	expanded, err := platformauthz.ExpandInheritance(authorization.PlatformPolicy()...)
	if err != nil {
		panic(err)
	}

	service := authorization.NewServiceRolePermissionCheckerFromSet(nil, expanded[authorization.ServiceUserRoleName])
	account := authorization.NewAccountRolePermissionCheckerFromSet(nil, expanded[authorization.AccountAdminRoleName])

	var reserved []string
	for _, table := range []map[string][]platformauthz.Permission{
		dataprivacyhttp.Permissions(),
		mediaregistryhttp.Permissions(),
		operationshttp.Permissions(),
	} {
		for route, required := range table {
			for _, p := range required {
				if !service.HasPermission(p) && !account.HasPermission(p) {
					reserved = append(reserved, route)
					break
				}
			}
		}
	}

	slices.Sort(reserved)

	return reserved
}

func conformanceSubjectFor(ctx context.Context, user *identity.User, desiredAccount string) (*conformance.Subject, error) {
	token, err := loginForConformance(ctx, user, desiredAccount)
	if err != nil {
		return nil, err
	}

	conformanceTokens.Store(user.ID, token)

	return conformanceSubjectWithToken(ctx, user.ID, token)
}

// conformanceSubjectWithToken is userID, calling with token.
func conformanceSubjectWithToken(ctx context.Context, userID, token string) (*conformance.Subject, error) {
	ddbClient, err := buildAuthedGRPCClientWithBearerToken(token)
	if err != nil {
		return nil, err
	}
	defer func() { _ = ddbClient.Close() }()

	status, err := ddbClient.GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
	if err != nil {
		return nil, fmt.Errorf("reading the subject's active account: %w", err)
	}

	activeAccount := status.GetStatus().GetActiveAccountId()

	conn, err := dialConformance(client.WithBearerTokenCredentials(token))
	if err != nil {
		return nil, err
	}

	scopes := map[string]tenancy.Scope{}
	for _, surface := range accountScopedSurfaces {
		scopes[surface] = tenancy.Of(activeAccount)
	}

	return &conformance.Subject{
		Scope:     tenancy.Global(),
		Scopes:    scopes,
		UserID:    userID,
		AccountID: activeAccount,
		Conn:      conn,
		// The surfaces platform serves over HTTP, on the API server's router, at their
		// packages' default base paths.
		HTTP: &conformance.HTTPSurfaces{
			Client:        &http.Client{Transport: bearerTransport(token)},
			BaseURL:       httpTestServerAddress,
			DataPrivacy:   true,
			MediaRegistry: true,
			Operations:    true,
			// The API server runs an operations.Watcher, so the event stream is mounted.
			OperationEvents: true,
			// The authorization server's routes, mounted at the paths oauth2server fixes.
			OAuth2Server: true,
		},
		Surfaces: conformance.Surfaces{
			Audit:         auditclient.Wrap(conn),
			Billing:       billingclient.Wrap(conn),
			Comments:      commentsclient.Wrap(conn),
			Identity:      identityclient.Wrap(conn),
			IssueReports:  issuereportsclient.Wrap(conn),
			MediaRegistry: mediaregistryclient.Wrap(conn),
			Notifications: notificationsclient.Wrap(conn),
			OAuth2Clients: oauth2clientsclient.Wrap(conn),
			Passkeys:      passkeyspb.NewPasskeysServiceClient(conn),
			PasswordReset: passwordresetpb.NewPasswordResetServiceClient(conn),
			Settings:      settingsclient.Wrap(conn),
			SignIn:        signinclient.Wrap(conn),
			Waitlists:     waitlistsclient.Wrap(conn),
			Webhooks:      webhooksclient.Wrap(conn),
			// The operator halves, which audit's and sign-in's servers register beside them.
			AuditAdministration:  auditpb.NewAuditAdministrationServiceClient(conn),
			SignInAdministration: signinpb.NewSignInAdministrationServiceClient(conn),
		},
	}, nil
}

func loginForConformance(ctx context.Context, user *identity.User, desiredAccount string) (string, error) {
	code, err := totpCodeFor(user)
	if err != nil {
		return "", err
	}

	password := user.HashedPassword
	if user.Username == premadeAdminUser.Username {
		password = adminUserPassword
	}

	return localdev.FetchLoginTokenForUser(ctx, fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port), &signinpb.Credentials{
		Username:        user.Username,
		Password:        password,
		TotpCode:        code,
		ActiveAccountId: desiredAccount,
	})
}

// joinAccount makes user a member of accountID the way a person joins a household: its owner
// invites them, and they accept with the token that was mailed.
func joinAccount(ctx context.Context, user *identity.User, accountID string) error {
	directory, store, err := conformanceDirectory()
	if err != nil {
		return err
	}

	account, err := store.GetAccount(ctx, databaseClient.Reader(), tenancy.Global(), accountID)
	if err != nil {
		return err
	}

	invitation, err := directory.Invite(ctx, tenancy.Global(), &identity.Invitation{
		BelongsToAccount: accountID,
		FromUser:         account.OwnerUserID,
		ToEmail:          user.EmailAddress,
		ToName:           user.Username,
		// A peer rather than a junior: the suites' colleague is somebody with the same standing
		// in the tenant, so a refusal they meet is the rule under test rather than their role.
		Roles:     []string{authorization.AccountAdminRoleName},
		Token:     identifiers.New(),
		ExpiresAt: time.Now().Add(invitationLifetime).UTC(),
	})
	if err != nil {
		return err
	}

	_, err = directory.AcceptInvitation(ctx, tenancy.Global(), invitation.ID, invitation.Token, user.ID, "")

	return err
}

func conformanceDirectory() (*identity.Service, identity.Store, error) {
	return localdev.IdentityDirectory(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), databaseClient)
}

// conformanceAuditable records an entry in the account's chain, which is the chain the audit
// read resolves for a caller whose active account it is.
func conformanceAuditable(ctx context.Context, scope tenancy.Scope) (*conformance.Audited, error) {
	repo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, databaseClient)
	if err != nil {
		return nil, err
	}

	accountID := scope.Owner()
	entry := ddbaudit.NewEntry(identifiers.New(), accountID, "conformance_audited", identifiers.New(), platformaudit.EventOther)

	if err = databaseClient.WithTransaction(context.WithoutCancel(ctx), func(tx database.Tx) error {
		return repo.Record(ctx, tx, entry)
	}); err != nil {
		return nil, err
	}

	return &conformance.Audited{
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		ActorID:      entry.Actor.ID,
	}, nil
}

// conformanceCredentialed reports a fragment of the password digest registration stored.
func conformanceCredentialed(ctx context.Context, _ tenancy.Scope, userID string) (string, error) {
	_, store, err := conformanceDirectory()
	if err != nil {
		return "", err
	}

	user, err := store.GetUser(ctx, databaseClient.Reader(), tenancy.Global(), userID)
	if err != nil {
		return "", err
	}

	if len(user.HashedPassword) < 16 {
		return "", errors.New("the stored password digest is too short to search for")
	}

	return user.HashedPassword[len(user.HashedPassword)-16:], nil
}

// conformanceInvitationToken reads the token off the mail request identitystore.InvitationMailer
// queued once the invitation committed — the row the invitation mail is rendered from, and the one
// place the secret goes besides the invitee's inbox. Nothing in this suite relays the outbox, so it
// is still there.
func conformanceInvitationToken(ctx context.Context, _ tenancy.Scope, invitationID string) (string, error) {
	payloads, err := outboxPayloads(ctx,
		`convert_from(payload, 'UTF8') LIKE '%' || $1 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%'`,
		invitationID, ddbidentity.AccountInvitationMailRequestedEventType)
	if err != nil {
		return "", err
	}

	for _, payload := range payloads {
		if token := findStringKey(payload, identitykeys.AccountInvitationTokenKey); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no queued invitation mail names invitation %s", invitationID)
}

// findStringKey finds a string value under key at any depth of a JSON document.
func findStringKey(doc, key string) string {
	var v any
	if json.Unmarshal([]byte(doc), &v) != nil {
		return ""
	}

	var walk func(any) string
	walk = func(v any) string {
		switch x := v.(type) {
		case map[string]any:
			if s, ok := x[key].(string); ok && s != "" {
				return s
			}
			for _, child := range x {
				if s := walk(child); s != "" {
					return s
				}
			}
		case []any:
			for _, child := range x {
				if s := walk(child); s != "" {
					return s
				}
			}
		}

		return ""
	}

	return walk(v)
}

func conformanceEmailVerified(ctx context.Context, _ tenancy.Scope, userID string) error {
	_, store, err := conformanceDirectory()
	if err != nil {
		return err
	}

	return databaseClient.WithTransaction(ctx, func(tx database.Tx) error {
		return store.MarkUserEmailAddressProven(ctx, tx, tenancy.Global(), userID)
	})
}

// conformanceSubscribed writes a subscription the way the payment provider's webhook handler does.
func conformanceSubscribed(ctx context.Context, _ tenancy.Scope, accountID string) (*billing.Subscription, error) {
	now := time.Now().UTC().Truncate(time.Second)

	var subscription *billing.Subscription

	err := databaseClient.WithTransaction(context.WithoutCancel(ctx), func(tx database.Tx) error {
		plan, err := billingStore.CreateProduct(ctx, tx, tenancy.Global(), &billing.Product{
			Name:                  "conformance plan",
			Kind:                  billing.KindRecurring,
			Currency:              "USD",
			AmountCents:           1000,
			BillingIntervalMonths: 1,
			ExternalProductID:     "prod_" + identifiers.New(),
		})
		if err != nil {
			return err
		}

		subscription, err = billingStore.CreateSubscription(ctx, tx, tenancy.Global(), &billing.Subscription{
			BelongsToAccount:       accountID,
			ProductID:              plan.ID,
			ExternalSubscriptionID: "sub_" + identifiers.New(),
			Status:                 capitalism.SubscriptionStatusActive,
			CurrentPeriodStart:     now.Add(-24 * time.Hour),
			CurrentPeriodEnd:       now.Add(24 * time.Hour),
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// conformanceNotified files a notification through the decorated inbox the application's own
// announcements go through.
func conformanceNotified(ctx context.Context, _ tenancy.Scope, userID string) (string, error) {
	created, err := createUserNotification(ctx, userID)
	if err != nil {
		return "", err
	}

	return created.ID, nil
}

func totpCodeFor(user *identity.User) (string, error) {
	return totp.GenerateCode(strings.ToUpper(user.TwoFactorSecret), time.Now().UTC())
}

// conformanceCommentTarget writes a recipe for the comments suite to talk about.
//
// Recipes are checked targets here — a comment on one this application does not have is
// refused — so the suite cannot mint an identifier. The row goes in through the insert the
// recipe repository runs, without the steps a recipe created over the wire must carry: what the
// existence check reads is the recipe, and a full recipe is a dozen catalog rows the comments
// suite has no use for.
func conformanceCommentTarget(ctx context.Context, _ tenancy.Scope) (targetType, targetID string, err error) {
	id := identifiers.New()

	if err = mealplanninggenerated.New().CreateRecipe(ctx, databaseClient.Writer(), &mealplanninggenerated.CreateRecipeParams{
		ID:                   id,
		Name:                 "conformance recipe " + id,
		Slug:                 "conformance-recipe-" + id,
		MinEstimatedPortions: "1",
		PortionName:          "portion",
		PluralPortionName:    "portions",
		Status:               mealplanninggenerated.RecipeStatus(mealplanning.RecipeStatusApproved),
		YieldsComponentType:  mealplanninggenerated.ComponentTypeUnspecified,
		CreatedByUser:        premadeAdminUser.ID,
	}); err != nil {
		return "", "", err
	}

	return string(mealplanning.CommentTargetTypeRecipes), id, nil
}

// conformanceVerificationToken reads the verification link's secret off the newest event that
// queued a verification mail for the address — a registration, or a request for another link.
// The users row holds only a digest, so the event is the one place the secret survives.
func conformanceVerificationToken(ctx context.Context, _ tenancy.Scope, emailAddress string) (string, error) {
	var userID string
	if err := databaseClient.Reader().QueryRowContext(ctx,
		`SELECT id FROM ddb_identity_users WHERE email_address = $1`, emailAddress).Scan(&userID); err != nil {
		return "", fmt.Errorf("finding the user registered as %s: %w", emailAddress, err)
	}

	payloads, err := outboxPayloads(ctx, `convert_from(payload, 'UTF8') LIKE '%' || $1 || '%'`, userID)
	if err != nil {
		return "", err
	}

	// The link travels on this application's mail request alone — at registration, which signin
	// mails through SignInMailers once it commits, and on every request for another.
	for _, payload := range payloads {
		if token := findStringKey(payload, identitykeys.UserEmailVerificationTokenKey); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no queued event carries a verification link for %s", emailAddress)
}

var (
	waitlistConfirmLink     = regexp.MustCompile(`/waitlists/confirm\?t=([A-Za-z0-9_-]+)`)
	waitlistUnsubscribeLink = regexp.MustCompile(`/waitlists/unsubscribe\?t=([A-Za-z0-9_-]+)`)
)

// conformanceWaitlistLinks reads the two links out of the newest confirmation mail queued to
// contact. The suite's contacts are fresh per join, so the address alone names the signup.
func conformanceWaitlistLinks(ctx context.Context, _ tenancy.Scope, _, contact string) (*conformance.WaitlistLinks, error) {
	payloads, err := outboxPayloads(ctx,
		`topic = $1 AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%'`,
		apiServiceConfig.Queues.OutboundEmailsTopicName, contact)
	if err != nil {
		return nil, err
	}

	for _, payload := range payloads {
		confirm, unsubscribe := waitlistConfirmLink.FindStringSubmatch(payload), waitlistUnsubscribeLink.FindStringSubmatch(payload)
		if confirm != nil && unsubscribe != nil {
			return &conformance.WaitlistLinks{Confirm: confirm[1], Unsubscribe: unsubscribe[1]}, nil
		}
	}

	return nil, fmt.Errorf("no queued mail carries waitlist links for %s", contact)
}

// outboxPayloads reads the payloads of the outbox rows matching where, newest first. Nothing in
// this suite relays the outbox, so every row queued during the run is still there.
func outboxPayloads(ctx context.Context, where string, args ...any) ([]string, error) {
	rows, err := databaseClient.Reader().QueryContext(ctx,
		`SELECT convert_from(payload, 'UTF8') FROM outbox_messages WHERE `+where+` ORDER BY created_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var payloads []string
	for rows.Next() {
		var payload string
		if err = rows.Scan(&payload); err != nil {
			return nil, err
		}

		payloads = append(payloads, payload)
	}

	return payloads, rows.Err()
}

// conformanceTokens are the bearer tokens the subjects this run minted signed in with, by user,
// for the actions that have to act as somebody the suite names by identifier.
var conformanceTokens sync.Map

// bearerTransport carries token on every request an HTTP client makes.
type bearerTransport string

func (b bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+string(b))

	return http.DefaultTransport.RoundTrip(req)
}

// conformanceRegistered uploads an object as userID through this application's own upload path —
// platform's MediaRegistryService.UploadObject, which stores the bytes and registers the row — and
// reports it.
//
// The resource surface's assertions upload through that RPC themselves; the byte-serve's start
// here, because a deployment may store an object some other way and serve it all the same.
func conformanceRegistered(ctx context.Context, _ tenancy.Scope, userID string) (*conformance.RegisteredObject, error) {
	token, ok := conformanceTokens.Load(userID)
	if !ok {
		return nil, fmt.Errorf("no subject minted in this run is user %s", userID)
	}

	uploader, err := buildAuthedGRPCClientWithBearerToken(token.(string))
	if err != nil {
		return nil, err
	}
	defer func() { _ = uploader.Close() }()

	content := []byte("conformance object " + identifiers.New())

	stream, err := uploader.UploadObject(ctx)
	if err != nil {
		return nil, err
	}

	if err = stream.Send(&mediaregistrypb.UploadObjectRequest{
		Part: &mediaregistrypb.UploadObjectRequest_Header{Header: &mediaregistrypb.UploadObjectHeader{
			Name:        "conformance.png",
			ContentType: ddbuploadedmedia.MimeTypeImagePNG,
		}},
	}); err != nil {
		return nil, err
	}

	if err = stream.Send(&mediaregistrypb.UploadObjectRequest{Part: &mediaregistrypb.UploadObjectRequest_Chunk{Chunk: content}}); err != nil {
		return nil, err
	}

	uploaded, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	return &conformance.RegisteredObject{ID: uploaded.GetResult().GetId(), Content: content}, nil
}

// conformancePasswordResetToken reads the reset link's secret off the newest event that queued a
// reset mail for the address. The token store holds only a digest, so the event is the one place
// the secret survives.
func conformancePasswordResetToken(ctx context.Context, _ tenancy.Scope, emailAddress string) (string, error) {
	var userID string
	if err := databaseClient.Reader().QueryRowContext(ctx,
		`SELECT id FROM ddb_identity_users WHERE email_address = $1`, emailAddress).Scan(&userID); err != nil {
		return "", fmt.Errorf("finding the user registered as %s: %w", emailAddress, err)
	}

	payloads, err := outboxPayloads(ctx,
		`convert_from(payload, 'UTF8') LIKE '%' || $1 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%'`,
		userID, ddbidentity.PasswordResetTokenCreatedEventType)
	if err != nil {
		return "", err
	}

	for _, payload := range payloads {
		if secret := findStringKey(payload, authkeys.PasswordResetTokenSecretKey); secret != "" {
			return secret, nil
		}
	}

	return "", fmt.Errorf("no queued event carries a password reset link for %s", emailAddress)
}

// conformanceAuthorized approves an authorization request as userID, the way this deployment's
// authorization server is answered: a POST to /authorize carrying the person's bearer token, the
// request's parameters left in the query. The redirect is read, not followed.
func conformanceAuthorized(ctx context.Context, _ tenancy.Scope, userID, authorizeURL string) (string, error) {
	token, ok := conformanceTokens.Load(userID)
	if !ok {
		return "", fmt.Errorf("no subject minted in this run is user %s", userID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authorizeURL, http.NoBody)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token.(string))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient, err := localdev.NewNonRedirectingHTTPClient()
	if err != nil {
		return "", err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	location := res.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("the authorization server answered %d with no redirect", res.StatusCode)
	}

	return location, nil
}

// conformanceHandleReminder reports the username the newest reminder queued to an address names.
// The event names the user and the mail is rendered from their row, so the reminder is found by
// the user the address belongs to.
func conformanceHandleReminder(ctx context.Context, _ tenancy.Scope, emailAddress string) (string, error) {
	var userID, username string
	if err := databaseClient.Reader().QueryRowContext(ctx,
		`SELECT id, username FROM ddb_identity_users WHERE email_address = $1`, emailAddress).Scan(&userID, &username); err != nil {
		return "", fmt.Errorf("finding the user registered as %s: %w", emailAddress, err)
	}

	payloads, err := outboxPayloads(ctx,
		`convert_from(payload, 'UTF8') LIKE '%' || $1 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%'`,
		userID, ddbidentity.UsernameReminderRequestedEventType)
	if err != nil {
		return "", err
	}

	if len(payloads) == 0 {
		return "", fmt.Errorf("no username reminder was queued for %s", emailAddress)
	}

	return username, nil
}

// conformanceOperatedKind is work that waits to be cancelled, registered beside this
// application's own kinds so the Operated action has a kind operations.Service.Start accepts.
// A finished operation is one a cancellation leaves untouched, so an assertion that a refused
// cancellation changed nothing would pass whatever the surface did on one.
const (
	conformanceOperatedKind     = "conformance.operated"
	conformanceOperatedBackstop = 30 * time.Second
)

var registerConformanceOperatedKind = sync.OnceFunc(func() {
	operations.MustRegister(dataPrivacyFulfillment.OperationsRegistry, operations.Definition[struct{}]{
		Kind: conformanceOperatedKind,
		Run: func(ctx context.Context, _ struct{}, rep operations.Reporter) (*operations.Result, error) {
			backstop := time.NewTimer(conformanceOperatedBackstop)
			defer backstop.Stop()

			select {
			case <-rep.Cancelled():
			case <-backstop.C:
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			return nil, nil
		},
	})
})

// conformanceOperated starts an operation owned by the tenant, through the operations service
// the scheduler runs — the end of the path any of this application's start endpoints takes.
func conformanceOperated(ctx context.Context, scope tenancy.Scope) (string, error) {
	registerConformanceOperatedKind()

	op, err := dataPrivacyFulfillment.Operations.Start(ctx, conformanceOperatedKind, struct{}{}, operations.WithOwner(scope))
	if err != nil {
		return "", err
	}

	return op.ID, nil
}

// conformanceArtifactExpired runs dataprivacy's sweep at a clock a second past the request's
// expiry, over a store whose sweep sees that one request: at that clock, an unconfined sweep would
// expire every export the run has made, including ones a parallel assertion is reading.
func conformanceArtifactExpired(ctx context.Context, _ tenancy.Scope, requestID string) error {
	store := dataPrivacyFulfillment.Store

	req, err := store.Get(ctx, databaseClient.Reader(), nil, requestID)
	if err != nil {
		return fmt.Errorf("reading the request to expire: %w", err)
	}

	if req.ArtifactRef == "" || req.ExpiresAt.IsZero() {
		return fmt.Errorf("request %s names no artifact to expire", requestID)
	}

	sweeper, err := platformdataprivacy.NewSweeper(ctx,
		&platformdataprivacy.SweeperConfig{BatchSize: 1 << 16, DisableReap: true},
		&oneRequestStore{Store: store, requestID: requestID},
		platformdataprivacy.WithSweeperUploadManager(dataPrivacyFulfillment.Artifacts.UploadManager),
		platformdataprivacy.WithSweeperClock(stoppedClock{at: req.ExpiresAt.Add(time.Second)}),
	)
	if err != nil {
		return fmt.Errorf("building the sweeper: %w", err)
	}

	result, err := sweeper.Sweep(ctx)
	if err != nil {
		return fmt.Errorf("sweeping: %w", err)
	}

	if result.ArtifactsExpired != 1 {
		return fmt.Errorf("the sweep did not expire request %s's artifact", requestID)
	}

	return nil
}

// oneRequestStore is a dataprivacy.Store whose sweep sees one request and nothing else.
type oneRequestStore struct {
	platformdataprivacy.Store

	requestID string
}

func (o *oneRequestStore) ExpiringArtifacts(ctx context.Context, now time.Time, limit int) ([]*platformdataprivacy.Request, error) {
	due, err := o.Store.ExpiringArtifacts(ctx, now, limit)
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(due, func(req *platformdataprivacy.Request) bool { return req.ID != o.requestID }), nil
}

// LapseUnconfirmed lapses nothing: at a clock days ahead, every erasure a parallel assertion is
// waiting to confirm would go with it.
func (o *oneRequestStore) LapseUnconfirmed(context.Context, time.Time, int) (int64, error) {
	return 0, nil
}

// stoppedClock reads one moment, whenever it is asked.
type stoppedClock struct {
	clock.WallClock

	at time.Time
}

func (c stoppedClock) Now() time.Time { return c.at }

func (c stoppedClock) Since(t time.Time) time.Duration { return c.at.Sub(t) }
