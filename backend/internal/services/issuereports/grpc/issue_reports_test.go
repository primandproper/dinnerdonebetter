package grpc

import (
	"context"
	"testing"

	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
	issuereportsmock "github.com/primandproper/platform-go/v13/issuereports/mock"
	"github.com/primandproper/platform-go/v13/pointer"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// assertCode asserts that err reached the client as the given gRPC code.
func assertCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()

	require.Error(t, err)
	assert.Equal(t, expected, status.Code(err), "got %v", err)
}

func TestServiceImpl_CreateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("files the report as the session's user and account", func(t *testing.T) {
		t.Parallel()

		ctx, userID, accountID := buildSessionContextForTest(t)
		example := issuereportfakes.BuildFakeIssueReport()

		var (
			written         *issuereports.Report
			statusOnArrival issuereports.Status
		)
		store := &issuereportsmock.StoreMock{
			CreateReportFunc: func(_ context.Context, report *issuereports.Report) error {
				statusOnArrival = report.Status

				report.ID = fake.BuildFakeID()
				report.Status = issuereports.StatusOpen
				written = report

				return nil
			},
		}

		res, err := buildTestService(t, store, nil).CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{
			Input: converters.ConvertIssueReportToGRPCIssueReportCreationRequestInput(example),
		})
		require.NoError(t, err)
		require.NotNil(t, res.GetCreated())

		// The reporter and the scope come off the session, never off the request: a
		// report that could name either is one anybody could file as anybody, into
		// anybody's queue.
		assert.Equal(t, userID, written.Reporter)
		assert.Equal(t, ddbissuereports.Scope(accountID), written.Scope)
		assert.Equal(t, example.Kind, written.Kind)
		assert.Equal(t, example.Details, written.Details)
		assert.Equal(t, example.SubjectType, written.SubjectType)
		assert.Equal(t, example.SubjectID, written.SubjectID)

		// The status is the store's to assign: the creation input has no field for
		// it, because a report that started resolved is one nobody resolved.
		assert.Empty(t, statusOnArrival)
		assert.Equal(t, issuereports.StatusOpen.String(), res.GetCreated().GetStatus())
	})

	t.Run("refuses a report with nothing in it", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		store := &issuereportsmock.StoreMock{
			CreateReportFunc: func(context.Context, *issuereports.Report) error {
				return issuereports.ErrEmptyDetails
			},
		}

		_, err := buildTestService(t, store, nil).CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{
			Input: &issuereportssvc.IssueReportCreationRequestInput{Kind: "bug"},
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("refuses a request with no input", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		_, err := buildTestService(t, nil, nil).CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil, nil).CreateIssueReport(t.Context(), &issuereportssvc.CreateIssueReportRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_GetIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("reads within the session's account", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		example := issuereportfakes.BuildFakeIssueReportForScope(accountID)

		var readScope tenancy.Scope
		store := &issuereportsmock.StoreMock{
			GetReportFunc: func(_ context.Context, scope tenancy.Scope, _ string) (*issuereports.Report, error) {
				readScope = scope

				return example, nil
			},
		}

		res, err := buildTestService(t, store, nil).GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{
			IssueReportId: example.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, ddbissuereports.Scope(accountID), readScope)
		assert.Equal(t, example.ID, res.GetResult().GetId())
		assert.Equal(t, example.Kind, res.GetResult().GetKind())
	})

	// A report in another account is absent rather than forbidden. The alternative
	// tells the caller which report IDs exist in accounts they cannot see.
	t.Run("another account's report is not found", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		store := &issuereportsmock.StoreMock{
			GetReportFunc: func(context.Context, tenancy.Scope, string) (*issuereports.Report, error) {
				return nil, platformerrors.Wrap(issuereports.ErrReportNotFound, "reading")
			},
		}

		_, err := buildTestService(t, store, nil).GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
		})
		assertCode(t, err, codes.NotFound)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil, nil).GetIssueReport(t.Context(), &issuereportssvc.GetIssueReportRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_GetIssueReports(t *testing.T) {
	t.Parallel()

	t.Run("pages the session's account", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		page := issuereportfakes.BuildFakeIssueReportList(accountID)

		var readScope tenancy.Scope
		store := &issuereportsmock.StoreMock{
			ListReportsFunc: func(_ context.Context, scope tenancy.Scope, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.Report], error) {
				readScope = scope

				return page, nil
			},
		}

		res, err := buildTestService(t, store, nil).GetIssueReports(ctx, &issuereportssvc.GetIssueReportsRequest{})
		require.NoError(t, err)
		assert.Equal(t, ddbissuereports.Scope(accountID), readScope)
		assert.Len(t, res.GetResults(), len(page.Data))
		assert.NotNil(t, res.GetPagination())
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil, nil).GetIssueReports(t.Context(), &issuereportssvc.GetIssueReportsRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_GetIssueReportsByStatus(t *testing.T) {
	t.Parallel()

	t.Run("reads the queue the caller named", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		page := issuereportfakes.BuildFakeIssueReportList(accountID)

		var readStatus issuereports.Status
		store := &issuereportsmock.StoreMock{
			ListReportsByStatusFunc: func(_ context.Context, _ tenancy.Scope, status issuereports.Status, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.Report], error) {
				readStatus = status

				return page, nil
			},
		}

		// Mixed case on purpose: the status arrives from a console's form, and
		// "Open" and "open" are one queue rather than two.
		res, err := buildTestService(t, store, nil).GetIssueReportsByStatus(ctx, &issuereportssvc.GetIssueReportsByStatusRequest{
			Status: "Open",
		})
		require.NoError(t, err)
		assert.Equal(t, issuereports.StatusOpen, readStatus)
		assert.Len(t, res.GetResults(), len(page.Data))
	})

	// An unknown status is refused rather than answered with an empty page, because
	// an empty page is exactly what a misspelled queue looks like.
	t.Run("refuses a status nothing serves", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		_, err := buildTestService(t, nil, nil).GetIssueReportsByStatus(ctx, &issuereportssvc.GetIssueReportsByStatusRequest{
			Status: "closed",
		})
		assertCode(t, err, codes.InvalidArgument)
	})
}

