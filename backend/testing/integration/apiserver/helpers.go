package integration

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/fakes"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/outbox"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	adminUserPassword = "integration-tests-are-cool"

	nonexistentID = "00000000000000000000"
)

var (
	premadeAdminUser = &identity.User{
		ID:              identifiers.New(),
		TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		EmailAddress:    "integration_tests@example.email",
		Username:        "admin_user",
		HashedPassword:  adminUserPassword,
		ServiceRoles:    []string{authorization.ServiceUserRoleName},
	}

	adminClient client.Client
)

func getAccountIDForTest(t *testing.T, c client.Client) string {
	t.Helper()
	ctx := t.Context()

	status, err := c.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, status)
	require.NotEmpty(t, status.ActiveAccount)
	return status.ActiveAccount
}

func buildUnauthenticatedGRPCClientForTest(t *testing.T) client.Client {
	t.Helper()

	c, err := client.BuildUnauthenticatedGRPCClient(fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port))
	require.NoError(t, err)

	return c
}

func buildAuthedGRPCClient(ctx context.Context, token string) (client.Client, error) {
	c, err := localdev.BuildInsecureOAuthedGRPCClient(
		ctx,
		createdClientID,
		createdClientSecret,
		httpTestServerAddress,
		fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port),
		token,
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// buildAuthedGRPCClientWithBearerToken builds a client that sends the JWT directly as Bearer.
// Use this when the token has an account_id claim (e.g. from LoginForToken with DesiredAccountId)
// so that GetAuthStatus and other calls use that account as the active one.
func buildAuthedGRPCClientWithBearerToken(token string) (client.Client, error) {
	return client.BuildUnauthenticatedGRPCClientWithBearerToken(
		fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port),
		token,
	)
}

func hashStringToNumber(s string) uint64 {
	// Create a new FNV-1a 64-bit hash object
	h := fnv.New64a()

	// Write the bytes of the string into the hash object
	_, err := h.Write([]byte(s))
	if err != nil {
		// Handle error if necessary
		panic(err)
	}

	// Return the resulting hash value as a number (uint64)
	return h.Sum64()
}

func createServiceUserForTest(t *testing.T, verifyTOTP bool, in *auth.UserRegistrationInput) *identity.User {
	t.Helper()

	user, err := createServiceUser(t.Context(), verifyTOTP, in)
	require.NoError(t, err)

	return user
}

// createServiceUser registers somebody through the same RPC a sign-up form calls.
//
// RegisterUser on the auth service, not CreateUser on the identity service: the directory
// is platform's now and every method on it is behind a grant, so the one call a caller
// with no session makes is on the surface that has always served them.
func createServiceUser(ctx context.Context, verifyTOTP bool, in *auth.UserRegistrationInput) (*identity.User, error) {
	c, err := client.BuildUnauthenticatedGRPCClient(fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port))
	if err != nil {
		return nil, fmt.Errorf("initializing client: %w", err)
	}

	if in == nil {
		in = authfakes.BuildFakeUserRegistrationInput()
	}

	res, err := c.RegisterUser(ctx, &authsvc.RegisterUserRequest{
		Input: converters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(in),
	})
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	ucr := res.Created

	if verifyTOTP {
		if err = verifyTOTPSecretForUser(ctx, c, ucr.CreatedUserId, ucr.TwoFactorSecret); err != nil {
			return nil, fmt.Errorf("verifying totp code: %w", err)
		}
	}

	u := &identity.User{
		ID:              ucr.CreatedUserId,
		Username:        ucr.Username,
		EmailAddress:    ucr.EmailAddress,
		TwoFactorSecret: ucr.TwoFactorSecret,
		CreatedAt:       grpcconverters.ConvertPBTimestampToTime(ucr.CreatedAt),
		// this is a dirty trick to reuse this field to provide the password to the caller.
		HashedPassword: in.Password,
	}

	return u, nil
}

func verifyTOTPSecretForUser(ctx context.Context, c client.Client, userID, twoFactorSecret string) error {
	token, tokenErr := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
	if tokenErr != nil {
		return fmt.Errorf("generating totp code: %w", tokenErr)
	}

	if _, err := c.VerifyTOTPSecret(ctx, &authsvc.VerifyTOTPSecretRequest{
		TotpToken: token,
		UserId:    userID,
	}); err != nil {
		return fmt.Errorf("verifying totp code: %w", err)
	}

	return nil
}

