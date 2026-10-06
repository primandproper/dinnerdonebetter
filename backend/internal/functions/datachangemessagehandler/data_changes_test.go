package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	internalopsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	"github.com/primandproper/platform-go/v15/billing"
	identity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	msgqueuemock "github.com/primandproper/primitives-go/v2/messagequeue/mock"
	notifications "github.com/primandproper/primitives-go/v2/notifications/mobile"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envelopeFor wraps payload the way platform's Emitter does for the broker: the event named
// beside the payload, which is how a consumer tells identity.UserEvent's eleven events apart.
func envelopeFor(t *testing.T, eventType webhooks.EventType, payload any) *webhooks.Envelope {
	t.Helper()

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	return &webhooks.Envelope{EventType: eventType, ID: identifiers.New(), Payload: raw}
}

func rawEnvelopeFor(t *testing.T, eventType webhooks.EventType, payload any) []byte {
	t.Helper()

	raw, err := json.Marshal(envelopeFor(t, eventType, payload))
	require.NoError(t, err)

	return raw
}

// ownEvent is one of this application's own events, as its repositories emit it.
func ownEvent(t *testing.T, eventType, userID string, metadata map[string]any) *webhooks.Envelope {
	t.Helper()

	return envelopeFor(t, webhooks.EventType(eventType), &datachanges.Message{EventType: eventType, UserID: userID, Context: metadata})
}

// capturingPublisher replaces one of the handler's publishers with one that keeps what it was
// handed, so a test asserts on the mail or the push that was queued rather than on a call count.
func capturingPublisher() (publisher *msgqueuemock.PublisherMock, published *[]any) {
	var kept []any

	return &msgqueuemock.PublisherMock{
		PublishFunc: func(_ context.Context, msg any, _ ...messagequeue.PublishOption) error {
			kept = append(kept, msg)

			return nil
		},
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
		StopFunc:         func() {},
	}, &kept
}

func emailsPublished(t *testing.T, published []any) []*queuemessages.OutboundEmailMessage {
	t.Helper()

	out := make([]*queuemessages.OutboundEmailMessage, 0, len(published))
	for _, msg := range published {
		email, ok := msg.(*queuemessages.OutboundEmailMessage)
		require.True(t, ok, "published %T rather than an email", msg)
		out = append(out, email)
	}

	return out
}

func TestAsyncDataChangeMessageHandler_DataChangesEventHandler(T *testing.T) {
	// Rendering a mail reads the environment, which rules the subtests out of running in
	// parallel; the env var is set once here and inherited.
	T.Setenv("DINNER_DONE_BETTER_SERVICE_ENVIRONMENT", "testing")

	T.Run("routes a registration to the metrics it implies, and mails nothing", func(t *testing.T) {
		handler, directory, _, _, analyticsReporter, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		emails, published := capturingPublisher()
		handler.outboundEmailsPublisher = emails

		user := identityfakes.BuildFakeUser()
		raw := rawEnvelopeFor(t, identity.EventUserRegistered, &identity.UserEvent{
			UserID:    user.ID,
			AccountID: identifiers.New(),
		})

		directory.GetUserFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)

			return user, nil
		}
		analyticsReporter.EventOccurredFunc = func(_ context.Context, eventType, userID string, _ map[string]any) error {
			assert.Equal(t, identity.EventUserRegistered.String(), eventType)
			assert.Equal(t, user.ID, userID)

			return nil
		}
		analyticsReporter.AddUserFunc = func(_ context.Context, userID string, _ map[string]any) error {
			assert.Equal(t, user.ID, userID)

			return nil
		}

		require.NoError(t, handler.DataChangesEventHandler("data_changes")(t.Context(), raw))

		assert.Len(t, analyticsReporter.EventOccurredCalls(), 1)
		assert.Len(t, analyticsReporter.AddUserCalls(), 1)
		// The verification mail is signin's to send, through SignInMailers, once the
		// registration commits; it arrives as a mail request of its own.
		assert.Empty(t, emailsPublished(t, *published))
	})

	T.Run("an event nothing here acts on passes through", func(t *testing.T) {
		handler, directory, _, _, analyticsReporter, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		raw, err := json.Marshal(ownEvent(t, mealplanning.ValidIngredientUpdatedServiceEventType, identifiers.New(), nil))
		require.NoError(t, err)

		require.NoError(t, handler.DataChangesEventHandler("data_changes")(t.Context(), raw))

		assert.Empty(t, directory.GetUserCalls())
		assert.Empty(t, analyticsReporter.EventOccurredCalls())
	})

	T.Run("acknowledges the queue check's probe by its ID", func(t *testing.T) {
		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		probe, err := internalops.BuildQueueTestMessage("data_changes", "test-123", "")
		require.NoError(t, err)
		raw, err := json.Marshal(probe)
		require.NoError(t, err)

		repo := &internalopsmock.InternalOpsDataManagerMock{
			AcknowledgeQueueTestMessageFunc: func(_ context.Context, id string) error {
				assert.Equal(t, "test-123", id)

				return nil
			},
			PruneQueueTestMessagesFunc: func(context.Context, string) error { return nil },
		}
		handler.internalOpsRepo = repo

		require.NoError(t, handler.DataChangesEventHandler("data_changes")(t.Context(), raw))
		assert.Len(t, repo.AcknowledgeQueueTestMessageCalls(), 1)
	})

	T.Run("a body that is not an envelope is refused for good", func(t *testing.T) {
		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		err := handler.DataChangesEventHandler("data_changes")(t.Context(), []byte("invalid json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding event envelope")
	})

	T.Run("a bare payload naming no event is refused for good", func(t *testing.T) {
		// What the topic carried before platform-go #1130: a platform payload with nothing
		// saying which event it is. Something publishing one now is a wiring mistake to hear
		// about rather than skip.
		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		raw, err := json.Marshal(&identity.UserEvent{UserID: identifiers.New()})
		require.NoError(t, err)

		err = handler.DataChangesEventHandler("data_changes")(t.Context(), raw)
		require.Error(t, err)
		require.ErrorIs(t, err, webhooks.ErrEmptyEventType)
	})
}

