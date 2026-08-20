package oauth

import (
	"context"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth/generated"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const (
	o11yName = "oauth_db_client"
)

// repository is the oauth2 client registry's repo implementation.
//
// It holds no cryptography of its own any more. It used to encrypt the code, access and
// refresh columns of oauth2_client_tokens and keep an HMAC blind index beside each; those
// records are the platform authorization server's now, and it stores a SHA-256 digest of
// each credential rather than a reversible copy — so there is nothing here to decrypt.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	readDB            database.SQLQueryExecutor
}

// ProvideOAuthRepository provides a new repository.
func ProvideOAuthRepository(
	_ context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	_ *dbcfg.Config,
	client database.Client,
) oauth.Repository {
	return &repository{
		Client:            client,
		readDB:            client.Reader(),
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}
}
