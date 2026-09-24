package authentication

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	mockpublishers "github.com/primandproper/primitives-go/v2/messagequeue/mock"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildSignInHooksForTest(t *testing.T, publish func(context.Context, any, ...messagequeue.PublishOption) error) (signin.Hooks, *mockpublishers.PublisherMock) {
	t.Helper()

	publisher := &mockpublishers.PublisherMock{PublishFunc: publish}

	hooks, err := NewSignInHooks(publisher)
	require.NoError(t, err)

	return hooks, publisher
}

func TestNewSignInHooks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		hooks, err := NewSignInHooks(&mockpublishers.PublisherMock{})

		require.NoError(t, err)
		assert.NotNil(t, hooks)
	})

	T.Run("with no publisher", func(t *testing.T) {
		t.Parallel()

		hooks, err := NewSignInHooks(nil)

		require.ErrorIs(t, err, ErrNilSignInPublisher)
		assert.Nil(t, hooks)
	})
}

func TestSignInHooks_AfterAuthenticate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		hooks, publisher := buildSignInHooksForTest(t, func(context.Context, any, ...messagequeue.PublishOption) error { return nil })

		user := identityfakes.BuildFakeUser()
		account := identityfakes.BuildFakeAccountForUser(user.ID)

		err := hooks.AfterAuthenticate(ctx, nil, tenancy.Global(), &signin.Authentication{
			Principal: &identity.Principal{User: user, ActiveAccountID: account.ID},
		})
		require.NoError(t, err)

		require.Len(t, publisher.PublishCalls(), 1)
		assert.Equal(t, &audit.DataChangeMessage{
			EventType: ddbidentity.UserLoggedInServiceEventType,
			AccountID: account.ID,
			UserID:    user.ID,
		}, publisher.PublishCalls()[0].Data)
	})

	T.Run("for an administrative sign-in", func(t *testing.T) {
		t.Parallel()

		// The administrative door is a sign-in like any other, and was recorded as one
		// when ProcessLogin published the event itself.
		ctx := t.Context()
		hooks, publisher := buildSignInHooksForTest(t, func(context.Context, any, ...messagequeue.PublishOption) error { return nil })

		user := identityfakes.BuildFakeUser()

		err := hooks.AfterAuthenticate(ctx, nil, tenancy.Global(), &signin.Authentication{
			Principal:      &identity.Principal{User: user},
			Administrative: true,
		})
		require.NoError(t, err)

		require.Len(t, publisher.PublishCalls(), 1)
		assert.Equal(t, user.ID, publisher.PublishCalls()[0].Data.(*audit.DataChangeMessage).UserID)
	})

	T.Run("refuses the sign-in when the event cannot be published", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		publishErr := errors.New(identityfakes.BuildFakeUser().ID)
		hooks, _ := buildSignInHooksForTest(t, func(context.Context, any, ...messagequeue.PublishOption) error { return publishErr })

		err := hooks.AfterAuthenticate(ctx, nil, tenancy.Global(), &signin.Authentication{
			Principal: &identity.Principal{User: identityfakes.BuildFakeUser()},
		})

		require.ErrorIs(t, err, publishErr)
	})

	T.Run("with no principal", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		hooks, publisher := buildSignInHooksForTest(t, func(context.Context, any, ...messagequeue.PublishOption) error { return nil })

		err := hooks.AfterAuthenticate(ctx, nil, tenancy.Global(), &signin.Authentication{})

		require.Error(t, err)
		assert.Empty(t, publisher.PublishCalls())
	})
}