func TestAsyncDataChangeMessageHandler_handleDataChangeMessage(T *testing.T) {
	T.Parallel()

	T.Run("with nil event", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		err := handler.handleDataChangeMessage(t.Context(), nil)
		require.ErrorIs(t, err, errRequiredDataIsNil)
	})

	T.Run("attributes an account's event to the account's owner", func(t *testing.T) {
		t.Parallel()

		// A subscription is an account's and names nobody; the product counts it against
		// whoever owns the account.
		handler, directory, _, _, analyticsReporter, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		account := identityfakes.BuildFakeAccount()
		event := envelopeFor(t, billing.EventSubscriptionCreated, &billing.SubscriptionEvent{
			SubscriptionID: identifiers.New(), AccountID: account.ID, ProductID: "pro",
		})

		directory.GetAccountFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, accountID string) (*identity.Account, error) {
			assert.Equal(t, account.ID, accountID)

			return account, nil
		}
		analyticsReporter.EventOccurredFunc = func(_ context.Context, eventType, userID string, properties map[string]any) error {
			assert.Equal(t, billing.EventSubscriptionCreated.String(), eventType)
			assert.Equal(t, account.OwnerUserID, userID)
			assert.Equal(t, "pro", properties["productID"])

			return nil
		}

		require.NoError(t, handler.handleDataChangeMessage(t.Context(), event))
		assert.Len(t, analyticsReporter.EventOccurredCalls(), 1)
	})

	T.Run("reports nothing for an event off the allowlist", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, analyticsReporter, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		event := envelopeFor(t, identity.EventUserProfileUpdated, &identity.UserEvent{UserID: identifiers.New()})

		require.NoError(t, handler.handleDataChangeMessage(t.Context(), event))
		assert.Empty(t, analyticsReporter.EventOccurredCalls())
	})

	T.Run("reports nothing for an event nobody made", func(t *testing.T) {
		t.Parallel()

		// A background job's event names no user, and the analytics platform counts people.
		handler, _, _, _, analyticsReporter, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		event := ownEvent(t, mealplanning.MealPlanFinalizedServiceEventType, "", nil)

		require.NoError(t, handler.handleDataChangeMessage(t.Context(), event))
		assert.Empty(t, analyticsReporter.EventOccurredCalls())
	})

	T.Run("a payload that is not what its event promises is an error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		event := &webhooks.Envelope{EventType: identity.EventUserRegistered, Payload: json.RawMessage(`"not an object"`)}

		require.Error(t, handler.handleDataChangeMessage(t.Context(), event))
	})
}

func TestAsyncDataChangeMessageHandler_handleOutboundNotifications(T *testing.T) {
	T.Parallel()

	T.Run("with nil event", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		err := handler.handleOutboundNotifications(t.Context(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil data change event")
	})

	T.Run("a user the directory cannot find is an error", func(t *testing.T) {
		t.Parallel()

		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		expected := errors.New("user fetch error")
		directory.GetUserFunc = func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
			return nil, expected
		}

		event := envelopeFor(t, identity.EventUserRegistered, &identity.UserEvent{UserID: identifiers.New()})

		err := handler.handleOutboundNotifications(t.Context(), event)
		require.ErrorIs(t, err, expected)
		assert.Contains(t, err.Error(), "getting user")
	})

	T.Run("an unhandled event type reads nothing", func(t *testing.T) {
		t.Parallel()

		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		event := ownEvent(t, "unhandled.event.type", identifiers.New(), nil)

		require.NoError(t, handler.handleOutboundNotifications(t.Context(), event))
		assert.Empty(t, directory.GetUserCalls())
	})
}

