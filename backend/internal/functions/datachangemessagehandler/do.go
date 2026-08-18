package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexstamp"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v11/analytics"
	"github.com/primandproper/platform-go/v11/email"
	"github.com/primandproper/platform-go/v11/encoding"
	"github.com/primandproper/platform-go/v11/messagequeue"
	notifications "github.com/primandproper/platform-go/v11/notifications/mobile"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v11/search/sync"

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

// searchSyncers collects every index's Syncer, paired with the topic its events arrive on and
// with the Stamper it writes last_indexed_at through.
//
// The list is written out rather than discovered because it is the definition of which indexes
// this process consumes for: an index whose Syncer is missing here has a topic nobody reads,
// and the only symptom is search results that quietly stop moving.
//
// The Stamper is resolved by the index's name, which is the same string as the topic, so this
// gets the very object the Syncer was built over rather than a second buffer for the same
// index.
func searchSyncers(i do.Injector) []SearchSyncer {
	return []SearchSyncer{
		{
			Topic:  identityindexing.IndexTypeUsers,
			Handle: do.MustInvoke[*searchsync.Syncer[identityindexing.UserSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, identityindexing.IndexTypeUsers).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeMeals,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.MealSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeMeals).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeRecipes,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.RecipeSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeRecipes).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidIngredients,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidIngredientSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeValidIngredients).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidInstruments,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidInstrumentSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeValidInstruments).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidMeasurementUnits,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidMeasurementUnitSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeValidMeasurementUnits).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidPreparations,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidPreparationSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeValidPreparations).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidIngredientStates,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidIngredientStateSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeValidIngredientStates).Close,
		},
		{
			Topic:  mealplanningindexing.IndexTypeValidVessels,
			Handle: do.MustInvoke[*searchsync.Syncer[mealplanningindexing.ValidVesselSearchSubset]](i).Handle,
			Close:  do.MustInvokeNamed[*indexstamp.Stamper](i, mealplanningindexing.IndexTypeValidVessels).Close,
		},
	}
}
