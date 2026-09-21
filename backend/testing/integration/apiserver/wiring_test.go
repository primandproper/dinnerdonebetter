package integration

import (
	"context"
	"testing"
	"time"

	datachangemessagehandlerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/functions/data_change_message_handler"
	schedulerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/scheduler"
	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/platform-go/v14/dataprivacy/auditerasure"
	"github.com/primandproper/platform-go/v14/metering"
	"github.com/primandproper/platform-go/v14/operations"
	"github.com/primandproper/platform-go/v14/outbox"
	"github.com/primandproper/platform-go/v14/retention"
	"github.com/primandproper/platform-go/v14/saga"
	"github.com/primandproper/platform-go/v14/webhooks"
	"github.com/primandproper/primitives-go/v2/jobs"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The two worker processes' containers are the one thing driving a worker directly cannot see.
//
// Every other test in this suite builds a worker's dependencies for it, which is exactly what
// makes those tests worth having and exactly what makes them blind here: a handler hung off the
// wrong pool, or a component that quietly stopped being registered, is invisible to a test that
// supplies the component itself. samber/do resolves lazily, so an absent registration is not a
// startup error either — it is an error the first time something asks, which for a background
// worker can be the first time a subject asks for their data.
//
// That is not hypothetical. The data privacy registry could not be built in the scheduler at all
// until this test was written: the container registered the settings repository but not the
// domain interface its collector asks for, so the fulfillment worker failed to resolve — and
// nothing said so, because nothing had ever resolved it.
//
// These tests resolve; they do not run. Running is what the rest of the suite does.

// buildSchedulerInjector stands up the scheduler's container over this suite's database and
// releases it when the test ends.
func buildSchedulerInjector(t *testing.T) do.Injector {
	t.Helper()

	i := schedulerbuild.BuildInjector(context.Background(), schedulerConfig)

	shutdownInjector(t, i)

	return i
}

// shutdownInjector releases a container's resources after the test.
func shutdownInjector(t *testing.T, i *do.RootScope) {
	t.Helper()

	t.Cleanup(func() {
		// A background context: t.Context() is already cancelled by the time cleanups run,
		// and these shutdowns close connection pools and queue goroutines.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if report := i.ShutdownWithContext(ctx); report != nil {
			assert.True(t, report.Succeed, "releasing the container: %v", report)
		}
	})
}

