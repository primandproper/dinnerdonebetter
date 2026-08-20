package audit

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v12/observability/logging"
)

func BuildDataChangeMessageFromContext(ctx context.Context, logger logging.Logger, eventType string, metadata map[string]any) *DataChangeMessage {
	sessionContext := sessions.FromContext(ctx)
	if sessionContext == nil {
		logger.WithValue("event_type", eventType).Info("failed to extract session data from context")
	}

	// The getters are nil-safe, so an absent session yields empty attribution rather than a panic.
	return &DataChangeMessage{
		EventType: eventType,
		Context:   metadata,
		UserID:    sessionContext.GetUserID(),
		AccountID: sessionContext.GetActiveAccountID(),
	}
}
