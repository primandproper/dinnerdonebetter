package settingsspike

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v14/callers"
	platformsettings "github.com/primandproper/platform-go/v14/settings"
	settingsgrpc "github.com/primandproper/platform-go/v14/settings/grpc"
	"github.com/primandproper/platform-go/v14/settings/settingspb"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This file began as a pass-2 spike asking what becomes of the AdminOnly
// enforcement this application performs when platform's settings surface is
// mounted. The answer was "it is lost": any member holding the value-write grant
// — which every self-service user must hold, or they cannot set their own
// preferences — could set a value for a setting the catalog had reserved.
//
// platform closed it. settings/grpc now asks PermissionWriteAdminValues inside
// the handler, against the definition it read, because the fact that decides the
// question is in the request body and not in the method name: SetValue is one
// method serving both kinds of setting. A per-method grant could not have
// expressed it, and neither could SubjectAuthorizer, which is handed a caller
// and a subject and never learns which definition was named.
//
// What is left here is the regression test, from both directions. Refusing is
// only half of correct — a rule that refused every member every setting would
// pass a one-sided test and remove self-service.

const tablePrefix = ddbsettings.TablePrefix

// TestMain migrates the one database this package's tests share. The settings
// tables are already in this application's migrations — settings was adopted as
// a store in #1379 — so nothing extra is rendered here.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}

// selfServiceOnly is the SubjectAuthorizer platform's own documentation gives
// for a deployment like this one: a caller may answer their own settings and
// nobody else's.
//
// It is deliberately the same in every case below, because it is not what
// separates them. Whose settings these are and whether this setting is one an
// ordinary member may answer are two different questions, and this one only ever
// answers the first.
func selfServiceOnly() settingsgrpc.SubjectAuthorizer {
	return settingsgrpc.SubjectAuthorizerFunc(
		func(_ context.Context, caller callers.Principal, subject platformsettings.Subject) error {
			if subject.Type == platformsettings.SubjectUser && subject.ID == caller.UserID() {
				return nil
			}

			return callers.ErrTargetNotPermitted
		})
}

// grantsOf is the GrantsExtractor for a caller holding exactly perms.
func grantsOf(perms ...platformauthz.Permission) platformauthz.GrantsExtractor {
	return func(context.Context) (platformauthz.Grants, bool) {
		return platformauthz.NewGrants(platformauthz.NewPermissionSet(perms...)), true
	}
}

type fixture struct {
	server settingspb.SettingsServiceServer
	store  platformsettings.Store
	db     database.Client
}

// buildFixture wires platform's settings surface with the grants a caller holds.
func buildFixture(t *testing.T, grants platformauthz.GrantsExtractor) *fixture {
	t.Helper()

	ctx := t.Context()

	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	db, err := postgres.NewDatabaseClient(ctx, config,
		postgres.WithLogger(loggingnoop.NewLogger()),
		postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NoError(t, err)

	store, err := platformsettings.NewSQLStore(db, platformsettings.WithTablePrefix(tablePrefix))
	require.NoError(t, err)

	opts := []settingsgrpc.Option{
		settingsgrpc.WithLogger(loggingnoop.NewLogger()),
		settingsgrpc.WithTracerProvider(tracingnoop.NewTracerProvider()),
		settingsgrpc.WithMetricsProvider(metricsnoop.NewMetricsProvider()),
	}
	if grants != nil {
		opts = append(opts, settingsgrpc.WithGrantsExtractor(grants))
	}

	server, err := settingsgrpc.NewServer(store, db, sessions.PrincipalFromContext, selfServiceOnly(), opts...)
	require.NoError(t, err)

	return &fixture{server: server, store: store, db: db}
}

// define adds one setting to the catalog and answers with its name.
func (f *fixture) define(t *testing.T, ctx context.Context, adminOnly bool) string {
	t.Helper()

	name := "spike_setting_" + identifiers.New()

	require.NoError(t, f.db.WithTransaction(ctx, func(tx database.Tx) error {
		_, err := f.store.CreateDefinition(ctx, tx, ddbsettings.Scope(), &platformsettings.Definition{
			ID:          identifiers.New(),
			Name:        name,
			Description: "a setting",
			Kind:        platformsettings.KindString,
			Enumeration: []string{"on", "off"},
			Default:     pointer.To("off"),
			AdminOnly:   adminOnly,
		})

		return err
	}))

	return name
}

// callerCtx is a signed-in member. Their session carries no service-admin role:
// the grant that decides these tests is the one the GrantsExtractor supplies.
func callerCtx(t *testing.T, ctx context.Context, db database.Client) (signedIn context.Context, userID string) {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())

	return sessions.AttachToContext(ctx, &sessions.ContextData{
		ActiveAccountID: identifiers.New(),
		Requester:       sessions.RequesterInfo{UserID: user.ID},
	}), user.ID
}

