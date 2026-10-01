package grpc

import (
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
)

// AnonymousMethods are every RPC on this service, each reached without a session: a sign-up
// has nobody to be yet, and an exchange's authority is the refresh token it carries, which
// is presented exactly when the access token it would have been sent with has expired.
func AnonymousMethods() []string {
	return []string{
		authsvc.AuthService_RegisterUser_FullMethodName,
		authsvc.AuthService_ExchangeToken_FullMethodName,
	}
}
