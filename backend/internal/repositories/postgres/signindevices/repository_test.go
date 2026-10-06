package signindevices

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/devices"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from. See pgtesting.RunTestsWithSharedDatabase.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}

func buildRepositoryForTest(t *testing.T) (*Repository, database.Client) {
	t.Helper()

	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(t.Context(), config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NoError(t, err)

	return ProvideSignInDevicesRepository(tracingnoop.NewTracerProvider(), pgc), pgc
}

func buildInertRepositoryForTest() *Repository {
	return ProvideSignInDevicesRepository(tracingnoop.NewTracerProvider(), &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
		WriterFunc: func() database.SQLQueryExecutor { return nil },
	})
}

func buildDeviceForTest(userID string, expiresAt time.Time) *devices.Device {
	return &devices.Device{
		FamilyID:   identifiers.New(),
		UserID:     userID,
		IPAddress:  gofakeit.IPv4Address(),
		UserAgent:  gofakeit.UserAgent(),
		DeviceName: gofakeit.Word(),
		ExpiresAt:  expiresAt,
	}
}

func TestRepository_RecordSignInDevice(T *testing.T) {
	T.Parallel()

	T.Run("with no device", func(t *testing.T) {
		t.Parallel()

		assert.ErrorIs(t, buildInertRepositoryForTest().RecordSignInDevice(t.Context(), nil, nil), platformerrors.ErrNilInputParameter)
	})

	T.Run("with no login", func(t *testing.T) {
		t.Parallel()

		device := buildDeviceForTest(identifiers.New(), time.Now())
		device.FamilyID = ""

		assert.ErrorIs(t, buildInertRepositoryForTest().RecordSignInDevice(t.Context(), nil, device), platformerrors.ErrInvalidIDProvided)
	})
}

func TestRepository_Integration(T *testing.T) {
	T.Parallel()

	T.Run("records a device and reads it back for its login", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo, db := buildRepositoryForTest(t)
		user := pgtesting.CreateUserForTest(t, nil, db.Writer())

		device := buildDeviceForTest(user.ID, time.Now().Add(time.Hour))
		other := buildDeviceForTest(user.ID, time.Now().Add(time.Hour))
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), device))
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), other))

		found, err := repo.GetSignInDevicesForFamilies(ctx, user.ID, []string{device.FamilyID, identifiers.New()})
		require.NoError(t, err)
		require.Len(t, found, 1)

		assert.Equal(t, device.FamilyID, found[0].FamilyID)
		assert.Equal(t, device.IPAddress, found[0].IPAddress)
		assert.Equal(t, device.UserAgent, found[0].UserAgent)
		assert.Equal(t, device.DeviceName, found[0].DeviceName)
	})

	T.Run("a refresh renews the login's row rather than adding one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo, db := buildRepositoryForTest(t)
		user := pgtesting.CreateUserForTest(t, nil, db.Writer())

		device := buildDeviceForTest(user.ID, time.Now().Add(time.Hour))
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), device))

		renewed := buildDeviceForTest(user.ID, time.Now().Add(2*time.Hour))
		renewed.FamilyID = device.FamilyID
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), renewed))

		all, err := repo.GetSignInDevicesForUser(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, all, 1)

		assert.Equal(t, renewed.IPAddress, all[0].IPAddress)
		assert.Equal(t, renewed.UserAgent, all[0].UserAgent)
	})

	T.Run("does not answer one person's login to another", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo, db := buildRepositoryForTest(t)
		owner := pgtesting.CreateUserForTest(t, nil, db.Writer())
		stranger := pgtesting.CreateUserForTest(t, nil, db.Writer())

		device := buildDeviceForTest(owner.ID, time.Now().Add(time.Hour))
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), device))

		// Nor does a write naming somebody else move the row to them.
		moved := buildDeviceForTest(stranger.ID, time.Now().Add(time.Hour))
		moved.FamilyID = device.FamilyID
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), moved))

		found, err := repo.GetSignInDevicesForFamilies(ctx, stranger.ID, []string{device.FamilyID})
		require.NoError(t, err)
		assert.Empty(t, found)

		kept, err := repo.GetSignInDevicesForFamilies(ctx, owner.ID, []string{device.FamilyID})
		require.NoError(t, err)
		require.Len(t, kept, 1)
		assert.Equal(t, device.IPAddress, kept[0].IPAddress)
	})

	T.Run("sweeps the devices of logins that can no longer be alive", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo, db := buildRepositoryForTest(t)
		user := pgtesting.CreateUserForTest(t, nil, db.Writer())

		live := buildDeviceForTest(user.ID, time.Now().Add(time.Hour))
		lapsed := buildDeviceForTest(user.ID, time.Now().Add(-time.Hour))
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), live))
		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), lapsed))

		swept, err := repo.Sweep(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), swept)

		left, err := repo.GetSignInDevicesForUser(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, left, 1)
		assert.Equal(t, live.FamilyID, left[0].FamilyID)
	})

	T.Run("erasing a user takes their devices", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo, db := buildRepositoryForTest(t)
		user := pgtesting.CreateUserForTest(t, nil, db.Writer())

		require.NoError(t, repo.RecordSignInDevice(ctx, db.Writer(), buildDeviceForTest(user.ID, time.Now().Add(time.Hour))))

		_, err := db.Writer().ExecContext(ctx, "DELETE FROM ddb_identity_users WHERE id = $1", user.ID)
		require.NoError(t, err)

		left, err := repo.GetSignInDevicesForUser(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, left)
	})
}
