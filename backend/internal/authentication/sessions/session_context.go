package sessions

import (
	"context"
	"encoding/gob"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformkeys "github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/observability/logging"
)

func init() {
	gob.Register(&ContextData{})
}

// contextKey is the unexported type of this package's context keys. It is unexported and
// distinct so a value stored here cannot collide with one stored by any other package, which
// is what the string-typed key platform-go v10 removed could not promise. The name is on the
// value rather than in a field so that a key printed in a panic or a dump still says what it is.
type contextKey string

// SessionContextDataKey keys the session data on a context.
const SessionContextDataKey contextKey = "session_context_data"

// AttachToContext returns a copy of ctx carrying x as its session data. This is the only supported
// way to populate the session context key, so that reads and writes stay in one place.
func AttachToContext(ctx context.Context, x *ContextData) context.Context {
	return context.WithValue(ctx, SessionContextDataKey, x)
}

// FromContext returns the ContextData attached to ctx, or nil if none is present.
//
// Every route the authentication interceptor guards has ContextData attached before the handler
// runs, so handlers behind it can treat a nil return as impossible. Use this on the routes that
// are deliberately reachable without authentication, where absence is an expected state rather
// than an error; everywhere else, prefer RequireFromContext.
func FromContext(ctx context.Context) *ContextData {
	sessionCtxData, ok := ctx.Value(SessionContextDataKey).(*ContextData)
	if !ok {
		return nil
	}

	return sessionCtxData
}

// RequireFromContext returns the ContextData attached to ctx, or ErrAuthenticationNotFound when
// none is present. The registered gRPC and HTTP error mappers translate that sentinel into an
// unauthenticated response, so callers do not need to classify it themselves.
func RequireFromContext(ctx context.Context) (*ContextData, error) {
	if sessionCtxData := FromContext(ctx); sessionCtxData != nil {
		return sessionCtxData, nil
	}

	return nil, ErrAuthenticationNotFound
}

// ContextData represents what we encode in our passwords cookies.
type ContextData struct {
	_ struct{} `json:"-"`

	AccountPermissions map[string]authorization.AccountRolePermissionsChecker `json:"-"`
	Requester          RequesterInfo                                          `json:"-"`
	ActiveAccountID    string                                                 `json:"-"`
	SessionID          string                                                 `json:"-"`
}

// RequesterInfo contains data relevant to the user making a request.
type RequesterInfo struct {
	_ struct{} `json:"-"`

	ServicePermissions       authorization.ServiceRolePermissionChecker `json:"-"`
	AccountStatus            string                                     `json:"-"`
	AccountStatusExplanation string                                     `json:"-"`
	UserID                   string                                     `json:"-"`
	EmailAddress             string                                     `json:"-"`
	Username                 string                                     `json:"-"`
}

// The getters below are nil-safe so that callers on unauthenticated routes can read through a nil
// ContextData without a preceding nil check. A nil receiver yields the zero value, which is the
// same thing an anonymous caller would produce.

// GetUserID is a simple getter.
func (x *ContextData) GetUserID() string {
	if x == nil {
		return ""
	}

	return x.Requester.UserID
}

// GetEmailAddress is a simple getter.
func (x *ContextData) GetEmailAddress() string {
	if x == nil {
		return ""
	}

	return x.Requester.EmailAddress
}

// GetUsername is a simple getter.
func (x *ContextData) GetUsername() string {
	if x == nil {
		return ""
	}

	return x.Requester.Username
}

// GetServicePermissions is a simple getter.
// GetServicePermissions is nil-safe in both directions: a nil ContextData and one whose
// Requester carries no checker both answer with a checker over no roles and no permissions.
//
// A nil interface would satisfy the "yields the zero value" promise above only until
// somebody called a method on it, which is the whole point of reading through this on an
// unauthenticated route — every caller in the tree calls IsServiceAdmin or HasPermission on
// the result immediately, and a nil return makes each of those a panic on exactly the
// request that has no session. The empty checker says no to both, which is what an
// anonymous caller should hear.
func (x *ContextData) GetServicePermissions() authorization.ServiceRolePermissionChecker {
	if x == nil || x.Requester.ServicePermissions == nil {
		return authorization.NewServiceRolePermissionChecker(nil, nil)
	}

	return x.Requester.ServicePermissions
}

// GetActiveAccountID is a simple getter.
func (x *ContextData) GetActiveAccountID() string {
	if x == nil {
		return ""
	}

	return x.ActiveAccountID
}

// GetSessionID is a simple getter.
func (x *ContextData) GetSessionID() string {
	if x == nil {
		return ""
	}

	return x.SessionID
}

// AccountRolePermissionsChecker returns the relevant AccountRolePermissionsChecker.
func (x *ContextData) AccountRolePermissionsChecker() authorization.AccountRolePermissionsChecker {
	if x != nil {
		if checker, ok := x.AccountPermissions[x.ActiveAccountID]; ok {
			return checker
		}
	}

	return authorization.NewAccountRolePermissionChecker(nil)
}

// ServiceRolePermissionChecker returns the relevant ServiceRolePermissionChecker.
func (x *ContextData) ServiceRolePermissionChecker() authorization.ServiceRolePermissionChecker {
	return x.GetServicePermissions()
}

// AttachToLogger provides a consistent way to attach a ContextData object to a logger.
func (x *ContextData) AttachToLogger(logger logging.Logger) logging.Logger {
	if x != nil {
		logger = logger.WithValue(platformkeys.RequesterIDKey, x.GetUserID()).
			WithValue(platformkeys.ActiveAccountIDKey, x.ActiveAccountID)
	}

	return logger
}