func setValue(ctx context.Context, server settingspb.SettingsServiceServer, userID, name, value string) error {
	_, err := server.SetValue(ctx, &settingspb.SetValueRequest{
		Subject: &settingspb.SettingSubject{Type: string(platformsettings.SubjectUser), Id: userID},
		Name:    name,
		Value:   &settingspb.TypedValue{Value: &settingspb.TypedValue_StringValue{StringValue: value}},
	})

	return err
}

// TestAdminOnly_IsEnforcedForValueWrites pins the refusal and, as importantly,
// pins that it does not over-refuse.
//
// The method grant cannot tell these cases apart: every one of them is SetValue
// under settings.values.write, which every self-service member holds. What
// separates them is the definition each names and the grant the caller carries.
func TestAdminOnly_IsEnforcedForValueWrites(T *testing.T) {
	T.Parallel()

	T.Run("a member without the admin grant is refused a reserved setting", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		f := buildFixture(t, grantsOf(settingsgrpc.PermissionWriteValues))
		caller, userID := callerCtx(t, ctx, f.db)

		name := f.define(t, ctx, true)

		err := setValue(caller, f.server, userID, name, "on")
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err),
			"a reserved setting refused for want of a grant is a refusal, not a failure")

		_, readErr := f.store.GetValue(ctx, f.db.Reader(), ddbsettings.Scope(),
			platformsettings.Subject{Type: platformsettings.SubjectUser, ID: userID}, name)
		require.Error(t, readErr, "nothing should have been stored")
	})

	T.Run("the same member sets an ordinary setting", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		f := buildFixture(t, grantsOf(settingsgrpc.PermissionWriteValues))
		caller, userID := callerCtx(t, ctx, f.db)

		// Identical call, identical grant, identical authorizer. Only the
		// definition differs, which is the whole point of enforcing it here rather
		// than on the method.
		name := f.define(t, ctx, false)

		require.NoError(t, setValue(caller, f.server, userID, name, "on"),
			"self-service must survive the fix, or it has taken more than it closed")

		stored, err := f.store.GetValue(ctx, f.db.Reader(), ddbsettings.Scope(),
			platformsettings.Subject{Type: platformsettings.SubjectUser, ID: userID}, name)
		require.NoError(t, err)
		assert.Equal(t, "on", stored.Raw)
	})

	T.Run("a caller holding the admin grant sets a reserved setting", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		f := buildFixture(t, grantsOf(settingsgrpc.PermissionWriteValues, settingsgrpc.PermissionWriteAdminValues))
		caller, userID := callerCtx(t, ctx, f.db)

		name := f.define(t, ctx, true)

		require.NoError(t, setValue(caller, f.server, userID, name, "on"),
			"the grant exists so that somebody can write these")

		stored, err := f.store.GetValue(ctx, f.db.Reader(), ddbsettings.Scope(),
			platformsettings.Subject{Type: platformsettings.SubjectUser, ID: userID}, name)
		require.NoError(t, err)
		assert.Equal(t, "on", stored.Raw)
	})

	// platform rules that clearing is writing: taking an administrator's answer
	// back returns the setting to its default, which decides it for the subject
	// exactly as naming a value does. Pinned because it is the half of the rule a
	// reading of "write" could plausibly have missed.
	T.Run("clearing a reserved setting is refused too", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		f := buildFixture(t, grantsOf(settingsgrpc.PermissionWriteValues))
		caller, userID := callerCtx(t, ctx, f.db)

		name := f.define(t, ctx, true)

		_, err := f.server.ClearValue(caller, &settingspb.ClearValueRequest{
			Subject: &settingspb.SettingSubject{Type: string(platformsettings.SubjectUser), Id: userID},
			Name:    name,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	// A server built without a GrantsExtractor grants nothing, so a reserved
	// setting is refused rather than served. Fail-closed is the right default for
	// a write; see archived.go for why a read narrows instead.
	T.Run("no grants extractor refuses rather than permits", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		f := buildFixture(t, nil)
		caller, userID := callerCtx(t, ctx, f.db)

		name := f.define(t, ctx, true)

		err := setValue(caller, f.server, userID, name, "on")
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}