func createClientForUser(ctx context.Context, user *identity.User) (client.Client, error) {
	token, err := fetchLoginTokenForUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("fetching token for user %s: %w", user.Username, err)
	}

	oauthedClient, err := buildAuthedGRPCClient(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("building oauthed client: %w", err)
	}

	return oauthedClient, nil
}

func buildUserRegistrationInputForTest(t *testing.T) *auth.UserRegistrationInput {
	t.Helper()

	return &auth.UserRegistrationInput{
		EmailAddress:          fmt.Sprintf("test+%d@whatever.com", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
		FirstName:             fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
		AccountName:           fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
		LastName:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
		Password:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
		Username:              fmt.Sprintf("test_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
		AcceptedPrivacyPolicy: true,
		AcceptedTOS:           true,
	}
}

func createUserAndClientForTest(t *testing.T) (*identity.User, client.Client) {
	t.Helper()

	return createUserAndClientForTestWithRegistrationInput(t, buildUserRegistrationInputForTest(t))
}

func createUserAndClientForTestWithRegistrationInput(t *testing.T, input *auth.UserRegistrationInput) (*identity.User, client.Client) {
	t.Helper()

	ctx := t.Context()

	user := createServiceUserForTest(t, true, input)
	oauthedClient, err := buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, user))
	require.NoError(t, err)

	return user, oauthedClient
}

func fetchLoginTokenForUserForTest(t *testing.T, user *identity.User) string {
	t.Helper()
	ctx := t.Context()

	rv, err := fetchLoginTokenForUser(ctx, user)
	require.NoError(t, err)

	return rv
}

func generateTOTPCodeForUserForTest(t *testing.T, user *identity.User) string {
	t.Helper()

	code, err := totp.GenerateCode(strings.ToUpper(user.TwoFactorSecret), time.Now().UTC())
	require.NoError(t, err)

	return code
}

func fetchLoginTokenForUser(ctx context.Context, user *identity.User) (string, error) {
	code, err := totp.GenerateCode(strings.ToUpper(user.TwoFactorSecret), time.Now().UTC())
	if err != nil {
		return "", err
	}

	loginInput := &authsvc.UserLoginInput{
		Username:  user.Username,
		Password:  user.HashedPassword,
		TotpToken: code,
	}

	// wretched hack that unfortunately works
	if user.Username == premadeAdminUser.Username {
		loginInput.Password = adminUserPassword
	}

	return localdev.FetchLoginTokenForUser(ctx, fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port), loginInput)
}

//// ChatGPT Zone

const (
	nilStr = "nil"
)

type compareOptions struct {
	// Ignore any field with these names at any depth (e.g., "LastUpdatedAt").
	IgnoreFieldNames map[string]struct{}
	// Only exported fields are considered (safe for cross-package types).
	ExportedOnly bool
}

// assertRoughEquality reports whether a and b are deeply equal after ignoring fields by name at any depth.
// Works across different struct types as long as exported field names/structure align.
func assertRoughEquality[T any](t *testing.T, expected, actual T, ignoreFieldNames ...string) {
	t.Helper()

	opts := compareOptions{
		IgnoreFieldNames: toSet(ignoreFieldNames),
		ExportedOnly:     true,
	}
	ma := flattenComparable(expected, opts)
	mb := flattenComparable(actual, opts)
	diff := diffMaps(ma, mb)

	if len(diff) == 0 {
		func() { /* some no-op to set expected breakpoint on */ }()
	}

	assert.Empty(t, diff, "diffs: %+v", diff)
}

func toSet(xs []string) map[string]struct{} {
	m := make(map[string]struct{}, len(xs))
	for _, x := range xs {
		m[x] = struct{}{}
	}
	return m
}

