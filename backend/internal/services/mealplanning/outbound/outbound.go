/*
Package outbound turns the meal planning domain's events into the mail they imply.

It is the domain's entry in the async data change handler's list of outbound notification
handlers (internal/functions/datachangemessagehandler), and the one place this domain decides
which of its events somebody is told about. Today that is one: a meal plan created by a person
tells every verified member of the household. The handler ranges over the list and names no
domain; what it knows of this one is the entry internal/build hands it.
*/
package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningemails "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/emails"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/samber/do/v2"
)

const (
	o11yName = "mealplanning_outbound_notifications"

	// emailTypeMealPlanCreated labels the mail in the handler's logs.
	emailTypeMealPlanCreated = "meal plan created"
)

// Notifier builds the mail this domain's events imply.
type Notifier struct {
	tracer    tracing.Tracer
	logger    logging.Logger
	repo      mealplanning.Repository
	directory platformidentity.Store
	db        database.Client
	baseURL   string
}

// NewNotifier builds a Notifier reading meal plans through repo and the household through
// directory. baseURL is the public address the mail's links point into.
func NewNotifier(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	repo mealplanning.Repository,
	directory platformidentity.Store,
	db database.Client,
	baseURL string,
) (*Notifier, error) {
	switch {
	case repo == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil meal planning repository")
	case directory == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil identity store")
	case db == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client")
	}

	return &Notifier{
		tracer:    tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:    logging.NewNamedLogger(logger, o11yName),
		repo:      repo,
		directory: directory,
		db:        db,
		baseURL:   baseURL,
	}, nil
}

// Register provides the Notifier, over the repository and the identity store the process
// registered and the base URL the handler's config names.
func Register(i do.Injector) {
	do.Provide[*Notifier](i, func(i do.Injector) (*Notifier, error) {
		return NewNotifier(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*config.AsyncMessageHandlerConfig](i).BaseURL,
		)
	})
}

// Handle answers the handler's question for one event: whether it is this domain's to notify
// about, and if so the mail to send. It has the shape of
// datachangemessagehandler.OutboundNotificationHandler.
func (n *Notifier) Handle(ctx context.Context, event *webhooks.Envelope) (handled bool, emailType string, emails []*queuemessages.OutboundEmailMessage, err error) {
	if event == nil || event.EventType != webhooks.EventType(mealplanning.MealPlanCreatedServiceEventType) {
		return false, "", nil, nil
	}

	var changeMessage datachanges.Message
	if err = json.Unmarshal(event.Payload, &changeMessage); err != nil {
		return true, emailTypeMealPlanCreated, nil, fmt.Errorf("decoding the payload of a %q event: %w", event.EventType, err)
	}

	// A meal plan nobody made — a background job's — announces itself to nobody.
	if changeMessage.UserID == "" {
		return true, emailTypeMealPlanCreated, nil, nil
	}

	emails, err = n.mealPlanCreated(ctx, &changeMessage)
	if err != nil {
		return true, emailTypeMealPlanCreated, nil, err
	}

	return true, emailTypeMealPlanCreated, emails, nil
}

// mealPlanCreated builds one mail per verified member of the household the plan belongs to.
func (n *Notifier) mealPlanCreated(ctx context.Context, changeMessage *datachanges.Message) ([]*queuemessages.OutboundEmailMessage, error) {
	ctx, span := n.tracer.StartSpan(ctx)
	defer span.End()

	logger := n.logger.WithValue("event_type", changeMessage.EventType)

	mealPlanID, ok := changeMessage.Context[mealplanningkeys.MealPlanIDKey].(string)
	if !ok || mealPlanID == "" || changeMessage.AccountID == "" {
		return nil, observability.PrepareError(fmt.Errorf("meal plan created event requires meal_plan.id and accountID in context"), span, "publishing meal plan created email")
	}

	mealPlan, err := n.repo.GetMealPlan(ctx, mealPlanID, changeMessage.AccountID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting meal plan for created email")
	}

	if mealPlan == nil {
		return nil, observability.PrepareError(fmt.Errorf("meal plan is nil"), span, "publishing meal plan created email")
	}

	// The roster, paged to the end. The read this replaced answered with an Account
	// carrying its whole membership list, so a household past the first page would have
	// had its later members silently left off the mailing.
	members, err := ddbidentity.MembersOfAccount(ctx, n.directory, n.db.Reader(), mealPlan.BelongsToAccount)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting account members")
	}

	users, err := n.directory.ListUsersByIDs(ctx, n.db.Reader(), tenancy.Global(), members)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting account members")
	}

	var outbound []*queuemessages.OutboundEmailMessage

	for _, member := range users {
		if member.EmailAddressVerifiedAt == nil {
			continue
		}

		msg, emailErr := mealplanningemails.BuildMealPlanCreatedEmail(member, mealPlan, n.baseURL)
		if emailErr != nil {
			return nil, observability.PrepareAndLogError(emailErr, logger, span, "building meal plan created email")
		}

		outbound = append(outbound, msg)
	}

	return outbound, nil
}