// TestWorkerWiring_Scheduler resolves everything the scheduler process runs.
//
// One entry per thing that is started or ticked in that process. A registration deleted, or a
// dependency added to one of these constructors without being registered beside it, is a
// resolution failure here rather than a crash loop in an environment.
func TestWorkerWiring_Scheduler(T *testing.T) {
	T.Parallel()

	T.Run("resolves every component it runs", func(t *testing.T) {
		t.Parallel()

		i := buildSchedulerInjector(t)

		// The loops, each of which service.New starts and cmd/ddb shuts down.
		require.NotNil(t, do.MustInvoke[*jobs.Scheduler](i))
		require.NotNil(t, do.MustInvoke[*outbox.Relay](i))
		require.NotNil(t, do.MustInvoke[*saga.Worker](i))
		require.NotNil(t, do.MustInvoke[*webhooks.Worker](i))
		require.NotNil(t, do.MustInvoke[*operations.Worker](i))

		// The scheduled jobs. Each is resolved from inside its own closure at tick time rather
		// than at registration, so building the Scheduler above proves nothing about them — a
		// job whose dependency stopped being registered would tick, panic, and be recorded as a
		// failed run for as long as nobody read the metric.
		require.NotNil(t, do.MustInvoke[*mealplanfinalization.Starter](i))
		require.NotNil(t, do.MustInvoke[*mealplantasknotifications.Worker](i))
		require.NotNil(t, do.MustInvoke[*queuetest.Job](i))
		require.NotNil(t, do.MustInvoke[*platformdataprivacy.Sweeper](i))
		require.NotNil(t, do.MustInvoke[*retention.Sweeper](i))
		require.NotNil(t, do.MustInvoke[*metering.Flusher](i))
	})

	T.Run("registers every data privacy collector and eraser", func(t *testing.T) {
		t.Parallel()

		i := buildSchedulerInjector(t)

		registry := do.MustInvoke[*platformdataprivacy.Registry](i)

		// The exported document's sections are exactly these keys, minus the ones the subject
		// happens to hold nothing under — so a domain that stopped being registered produces an
		// export that is complete by its own manifest and missing a domain's worth of somebody's
		// data. There is no other place that would notice.
		//
		// It notices a domain that stops being registered. It did not notice three that never
		// were: passkeys, password reset tokens and registered OAuth2 clients were absent from
		// the registry and therefore absent from this list, which is written from what is
		// registered rather than from what holds a subject's data. A list like this pins a set
		// against drift and cannot tell you the set was wrong to begin with — the three were
		// found by reading platform's privacy packages against this one, not by a failure here.
		// There is no webhooks key, and its absence is asserted by this list being exact:
		// nothing in that domain names a person, so a collector over it answered a question
		// about a subject with an account's delivery configuration. See docs/data-privacy.md
		// for the one obligation that leaves here, which retention discharges.
		assert.ElementsMatch(t, []string{
			ddbdataprivacy.CollectorKeyIdentity,
			ddbdataprivacy.CollectorKeyMealPlanning,
			ddbdataprivacy.CollectorKeySettings,
			ddbdataprivacy.CollectorKeyNotifications,
			ddbdataprivacy.CollectorKeyBilling,
			ddbdataprivacy.CollectorKeyAuditLog,
			ddbdataprivacy.CollectorKeyIssueReports,
			ddbdataprivacy.CollectorKeyMediaRegistry,
			ddbdataprivacy.CollectorKeyWaitlists,
			ddbdataprivacy.CollectorKeyComments,
			ddbdataprivacy.CollectorKeyPasskeys,
			ddbdataprivacy.CollectorKeyPasswordReset,
			ddbdataprivacy.CollectorKeyOAuth2Clients,
		}, registry.CollectorKeys())

		// The erasers are the destructive half, and the audit one is a policy decision that is
		// on in this deployment. An eraser missing here is data that survives a right-to-be-
		// forgotten request.
		//
		// There are ten where there used to be five, and the five that arrived are the ones
		// privacyadapters registers alongside their collectors: a domain's adapter builds
		// both halves. Four of them delete rows the identity cascade would have taken
		// anyway — settings values, issue reports, passkeys, reset tokens.
		//
		// That redundancy reverses what docs/data-privacy.md used to argue, and the reason
		// it reverses is in this repository's own history. Those statements are platform's
		// rather than ours, so the "eleven statements that can only agree with the one that
		// ran first" objection does not apply; and a cascade is exactly the thing that goes
		// missing silently. ddb_webauthn_credentials lost its foreign key during the
		// identity adoption and erasure stopped reaching a deleted user's passkeys, with
		// nothing raising — a registered eraser would have kept working through it.
		assert.ElementsMatch(t, []string{
			ddbdataprivacy.CollectorKeyComments,
			ddbdataprivacy.CollectorKeyWaitlists,
			ddbdataprivacy.CollectorKeyOAuth2Clients,
			ddbdataprivacy.CollectorKeySettings,
			ddbdataprivacy.CollectorKeyIssueReports,
			ddbdataprivacy.CollectorKeyMediaRegistry,
			ddbdataprivacy.CollectorKeyPasskeys,
			ddbdataprivacy.CollectorKeyPasswordReset,
			ddbdataprivacy.EraserKeyIdentity,
			auditerasure.DefaultKey,
		}, registry.EraserKeys())
	})
}

// TestWorkerWiring_AsyncMessageHandler resolves the data change consumer.
//
// It is one component rather than a list, because that process is one component: a handler the
// pools hand messages to. What this catches is the same thing — a dependency added to the handler
// without being registered — in the process where the symptom would be a broker topic nothing
// drains.
func TestWorkerWiring_AsyncMessageHandler(T *testing.T) {
	T.Parallel()

	T.Run("resolves its handler", func(t *testing.T) {
		t.Parallel()

		i := datachangemessagehandlerbuild.BuildInjector(context.Background(), asyncMessageHandlerConfig)
		shutdownInjector(t, i)

		require.NotNil(t, do.MustInvoke[*datachangemessagehandler.AsyncDataChangeMessageHandler](i))
	})
}
