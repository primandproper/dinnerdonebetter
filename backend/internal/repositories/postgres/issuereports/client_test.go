package issue_reports

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v7/database"
	mockdatabase "github.com/primandproper/platform-go/v7/database/mock"
	"github.com/primandproper/platform-go/v7/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3
)

func buildDatabaseClientForTest(t *testing.T) (*repository, audit.Repository) {
	t.Helper()

	ctx := t.Context()
	db, config := pgtesting.BuildDatabaseContainerForTest(t)
	migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
	require.NoError(t, err)
	require.NoError(t, migrator.Migrate(ctx, db))

	pgc, err := postgres.NewDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	c := ProvideIssueReportsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	require.NoError(t, err)

	return c.(*repository), auditLogEntryRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideIssueReportsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }})

	return c.(*repository)
}
