package waitlists

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v15/callers"
	platformwaitlists "github.com/primandproper/platform-go/v15/waitlists"
	waitlistsmock "github.com/primandproper/platform-go/v15/waitlists/mock"
	"github.com/primandproper/primitives-go/v2/database"
	databasemock "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signedIn is a request from userID holding perms service-wide.
func signedIn(t *testing.T, userID string, perms ...authorization.Permission) (context.Context, callers.Principal) {
	t.Helper()

	ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:             userID,
			ServicePermissions: authorization.NewServiceRolePermissionChecker(nil, perms),
		},
	})

	principal, ok := sessions.PrincipalFromContext(ctx)
	require.True(t, ok)

	return ctx, principal
}

// signupBy is an authorizer over a store holding one signup, made by subjectID.
func signupBy(subjectID string) *signupFixture {
	signup := &platformwaitlists.Signup{
		ID:      fake.BuildFakeID(),
		ListID:  fake.BuildFakeID(),
		Subject: platformwaitlists.Subject{Type: platformwaitlists.SubjectUser, ID: subjectID},
	}

	store := &waitlistsmock.SignupStoreMock{
		GetSignupFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, string) (*platformwaitlists.Signup, error) {
			return signup, nil
		},
	}
	db := &databasemock.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }}

	return &signupFixture{signup: signup, authorizer: ownSignupOrAdmin(store, db, sessions.GrantsFromContext)}
}

type signupFixture struct {
	_ struct{} `json:"-"`

	signup     *platformwaitlists.Signup
	authorizer interface {
		AuthorizeWithdrawal(ctx context.Context, caller callers.Principal, scope tenancy.Scope, listID, signupID string) error
		AuthorizeSubjectRead(ctx context.Context, caller callers.Principal, scope tenancy.Scope, subject platformwaitlists.Subject) error
	}
}

func (f *signupFixture) withdraw(ctx context.Context, caller callers.Principal) error {
	return f.authorizer.AuthorizeWithdrawal(ctx, caller, tenancy.Global(), f.signup.ListID, f.signup.ID)
}

func TestOwnSignupOrAdmin_Withdrawal(T *testing.T) {
	T.Parallel()

	T.Run("permits the person who joined", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		ctx, principal := signedIn(t, userID)

		assert.NoError(t, signupBy(userID).withdraw(ctx, principal))
	})

	T.Run("refuses somebody else's signup to a caller without the grant", func(t *testing.T) {
		t.Parallel()

		// Reading every signup is not taking anybody off a list.
		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.ReadWaitlistSignupsPermission)

		assert.ErrorIs(t, signupBy(fake.BuildFakeID()).withdraw(ctx, principal), callers.ErrTargetNotPermitted)
	})

	T.Run("permits somebody else's signup to a holder of the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.WithdrawAnyWaitlistSignupsPermission)

		assert.NoError(t, signupBy(fake.BuildFakeID()).withdraw(ctx, principal))
	})

	T.Run("a service admin holds the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.ServiceAdminPermissions...)

		assert.NoError(t, signupBy(fake.BuildFakeID()).withdraw(ctx, principal))
	})

	T.Run("refuses an anonymous request naming a signup by id", func(t *testing.T) {
		t.Parallel()

		assert.ErrorIs(t, signupBy(fake.BuildFakeID()).withdraw(t.Context(), nil), callers.ErrTargetNotPermitted)
	})
}

func TestOwnSignupOrAdmin_SubjectRead(T *testing.T) {
	T.Parallel()

	read := func(ctx context.Context, caller callers.Principal, subjectID string) error {
		return signupBy(subjectID).authorizer.AuthorizeSubjectRead(ctx, caller, tenancy.Global(),
			platformwaitlists.Subject{Type: platformwaitlists.SubjectUser, ID: subjectID})
	}

	T.Run("permits a caller asking about themselves", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		ctx, principal := signedIn(t, userID, authorization.ReadOwnWaitlistSignupsPermission)

		assert.NoError(t, read(ctx, principal, userID))
	})

	T.Run("refuses somebody else to a caller without the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.AccountMemberPermissions...)

		assert.ErrorIs(t, read(ctx, principal, fake.BuildFakeID()), callers.ErrTargetNotPermitted)
	})

	T.Run("permits somebody else to a holder of the signup read", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.ReadWaitlistSignupsPermission)

		assert.NoError(t, read(ctx, principal, fake.BuildFakeID()))
	})

	T.Run("a service admin holds the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.ServiceAdminPermissions...)

		assert.NoError(t, read(ctx, principal, fake.BuildFakeID()))
	})

	T.Run("refuses a request with no session", func(t *testing.T) {
		t.Parallel()

		assert.ErrorIs(t, read(t.Context(), nil, fake.BuildFakeID()), callers.ErrTargetNotPermitted)
	})
}
