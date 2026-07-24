package integration

import (
	"context"
	"fmt"
	"hash/fnv"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/localdev"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v6/identifiers"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func createServiceUserForTest(t *testing.T, verifyTOTP bool, in *identity.UserRegistrationInput) *identity.User {
	t.Helper()

	user, err := createServiceUser(t.Context(), verifyTOTP, in)
	require.NoError(t, err)

	return user
}

func createServiceUser(ctx context.Context, verifyTOTP bool, in *identity.UserRegistrationInput) (*identity.User, error) {
	c, err := client.BuildUnauthenticatedGRPCClient(fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port))
	if err != nil {
		return nil, fmt.Errorf("initializing client: %w", err)
	}

	if in == nil {
		in = identityfakes.BuildFakeUserCreationInput()
	}
	input := converters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(in)

	res, err := c.CreateUser(ctx, &identitysvc.CreateUserRequest{Input: input})
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

func buildUserRegistrationInputForTest(t *testing.T) *identity.UserRegistrationInput {
	t.Helper()

	return &identity.UserRegistrationInput{
		Birthday:              new(time.Now()),
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

func createUserAndClientForTestWithRegistrationInput(t *testing.T, input *identity.UserRegistrationInput) (*identity.User, client.Client) {
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
	code, err := user.GenerateTOTPCode()
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

	assert.True(t, len(diff) == 0, "diffs: %+v", diff)
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
