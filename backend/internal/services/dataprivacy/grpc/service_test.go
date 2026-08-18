package grpc

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"

	platformdataprivacy "github.com/primandproper/platform-go/v11/dataprivacy"
	dataprivacymock "github.com/primandproper/platform-go/v11/dataprivacy/mock"
	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T, requests *dataprivacymock.ServiceMock) *serviceImpl {
	t.Helper()

	if requests == nil {
		requests = &dataprivacymock.ServiceMock{}
	}

	return &serviceImpl{
		tracer:   tracing.NewTracerForTest(t.Name()),
		logger:   loggingnoop.NewLogger(),
		requests: requests,
	}
}

func sessionContextForUser(t *testing.T, userID string) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester: sessions.RequesterInfo{UserID: userID},
	})
}

func TestNewDataPrivacyService(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, NewDataPrivacyService(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&dataprivacymock.ServiceMock{},
	))
}

func TestServiceImpl_AggregateUserDataReport(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()
		requestID := identifiers.New()

		requests := &dataprivacymock.ServiceMock{
			SubmitFunc: func(_ context.Context, subject platformdataprivacy.Subject, requestType platformdataprivacy.RequestType) (*platformdataprivacy.Request, error) {
				assert.Equal(t, userID, subject.ID)
				assert.Equal(t, platformdataprivacy.SubjectUser, subject.Type)
				assert.Equal(t, platformdataprivacy.RequestExport, requestType)
				// Empty scope, which means every scope the subject appears in — what a
				// plain "give me my data" asks for.
				assert.Empty(t, subject.Scope)

				return &platformdataprivacy.Request{
					ID:      requestID,
					Subject: subject,
					Type:    requestType,
					Status:  platformdataprivacy.StatusInProgress,
				}, nil
			},
		}

		res, err := buildTestService(t, requests).AggregateUserDataReport(
			sessionContextForUser(t, userID),
			&dataprivacysvc.AggregateUserDataReportRequest{},
		)

		require.NoError(t, err)
		assert.Equal(t, requestID, res.GetRequest().GetId())
		assert.Equal(t, string(platformdataprivacy.StatusInProgress), res.GetRequest().GetStatus())
	})

	T.Run("without a session", func(t *testing.T) {
		t.Parallel()

		res, err := buildTestService(t, nil).AggregateUserDataReport(
			t.Context(),
			&dataprivacysvc.AggregateUserDataReportRequest{},
		)

		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestServiceImpl_DestroyAllUserData(T *testing.T) {
	T.Parallel()

	T.Run("queues an erasure rather than performing one", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()

		requests := &dataprivacymock.ServiceMock{
			SubmitFunc: func(_ context.Context, subject platformdataprivacy.Subject, requestType platformdataprivacy.RequestType) (*platformdataprivacy.Request, error) {
				assert.Equal(t, platformdataprivacy.RequestErasure, requestType)

				return &platformdataprivacy.Request{
					ID:      identifiers.New(),
					Subject: subject,
					Type:    requestType,
					Status:  platformdataprivacy.StatusInProgress,
				}, nil
			},
		}

		res, err := buildTestService(t, requests).DestroyAllUserData(
			sessionContextForUser(t, userID),
			&dataprivacysvc.DestroyAllUserDataRequest{},
		)

		require.NoError(t, err)
		// Pending, not completed. The deletion happens in the fulfillment worker, inside
		// one transaction shared by every domain's eraser.
		assert.Equal(t, string(platformdataprivacy.StatusInProgress), res.GetRequest().GetStatus())
		assert.Equal(t, "erasure", res.GetRequest().GetRequestType())
	})
}

func TestServiceImpl_FetchUserDataReport(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()
		requestID := identifiers.New()
		body := `{"identity":{"user":{}}}`

		requests := &dataprivacymock.ServiceMock{
			GetFunc: func(context.Context, string) (*platformdataprivacy.Request, error) {
				return &platformdataprivacy.Request{
					ID:      requestID,
					Subject: platformdataprivacy.Subject{ID: userID},
					Status:  platformdataprivacy.StatusCompleted,
				}, nil
			},
			OpenFunc: func(_ context.Context, actualRequestID string) (io.ReadCloser, error) {
				assert.Equal(t, requestID, actualRequestID)

				return io.NopCloser(strings.NewReader(body)), nil
			},
		}

		res, err := buildTestService(t, requests).FetchUserDataReport(
			sessionContextForUser(t, userID),
			&dataprivacysvc.FetchUserDataReportRequest{DataPrivacyRequestId: requestID},
		)

		require.NoError(t, err)
		assert.JSONEq(t, body, string(res.GetArtifact()))
	})

	T.Run("refuses another subject's artifact without opening it", func(t *testing.T) {
		t.Parallel()

		// The ownership check happens against the request row, before a byte is read. An
		// Open call here would mean the artifact was fetched for somebody it is not
		// about — and NotFound rather than PermissionDenied, because a distinct denial
		// would confirm the request exists.
		requests := &dataprivacymock.ServiceMock{
			GetFunc: func(context.Context, string) (*platformdataprivacy.Request, error) {
				return &platformdataprivacy.Request{
					ID:      identifiers.New(),
					Subject: platformdataprivacy.Subject{ID: identifiers.New()},
					Status:  platformdataprivacy.StatusCompleted,
				}, nil
			},
		}

		res, err := buildTestService(t, requests).FetchUserDataReport(
			sessionContextForUser(t, identifiers.New()),
			&dataprivacysvc.FetchUserDataReportRequest{DataPrivacyRequestId: identifiers.New()},
		)

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Empty(t, requests.OpenCalls())
	})

	T.Run("an expired artifact is NotFound rather than an internal error", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()

		requests := &dataprivacymock.ServiceMock{
			GetFunc: func(context.Context, string) (*platformdataprivacy.Request, error) {
				return &platformdataprivacy.Request{
					ID:      identifiers.New(),
					Subject: platformdataprivacy.Subject{ID: userID},
					Status:  platformdataprivacy.StatusExpired,
				}, nil
			},
			OpenFunc: func(context.Context, string) (io.ReadCloser, error) {
				return nil, platformerrors.Wrap(platformdataprivacy.ErrArtifactUnavailable, "expired")
			},
		}

		res, err := buildTestService(t, requests).FetchUserDataReport(
			sessionContextForUser(t, userID),
			&dataprivacysvc.FetchUserDataReportRequest{DataPrivacyRequestId: identifiers.New()},
		)

		assert.Nil(t, res)
		// FailedPrecondition, not the NotFound this service asks for. platform-go v10 maps its
		// own sentinels to gRPC codes centrally, and that mapping wins over the code a caller
		// passes — ErrArtifactUnavailable is FailedPrecondition there. This is a client-visible
		// change from v9 and worth revisiting: "your export expired" reads more like NotFound
		// than like a precondition the client can fix and retry.
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	})
}

