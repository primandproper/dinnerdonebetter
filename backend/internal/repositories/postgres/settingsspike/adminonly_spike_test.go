package settingsspike

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v14/callers"
	platformsettings "github.com/primandproper/platform-go/v14/settings"
	settingsgrpc "github.com/primandproper/platform-go/v14/settings/grpc"
	"github.com/primandproper/platform-go/v14/settings/settingspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file is a pass-2 spike, not an adoption. It answers one question about
// mounting platform-go v14's settings surface: what becomes of the AdminOnly
// enforcement this application's own service performs.
//
// platform records AdminOnly and deliberately does not enforce it —
// settings.Definition says so outright: "It is recorded rather than enforced —
// this package has no notion of who is calling, and a store that pretended to
// would be an authorization check in the wrong layer. What it is for is the
// caller's own check."
//
// This application is that caller, and internal/services/settings/grpc is where
// the check lives today: a non-admin reaching for an admin-only setting is
// refused, on the read path and on the write path.
//
// The question is whether platform's own gRPC surface gives that check a place
// to stand. The test below is written to pass if it does not.

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
// for a deployment like this one, copied from settings/grpc/authorizer.go:60.
//
// It is the whole of the authorization this surface offers over a value write.
func selfServiceOnly() settingsgrpc.SubjectAuthorizer {
	return settingsgrpc.SubjectAuthorizerFunc(
		func(_ context.Context, caller callers.Principal, subject platformsettings.Subject) error {
			if subject.Type == platformsettings.SubjectUser && subject.ID == caller.UserID() {
				return nil
			}

			return callers.ErrTargetNotPermitted
		})
}

// TestSpike_PlatformServerDoesNotEnforceAdminOnly demonstrates the gap.
//
// A caller who is not a service administrator sets a value for a definition
// marked AdminOnly, for themselves, through platform's surface wired the way
// platform documents. It succeeds.
//
// The method grant does not stop them: SetValue requires
// settings.values.write, which every self-service user must hold in order to
// set any of their own settings at all. The SubjectAuthorizer does not stop
// them either, and cannot: it is handed the subject and never the definition,
// and it runs before the definition is read — see settings/grpc/values.go:69,
// where authorizeSubject is called above the transaction that then does
// GetDefinitionByName.
//
// So there is no seam on this surface where a consumer can apply the check
// AdminOnly's own documentation asks them to apply.
func TestSpike_PlatformServerDoesNotEnforceAdminOnly(T *testing.T) {
	T.Parallel()

	T.Run("a non-admin sets an admin-only setting through platform's surface", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		_, config := pgtesting.NewIsolatedDatabaseForTest(t)

		db, err := postgres.NewDatabaseClient(ctx, config,
			postgres.WithLogger(loggingnoop.NewLogger()),
			postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
		require.NoError(t, err)

		store, err := platformsettings.NewSQLStore(db,
			platformsettings.WithTablePrefix(ddbsettings.TablePrefix))
		require.NoError(t, err)

		server, err := settingsgrpc.NewServer(store, db, sessions.PrincipalFromContext, selfServiceOnly(),
			settingsgrpc.WithLogger(loggingnoop.NewLogger()),
			settingsgrpc.WithTracerProvider(tracingnoop.NewTracerProvider()),
			settingsgrpc.WithMetricsProvider(metricsnoop.NewMetricsProvider()))
		require.NoError(t, err)

		// A setting the operator marked as theirs to choose, not the user's.
		name := "spike_admin_only_" + identifiers.New()

		require.NoError(t, db.WithTransaction(ctx, func(tx database.Tx) error {
			_, createErr := store.CreateDefinition(ctx, tx, ddbsettings.Scope(), &platformsettings.Definition{
				ID:          identifiers.New(),
				Name:        name,
				Description: "only an administrator may set this",
				Kind:        platformsettings.KindString,
				Enumeration: []string{"on", "off"},
				Default:     pointer.To("off"),
				AdminOnly:   true,
			})

			return createErr
		}))

		// An ordinary user: the plain service-user role, holding no service-admin
		// permissions at all.
		user := pgtesting.CreateUserForTest(t, nil, db.Writer())
		ordinary := authorization.NewServiceRolePermissionChecker(
			[]string{authorization.ServiceUserRole.String()}, nil)
		require.False(t, ordinary.IsServiceAdmin(),
			"the role this caller holds must not be a service admin, or the test proves nothing")

		// The method grant admits this caller, which is the half of the story a
		// handler-level test would otherwise assume. This application's
		// authorization interceptor is wired and fail-closed — a method absent from
		// the aggregated map is refused outright
		// (internal/services/auth/grpc/interceptors/authn_interceptor.go:315) — so
		// the question is never "is the grant checked" but "what can the grant
		// say". It says settings.values.write, which every self-service user must
		// hold to set any of their own settings, and which this caller holds.
		member := authorization.NewAccountRolePermissionChecker(authorization.AccountMemberPermissions)
		require.True(t, member.HasPermission(authorization.CreateSettingValuesPermission),
			"an ordinary member must hold the value-write grant, or this surface is unusable for self-service")

		callerCtx := sessions.AttachToContext(ctx, &sessions.ContextData{
			ActiveAccountID:    identifiers.New(),
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{},
			Requester: sessions.RequesterInfo{
				UserID:             user.ID,
				ServicePermissions: ordinary,
			},
		})

		response, err := server.SetValue(callerCtx, &settingspb.SetValueRequest{
			Subject: &settingspb.SettingSubject{
				Type: string(platformsettings.SubjectUser),
				Id:   user.ID,
			},
			Name:  name,
			Value: &settingspb.TypedValue{Value: &settingspb.TypedValue_StringValue{StringValue: "on"}},
		})

		// This is the finding. The write goes through.
		require.NoError(t, err, "platform's surface accepted the write")
		require.NotNil(t, response)

		stored, err := store.GetValue(ctx, db.Reader(), ddbsettings.Scope(),
			platformsettings.Subject{Type: platformsettings.SubjectUser, ID: user.ID}, name)
		require.NoError(t, err)
		assert.Equal(t, "on", stored.Raw,
			"a non-admin chose the value of an admin-only setting, which is what AdminOnly exists to prevent")
	})
}
