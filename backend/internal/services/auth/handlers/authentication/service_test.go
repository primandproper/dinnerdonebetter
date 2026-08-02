package authentication

import (
	"context"
	"encoding/base64"
	"testing"

	authcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/config"
	mockauthn "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/mock"
	identitymanagermock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	oauthmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/mock"
	queuescfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/queues/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	tokenscfg "github.com/primandproper/platform-go/v9/authentication/tokens/config"
	mocktotp "github.com/primandproper/platform-go/v9/authentication/totp/mock"
	"github.com/primandproper/platform-go/v9/messagequeue"
	mockpublishers "github.com/primandproper/platform-go/v9/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	ctx := t.Context()
	logger := loggingnoop.NewLogger()

	cfg := &Config{
		Tokens: authcfg.TokensConfig{
			Config: tokenscfg.Config{
				Provider:                tokenscfg.ProviderJWT,
				Audience:                "",
				Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
			},
		},
	}
	queueCfg := &queuescfg.Config{DataChangesTopicName: "data_changes"}

	pp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishFunc:      func(_ context.Context, _ any) error { return nil },
				PublishAsyncFunc: func(_ context.Context, _ any) {},
				StopFunc:         func() {},
			}, nil
		},
	}

	s, err := ProvideService(
		ctx,
		logger,
		cfg,
		&mockauthn.AuthenticatorMock{},
		&mocktotp.VerifierMock{},
		&oauthmock.RepositoryMock{},
		&identitymanagermock.IdentityDataManagerMock{},
		tracingnoop.NewTracerProvider(),
		pp,
		queueCfg,
	)
	require.NoError(t, err)

	return s.(*service)
}

func TestProvideService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		cfg := &Config{
			Tokens: authcfg.TokensConfig{
				Config: tokenscfg.Config{
					Provider:                tokenscfg.ProviderJWT,
					Audience:                "",
					Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
				},
			},
		}
		queueCfg := &queuescfg.Config{DataChangesTopicName: "data_changes"}

		pp := &mockpublishers.PublisherProviderMock{
			NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
				return &mockpublishers.PublisherMock{
					PublishFunc:      func(_ context.Context, _ any) error { return nil },
					PublishAsyncFunc: func(_ context.Context, _ any) {},
					StopFunc:         func() {},
				}, nil
			},
		}

		s, err := ProvideService(
			ctx,
			logger,
			cfg,
			&mockauthn.AuthenticatorMock{},
			&mocktotp.VerifierMock{},
			&oauthmock.RepositoryMock{},
			&identitymanagermock.IdentityDataManagerMock{},
			tracingnoop.NewTracerProvider(),
			pp,
			queueCfg,
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})
}