func TestAsyncDataChangeMessageHandler_handleIdentityOutboundNotification(T *testing.T) {
	T.Setenv("DINNER_DONE_BETTER_SERVICE_ENVIRONMENT", "testing")

	// pushesFrom swaps in a push publisher that keeps what it was handed.
	pushesFrom := func(handler *AsyncDataChangeMessageHandler) *[]any {
		pushes, published := capturingPublisher()
		handler.mobileNotificationsPublisher = pushes

		return published
	}

	T.Run("a registration through an invitation also tells the household", func(t *testing.T) {
		handler, directory, _, _, analyticsReporter, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		analyticsReporter.AddUserFunc = func(context.Context, string, map[string]any) error { return nil }

		joined, other := identityfakes.BuildFakeUser(), identityfakes.BuildFakeUser()
		accountID := identifiers.New()
		published := pushesFrom(handler)

		directory.GetUserFunc = returning(joined)
		directory.ListAccountMembersFunc = rosterOf(t, accountID, joined, other)

		event := envelopeFor(t, identity.EventUserRegistered, &identity.UserEvent{
			UserID: joined.ID, AccountID: accountID, InvitationID: identifiers.New(),
		})

		handled, _, emails, err := handler.handleIdentityOutboundNotification(t.Context(), event)
		require.NoError(t, err)
		assert.True(t, handled)
		assert.Empty(t, emails)

		require.Len(t, *published, 1)
		push, ok := (*published)[0].(*notifications.MobileNotificationRequest)
		require.True(t, ok)
		assert.Equal(t, []string{other.ID}, push.RecipientUserIDs)
		assert.Equal(t, joined.ID, push.Context[ddbidentity.ExcludedUserIDContextKey])
	})

	T.Run("an acceptance tells the household who joined", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		joined, other := identityfakes.BuildFakeUser(), identityfakes.BuildFakeUser()
		accountID := identifiers.New()
		published := pushesFrom(handler)

		directory.GetUserFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*identity.User, error) {
			assert.Equal(t, joined.ID, userID)

			return joined, nil
		}
		directory.ListAccountMembersFunc = rosterOf(t, accountID, joined, other)

		acceptor := joined.ID
		event := envelopeFor(t, identity.EventInvitationAccepted, &identity.InvitationEvent{
			InvitationID: identifiers.New(), AccountID: accountID, FromUser: other.ID, ToUser: &acceptor,
		})

		handled, _, emails, err := handler.handleIdentityOutboundNotification(t.Context(), event)
		require.NoError(t, err)
		assert.True(t, handled)
		assert.Empty(t, emails)

		require.Len(t, *published, 1)
		push, ok := (*published)[0].(*notifications.MobileNotificationRequest)
		require.True(t, ok)
		assert.Equal(t, []string{other.ID}, push.RecipientUserIDs)
		assert.Contains(t, push.Body, joined.DisplayName)
	})

	T.Run("an acceptance naming nobody is refused", func(t *testing.T) {
		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		event := envelopeFor(t, identity.EventInvitationAccepted, &identity.InvitationEvent{
			InvitationID: identifiers.New(), AccountID: identifiers.New(), FromUser: identifiers.New(),
		})

		_, _, _, err := handler.handleIdentityOutboundNotification(t.Context(), event)
		require.Error(t, err)
	})

	T.Run("an invitation is mailed with the token off its mail request", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		sender := identityfakes.BuildFakeUser()
		invitation := identityfakes.BuildFakeInvitationFromUserToAccount(sender.ID, identifiers.New())
		// As the row reads back: the column holds a digest and no read fills the secret in.
		invitation.Token = ""

		directory.GetUserFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*identity.User, error) {
			assert.Equal(t, sender.ID, userID)

			return sender, nil
		}
		directory.GetInvitationFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, invitationID string) (*identity.Invitation, error) {
			assert.Equal(t, invitation.ID, invitationID)

			return invitation, nil
		}

		event := ownEvent(t, ddbidentity.AccountInvitationMailRequestedEventType, sender.ID, map[string]any{
			identitykeys.AccountInvitationIDKey:    invitation.ID,
			identitykeys.AccountInvitationTokenKey: "invitation-secret",
		})

		handled, emailType, emails, err := handler.handleIdentityOutboundNotification(t.Context(), event)
		require.NoError(t, err)
		assert.True(t, handled)
		assert.Equal(t, "account invitation created", emailType)
		require.Len(t, emails, 1)
		assert.Equal(t, invitation.ToEmail, emails[0].ToAddress)
		assert.Contains(t, emails[0].HTMLContent, "invitation-secret")
	})

	T.Run("an invitation carrying no token is refused rather than mailed as a dead link", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		directory.GetUserFunc = returning(identityfakes.BuildFakeUser())
		event := ownEvent(t, ddbidentity.AccountInvitationMailRequestedEventType, identifiers.New(), map[string]any{
			identitykeys.AccountInvitationIDKey: identifiers.New(),
		})

		_, _, _, err := handler.handleIdentityOutboundNotification(t.Context(), event)
		require.Error(t, err)
		assert.Empty(t, directory.GetInvitationCalls())
	})

	T.Run("a reset request is mailed with the secret it carries", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		user := verifiedUser()
		directory.GetUserFunc = returning(user)

		event := ownEvent(t, ddbidentity.PasswordResetTokenCreatedEventType, user.ID, map[string]any{
			authkeys.PasswordResetTokenIDKey:     identifiers.New(),
			authkeys.PasswordResetTokenSecretKey: "reset-secret",
		})

		_, emailType, emails, err := handler.handleIdentityOutboundNotification(t.Context(), event)
		require.NoError(t, err)
		assert.Equal(t, "password reset request", emailType)
		require.Len(t, emails, 1)
		assert.Contains(t, emails[0].HTMLContent, "reset-secret")
	})

	T.Run("a reset request carrying no secret is refused", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		user := verifiedUser()
		directory.GetUserFunc = returning(user)

		_, _, _, err := handler.handleIdentityOutboundNotification(t.Context(), ownEvent(t, ddbidentity.PasswordResetTokenCreatedEventType, user.ID, nil))
		require.Error(t, err)
	})

	T.Run("a redeemed reset tells the owner", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		user := verifiedUser()
		directory.GetUserFunc = returning(user)

		event := envelopeFor(t, passwordreset.EventTokenRedeemed, &passwordreset.TokenEvent{TokenID: identifiers.New(), UserID: user.ID})

		_, emailType, emails, err := handler.handleIdentityOutboundNotification(t.Context(), event)
		require.NoError(t, err)
		assert.Equal(t, "password reset token redeemed", emailType)
		assert.Len(t, emails, 1)
	})

	T.Run("a password change through either door tells the owner", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		user := verifiedUser()
		directory.GetUserFunc = returning(user)

		for _, event := range []*webhooks.Envelope{
			envelopeFor(t, signin.EventPasswordUpdated, &signin.UserEvent{UserID: user.ID}),
			envelopeFor(t, identity.EventUserPasswordChanged, &identity.UserEvent{UserID: user.ID}),
		} {
			_, emailType, emails, err := handler.handleIdentityOutboundNotification(t.Context(), event)
			require.NoError(t, err)
			assert.Equal(t, "password changed", emailType)
			assert.Len(t, emails, 1)
		}
	})

	T.Run("another verification link and a handle reminder are mailed", func(t *testing.T) {
		handler, directory, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		// Unproven for the link — a verification mail to a proven address is refused — and
		// proven for the reminder, which goes only to an address somebody has proven.
		unverified, verified := identityfakes.BuildFakeUser(), verifiedUser()

		directory.GetUserFunc = returning(unverified)
		_, _, emails, err := handler.handleIdentityOutboundNotification(t.Context(),
			ownEvent(t, ddbidentity.UserEmailAddressVerificationEmailRequestedEventType, unverified.ID, map[string]any{identitykeys.UserEmailVerificationTokenKey: "again"}))
		require.NoError(t, err)
		require.Len(t, emails, 1)
		assert.Contains(t, emails[0].HTMLContent, "again")

		user := verified
		directory.GetUserFunc = returning(user)
		_, _, emails, err = handler.handleIdentityOutboundNotification(t.Context(),
			ownEvent(t, ddbidentity.UsernameReminderRequestedEventType, user.ID, nil))
		require.NoError(t, err)
		require.Len(t, emails, 1)
		assert.Contains(t, strings.ToLower(emails[0].HTMLContent), strings.ToLower(user.Username))
	})

	T.Run("an event about something else is not this handler's", func(t *testing.T) {
		handler, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		handled, _, _, err := handler.handleIdentityOutboundNotification(t.Context(),
			ownEvent(t, mealplanning.RecipeCreatedServiceEventType, identifiers.New(), nil))
		require.NoError(t, err)
		assert.False(t, handled)
	})
}

// verifiedUser is a recipient the owner-facing mails will address: those refuse an address nobody
// has proven, and the fake's is unproven.
func verifiedUser() *identity.User {
	user := identityfakes.BuildFakeUser()
	now := time.Now()
	user.EmailAddressVerifiedAt = &now

	return user
}

// returning is a GetUser that answers with user whoever is asked for.
func returning(user *identity.User) func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
	return func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
		return user, nil
	}
}

// rosterOf is a ListAccountMembers over one page holding exactly these members.
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
