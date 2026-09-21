package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Neither of these is a constant this package otherwise has, precisely because no builder
// here generates queries against either table — the session table never had one, and the
// identity adoption took the user tables to platform and deleted the builders that did.
// That is what put both among the tables the old registry-derived TRUNCATE missed.
const (
	usersTableName    = "users"
	sessionsTableName = "sessions"
)

func Test_buildMaintenanceQueries(T *testing.T) {
	T.Parallel()

	T.Run("truncates whatever the schema holds rather than a list fixed at generation time", func(t *testing.T) {
		t.Parallel()

		queries := buildMaintenanceQueries(postgres)
		require.NotEmpty(t, queries)

		destroy := queries[0]
		require.Equal(t, "DestroyAllData", destroy.Annotation.Name)

		assert.Contains(t, destroy.Content, "FROM pg_tables")
		assert.Contains(t, destroy.Content, "schemaname = 'public'")
		// The migrator's bookkeeping is the one table left standing.
		assert.Contains(t, destroy.Content, "tablename <> 'goose_db_version'")

		// The regression this guards: a table named here is a table somebody has to
		// remember to add, which is how eleven of them went missing. Nothing else in this
		// statement is a table name, so finding one means the list came back.
		for _, table := range []string{
			usersTableName,
			sessionsTableName,
			recipesTableName,
			mealsTableName,
		} {
			assert.NotContains(t, destroy.Content, table)
		}
	})

	T.Run("with unsupported database", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, buildMaintenanceQueries("mariadb"))
	})
}
