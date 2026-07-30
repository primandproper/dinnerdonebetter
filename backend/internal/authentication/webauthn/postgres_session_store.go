package webauthn

import (
	"context"
	"database/sql"
	"time"

	"github.com/primandproper/platform-go/v8/database"
	"github.com/primandproper/platform-go/v8/encoding"
	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	o11yName                = "webauthn_session_store"
	postgresCleanupInterval = 5 * time.Minute
)

// PostgresSessionStore is a PostgreSQL-backed session store. Suitable for multi-instance deployments.
type PostgresSessionStore struct {
	client  database.Client
	logger  logging.Logger
	tracer  tracing.Tracer
	encoder encoding.ServerEncoderDecoder
}

// NewPostgresSessionStore creates a new PostgreSQL-backed session store.
func NewPostgresSessionStore(client database.Client, logger logging.Logger, tracerProvider tracing.TracerProvider) *PostgresSessionStore {
	encoder := encoding.NewServerEncoderDecoder(encoding.ContentTypeJSON, encoding.WithLogger(logger), encoding.WithTracerProvider(tracerProvider))
	s := &PostgresSessionStore{
		client:  client,
		logger:  logging.NewNamedLogger(logger, o11yName),
		tracer:  tracing.NewNamedTracer(tracerProvider, o11yName),
		encoder: encoder,
	}
	go s.cleanupLoop()
	return s
}

func (s *PostgresSessionStore) cleanupLoop() {
	ticker := time.NewTicker(postgresCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, span := s.tracer.StartSpan(context.Background())
		db := s.client.Writer()
		_, err := db.ExecContext(ctx, "DELETE FROM webauthn_sessions WHERE expires_at < NOW()")
		span.End()
		if err != nil {
			s.logger.Error("failed to cleanup expired webauthn sessions", err)
		}
	}
}

// SaveSession stores session data keyed by challenge.
func (s *PostgresSessionStore) SaveSession(ctx context.Context, challenge string, session *webauthn.SessionData, ttl time.Duration) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, "webauthn.challenge", challenge)
	logger := s.logger.WithValue("challenge", challenge)

	data := s.encoder.MustEncodeJSON(ctx, session)
	expiresAt := s.client.CurrentTime().Add(ttl)
	db := s.client.Writer()

	_, err := db.ExecContext(ctx, `
		INSERT INTO webauthn_sessions (challenge, session_data, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (challenge) DO UPDATE SET
			session_data = EXCLUDED.session_data,
			expires_at = EXCLUDED.expires_at
	`, challenge, data, expiresAt)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "saving webauthn session")
	}
	return nil
}

// DeleteSession removes the session data for a challenge. Called immediately after a successful
// assertion so the one-time challenge can't be replayed within its TTL.
func (s *PostgresSessionStore) DeleteSession(ctx context.Context, challenge string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, "webauthn.challenge", challenge)
	logger := s.logger.WithValue("challenge", challenge)

	db := s.client.Writer()

	if _, err := db.ExecContext(ctx, "DELETE FROM webauthn_sessions WHERE challenge = $1", challenge); err != nil { // #nosec G701 -- challenge is parameterized via $1
		return observability.PrepareAndLogError(err, logger, span, "deleting webauthn session")
	}

	return nil
}

// GetSession retrieves session data by challenge.
func (s *PostgresSessionStore) GetSession(ctx context.Context, challenge string) (*webauthn.SessionData, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, "webauthn.challenge", challenge)
	logger := s.logger.WithValue("challenge", challenge)

	db := s.client.Writer()

	var data []byte
	var expiresAt time.Time
	err := db.QueryRowContext(ctx, `
		SELECT session_data, expires_at
		FROM webauthn_sessions
		WHERE challenge = $1
	`, challenge).Scan(&data, &expiresAt) // #nosec G701 -- challenge is parameterized via $1
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("webauthn session not found")
			return nil, platformerrors.New("session not found")
		}
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webauthn session")
	}

	if s.client.CurrentTime().After(expiresAt) {
		if _, delErr := db.ExecContext(ctx, "DELETE FROM webauthn_sessions WHERE challenge = $1", challenge); delErr != nil { // #nosec G701 -- challenge is parameterized via $1
			s.logger.Error("failed to delete expired webauthn session", delErr)
		}
		logger.Debug("webauthn session expired")
		return nil, platformerrors.New("session expired")
	}

	var session webauthn.SessionData
	if decodeErr := s.encoder.DecodeBytes(ctx, data, &session); decodeErr != nil {
		return nil, observability.PrepareAndLogError(decodeErr, logger, span, "decoding webauthn session data")
	}

	return &session, nil
}
