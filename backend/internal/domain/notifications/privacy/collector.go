// Package privacy is the notifications domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const o11yName = "notifications_privacy_collector"

// Collector collects in-app notifications sent to a subject.
type Collector struct {
	repo   notifications.Repository
	tracer tracing.Tracer
	logger logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the notifications collector.
func NewCollector(repo notifications.Repository, logger logging.Logger, tracerProvider tracing.Provider) *Collector {
	return &Collector{
		repo:   repo,
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
//
// One read, so this looks like a platformdataprivacy.CollectorFor and is not.
// That constructor encodes the rows themselves; this section is a
// notifications.UserDataCollection wrapping them, which is the shape the gRPC
// converters read and so is not this collector's to flatten.
// The request scope is not consulted. This application's subject access
// requests name a person rather than a tenant, and the rows below are
// reached by subject id; narrowing to one scope would under-report.
func (c *Collector) Collect(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	notifs, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[notifications.UserNotification], error) {
		return c.repo.GetUserNotifications(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, c.logger.WithSpan(span), span, "fetching user notifications")
	}

	return platformdataprivacy.Fragment(len(notifs) > 0, &notifications.UserDataCollection{Data: notifs})
}
