package settings

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/pointer"
	settings "github.com/primandproper/platform-go/v13/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seededDefinitionsCount is what migration 41 leaves in the catalog:
// user_temperature_unit, re-seeded there from the table this store replaced.
const seededDefinitionsCount = 1

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from, so that a test costs a database
// clone rather than a container start plus a migration replay. See
// pgtesting.RunTestsWithSharedDatabase.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}

// buildDatabaseClientForTest builds the store over a real database.
func buildDatabaseClientForTest(t *testing.T) (settings.Store, audit.Repository, database.SQLQueryExecutor) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	c, err := ProvideSettingsRepository(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		auditLogEntryRepo,
		pgc,
		nil,
	)
	require.NoError(t, err)

	return c, auditLogEntryRepo, pgc.Writer()
}

// subjectForTest creates a user and an account for them, and returns the user.
//
// The account is not incidental: ddb_settings_values has a foreign key to users,
// and the audit entry a value write records is read back through the audit chain.
func subjectForTest(t *testing.T, writer database.SQLQueryExecutor) string {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, writer)
	pgtesting.CreateAccountForTest(t, nil, user.ID, writer)

	return user.ID
}

// definitionForTest adds one setting to the catalog.
func definitionForTest(t *testing.T, ctx context.Context, dbc settings.Store) *settings.Definition {
	t.Helper()

	definition, err := dbc.CreateDefinition(ctx, ddbsettings.Scope(), fakes.BuildFakeSettingDefinition())
	require.NoError(t, err)

	return definition
}

func TestRepository_Integration_SettingDefinitions(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, _ := buildDatabaseClientForTest(t)
	scope := ddbsettings.Scope()

	example := fakes.BuildFakeSettingDefinition()

	created, err := dbc.CreateDefinition(ctx, scope, example)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.ElementsMatch(t, example.Enumeration, created.Enumeration)

	// A definition belongs to nobody, so its entries are recorded under the
	// unattributed actor — the same shape the table this replaced recorded under.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeSettingDefinitions, RelevantID: created.ID},
	})

	fetched, err := dbc.GetDefinition(ctx, scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, example.Name, fetched.Name)
	assert.Equal(t, example.Description, fetched.Description)

	// The name is the handle every value-side call takes, and it finds the same row.
	byName, err := dbc.GetDefinitionByName(ctx, scope, created.Name)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byName.ID)

	// The seeded setting is in the catalog beside this one.
	page, err := dbc.ListDefinitions(ctx, scope, nil)
	require.NoError(t, err)
	require.Len(t, page.Data, seededDefinitionsCount+1)

	fetched.Description = "renamed"
	require.NoError(t, dbc.UpdateDefinition(ctx, scope, fetched))

	updated, err := dbc.GetDefinition(ctx, scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Description)
	assert.NotNil(t, updated.LastUpdatedAt)

	require.NoError(t, dbc.ArchiveDefinition(ctx, scope, created.ID))

	afterArchive, err := dbc.GetDefinition(ctx, scope, created.ID)
	require.Error(t, err)
	assert.Nil(t, afterArchive)
	require.ErrorIs(t, err, settings.ErrDefinitionNotFound)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeSettingDefinitions, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeSettingDefinitions, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeSettingDefinitions, RelevantID: created.ID},
	})
}

// TestRepository_Integration_ArchivingKeepsTheNameClaimed pins that a retired
// setting does not free its name. The values written under it are still stored,
// and a second definition inheriting the name would inherit them.
func TestRepository_Integration_ArchivingKeepsTheNameClaimed(t *testing.T) {
	ctx := t.Context()
	dbc, _, _ := buildDatabaseClientForTest(t)
	scope := ddbsettings.Scope()

	example := fakes.BuildFakeSettingDefinition()

	created, err := dbc.CreateDefinition(ctx, scope, example)
	require.NoError(t, err)

	require.NoError(t, dbc.ArchiveDefinition(ctx, scope, created.ID))

	second := fakes.BuildFakeSettingDefinition()
	second.Name = example.Name

	_, err = dbc.CreateDefinition(ctx, scope, second)
	require.Error(t, err)
	require.ErrorIs(t, err, settings.ErrDefinitionNameTaken)
}

