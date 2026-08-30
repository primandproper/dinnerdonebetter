package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v13/database"
	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const eraserO11yName = "identity_privacy_eraser"

// Eraser deletes a subject's user record, and with it everything the schema
// cascades from one.
//
// It is the only application-level eraser registered, and that is a claim worth
// stating rather than leaving to be inferred from the absence of files. Every
// belongs_to_user and belongs_to_account foreign key in this schema carries ON
// DELETE CASCADE, so a domain eraser of its own would issue a DELETE for rows
// Postgres has already removed — eleven statements that can only agree with the
// one that ran first, and eleven places for that agreement to rot.
//
// The two things that make a per-domain eraser worth writing are retention and
// anonymization: data that must be kept under some legal basis, or that a
// foreign key still points at. Neither applies here yet. The moment one does —
// the likeliest candidate being payment records, which tax law generally
// requires be retained for years — that domain registers its own Eraser under
// its own key, reports what it kept and why, and this one stops being the whole
// story. Nothing about this file has to change for that to happen, which is the
// point of registering erasers separately from collectors.
//
// The audit log is already that exception, registered by platform-go rather than
// here. See internal/domain/audit/privacy/eraser.go.
type Eraser struct {
	repo   identity.Repository
	tracer tracing.Tracer
	logger logging.Logger
}

var _ platformdataprivacy.Eraser = (*Eraser)(nil)

// NewEraser builds the identity eraser.
func NewEraser(repo identity.Repository, logger logging.Logger, tracerProvider tracing.Provider) *Eraser {
	return &Eraser{
		repo:   repo,
		tracer: tracing.NewNamedTracer(tracerProvider, eraserO11yName),
		logger: logging.NewNamedLogger(logger, eraserO11yName),
	}
}

// Erase implements platformdataprivacy.Eraser.
//
// It uses the executor it is given rather than a handle of its own, which is
// what makes the erasure atomic across domains: this DELETE, the audit scope
// deletions, and the bookkeeping that records the request completed all commit
// together or not at all.
//
// The Deleted count is user rows, not the cascade's true total. Postgres does
// not report what a cascade removed, and inventing a number by counting every
// affected table before the delete would cost a dozen queries to produce a
// figure nobody can act on. One is the honest answer to "how many subjects were
// erased".
func (e *Eraser) Erase(
	ctx context.Context,
	q database.Tx,
	subject platformdataprivacy.Subject,
) (platformdataprivacy.ErasureOutcome, error) {
	ctx, span := e.tracer.StartSpan(ctx)
	defer span.End()

	logger := e.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, subject.ID)

	deleted, err := e.repo.EraseUser(ctx, q, subject.ID)
	if err != nil {
		return platformdataprivacy.ErasureOutcome{}, observability.PrepareAndLogError(err, logger, span, "erasing user")
	}

	return platformdataprivacy.ErasureOutcome{Deleted: deleted}, nil
}
