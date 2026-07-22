package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	datachangemessagehandlerbuild "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/functions/data_change_message_handler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"

	"github.com/primandproper/platform-go/v5/observability/metrics"
	"github.com/primandproper/platform-go/v5/observability/tracing"

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

	ctx := context.Background()

	injector := datachangemessagehandlerbuild.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so a short-lived pod exports its spans/metrics before terminating.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if shutdownErr := do.MustInvoke[metrics.Provider](injector).Shutdown(shutdownCtx); shutdownErr != nil {
			log.Printf("error shutting down metrics: %v", shutdownErr)
		}
		if flushErr := do.MustInvoke[tracing.TracerProvider](injector).ForceFlush(shutdownCtx); flushErr != nil {
			log.Printf("error flushing traces: %v", flushErr)
		}
	}()

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

	if err = dataChangeMessageHandler.ConsumeMessages(ctx, stopChan, errorsChan); err != nil {
		log.Fatal(err)
	}

	// Block until the first shutdown signal arrives.
	<-signalChan

	// Closing stopChan broadcasts to every consumer; then wait for in-flight handlers to drain
	// before the deferred telemetry flush runs and main returns.
	close(stopChan)

	drainCtx, cancelDrain := context.WithTimeout(ctx, 30*time.Second)
	defer cancelDrain()
	dataChangeMessageHandler.WaitForConsumers(drainCtx)
}