func TestServiceImpl_GetIssueReportsBySubject(t *testing.T) {
	t.Parallel()

	t.Run("by subject type", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		page := issuereportfakes.BuildFakeIssueReportList(accountID)

		var readType string
		store := &issuereportsmock.StoreMock{
			ListReportsBySubjectTypeFunc: func(_ context.Context, _ tenancy.Scope, subjectType string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.Report], error) {
				readType = subjectType

				return page, nil
			},
		}

		res, err := buildTestService(t, store, nil).GetIssueReportsBySubjectType(ctx, &issuereportssvc.GetIssueReportsBySubjectTypeRequest{
			SubjectType: "recipes",
		})
		require.NoError(t, err)
		assert.Equal(t, "recipes", readType)
		assert.Len(t, res.GetResults(), len(page.Data))
	})

	t.Run("for one subject", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		page := issuereportfakes.BuildFakeIssueReportList(accountID)
		subjectID := fake.BuildFakeID()

		var readType, readID string
		store := &issuereportsmock.StoreMock{
			ListReportsForSubjectFunc: func(_ context.Context, _ tenancy.Scope, subjectType, id string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.Report], error) {
				readType, readID = subjectType, id

				return page, nil
			},
		}

		res, err := buildTestService(t, store, nil).GetIssueReportsForSubject(ctx, &issuereportssvc.GetIssueReportsForSubjectRequest{
			SubjectType: "recipes",
			SubjectId:   subjectID,
		})
		require.NoError(t, err)
		assert.Equal(t, "recipes", readType)
		assert.Equal(t, subjectID, readID)
		assert.Len(t, res.GetResults(), len(page.Data))
	})
}

func TestServiceImpl_UpdateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("merges the update onto the stored report", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		example := issuereportfakes.BuildFakeIssueReportForScope(accountID)

		var written *issuereports.Report
		store := &issuereportsmock.StoreMock{
			GetReportFunc: func(context.Context, tenancy.Scope, string) (*issuereports.Report, error) {
				return example, nil
			},
			UpdateReportFunc: func(_ context.Context, report *issuereports.Report) error {
				written = report

				return nil
			},
		}

		newDetails := fake.BuildFakeString()

		res, err := buildTestService(t, store, nil).UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: example.ID,
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{Details: pointer.To(newDetails)},
		})
		require.NoError(t, err)
		assert.Equal(t, newDetails, written.Details)

		// The fields the client did not send survive. A whole-row write built from
		// the request alone would blank every one of them.
		assert.Equal(t, example.Kind, written.Kind)
		assert.Equal(t, example.SubjectType, written.SubjectType)
		assert.Equal(t, newDetails, res.GetUpdated().GetDetails())
	})

	// The lifecycle has one door, and it is not this one.
	t.Run("cannot move the status", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		example := issuereportfakes.BuildFakeIssueReportForScope(accountID)

		var written *issuereports.Report
		store := &issuereportsmock.StoreMock{
			GetReportFunc: func(context.Context, tenancy.Scope, string) (*issuereports.Report, error) {
				return example, nil
			},
			UpdateReportFunc: func(_ context.Context, report *issuereports.Report) error {
				written = report

				return nil
			},
		}

		_, err := buildTestService(t, store, nil).UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: example.ID,
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{Kind: pointer.To("billing")},
		})
		require.NoError(t, err)
		assert.Equal(t, issuereports.StatusOpen, written.Status)
	})

	t.Run("another account's report is not found", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		store := &issuereportsmock.StoreMock{
			GetReportFunc: func(context.Context, tenancy.Scope, string) (*issuereports.Report, error) {
				return nil, issuereports.ErrReportNotFound
			},
		}

		_, err := buildTestService(t, store, nil).UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{Details: pointer.To("x")},
		})
		assertCode(t, err, codes.NotFound)
	})
}

