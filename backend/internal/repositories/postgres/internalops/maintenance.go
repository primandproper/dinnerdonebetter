package internalops

import (
	"context"

	"github.com/primandproper/platform-go/v7/observability"
)

// DeleteExpiredOAuth2ClientTokens deletes expired oauth2 client tokens.
func (q *repository) DeleteExpiredOAuth2ClientTokens(ctx context.Context) (int64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	deleted, err := q.generatedQuerier.DeleteExpiredOAuth2ClientTokens(ctx, q.writeDB)
	if err != nil {
		return 0, observability.PrepareError(err, span, "deleting expired oauth2 client tokens")
	}

	q.logger.Info("deleted expired oauth2 client tokens")

	return deleted, nil
}
