package auth

import (
	"context"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/sessions"
	sessionsdatabase "github.com/primandproper/platform-go/v13/sessions/database"
)

const (
	sessionsO11yName = "user_session_store"

	// resourceTypeUserSessions is the audit log's name for a session, which is this
	// application's vocabulary rather than the table's — the table is ddb_sessions. It is
	// unchanged from what the store this replaced recorded, so an investigation reading
	// across the changeover reads one resource type rather than two.
	resourceTypeUserSessions = "user_sessions"
)

var _ auth.SessionStore = (*auditedUserSessionStore)(nil)

// auditedUserSessionStore is platform-go's session store with this repository's audit log
// wrapped around establishment and revocation.
//
// The store itself is the platform's, whole. What it replaces here was a table of its own
// keyed by the JTI of the token issued alongside a session, with a revoked_at column that
// only the reads knew to filter on — which is to say a second account of which sessions
// were live, kept beside the one the sign-in wrote, and wrong the moment they disagreed.
// The platform's row *is* the session: a revocation removes it, and there is nothing left
// for a read to be read past.
//
// What the platform has no opinion about is who wants a record of it. "When did this
// account sign in, from what, and who ended those sessions" is a question asked months
// after the rows are gone — the db-cleaner deletes an expired session, and a revoked one
// never reaches the sweep at all — so the two facts are written somewhere that keeps them.
//
// Get and Save are the embedded store's, unrecorded. Get happens on every authenticated
// request, and an entry per call would bury every entry that means something; Save is the
// token rotation that follows a refresh, which is bookkeeping about a session rather than
// a thing that happened to one.
type auditedUserSessionStore struct {
	auth.SessionStore

	db                database.Client
	auditLogEntryRepo audit.Repository
	tracer            tracing.Tracer
	logger            logging.Logger
}

// ProvideUserSessionBackend builds the platform's SQL session backend over this
// deployment's database.
//
// It is the concrete backend rather than the sessions.Backend seam, because the db-cleaner
// job needs Sweep and that is deliberately not on the interface — a caller who chose the
// cache has nothing to sweep.
//
// No sweeper goroutine is started, which is the same call the authorization server's and
// the password reset store's tables make: one scheduled sweep for the deployment rather
// than one per replica, each running the same full-table delete on its own timer. See
// services/oauth/workers/db_cleaner.
//
// The prefix is the domain's rather than a configured one, because it has to be the prefix
// migration 35 created the table under — a prefix that differs between the DDL and the
// store is a service that comes up clean and cannot find a table.
func ProvideUserSessionBackend(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
) (*sessionsdatabase.Backend[auth.SessionPayload], error) {
	return sessionsdatabase.NewBackend[auth.SessionPayload](
		&sessionsdatabase.Config{TablePrefix: auth.TablePrefix},
		client,
		sessionsdatabase.WithLogger(logging.NewNamedLogger(logger, sessionsO11yName)),
		sessionsdatabase.WithTracerProvider(tracerProvider),
		sessionsdatabase.WithMetricsProvider(metricsProvider),
	)
}

// ProvideUserSessionStore builds the session store the API server uses: the platform's over
// the table above, with this repository's audit log around it.
//
// The expiry policy is the configuration's, and a zero field is left off rather than passed
// as zero, because zero is meaningful to two of the three options — a zero timeout disables
// that timeout, and a zero touch interval refreshes the idle deadline on every read. Left
// off, the store applies the defaults sessions.Policy documents.
func ProvideUserSessionStore(
	cfg *authcfg.SessionsConfig,
	backend *sessionsdatabase.Backend[auth.SessionPayload],
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
) (auth.SessionStore, error) {
	namedLogger := logging.NewNamedLogger(logger, sessionsO11yName)

	opts := []sessions.Option{
		sessions.WithLogger(namedLogger),
		sessions.WithTracerProvider(tracerProvider),
		sessions.WithMetricsProvider(metricsProvider),
	}

	if cfg.AbsoluteTimeout > 0 {
		opts = append(opts, sessions.WithAbsoluteTimeout(cfg.AbsoluteTimeout))
	}

	if cfg.IdleTimeout > 0 {
		opts = append(opts, sessions.WithIdleTimeout(cfg.IdleTimeout))
	}

	if cfg.TouchInterval > 0 {
		opts = append(opts, sessions.WithTouchInterval(cfg.TouchInterval))
	}

	store, err := sessions.NewStore[auth.SessionPayload](backend, opts...)
	if err != nil {
		return nil, observability.PrepareError(err, nil, "building the user session store")
	}

	return &auditedUserSessionStore{
		SessionStore:      store,
		db:                client,
		auditLogEntryRepo: auditLogEntryRepo,
		tracer:            tracing.NewNamedTracer(tracerProvider, sessionsO11yName),
		logger:            namedLogger,
	}, nil
}

