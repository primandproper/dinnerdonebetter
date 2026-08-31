/*
Package push turns a set of user IDs into pushes actually delivered to devices.

It is the half of mobile notifications that has nothing to do with why the
notification exists. Two things decide that a push is owed — a message arriving
on the mobile notifications topic, and a meal plan task claimed from the work
queue — and both then need the same three steps: read the recipients' device
tokens, send to each, and archive the tokens Apple says are dead.

Those steps used to live as methods on the async message handler, which was the
only process that pushed. Once a second process needed them, leaving them there
would have meant either a second copy or routing scheduled work through a topic
to reach them, so they are a component instead. Nothing here knows what the
notification is about.
*/
package push

import (
	"context"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/platform-go/v13/filtering"
	platformnotifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const o11yName = "push_fanout"

// badDeviceTokenMarker is what APNs calls a token it will never accept again. It is matched in
// the error's text because that is the only form the sender surfaces it in.
const badDeviceTokenMarker = "BadDeviceToken"

const (
	statusSuccess = "success"
	statusFailure = "failure"
)

// tokenReader is the slice of the notifications data manager this package uses: the device
// tokens a user has registered, and the ability to retire one the platform has rejected.
type tokenReader interface {
	GetUserDeviceTokens(ctx context.Context, userID string, filter *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[notifications.UserDeviceToken], error)
	ArchiveUserDeviceToken(ctx context.Context, userID, tokenID string) error
}

// Fanout sends one message to every device belonging to a set of users.
type Fanout struct {
	logger logging.Logger
	tokens tokenReader
	sender platformnotifications.PushNotificationSender

	sentCounter           metrics.Int64Counter
	tokensArchivedCounter metrics.Int64Counter
}

// NewFanout builds a Fanout.
func NewFanout(
	logger logging.Logger,
	tokens tokenReader,
	sender platformnotifications.PushNotificationSender,
	metricsProvider metrics.Provider,
) (*Fanout, error) {
	sentCounter, err := metricsProvider.NewInt64Counter("push_notifications_sent_total")
	if err != nil {
		return nil, err
	}

	tokensArchivedCounter, err := metricsProvider.NewInt64Counter("bad_device_tokens_archived_total")
	if err != nil {
		return nil, err
	}

	return &Fanout{
		logger:                logging.NewNamedLogger(logger, o11yName),
		tokens:                tokens,
		sender:                sender,
		sentCounter:           sentCounter,
		tokensArchivedCounter: tokensArchivedCounter,
	}, nil
}

// Result reports what one fan-out reached.
//
// Devices and Delivered are separate numbers because the two ways of reaching nobody are not the
// same fact and a caller has to tell them apart. Nought devices means the recipients have never
// registered one, which no amount of retrying changes; devices but nothing delivered means every
// send failed, which retrying usually does change.
type Result struct {
	// Devices counts the registered devices this fan-out tried to reach.
	Devices int
	// Delivered counts how many of them accepted the push.
	Delivered int
}

// Reached reports whether at least one device took the message.
func (r Result) Reached() bool { return r.Delivered > 0 }

// Unreachable reports that the recipients have no device to push to at all — as distinct from
// having devices that all refused.
func (r Result) Unreachable() bool { return r.Devices == 0 }

// Send delivers msg to every device registered to any of userIDs.
//
// A send that fails for one device does not stop the rest: a user with a stale iPad and a live
// phone should still get the notification on the phone. What the caller does about a partial or
// total failure is the caller's, which is why the counts come back rather than a verdict.
//
// Recipients with no registered devices are not an error. Somebody who has never opened the app
// on a phone is owed nothing, so that is an empty Result rather than a failure to retry.
func (f *Fanout) Send(ctx context.Context, requestType string, userIDs []string, msg platformnotifications.PushMessage) (Result, error) {
	if len(userIDs) == 0 {
		return Result{}, nil
	}

	deviceTokens, err := f.collect(ctx, userIDs)
	if err != nil {
		return Result{}, err
	}

	logger := f.logger.WithValue("recipient_count", len(userIDs)).WithValue("device_count", len(deviceTokens))

	result := Result{Devices: len(deviceTokens)}
	for _, token := range deviceTokens {
		if f.send(ctx, logger, requestType, token, msg) {
			result.Delivered++
		}
	}

	return result, nil
}

// collect gathers every device token registered to any of userIDs.
func (f *Fanout) collect(ctx context.Context, userIDs []string) ([]*notifications.UserDeviceToken, error) {
	var tokens []*notifications.UserDeviceToken

	filter := filtering.DefaultQueryFilter()
	for _, userID := range userIDs {
		result, err := f.tokens.GetUserDeviceTokens(ctx, userID, filter, nil)
		if err != nil {
			return nil, err
		}

		for _, token := range result.Data {
			if token != nil && token.DeviceToken != "" {
				tokens = append(tokens, token)
			}
		}
	}

	return tokens, nil
}

// send pushes to one device, archiving the token if the platform says it is dead.
func (f *Fanout) send(
	ctx context.Context,
	logger logging.Logger,
	requestType string,
	token *notifications.UserDeviceToken,
	msg platformnotifications.PushMessage,
) bool {
	sendErr := f.sender.SendPush(ctx, token.Platform, token.DeviceToken, msg)
	if sendErr == nil {
		f.sentCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("request_type", requestType),
			attribute.String("status", statusSuccess),
		))

		return true
	}

	l := logger.WithValue("user_device_token_id", token.ID)
	l.WithValue("error", sendErr).Error("sending push notification to device", sendErr)
	f.sentCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("request_type", requestType),
		attribute.String("status", statusFailure),
	))

	// The token is gone for good, not failing transiently, so it is retired rather than
	// retried. Left in place it costs a doomed send on every future notification to this user.
	if strings.Contains(sendErr.Error(), badDeviceTokenMarker) {
		if archiveErr := f.tokens.ArchiveUserDeviceToken(ctx, token.BelongsToUser, token.ID); archiveErr != nil {
			l.Error("archiving invalid device token", archiveErr)
		} else {
			f.tokensArchivedCounter.Add(ctx, 1)
			l.Info("archived invalid device token (BadDeviceToken from APNs)")
		}
	}

	return false
}
