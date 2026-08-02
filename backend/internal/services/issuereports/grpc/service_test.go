package grpc

import (
	"testing"

	commentsmanagermock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/manager/mock"
	issuereportmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/mock"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		issueReportsManager := &issuereportmock.RepositoryMock{}
		commentsManager := &commentsmanagermock.CommentsDataManagerMock{}

		service := NewService(logger, tracerProvider, issueReportsManager, commentsManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*issuereportssvc.IssueReportsServiceServer)(nil), service)
	})
}
