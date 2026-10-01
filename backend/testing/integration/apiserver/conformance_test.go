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
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"
	notificationfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/fakes"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	mealplanninggenerated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v14/audit/auditpb"
	auditclient "github.com/primandproper/platform-go/v14/audit/grpc/client"
	oauth2clientsclient "github.com/primandproper/platform-go/v14/authentication/oauth2clients/grpc/client"
	"github.com/primandproper/platform-go/v14/authentication/passkeys/passkeyspb"
	"github.com/primandproper/platform-go/v14/authentication/passwordreset/passwordresetpb"
	signinclient "github.com/primandproper/platform-go/v14/authentication/signin/grpc/client"
	"github.com/primandproper/platform-go/v14/authentication/signin/signinpb"
	"github.com/primandproper/platform-go/v14/billing"
	billingclient "github.com/primandproper/platform-go/v14/billing/grpc/client"
	commentsclient "github.com/primandproper/platform-go/v14/comments/grpc/client"
	"github.com/primandproper/platform-go/v14/conformance"
	conformanceall "github.com/primandproper/platform-go/v14/conformance/all"
	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	dataprivacyhttp "github.com/primandproper/platform-go/v14/dataprivacy/http"
	"github.com/primandproper/platform-go/v14/identity"
	identityclient "github.com/primandproper/platform-go/v14/identity/grpc/client"
	issuereportsclient "github.com/primandproper/platform-go/v14/issuereports/grpc/client"
	mediaregistryhttp "github.com/primandproper/platform-go/v14/mediaregistry/http"
	notificationsclient "github.com/primandproper/platform-go/v14/notifications/grpc/client"
	"github.com/primandproper/platform-go/v14/operations"
	operationshttp "github.com/primandproper/platform-go/v14/operations/http"
	settingsclient "github.com/primandproper/platform-go/v14/settings/grpc/client"
	waitlistsclient "github.com/primandproper/platform-go/v14/waitlists/grpc/client"
	webhooksclient "github.com/primandproper/platform-go/v14/webhooks/grpc/client"
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
		VisitorScope: new(ddbidentity.Scope()),
		// Settings here have one subject type, the user — see internal/domain/settings — so no
		// caller resolves an account's setting, their own included.
		AccountSettingsUnresolved: true,
		// The interceptor checks a platform sign-in token's login on every request.
		ImmediateRevocation: true,
		// GetPrincipal is built with identitygrpc.WithPermissionResolver.
		PrincipalPermissions: true,
		// authentication.RegistrationPolicy registers somebody in good standing and issues them a
		// second factor, unproven until they answer it.
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

	if _, err = directory.SetUserServiceRoles(ctx, ddbidentity.Scope(), user.ID,
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
		account, accountErr := store.GetAccount(ctx, databaseClient.Reader(), ddbidentity.Scope(), req.Scope.Owner())
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

	account, err := store.GetAccount(ctx, databaseClient.Reader(), ddbidentity.Scope(), accountID)
	if err != nil {
		return err
	}

	invitation, err := directory.Invite(ctx, ddbidentity.Scope(), &identity.Invitation{
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

	_, err = directory.AcceptInvitation(ctx, ddbidentity.Scope(), invitation.ID, invitation.Token, user.ID, "")

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
	entry := &ddbaudit.AuditLogEntry{
		BelongsToAccount: &accountID,
		BelongsToUser:    identifiers.New(),
		ResourceType:     "conformance_audited",
		RelevantID:       identifiers.New(),
		EventType:        ddbaudit.AuditLogEventTypeOther,
	}

	if err = databaseClient.WithTransaction(context.WithoutCancel(ctx), func(tx database.Tx) error {
		return repo.Record(ctx, tx, entry)
	}); err != nil {
		return nil, err
	}

	return &conformance.Audited{
		ResourceType: entry.ResourceType,
		ResourceID:   entry.RelevantID,
		ActorID:      entry.BelongsToUser,
	}, nil
}

// conformanceCredentialed reports a fragment of the password digest registration stored.
func conformanceCredentialed(ctx context.Context, _ tenancy.Scope, userID string) (string, error) {
	_, store, err := conformanceDirectory()
	if err != nil {
		return "", err
	}

	user, err := store.GetUser(ctx, databaseClient.Reader(), ddbidentity.Scope(), userID)
	if err != nil {
		return "", err
	}

	if len(user.HashedPassword) < 16 {
		return "", errors.New("the stored password digest is too short to search for")
	}

	return user.HashedPassword[len(user.HashedPassword)-16:], nil
}

// conformanceInvitationToken reads the token off the event the invitation queued — the row the
// invitation mail is rendered from. Nothing in this suite relays the outbox, so it is still there.
func conformanceInvitationToken(ctx context.Context, _ tenancy.Scope, invitationID string) (string, error) {
	rows, err := databaseClient.Reader().QueryContext(ctx,
		`SELECT convert_from(payload, 'UTF8') FROM outbox_messages WHERE convert_from(payload, 'UTF8') LIKE '%' || $1 || '%'`,
		invitationID)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var payload string
		if err = rows.Scan(&payload); err != nil {
			return "", err
		}

		if token := findStringKey(payload, identitykeys.AccountInvitationTokenKey); token != "" {
			return token, nil
		}
	}

	if err = rows.Err(); err != nil {
		return "", err
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
		return store.MarkUserEmailAddressProven(ctx, tx, ddbidentity.Scope(), userID)
	})
}