// flattenComparable produces a deterministic, comparable map[path]string for any value.
// It skips fields listed in opts.IgnoreFieldNames (matched by the field name at any depth),
// only includes exported fields when opts.ExportedOnly is true, and handles cycles.
func flattenComparable(v any, opts compareOptions) map[string]string {
	out := make(map[string]string)
	visited := make(map[uintptr]struct{})
	var walk func(rv reflect.Value, path []string)

	shouldIgnoreField := func(fieldName string) bool {
		_, ok := opts.IgnoreFieldNames[fieldName]
		return ok
	}

	join := func(path []string) string {
		return strings.Join(path, ".")
	}

	// Handle time.Time especially for stable representation.
	writeTime := func(tv time.Time, path []string) {
		// Use RFC3339Nano for human-readable + stable, or UnixNano if you prefer strict numeric
		out[join(path)] = tv.UTC().Format(time.RFC3339Nano)
	}

	writeLeaf := func(rv reflect.Value, path []string) {
		// Convert leaf value to a stable string.
		switch rv.Kind() {
		case reflect.String:
			out[join(path)] = rv.String()
		case reflect.Bool:
			out[join(path)] = strconv.FormatBool(rv.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			out[join(path)] = strconv.FormatInt(rv.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			out[join(path)] = strconv.FormatUint(rv.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			out[join(path)] = strconv.FormatFloat(rv.Float(), 'g', -1, rv.Type().Bits())
		case reflect.Complex64, reflect.Complex128:
			c := rv.Complex()
			out[join(path)] = fmt.Sprintf("(%g+%gi)", real(c), imag(c))
		default:
			// Fallback
			out[join(path)] = fmt.Sprintf("%v", rv.Interface())
		}
	}

	walk = func(rv reflect.Value, path []string) {
		if !rv.IsValid() {
			out[join(path)] = nilStr
			return
		}

		// Unwrap interfaces
		if rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				out[join(path)] = nilStr
				return
			}
			rv = rv.Elem()
		}

		// Follow pointers with cycle detection
		if rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				out[join(path)] = nilStr
				return
			}
			ptr := rv.Pointer()
			if ptr != 0 {
				if _, seen := visited[ptr]; seen {
					// Prevent cycles
					out[join(path)] = "<cycle>"
					return
				}
				visited[ptr] = struct{}{}
			}
			walk(rv.Elem(), path)
			return
		}

		// time.Time special case
		if rv.Type() == reflect.TypeFor[time.Time]() {
			if x, ok := rv.Interface().(time.Time); ok {
				writeTime(x, path)
			}
			return
		}

		switch rv.Kind() {
		case reflect.Struct:
			rt := rv.Type()
			for i := 0; i < rv.NumField(); i++ {
				sf := rt.Field(i)
				// Skip unexported fields if requested
				if opts.ExportedOnly && sf.PkgPath != "" {
					continue
				}
				if shouldIgnoreField(sf.Name) {
					continue
				}
				walk(rv.Field(i), append(path, sf.Name))
			}

		case reflect.Slice, reflect.Array:
			l := rv.Len()
			for i := range l {
				walk(rv.Index(i), append(path, fmt.Sprintf("[%d]", i)))
			}

		case reflect.Map:
			if rv.IsNil() {
				out[join(path)] = nilStr
				return
			}
			keys := rv.MapKeys()
			// Sort keys deterministically by their string form
			sort.Slice(keys, func(i, j int) bool {
				return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
			})
			for _, k := range keys {
				kv := rv.MapIndex(k)
				walk(kv, append(path, fmt.Sprintf("[%v]", k.Interface())))
			}

		default:
			// Basic leaf
			writeLeaf(rv, path)
		}
	}

	walk(reflect.ValueOf(v), nil)
	return out
}

func diffMaps(a, b map[string]string) map[string][2]string {
	diff := make(map[string][2]string)
	keys := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}
	for k := range keys {
		av, aok := a[k]
		bv, bok := b[k]
		switch {
		case aok && !bok:
			diff[k] = [2]string{av, "<absent>"}
		case !aok && bok:
			diff[k] = [2]string{"<absent>", bv}
		case aok && av != bv:
			diff[k] = [2]string{av, bv}
		}
	}
	return diff
}

// requireStreamSend asserts that sending on a client stream did not fail, treating io.EOF as
// success.
//
// gRPC reports a stream the server has already finished as io.EOF from Send, not as the status:
// "the stream is done, ask CloseAndRecv why". A test that provokes an immediate rejection — an
// unauthenticated upload, where the interceptor answers before the first message is read — is
// therefore racing. If the rejection lands first, Send returns io.EOF; if the send lands first,
// Send succeeds and the rejection arrives at CloseAndRecv. Both are the server behaving
// correctly, and only CloseAndRecv can tell the test which error it actually got, so the
// assertion belongs there rather than here.
func requireStreamSend(t *testing.T, err error) {
	t.Helper()

	if errors.Is(err, io.EOF) {
		return
	}

	require.NoError(t, err)
}

