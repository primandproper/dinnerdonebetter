package manager

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	oauthkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/keys"
	oauthmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v6/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"
	"github.com/primandproper/platform-go/v6/random"
	randommock "github.com/primandproper/platform-go/v6/random/mock"
	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildOAuthManagerForTest(t *testing.T) *manager {
	t.Helper()

	ctx := t.Context()
	repo := &oauthmock.RepositoryMock{}
	queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: t.Name()}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{}, nil
		},
	}

	sessionData := &sessions.ContextData{}
	sessionData.ActiveAccountID = "account-1"
	sessionData.Requester.UserID = "user-1"
	sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
		return sessionData, nil
	}

	secretGen := random.NewGenerator(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

	m, err := NewOAuth2Manager(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		secretGen,
		sessionFetcher,
		mpp,
		repo,
		queueCfg,
	)
	require.NoError(t, err)

	return m.(*manager)
}

func setupExpectationsForOAuthManager(
	manager *manager,
	repoSetupFunc func(repo *oauthmock.RepositoryMock),
	secretGenSetupFunc func(gen *randommock.GeneratorMock),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &oauthmock.RepositoryMock{}
	if repoSetupFunc != nil {
		repoSetupFunc(repo)
	}
	manager.oauthRepository = repo

	expectations := []any{repo}

	sg := &randommock.GeneratorMock{
		GenerateHexEncodedStringFunc: func(_ context.Context, _ int) (string, error) { return "", nil },
	}
	if secretGenSetupFunc != nil {
		secretGenSetupFunc(sg)
	}
	manager.secretGenerator = sg

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return expectations
}

func TestOAuth2Manager_CreateOAuth2Client(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		om := buildOAuthManagerForTest(t)

		input := fakes.BuildFakeOAuth2ClientCreationRequestInput()
		expected := fakes.BuildFakeOAuth2Client()

		const plaintextSecret = "eeddccbb55443322eeddccbb55443322"

		expectations := setupExpectationsForOAuthManager(
			om,
			func(repo *oauthmock.RepositoryMock) {
				repo.On(reflection.GetMethodName(repo.CreateOAuth2Client), testutils.ContextMatcher, mock.MatchedBy(func(in *oauth.OAuth2ClientDatabaseCreationInput) bool {
					// only the digest of the generated secret may be persisted.
					return in.Name == input.Name && in.Description == input.Description && in.ClientID != "" && in.ClientSecret == oauth.HashClientSecret(plaintextSecret)
				})).Return(expected, nil)
			},
			func(gen *randommock.GeneratorMock) {
				callCount := 0
				gen.GenerateHexEncodedStringFunc = func(_ context.Context, _ int) (string, error) {
					callCount++
					if callCount == 1 {
						return "aabbccdd11223344aabbccdd11223344", nil
					}
					return plaintextSecret, nil
				}
			},
			map[string][]string{
				oauth.OAuth2ClientCreatedServiceEventType: {oauthkeys.OAuth2ClientIDKey},
			},
		)

		actual, err := om.CreateOAuth2Client(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
		// the caller gets the plaintext secret exactly once, at creation time.
		assert.Equal(t, plaintextSecret, actual.ClientSecret)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestOAuth2Manager_ArchiveOAuth2Client(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		om := buildOAuthManagerForTest(t)

		oauth2ClientID := fakes.BuildFakeID()

		expectations := setupExpectationsForOAuthManager(
			om,
			func(repo *oauthmock.RepositoryMock) {
				repo.On(reflection.GetMethodName(repo.ArchiveOAuth2Client), testutils.ContextMatcher, oauth2ClientID).Return(nil)
			},
			nil,
			map[string][]string{
				oauth.OAuth2ClientArchivedServiceEventType: {oauthkeys.OAuth2ClientIDKey},
			},
		)

		err := om.ArchiveOAuth2Client(ctx, oauth2ClientID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
