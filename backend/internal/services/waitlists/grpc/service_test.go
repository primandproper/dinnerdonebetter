package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	"github.com/primandproper/platform-go/v13/clock"
	"github.com/primandproper/platform-go/v13/fake"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	waitlistsmock "github.com/primandproper/platform-go/v13/waitlists/mock"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

// buildTestService builds a service backed by the given store mock. A nil mock gets
// an unconfigured one, which panics if any of its methods are called — so a test that
// reaches the store where it did not arrange to fails loudly rather than on a nil
// result.
func buildTestService(t *testing.T, store *waitlistsmock.StoreMock) *serviceImpl {
	t.Helper()

	if store == nil {
		store = &waitlistsmock.StoreMock{}
	}

	return &serviceImpl{
		tracer:    tracing.NewTracerForTest(t.Name()),
		logger:    loggingnoop.NewLogger(),
		waitlists: store,
		clock:     clock.NewClock(),
	}
}

// requester is who a test is making its requests as.
type requester struct {
	ctx     context.Context
	userID  string
	email   string
	account string
}

// userContextForTest returns an ordinary signed-in user: no service role beyond
// the default one, and an address of their own.
func userContextForTest(t *testing.T) requester {
	t.Helper()

	return contextFor(t, authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil))
}

// adminContextForTest returns a service admin, which is who administers the
// catalog and works the queue.
func adminContextForTest(t *testing.T) requester {
	t.Helper()

	return contextFor(t, authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceAdminRole.String()}, authorization.ServiceAdminPermissions))
}

func contextFor(t *testing.T, permissions authorization.ServiceRolePermissionChecker) requester {
	t.Helper()

	r := requester{
		userID:  fake.BuildFakeID(),
		email:   gofakeit.Email(),
		account: fake.BuildFakeID(),
	}

	r.ctx = sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: r.account,
		Requester: sessions.RequesterInfo{
			UserID:             r.userID,
			EmailAddress:       r.email,
			ServicePermissions: permissions,
		},
	})

	return r
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&waitlistsmock.StoreMock{},
		)

		assert.NotNil(t, service)
		assert.Implements(t, (*waitlistssvc.WaitlistsServiceServer)(nil), service)
	})
}

func TestProvideMethodPermissions(t *testing.T) {
	t.Parallel()

	t.Run("covers every method the service serves", func(t *testing.T) {
		t.Parallel()

		permissions := ProvideMethodPermissions()

		// A method with no entry is refused by the interceptor, so an RPC added
		// without one is an RPC nobody can call. Listing them here is what makes that
		// a test failure rather than a support ticket.
		for _, method := range []string{
			waitlistssvc.WaitlistsService_CreateWaitlist_FullMethodName,
			waitlistssvc.WaitlistsService_GetWaitlist_FullMethodName,
			waitlistssvc.WaitlistsService_GetWaitlists_FullMethodName,
			waitlistssvc.WaitlistsService_GetOpenWaitlists_FullMethodName,
			waitlistssvc.WaitlistsService_UpdateWaitlist_FullMethodName,
			waitlistssvc.WaitlistsService_ArchiveWaitlist_FullMethodName,
			waitlistssvc.WaitlistsService_WaitlistIsOpen_FullMethodName,
			waitlistssvc.WaitlistsService_JoinWaitlist_FullMethodName,
			waitlistssvc.WaitlistsService_GetWaitlistSignup_FullMethodName,
			waitlistssvc.WaitlistsService_GetWaitlistSignupsForWaitlist_FullMethodName,
			waitlistssvc.WaitlistsService_UpdateWaitlistSignup_FullMethodName,
			waitlistssvc.WaitlistsService_InviteWaitlistSignup_FullMethodName,
			waitlistssvc.WaitlistsService_ConvertWaitlistSignup_FullMethodName,
			waitlistssvc.WaitlistsService_WithdrawFromWaitlist_FullMethodName,
			waitlistssvc.WaitlistsService_ArchiveWaitlistSignup_FullMethodName,
		} {
			assert.NotEmpty(t, permissions[method], method)
		}
	})
}