// conformanceSubscribed writes a subscription the way the payment provider's webhook handler does.
func conformanceSubscribed(ctx context.Context, _ tenancy.Scope, accountID string) (*billing.Subscription, error) {
	now := time.Now().UTC().Truncate(time.Second)

	var subscription *billing.Subscription

	err := databaseClient.WithTransaction(context.WithoutCancel(ctx), func(tx database.Tx) error {
		plan, err := billingStore.CreateProduct(ctx, tx, ddbpayments.Scope(), &billing.Product{
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

		subscription, err = billingStore.CreateSubscription(ctx, tx, ddbpayments.Scope(), &billing.Subscription{
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

// conformanceNotified files a notification through the repository the application's own
// announcements go through.
func conformanceNotified(ctx context.Context, _ tenancy.Scope, userID string) (string, error) {
	input := converters.ConvertUserNotificationToUserNotificationDatabaseCreationInput(notificationfakes.BuildFakeUserNotification())
	input.BelongsToUser = userID

	created, err := notifsRepo.CreateUserNotification(ctx, input)
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
// UploadedMediaService.Upload, which stores the bytes and registers the row — and reports it.
func conformanceRegistered(ctx context.Context, _ tenancy.Scope, userID string) (*conformance.RegisteredObject, error) {
	token, ok := conformanceTokens.Load(userID)
	if !ok {
		return nil, fmt.Errorf("no subject minted in this run is user %s", userID)
	}

	uploader, err := buildAuthedGRPCClientWithBearerToken(token.(string))
	if err != nil {
		return nil, err
	}

	content := []byte("conformance object " + identifiers.New())

	stream, err := uploader.Upload(ctx)
	if err != nil {
		return nil, err
	}

	if err = stream.Send(&uploadedmediasvc.UploadRequest{
		Payload: &uploadedmediasvc.UploadRequest_Metadata{Metadata: &uploadedmediasvc.UploadMetadata{
			ObjectName:  "conformance.png",
			ContentType: ddbuploadedmedia.MimeTypeImagePNG,
		}},
	}); err != nil {
		return nil, err
	}

	if err = stream.Send(&uploadedmediasvc.UploadRequest{Payload: &uploadedmediasvc.UploadRequest_Chunk{Chunk: content}}); err != nil {
		return nil, err
	}

	uploaded, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	// The upload answers with where the object is stored and not with its row, so the row is
	// found by that key among the uploader's own.
	mine, err := uploader.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	for _, object := range mine.GetResults() {
		if object.GetObjectKey() == uploaded.GetObjectUrl() {
			return &conformance.RegisteredObject{ID: object.GetId(), Content: content}, nil
		}
	}

	return nil, fmt.Errorf("the upload stored %s and no row of the uploader's names it", uploaded.GetObjectUrl())
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
