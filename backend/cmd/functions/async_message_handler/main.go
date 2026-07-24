package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	datachangemessagehandlerbuild "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/functions/data_change_message_handler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

var (
	nonWebhookEventTypes = []string{
		identity.UserSignedUpServiceEventType,
		identity.UserArchivedServiceEventType,
		identity.TwoFactorSecretVerifiedServiceEventType,
		identity.TwoFactorDeactivatedServiceEventType,
		identity.TwoFactorSecretChangedServiceEventType,
		identity.PasswordResetTokenCreatedEventType,
		identity.PasswordResetTokenRedeemedEventType,
		identity.PasswordChangedEventType,
		identity.EmailAddressChangedEventType,
		identity.UsernameChangedEventType,
		identity.UserDetailsChangedEventType,
		identity.UsernameReminderRequestedEventType,
		identity.UserLoggedInServiceEventType,
		identity.UserLoggedOutServiceEventType,
		identity.UserChangedActiveAccountServiceEventType,
		identity.UserEmailAddressVerifiedEventType,
		identity.UserEmailAddressVerificationEmailRequestedEventType,
		identity.AccountInvitationAcceptedServiceEventType,
		identity.AccountMemberRemovedServiceEventType,
		identity.AccountMembershipPermissionsUpdatedServiceEventType,
		identity.AccountOwnershipTransferredServiceEventType,
		oauth.OAuth2ClientCreatedServiceEventType,
		oauth.OAuth2ClientArchivedServiceEventType,
	}
)

func main() {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.AsyncMessageHandlerConfig]()
	if err != nil {
		log.Fatalf("error getting config: %v", err)
	}
	cfg.Database.RunMigrations = false

	// run owns every defer (telemetry flush, drain cancel), so a fatal consume error can
	// exit non-zero here without skipping them.
	if err = run(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg *config.AsyncMessageHandlerConfig) error {
	injector := datachangemessagehandlerbuild.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so a short-lived pod exports its spans/metrics before terminating.
	defer telemetry.Flush(ctx, injector)

	dataChangeMessageHandler := do.MustInvoke[*datachangemessagehandler.AsyncDataChangeMessageHandler](injector)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	stopChan := make(chan bool)
	errorsChan := make(chan error)

	dataChangeMessageHandler.SetNonWebhookEventTypes(nonWebhookEventTypes)

	if err := dataChangeMessageHandler.ConsumeMessages(ctx, stopChan, errorsChan); err != nil {
		return err
	}

	// Block until the first shutdown signal arrives.
	<-signalChan

	// Closing stopChan broadcasts to every consumer; then wait for in-flight handlers to drain
	// before the deferred telemetry flush runs and main returns.
	close(stopChan)

	drainCtx, cancelDrain := context.WithTimeout(ctx, 30*time.Second)
	defer cancelDrain()
	dataChangeMessageHandler.WaitForConsumers(drainCtx)

	return nil
}
