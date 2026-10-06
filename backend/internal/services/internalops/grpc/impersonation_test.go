package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	msgconfig "github.com/primandproper/primitives-go/v2/messagequeue/config"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeImpersonator answers from its fields and records who it was asked to impersonate as.
type fakeImpersonator struct {
	signIn    *signin.SignIn
	err       error
	operator  string
	subject   string
	accountID string
	called    bool
}

func (f *fakeImpersonator) IssueImpersonationToken(_ context.Context, _ tenancy.Scope, operatorID string, _ tenancy.Scope, subjectID, accountID string) (*signin.SignIn, error) {
	f.called = true
	f.operator, f.subject, f.accountID = operatorID, subjectID, accountID

	return f.signIn, f.err
}

func asCaller(ctx context.Context, userID, impersonatorID string) context.Context {
	return sessions.AttachToContext(ctx, &sessions.ContextData{
		Requester:      sessions.RequesterInfo{UserID: userID},
		ImpersonatorID: impersonatorID,
	})
}

func buildImpersonationService(impersonator Impersonator) *serviceImpl {
	return NewService(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), &msgconfig.Config{}, nil, impersonator).(*serviceImpl)
}

func TestServiceImpl_ImpersonateUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		operatorID, subjectID, accountID := identifiers.New(), identifiers.New(), identifiers.New()
		impersonator := &fakeImpersonator{signIn: &signin.SignIn{
			Token:     identifiers.New(),
			ExpiresAt: time.Now().Add(15 * time.Minute),
			Principal: &platformidentity.Principal{ActiveAccountID: accountID},
		}}

		res, err := buildImpersonationService(impersonator).ImpersonateUser(asCaller(t.Context(), operatorID, ""), &internalopssvc.ImpersonateUserRequest{
			SubjectId: subjectID,
			AccountId: accountID,
		})
		require.NoError(t, err)

		assert.Equal(t, impersonator.signIn.Token, res.GetToken())
		assert.Equal(t, accountID, res.GetAccountId())

		// The operator is the caller, never a field.
		assert.Equal(t, operatorID, impersonator.operator)
		assert.Equal(t, subjectID, impersonator.subject)
		assert.Equal(t, accountID, impersonator.accountID)
	})

	T.Run("refuses a caller with no session", func(t *testing.T) {
		t.Parallel()

		impersonator := &fakeImpersonator{}

		_, err := buildImpersonationService(impersonator).ImpersonateUser(t.Context(), &internalopssvc.ImpersonateUserRequest{SubjectId: identifiers.New()})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
		assert.False(t, impersonator.called)
	})

	T.Run("refuses a caller already acting through an impersonation", func(t *testing.T) {
		t.Parallel()

		impersonator := &fakeImpersonator{}

		_, err := buildImpersonationService(impersonator).ImpersonateUser(asCaller(t.Context(), identifiers.New(), identifiers.New()), &internalopssvc.ImpersonateUserRequest{
			SubjectId: identifiers.New(),
		})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.False(t, impersonator.called)
	})

	T.Run("refuses a request naming nobody", func(t *testing.T) {
		t.Parallel()

		impersonator := &fakeImpersonator{}

		_, err := buildImpersonationService(impersonator).ImpersonateUser(asCaller(t.Context(), identifiers.New(), ""), &internalopssvc.ImpersonateUserRequest{})
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.False(t, impersonator.called)
	})

	T.Run("passes along signin's refusal", func(t *testing.T) {
		t.Parallel()

		impersonator := &fakeImpersonator{err: signin.ErrImpersonationDisabled}

		_, err := buildImpersonationService(impersonator).ImpersonateUser(asCaller(t.Context(), identifiers.New(), ""), &internalopssvc.ImpersonateUserRequest{
			SubjectId: identifiers.New(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("answers the impersonation policy's refusal as PermissionDenied", func(t *testing.T) {
		t.Parallel()

		// As signin hands it back: the policy's error wrapped.
		impersonator := &fakeImpersonator{err: platformerrors.Wrap(authentication.ErrImpersonationNotPermitted, "the impersonation policy refused")}

		_, err := buildImpersonationService(impersonator).ImpersonateUser(asCaller(t.Context(), identifiers.New(), ""), &internalopssvc.ImpersonateUserRequest{
			SubjectId: identifiers.New(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}
