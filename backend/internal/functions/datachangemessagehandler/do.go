package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v13/analytics"
	"github.com/primandproper/platform-go/v13/email"
	"github.com/primandproper/platform-go/v13/encoding"
	"github.com/primandproper/platform-go/v13/messagequeue"
	notifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v13/search/sync"

	"github.com/samber/do/v2"
)

// RegisterAsyncDataChangeMessageHandler registers the async data change message handler with the injector.
func RegisterAsyncDataChangeMessageHandler(i do.Injector) {
	do.Provide[*AsyncDataChangeMessageHandler](i, func(i do.Injector) (*AsyncDataChangeMessageHandler, error) {
		return NewAsyncDataChangeMessageHandler(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[*config.AsyncMessageHandlerConfig](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[internalops.InternalOpsDataManager](i),
			do.MustInvoke[messagequeue.ConsumerProvider](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[analytics.EventReporter](i),
			do.MustInvoke[email.Emailer](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[encoding.ServerEncoderDecoder](i),
			searchSyncers(i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[auth.PasswordResetTokenDataManager](i),
			do.MustInvoke[notificationsmanager.NotificationsDataManager](i),
			do.MustInvoke[notifications.PushNotificationSender](i),
		)
	})
}

// searchSyncers collects every index's Syncer, paired with the topic its events arrive on.
//
// The list is written out rather than discovered because it is the definition of which indexes
// this process consumes for: an index whose Syncer is missing here has a topic nobody reads,
// and the only symptom is search results that quietly stop moving.
func searchSyncers(i do.Injector) []SearchSyncer {
	return []SearchSyncer{
		{
			Topic:  identityindexing.IndexTypeUsers,
			Handle: do.MustInvoke[*searchsync.Syncer[identityindexing.UserSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeMeals,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.MealSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeRecipes,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.RecipeSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidIngredients,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidIngredientSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidInstruments,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidInstrumentSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidMeasurementUnits,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidMeasurementUnitSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidPreparations,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidPreparationSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidIngredientStates,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidIngredientStateSearchSubset]](i).Handle,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidVessels,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidVesselSearchSubset]](i).Handle,
		},
	}
}
