package emaildeliverabilitytest

import (
	"context"
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	"github.com/primandproper/platform-go/v11/email"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

const (
	serviceName = "email_deliverability_test"
)

// JobParams holds configuration for the email deliverability test job.
type JobParams struct {
	RecipientEmailAddress string
	ServiceEnvironment    string
}

// Job sends a test email to verify deliverability.
type Job struct {
	emailer        email.Emailer
	logger         logging.Logger
	tracer         tracing.Tracer
	recipientEmail string
}

// NewJob creates a new email deliverability test job.
func NewJob(
	emailer email.Emailer,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	params *JobParams,
) (*Job, error) {
	recipientEmail := params.RecipientEmailAddress
	if recipientEmail == "" {
		return nil, fmt.Errorf("recipient email is required")
	}

	return &Job{
		emailer:        emailer,
		logger:         logging.NewNamedLogger(logger, serviceName),
		tracer:         tracing.NewNamedTracer(tracerProvider, serviceName),
		recipientEmail: recipientEmail,
	}, nil
}

// Do sends the deliverability test email.
func (j *Job) Do(ctx context.Context) error {
	ctx, span := j.tracer.StartSpan(ctx)
	defer span.End()

	sentAt := time.Now().UTC().Format(time.RFC3339)

	msg := &queuemessages.OutboundEmailMessage{
		OutboundEmailMessage: email.OutboundEmailMessage{
			ToAddress:   j.recipientEmail,
			FromAddress: branding.FromEmail,
			FromName:    branding.CompanyName,
			Subject:     fmt.Sprintf("%s – Email Deliverability Test", branding.CompanyName),
			HTMLContent: fmt.Sprintf("<p>Test sent at %s</p>", sentAt),
		},
	}

	if err := j.emailer.SendEmail(ctx, &msg.OutboundEmailMessage); err != nil {
		j.logger.Error("sending deliverability test email", err)
		return fmt.Errorf("sending email: %w", err)
	}

	j.logger.Info("deliverability test email sent successfully")
	return nil
}