func TestRepository_Integration_SettingValues(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbsettings.Scope()

	userID := subjectForTest(t, writer)
	subject := ddbsettings.SubjectFor(userID)
	definition := definitionForTest(t, ctx, dbc)
	chosen := definition.Enumeration[0]

	value, err := dbc.SetValue(ctx, scope, subject, definition.Name, chosen)
	require.NoError(t, err)
	assert.NotEmpty(t, value.ID)
	assert.Equal(t, chosen, value.Raw)
	assert.Equal(t, subject, value.Subject)

	// The entry belongs to the person whose setting it is, not to the request.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeSettingValues, RelevantID: value.ID},
	})

	fetched, err := dbc.GetValue(ctx, scope, subject, definition.Name)
	require.NoError(t, err)
	assert.Equal(t, value.ID, fetched.ID)

	forSubject, err := dbc.ListValuesForSubject(ctx, scope, subject, nil)
	require.NoError(t, err)
	require.Len(t, forSubject.Data, 1)

	forDefinition, err := dbc.ListValuesForDefinition(ctx, scope, definition.Name, nil)
	require.NoError(t, err)
	require.Len(t, forDefinition.Data, 1)

	// A second answer converges on the same row rather than writing another.
	second := definition.Enumeration[1]

	changed, err := dbc.SetValue(ctx, scope, subject, definition.Name, second)
	require.NoError(t, err)
	assert.Equal(t, value.ID, changed.ID)
	assert.Equal(t, second, changed.Raw)

	still, err := dbc.ListValuesForSubject(ctx, scope, subject, nil)
	require.NoError(t, err)
	require.Len(t, still.Data, 1)

	require.NoError(t, dbc.ClearValue(ctx, scope, subject, definition.Name))

	afterClear, err := dbc.GetValue(ctx, scope, subject, definition.Name)
	require.Error(t, err)
	assert.Nil(t, afterClear)
	require.ErrorIs(t, err, settings.ErrValueNotFound)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeSettingValues, RelevantID: value.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeSettingValues, RelevantID: value.ID},
	})
}

// TestRepository_Integration_ValueOutsideTheEnumerationIsRefused pins the rule a
// hand-rolled pair drifts on: the value is checked against the definition read
// inside the write's own transaction, so there is no window in which an illegal
// one lands.
func TestRepository_Integration_ValueOutsideTheEnumerationIsRefused(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)
	scope := ddbsettings.Scope()

	subject := ddbsettings.SubjectFor(subjectForTest(t, writer))
	definition := definitionForTest(t, ctx, dbc)

	_, err := dbc.SetValue(ctx, scope, subject, definition.Name, "not-in-the-enumeration")
	require.Error(t, err)
	require.ErrorIs(t, err, settings.ErrNotEnumerated)

	// And a value against a setting that does not exist at all is refused for a
	// different reason, which is the other half of the rule.
	_, err = dbc.SetValue(ctx, scope, subject, "no-such-setting", "anything")
	require.Error(t, err)
	require.ErrorIs(t, err, settings.ErrDefinitionNotFound)
}

// TestRepository_Integration_EditRefusesToStrandStoredValues is the guard this
// store owns that the pair it replaced did not.
//
// Narrowing an enumeration decides how every value already written is read.
// Applied, the answer somebody chose would still be in the table and every read
// of it would fail — a setting that works for most people and is broken for the
// ones who picked the value an administrator has just made illegal.
func TestRepository_Integration_EditRefusesToStrandStoredValues(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)
	scope := ddbsettings.Scope()

	subject := ddbsettings.SubjectFor(subjectForTest(t, writer))
	definition := definitionForTest(t, ctx, dbc)
	stranded := definition.Enumeration[1]

	_, err := dbc.SetValue(ctx, scope, subject, definition.Name, stranded)
	require.NoError(t, err)

	narrowed, err := dbc.GetDefinition(ctx, scope, definition.ID)
	require.NoError(t, err)
	narrowed.Enumeration = []string{definition.Enumeration[0]}
	narrowed.Default = pointer.To(definition.Enumeration[0])

	err = dbc.UpdateDefinition(ctx, scope, narrowed)
	require.Error(t, err)
	require.ErrorIs(t, err, settings.ErrStrandedValues)

	// The edit is refused rather than half-applied: the setting still admits the
	// value somebody chose.
	unchanged, err := dbc.GetDefinition(ctx, scope, definition.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, definition.Enumeration, unchanged.Enumeration)
}

