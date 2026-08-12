package manager

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	oauthmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/mock"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v10/messagequeue"
	mockpublishers "github.com/primandproper/platform-go/v10/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v10/observability/tracing/noop"
	"github.com/primandproper/platform-go/v10/random"
	randommock "github.com/primandproper/platform-go/v10/random/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildOAuthManagerForTest(t *testing.T) *manager {
	t.Helper()

	ctx := t.Context()
	repo := &oauthmock.RepositoryMock{}
	queueCfg := &queuescfg.Config{DataChangesTopicName: t.Name()}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{}, nil
		},
	}

	secretGen := random.NewGenerator(random.WithLogger(loggingnoop.NewLogger()), random.WithTracerProvider(tracingnoop.NewTracerProvider()))

	m, err := NewOAuth2Manager(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		secretGen,
		mpp,
		repo,
		queueCfg,
	)
	require.NoError(t, err)

	return m.(*manager)
}

// attachMocksToOAuthManager wires a configured repository mock, secret generator and a no-op data
// changes publisher into the manager under test. A nil secretGenerator gets a stub that returns an
// empty string.
func attachMocksToOAuthManager(manager *manager, repo *oauthmock.RepositoryMock, secretGenerator *randommock.GeneratorMock) {
	manager.oauthRepository = repo

	if secretGenerator == nil {
		secretGenerator = &randommock.GeneratorMock{
			GenerateHexEncodedStringFunc: func(_ context.Context, _ int) (string, error) { return "", nil },
		}
	}
	manager.secretGenerator = secretGenerator

	manager.dataChangesPublisher = &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}
}

// buildSessionContextForTest returns a context carrying session data for an arbitrary requester.
func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: fakes.BuildFakeID(),
		Requester:       sessions.RequesterInfo{UserID: fakes.BuildFakeID()},
	})
}

func TestOAuth2Manager_CreateOAuth2Client(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		om := buildOAuthManagerForTest(t)

		input := fakes.BuildFakeOAuth2ClientCreationRequestInput()
		expected := fakes.BuildFakeOAuth2Client()

		const plaintextSecret = "eeddccbb55443322eeddccbb55443322"

		repo := &oauthmock.RepositoryMock{
			CreateOAuth2ClientFunc: func(_ context.Context, in *oauth.OAuth2ClientDatabaseCreationInput) (*oauth.OAuth2Client, error) {
				assert.Equal(t, input.Name, in.Name)
				assert.Equal(t, input.Description, in.Description)
				assert.NotEmpty(t, in.ClientID)
				// only the digest of the generated secret may be persisted.
				assert.Equal(t, oauth.HashClientSecret(plaintextSecret), in.ClientSecret)

				return expected, nil
			},
		}

		callCount := 0
		secretGenerator := &randommock.GeneratorMock{
			GenerateHexEncodedStringFunc: func(_ context.Context, _ int) (string, error) {
				callCount++
				if callCount == 1 {
					return "aabbccdd11223344aabbccdd11223344", nil
				}

				return plaintextSecret, nil
			},
		}

		attachMocksToOAuthManager(om, repo, secretGenerator)

		actual, err := om.CreateOAuth2Client(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		// the caller gets the plaintext secret exactly once, at creation time.
		assert.Equal(t, plaintextSecret, actual.ClientSecret)

		assert.Len(t, repo.CreateOAuth2ClientCalls(), 1)
	})
}

func TestOAuth2Manager_ArchiveOAuth2Client(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		om := buildOAuthManagerForTest(t)

		oauth2ClientID := fakes.BuildFakeID()

		repo := &oauthmock.RepositoryMock{
			ArchiveOAuth2ClientFunc: func(_ context.Context, clientID string) error {
				assert.Equal(t, oauth2ClientID, clientID)

				return nil
			},
		}
		attachMocksToOAuthManager(om, repo, nil)

		err := om.ArchiveOAuth2Client(ctx, oauth2ClientID)
		require.NoError(t, err)

		assert.Len(t, repo.ArchiveOAuth2ClientCalls(), 1)
	})
}
