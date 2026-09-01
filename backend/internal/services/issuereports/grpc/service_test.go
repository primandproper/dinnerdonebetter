package grpc

import (
	"testing"

	issuereportmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/mock"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	commentsmock "github.com/primandproper/platform-go/v13/comments/mock"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		issueReportsManager := &issuereportmock.RepositoryMock{}
		commentStore := &commentsmock.StoreMock{}

		service := NewService(logger, tracerProvider, issueReportsManager, commentStore)

		assert.NotNil(t, service)
		assert.Implements(t, (*issuereportssvc.IssueReportsServiceServer)(nil), service)
	})
}
