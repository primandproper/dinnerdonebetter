// Package privacy is the webhooks domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	platformwebhooks "github.com/primandproper/platform-go/v14/webhooks"
	"github.com/primandproper/primitives-go/v2/database"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const o11yName = "webhooks_privacy_collector"

// Collector collects webhook data about a subject.
//
// It reads platform's endpoint store directly rather than through a repository
// of this application's, because there is no longer one: the endpoints are
// platform's and the three local tables are gone. platform ships a privacy
// package for eight of its nine domains and not for webhooks, with no stated
// reason, so this stays here until that is settled — see
// docs/v14-port-completion-findings.md, finding H.
type Collector struct {
	endpoints       platformwebhooks.Store
	db              database.Client
	resolveAccounts dataprivacy.AccountIDResolver
	tracer          tracing.Tracer
	logger          logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the webhooks collector.
func NewCollector(
	endpoints platformwebhooks.Store,
	db database.Client,
	resolveAccounts dataprivacy.AccountIDResolver,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) *Collector {
	return &Collector{
		endpoints:       endpoints,
		db:              db,
		resolveAccounts: resolveAccounts,
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:          logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
//
// Webhooks belong to an account rather than a person, so they are keyed by
// account in the fragment. A subject in three accounts gets three groups, which
// is the only rendering that lets them tell which endpoint belongs to which.
// The request scope is not consulted. This application's subject access
// requests name a person rather than a tenant, and the rows below are
// reached by subject id; narrowing to one scope would under-report.
func (c *Collector) Collect(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithSpan(span)

	accountIDs, err := c.resolveAccounts(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "resolving accounts")
	}

	collection := UserDataCollection{Data: map[string][]platformwebhooks.Endpoint{}}

	for _, accountID := range accountIDs {
		hooks, hooksErr := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformwebhooks.Endpoint], error) {
			return c.endpoints.ListEndpoints(ctx, c.db.Reader(), tenancy.Of(accountID), filter)
		})
		if hooksErr != nil {
			return nil, observability.PrepareAndLogError(hooksErr, logger, span, "fetching webhooks for account")
		}

		if len(hooks) > 0 {
			collection.Data[accountID] = hooks
		}
	}

	return platformdataprivacy.Fragment(len(collection.Data) > 0, &collection)
}

// UserDataCollection is what this domain contributes to a subject access
// artifact: the endpoints in each account the subject belongs to, keyed by
// account.
//
// It is declared here rather than in internal/domain/webhooks because the type
// it holds is platform's now, and the domain package no longer has a webhook of
// its own to collect.
//
// The signing secret does not travel. platform's Endpoint carries one and tags
// it `json:"-"`, so it is absent from the artifact by the field's own
// declaration rather than by anything this package remembers to do — which is
// the right place for it, since this collector marshals the whole endpoint.
type UserDataCollection struct {
	_ struct{} `json:"-"`

	Data map[string][]platformwebhooks.Endpoint `json:"data"`
}
