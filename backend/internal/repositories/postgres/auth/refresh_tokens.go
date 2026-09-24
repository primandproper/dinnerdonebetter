package auth

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v14/authentication/signin/refreshtokens"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	refreshTokensO11yName = "refresh_token_db_client"
)

// ProvideRefreshTokenSQLStore builds the platform's refresh token store over this
// deployment's database: where a sign-in made through platform's SignInService keeps the
// rotating credential that outlives its access token.
//
// It is the store itself rather than the signin.RefreshTokenStore seam, for the reason the
// password reset store is: the db-cleaner job wants Sweep, which the seam does not carry.
//
// No sweeper goroutine is started, which is the call every other token table here makes —
// one scheduled sweep for the deployment rather than one per replica. See
// services/oauth/workers/db_cleaner.
//
// There is no audited wrapper. The events worth recording about a sign-in are recorded by
// the sign-in hooks in internal/authentication, inside the same transaction as the mint;
// an entry per exchange would record every hour of every session and say nothing an
// investigation asks for. A detected reuse is the exception worth a record, and platform
// does not yet say one happened anywhere but in the error it returns.
func ProvideRefreshTokenSQLStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
) (*refreshtokens.SQLStore, error) {
	return refreshtokens.NewSQLStore(
		&refreshtokens.Config{TablePrefix: auth.TablePrefix},
		client,
		refreshtokens.WithLogger(logging.NewNamedLogger(logger, refreshTokensO11yName)),
		refreshtokens.WithTracerProvider(tracerProvider),
		refreshtokens.WithMetricsProvider(metricsProvider),
	)
}