// invitationLifetime is how long an invitation a test issues has left to run.
//
// Named rather than inlined because the type requires an expiry — a link that never
// expires is a bearer credential nobody can retire — and every test that issues one wants
// the same answer: far enough out that the test can answer it, short enough that a row
// left behind is not a standing key to somebody's household.
const invitationLifetime = time.Hour

// oneHourFromNow is the expiry every invitation in this suite carries.
func oneHourFromNow() *timestamppb.Timestamp {
	return timestamppb.New(time.Now().Add(invitationLifetime).UTC())
}

// inviteForTest issues an invitation the test knows the token of.
//
// Through the same service the server runs, hooks and all, so the audit entry and the
// outbox event are the ones a request would write. What it does not do is go over the
// wire, and that is the point: the token is deliberately never on a gRPC response — the
// column holds a digest and the only moment the secret exists is the write that minted it,
// which is why platform hands the unredacted invitation to the hook that queues the mail.
// A test that took the token off the response would be testing a response that must not
// carry one.
//
// Everything from the acceptance onwards is the real path. It is the same arrangement
// passwordResetStoreForTest exists for, and for the same reason.
func inviteForTest(t *testing.T, fromUserID, accountID, toEmail string, roles ...string) *identity.Invitation {
	t.Helper()
	ctx := t.Context()

	if len(roles) == 0 {
		roles = []string{authorization.AccountMemberRoleName}
	}

	invitation, err := identityDirectoryWithHooks(t).Invite(ctx, ddbidentity.Scope(), &identity.Invitation{
		BelongsToAccount: accountID,
		FromUser:         fromUserID,
		ToEmail:          toEmail,
		ToName:           t.Name(),
		Note:             t.Name(),
		Roles:            roles,
		Token:            identifiers.New(),
		ExpiresAt:        time.Now().Add(invitationLifetime).UTC(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, invitation.Token)

	return invitation
}

// identityDirectoryWithHooks builds the directory service this suite issues invitations
// through: the same store the server holds, with the same hooks, over the same database.
func identityDirectoryWithHooks(t *testing.T) *identity.Service {
	t.Helper()

	auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, databaseClient)
	require.NoError(t, err)

	outboxWriter, err := outbox.NewWriter(dialect.Postgres, outbox.WithWriterSideEffect(indexevents.SideEffectName, indexevents.SideEffect))
	require.NoError(t, err)

	store, err := identity.NewSQLStore(databaseClient, identity.WithTablePrefix(ddbidentity.TablePrefix))
	require.NoError(t, err)

	directory, err := identity.NewService(databaseClient, store,
		identity.WithHooks(identitystore.ProvideHooks(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			auditLogRepo,
			events.NewEmitter(outboxWriter, apiServiceConfig.Queues.DataChangesTopicName, nil, indexevents.SideEffect),
		)),
	)
	require.NoError(t, err)

	return directory
}

// selfIDForTest is the user id the client is authenticated as.
func selfIDForTest(t *testing.T, c client.Client) string {
	t.Helper()
	ctx := t.Context()

	self, err := c.GetSelf(ctx, &authsvc.GetSelfRequest{})
	require.NoError(t, err)
	require.NotNil(t, self.GetResult())

	return self.GetResult().GetId()
}

// verifyEmailAddressForTest marks a user's address proven.
//
// Listing the invitations addressed to you is gated on having proven the address, and the
// gate is right: anybody may claim any address at registration, so reading what was sent
// to one you have not proven would be an oracle over other people's invitations.
//
// It goes through Store.MarkUserEmailAddressProven, which exists for exactly a caller who
// proved the address without holding the link that was mailed for it. The link flow is a
// path of its own and is tested as one; what this asserts is that the address is proven,
// not how.
func verifyEmailAddressForTest(t *testing.T, userID string) {
	t.Helper()
	ctx := t.Context()

	store, err := identity.NewSQLStore(databaseClient, identity.WithTablePrefix(ddbidentity.TablePrefix))
	require.NoError(t, err)

	require.NoError(t, databaseClient.WithTransaction(ctx, func(tx database.Tx) error {
		return store.MarkUserEmailAddressProven(ctx, tx, ddbidentity.Scope(), userID)
	}))
}