func TestServiceImpl_GetDataPrivacyRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()
		requestID := identifiers.New()

		requests := &dataprivacymock.ServiceMock{
			GetFunc: func(_ context.Context, actualRequestID string) (*platformdataprivacy.Request, error) {
				assert.Equal(t, requestID, actualRequestID)

				return &platformdataprivacy.Request{
					ID:      requestID,
					Subject: platformdataprivacy.Subject{ID: userID},
					Type:    platformdataprivacy.RequestExport,
					Status:  platformdataprivacy.StatusCompleted,
				}, nil
			},
		}

		res, err := buildTestService(t, requests).GetDataPrivacyRequest(
			sessionContextForUser(t, userID),
			&dataprivacysvc.GetDataPrivacyRequestRequest{DataPrivacyRequestId: requestID},
		)

		require.NoError(t, err)
		assert.Equal(t, requestID, res.GetRequest().GetId())
	})

	T.Run("with another subject's request", func(t *testing.T) {
		t.Parallel()

		requests := &dataprivacymock.ServiceMock{
			GetFunc: func(context.Context, string) (*platformdataprivacy.Request, error) {
				return &platformdataprivacy.Request{
					ID:      identifiers.New(),
					Subject: platformdataprivacy.Subject{ID: identifiers.New()},
				}, nil
			},
		}

		res, err := buildTestService(t, requests).GetDataPrivacyRequest(
			sessionContextForUser(t, identifiers.New()),
			&dataprivacysvc.GetDataPrivacyRequestRequest{DataPrivacyRequestId: identifiers.New()},
		)

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestServiceImpl_ListDataPrivacyRequests(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()
		exampleRequest := &platformdataprivacy.Request{
			ID:      identifiers.New(),
			Subject: platformdataprivacy.Subject{ID: userID},
			Type:    platformdataprivacy.RequestExport,
			Status:  platformdataprivacy.StatusCompleted,
		}

		requests := &dataprivacymock.ServiceMock{
			ListFunc: func(_ context.Context, subject platformdataprivacy.Subject, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformdataprivacy.Request], error) {
				// Scoped to the requester, so a subject can only ever see what has been
				// asked in their own name.
				assert.Equal(t, userID, subject.ID)

				return filtering.NewQueryFilteredResult(
					[]*platformdataprivacy.Request{exampleRequest},
					1, 1,
					func(r *platformdataprivacy.Request) string { return r.ID },
					filtering.DefaultQueryFilter(),
				), nil
			},
		}

		res, err := buildTestService(t, requests).ListDataPrivacyRequests(
			sessionContextForUser(t, userID),
			&dataprivacysvc.ListDataPrivacyRequestsRequest{},
		)

		require.NoError(t, err)
		require.Len(t, res.GetData(), 1)
		assert.Equal(t, exampleRequest.ID, res.GetData()[0].GetId())
	})

	T.Run("with error listing", func(t *testing.T) {
		t.Parallel()

		requests := &dataprivacymock.ServiceMock{
			ListFunc: func(context.Context, platformdataprivacy.Subject, *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformdataprivacy.Request], error) {
				return nil, platformerrors.New("blah")
			},
		}

		res, err := buildTestService(t, requests).ListDataPrivacyRequests(
			sessionContextForUser(t, identifiers.New()),
			&dataprivacysvc.ListDataPrivacyRequestsRequest{},
		)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}