func TestServiceImpl_TransitionIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("passes the guard the caller named", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)
		example := issuereportfakes.BuildFakeIssueReportForScope(accountID)
		resolution := "fixed in the next release"

		var gotFrom, gotTo issuereports.Status
		var gotResolution string
		store := &issuereportsmock.StoreMock{
			TransitionReportFunc: func(_ context.Context, _ tenancy.Scope, _ string, from, to issuereports.Status, note string) (*issuereports.Report, error) {
				gotFrom, gotTo, gotResolution = from, to, note

				moved := *example
				moved.Status = to
				moved.Resolution = note

				return &moved, nil
			},
		}

		res, err := buildTestService(t, store, nil).TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: example.ID,
			FromStatus:    "open",
			ToStatus:      "resolved",
			Resolution:    resolution,
		})
		require.NoError(t, err)
		assert.Equal(t, issuereports.StatusOpen, gotFrom)
		assert.Equal(t, issuereports.StatusResolved, gotTo)
		assert.Equal(t, resolution, gotResolution)
		assert.Equal(t, issuereports.StatusResolved.String(), res.GetResult().GetStatus())
		assert.Equal(t, resolution, res.GetResult().GetResolution())
	})

	// Two triagers deciding the same report: the second is told the report moved
	// rather than silently overwriting the first one's note. Aborted, not Internal,
	// because re-reading and deciding again is a sensible thing for the client to do.
	t.Run("a lost guard is a conflict", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		store := &issuereportsmock.StoreMock{
			TransitionReportFunc: func(context.Context, tenancy.Scope, string, issuereports.Status, issuereports.Status, string) (*issuereports.Report, error) {
				return nil, issuereports.ErrStatusConflict
			},
		}

		_, err := buildTestService(t, store, nil).TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
			FromStatus:    "open",
			ToStatus:      "resolved",
		})
		assertCode(t, err, codes.Aborted)
	})

	// A move the lifecycle does not admit is a precondition failure: the request was
	// well-formed and the shape of the lifecycle refused it.
	t.Run("a move the lifecycle refuses", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		store := &issuereportsmock.StoreMock{
			TransitionReportFunc: func(context.Context, tenancy.Scope, string, issuereports.Status, issuereports.Status, string) (*issuereports.Report, error) {
				return nil, issuereports.ErrInvalidStatusTransition
			},
		}

		_, err := buildTestService(t, store, nil).TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
			FromStatus:    "acknowledged",
			ToStatus:      "open",
		})
		assertCode(t, err, codes.FailedPrecondition)
	})

	t.Run("refuses a status nothing serves", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)
		s := buildTestService(t, nil, nil)

		_, err := s.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
			FromStatus:    "unfiled",
			ToStatus:      "resolved",
		})
		assertCode(t, err, codes.InvalidArgument)

		_, err = s.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
			FromStatus:    "open",
			ToStatus:      "closed",
		})
		assertCode(t, err, codes.InvalidArgument)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil, nil).TransitionIssueReport(t.Context(), &issuereportssvc.TransitionIssueReportRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}

func TestServiceImpl_ArchiveIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("archives within the session's account", func(t *testing.T) {
		t.Parallel()

		ctx, _, accountID := buildSessionContextForTest(t)

		var readScope tenancy.Scope
		store := &issuereportsmock.StoreMock{
			ArchiveReportFunc: func(_ context.Context, scope tenancy.Scope, _ string) error {
				readScope = scope

				return nil
			},
		}

		_, err := buildTestService(t, store, nil).ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
		})
		require.NoError(t, err)
		assert.Equal(t, ddbissuereports.Scope(accountID), readScope)
	})

	t.Run("another account's report is not found", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		store := &issuereportsmock.StoreMock{
			ArchiveReportFunc: func(context.Context, tenancy.Scope, string) error {
				return issuereports.ErrReportNotFound
			},
		}

		_, err := buildTestService(t, store, nil).ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
		})
		assertCode(t, err, codes.NotFound)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil, nil).ArchiveIssueReport(t.Context(), &issuereportssvc.ArchiveIssueReportRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}