// NewFor establishes a session and records the sign-in that established it.
func (s *auditedUserSessionStore) NewFor(
	ctx context.Context,
	holder sessions.Holder,
	metadata sessions.Metadata,
	data *auth.SessionPayload,
) (*auth.UserSession, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	session, err := s.SessionStore.NewFor(ctx, holder, metadata, data)
	if err != nil {
		return nil, err
	}

	if err = s.record(ctx, span, holder.Principal, audit.AuditLogEventTypeCreated, session.ID); err != nil {
		return nil, err
	}

	return session, nil
}

// Revoke ends one of a holder's sessions and records that it was ended.
func (s *auditedUserSessionStore) Revoke(ctx context.Context, holder sessions.Holder, id string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if err := s.SessionStore.Revoke(ctx, holder, id); err != nil {
		return err
	}

	return s.record(ctx, span, holder.Principal, audit.AuditLogEventTypeArchived, id)
}

// RevokeAll ends every session a holder holds and records each one that ended.
func (s *auditedUserSessionStore) RevokeAll(ctx context.Context, holder sessions.Holder) (int, error) {
	return s.revokeMany(ctx, holder, "")
}

// RevokeAllExcept ends every session a holder holds but one, and records each one that
// ended.
func (s *auditedUserSessionStore) RevokeAllExcept(ctx context.Context, holder sessions.Holder, keepID string) (int, error) {
	return s.revokeMany(ctx, holder, keepID)
}

// revokeMany is the body both bulk revocations share.
//
// The identifiers are read before the revocation rather than reported by it, because the
// store reports a count and an audit entry naming no session says only that something
// happened. The read is the same indexed one the security page makes, over a set that is a
// handful of rows.
//
// Reading first means a session established between the read and the revocation is ended
// without an entry — the only direction this can be wrong, since the revocation is what
// runs second. That is the same trade the record below makes, for the same reason: an
// audit log that occasionally misses an event is worth more than one that reports events
// which did not occur.
func (s *auditedUserSessionStore) revokeMany(ctx context.Context, holder sessions.Holder, keepID string) (int, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	listed, err := s.List(ctx, holder, keepID)
	if err != nil {
		return 0, observability.PrepareAndLogError(err, s.logger, span, "listing sessions ahead of revocation")
	}

	revoked, err := s.SessionStore.RevokeAllExcept(ctx, holder, keepID)
	if err != nil {
		return 0, err
	}

	for _, session := range listed {
		if session.ID == keepID {
			continue
		}

		if err = s.record(ctx, span, holder.Principal, audit.AuditLogEventTypeArchived, session.ID); err != nil {
			return revoked, err
		}
	}

	return revoked, nil
}

// record writes one audit entry for a session, in a transaction of its own.
//
// Of its own, because the write it describes has already committed inside the platform
// store, whose transaction nothing outside that package can join. So the pair is not
// atomic, and the gap has a direction: a crash between them loses the entry, never the
// session change. That is the right way round.
//
// A failure to record is returned rather than swallowed. A sign-in or a sign-out the log
// has no record of is precisely the one an investigation needs, so refusing the operation
// is better than completing it silently — the caller signs in again, or presses the button
// a second time.
func (s *auditedUserSessionStore) record(ctx context.Context, span tracing.Span, userID, eventType, sessionID string) error {
	tracing.AttachToSpan(span, authkeys.UserSessionIDKey, sessionID)

	if err := s.db.WithTransaction(ctx, func(tx database.Tx) error {
		return s.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUserSessions,
			RelevantID:    sessionID,
			EventType:     eventType,
			BelongsToUser: userID,
		})
	}); err != nil {
		return observability.PrepareAndLogError(err, s.logger, span, "recording user session audit log entry")
	}

	return nil
}
