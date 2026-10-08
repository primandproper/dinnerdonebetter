package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "https://example.test"

func buildTestNotifier(t *testing.T) (*Notifier, *mealplanningmock.RepositoryMock, *identitymock.StoreMock) {
	t.Helper()

	repo := &mealplanningmock.RepositoryMock{}
	directory := &identitymock.StoreMock{}

	notifier, err := NewNotifier(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), repo, directory, &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
	}, baseURL)
	require.NoError(t, err)

	return notifier, repo, directory
}

// mealPlanCreated is the event the repository emits when somebody makes a meal plan.
func mealPlanCreated(t *testing.T, userID, accountID, mealPlanID string) *webhooks.Envelope {
	t.Helper()

	payload, err := json.Marshal(&datachanges.Message{
		EventType: mealplanning.MealPlanCreatedServiceEventType,
		UserID:    userID,
		AccountID: accountID,
		Context:   map[string]any{mealplanningkeys.MealPlanIDKey: mealPlanID},
	})
	require.NoError(t, err)

	return &webhooks.Envelope{EventType: webhooks.EventType(mealplanning.MealPlanCreatedServiceEventType), Payload: payload}
}

// rosterOf answers the account's membership page with members.
func rosterOf(t *testing.T, accountID string, members ...*identity.User) func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.MembershipWithUser], error) {
	t.Helper()

	return func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, id string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.MembershipWithUser], error) {
		assert.Equal(t, accountID, id)

		page := &filtering.QueryFilteredResult[identity.MembershipWithUser]{}
		for _, member := range members {
			page.Data = append(page.Data, &identity.MembershipWithUser{User: member})
		}

		return page, nil
	}
}

// usersByID answers ListUsersByIDs with the users whose IDs were asked for.
func usersByID(users ...*identity.User) func(context.Context, database.SQLQueryExecutor, tenancy.Scope, []string) ([]*identity.User, error) {
	return func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, ids []string) ([]*identity.User, error) {
		var out []*identity.User

		for _, user := range users {
			for _, id := range ids {
				if user.ID == id {
					out = append(out, user)
				}
			}
		}

		return out, nil
	}
}

func verifiedUser() *identity.User {
	user := identityfakes.BuildFakeUser()
	now := time.Now()
	user.EmailAddressVerifiedAt = &now

	return user
}

func TestNewNotifier(T *testing.T) {
	T.Parallel()

	T.Run("refuses a nil repository", func(t *testing.T) {
		t.Parallel()

		_, err := NewNotifier(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &identitymock.StoreMock{}, &mockdatabase.ClientMock{}, baseURL)
		require.Error(t, err)
	})
}

func TestNotifier_Handle(T *testing.T) {
	T.Parallel()

	T.Run("a meal plan somebody made mails every verified member of the household", func(t *testing.T) {
		t.Parallel()

		notifier, repo, directory := buildTestNotifier(t)

		mealPlan := mealplanningfakes.BuildFakeMealPlan()
		verified, unverified := verifiedUser(), identityfakes.BuildFakeUser()
		unverified.EmailAddressVerifiedAt = nil

		repo.GetMealPlanFunc = func(_ context.Context, mealPlanID, accountID string) (*mealplanning.MealPlan, error) {
			assert.Equal(t, mealPlan.ID, mealPlanID)
			assert.Equal(t, mealPlan.BelongsToAccount, accountID)

			return mealPlan, nil
		}
		directory.ListAccountMembersFunc = rosterOf(t, mealPlan.BelongsToAccount, verified, unverified)
		directory.ListUsersByIDsFunc = usersByID(verified, unverified)

		handled, emailType, emails, err := notifier.Handle(t.Context(), mealPlanCreated(t, verified.ID, mealPlan.BelongsToAccount, mealPlan.ID))
		require.NoError(t, err)
		assert.True(t, handled)
		assert.Equal(t, emailTypeMealPlanCreated, emailType)

		// One mail, to the member whose address is proven; the other is told nothing.
		require.Len(t, emails, 1)
		assert.Equal(t, verified.EmailAddress, emails[0].ToAddress)
	})

	T.Run("a meal plan nobody made announces itself to nobody", func(t *testing.T) {
		t.Parallel()

		notifier, repo, _ := buildTestNotifier(t)

		handled, _, emails, err := notifier.Handle(t.Context(), mealPlanCreated(t, "", fake.BuildFakeID(), fake.BuildFakeID()))
		require.NoError(t, err)
		assert.True(t, handled)
		assert.Empty(t, emails)
		assert.Empty(t, repo.GetMealPlanCalls())
	})

	T.Run("an event naming no meal plan is an error", func(t *testing.T) {
		t.Parallel()

		notifier, _, _ := buildTestNotifier(t)

		handled, _, _, err := notifier.Handle(t.Context(), mealPlanCreated(t, fake.BuildFakeID(), fake.BuildFakeID(), ""))
		require.Error(t, err)
		assert.True(t, handled)
	})

	T.Run("a meal plan the repository cannot read is an error", func(t *testing.T) {
		t.Parallel()

		notifier, repo, _ := buildTestNotifier(t)

		repo.GetMealPlanFunc = func(context.Context, string, string) (*mealplanning.MealPlan, error) {
			return nil, errors.New("blah")
		}

		handled, _, _, err := notifier.Handle(t.Context(), mealPlanCreated(t, fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID()))
		require.Error(t, err)
		assert.True(t, handled)
	})

	T.Run("a payload that is not what its event promises is an error", func(t *testing.T) {
		t.Parallel()

		notifier, _, _ := buildTestNotifier(t)

		handled, _, _, err := notifier.Handle(t.Context(), &webhooks.Envelope{
			EventType: webhooks.EventType(mealplanning.MealPlanCreatedServiceEventType),
			Payload:   json.RawMessage(`"not an object"`),
		})
		require.Error(t, err)
		assert.True(t, handled)
	})

	T.Run("an event about something else is not this domain's", func(t *testing.T) {
		t.Parallel()

		notifier, _, _ := buildTestNotifier(t)

		handled, _, _, err := notifier.Handle(t.Context(), &webhooks.Envelope{EventType: identity.EventUserRegistered})
		require.NoError(t, err)
		assert.False(t, handled)

		handled, _, _, err = notifier.Handle(t.Context(), nil)
		require.NoError(t, err)
		assert.False(t, handled)
	})
}