// TestRepository_Integration_ResolutionHasThreeAnswers pins the tri-state, which
// is the whole reason a resolution is a value rather than a getter with a
// fallback.
func TestRepository_Integration_ResolutionHasThreeAnswers(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)
	scope := ddbsettings.Scope()

	subject := ddbsettings.SubjectFor(subjectForTest(t, writer))

	// A setting with a default, which nobody has answered.
	defaulted := definitionForTest(t, ctx, dbc)

	// A setting with no default at all, which is the state a plain string column
	// has nowhere to put.
	undefaulted := fakes.BuildFakeSettingDefinition()
	undefaulted.Default = nil
	undefaulted.Enumeration = nil

	created, err := dbc.CreateDefinition(ctx, scope, undefaulted)
	require.NoError(t, err)

	fromDefault, err := dbc.Resolve(ctx, scope, subject, defaulted.Name)
	require.NoError(t, err)
	assert.Equal(t, settings.SourceDefault, fromDefault.Source)
	assert.Equal(t, *defaulted.Default, fromDefault.Raw)

	unset, err := dbc.Resolve(ctx, scope, subject, created.Name)
	require.NoError(t, err)
	assert.Equal(t, settings.SourceUnset, unset.Source)
	assert.Empty(t, unset.Raw)

	// The unset one is an error from a typed read rather than the kind's zero,
	// which is what stops a caller from acting on a decision nobody made.
	_, err = unset.String()
	require.ErrorIs(t, err, settings.ErrSettingUnset)

	_, err = dbc.SetValue(ctx, scope, subject, defaulted.Name, defaulted.Enumeration[1])
	require.NoError(t, err)

	fromSubject, err := dbc.Resolve(ctx, scope, subject, defaulted.Name)
	require.NoError(t, err)
	assert.Equal(t, settings.SourceSubject, fromSubject.Source)
	assert.Equal(t, defaulted.Enumeration[1], fromSubject.Raw)

	// Clearing puts them back on the default rather than leaving them unanswered.
	require.NoError(t, dbc.ClearValue(ctx, scope, subject, defaulted.Name))

	backToDefault, err := dbc.Resolve(ctx, scope, subject, defaulted.Name)
	require.NoError(t, err)
	assert.Equal(t, settings.SourceDefault, backToDefault.Source)

	// ResolveAll answers the whole catalog, the settings nobody has touched
	// included — which is what a preferences page renders.
	all, err := dbc.ResolveAll(ctx, scope, subject)
	require.NoError(t, err)
	assert.Len(t, all, seededDefinitionsCount+2)
}

// TestRepository_Integration_ErasingAUserTakesTheirSettings pins the foreign key
// migration 41 re-creates.
//
// It is what keeps the single identity eraser covering this domain: the table
// this replaced carried belongs_to_user REFERENCES users ON DELETE CASCADE, and
// a preference that outlived the person who chose it would be personal data no
// erasure reaches. See internal/domain/settings for why one subject type is what
// makes the key possible at all.
func TestRepository_Integration_ErasingAUserTakesTheirSettings(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)
	scope := ddbsettings.Scope()

	userID := subjectForTest(t, writer)
	subject := ddbsettings.SubjectFor(userID)
	definition := definitionForTest(t, ctx, dbc)

	_, err := dbc.SetValue(ctx, scope, subject, definition.Name, definition.Enumeration[0])
	require.NoError(t, err)

	_, err = writer.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	require.NoError(t, err)

	gone, err := dbc.ListValuesForSubject(ctx, scope, subject, nil)
	require.NoError(t, err)
	assert.Empty(t, gone.Data)

	// The catalog is untouched: a definition belongs to nobody.
	stillDefined, err := dbc.GetDefinition(ctx, scope, definition.ID)
	require.NoError(t, err)
	assert.Equal(t, definition.ID, stillDefined.ID)
}

// TestRepository_Integration_TheSeededSettingSurvivedTheMigration pins what
// migration 41 carried across: the one setting this application ships with, at
// the id a client may already hold, under the kind platform's store understands.
func TestRepository_Integration_TheSeededSettingSurvivedTheMigration(t *testing.T) {
	ctx := t.Context()
	dbc, _, _ := buildDatabaseClientForTest(t)

	seeded, err := dbc.GetDefinitionByName(ctx, ddbsettings.Scope(), "user_temperature_unit")
	require.NoError(t, err)

	assert.Equal(t, "d6me6i4n9qd3gcf5j1p0", seeded.ID)
	assert.Equal(t, settings.KindString, seeded.Kind)
	require.NotNil(t, seeded.Default)
	assert.Equal(t, "fahrenheit", *seeded.Default)
	// The pipe-delimited column became rows, and they come back sorted.
	assert.Equal(t, []string{"celsius", "fahrenheit"}, seeded.Enumeration)
	assert.False(t, seeded.AdminOnly)
}
