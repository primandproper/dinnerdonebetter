package authentication

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/stretchr/testify/require"
)

type authServiceHTTPRoutesTestHelper struct {
	ctx                 context.Context
	req                 *http.Request
	res                 *httptest.ResponseRecorder
	sessionCtxData      *sessions.ContextData
	service             *service
	exampleUser         *identity.User
	exampleAccount      *identity.Account
	examplePermCheckers map[string]authorization.AccountRolePermissionsChecker
	exampleLoginInput   *auth.UserLoginInput
}

func (helper *authServiceHTTPRoutesTestHelper) setContextFetcher(t *testing.T) {
	t.Helper()

	sessionCtxData := &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:                   helper.exampleUser.ID,
			AccountStatus:            helper.exampleUser.AccountStatus,
			AccountStatusExplanation: helper.exampleUser.AccountStatusExplanation,
			ServicePermissions:       authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil),
		},
		ActiveAccountID:    helper.exampleAccount.ID,
		AccountPermissions: helper.examplePermCheckers,
	}

	helper.sessionCtxData = sessionCtxData
}

func buildTestHelper(t *testing.T) *authServiceHTTPRoutesTestHelper {
	t.Helper()

	helper := &authServiceHTTPRoutesTestHelper{}

	helper.ctx = t.Context()
	helper.service = buildTestService(t)
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleLoginInput = authfakes.BuildFakeUserLoginInputFromUser(helper.exampleUser)

	helper.examplePermCheckers = map[string]authorization.AccountRolePermissionsChecker{
		helper.exampleAccount.ID: authorization.NewAccountRolePermissionChecker(nil),
	}

	helper.setContextFetcher(t)

	var err error

	helper.res = httptest.NewRecorder()
	helper.req, err = http.NewRequestWithContext(
		helper.ctx,
		http.MethodGet,
		"https://whatever.whocares.gov",
		http.NoBody,
	)
	require.NotNil(t, helper.req)
	require.NoError(t, err)

	return helper
}
