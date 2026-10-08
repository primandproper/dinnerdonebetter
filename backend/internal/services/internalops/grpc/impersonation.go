package grpc

import (
	"context"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	// The mappers that give signin's refusals, and this application's impersonation policy's,
	// their codes rather than Internal.
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/errors"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	platformkeys "github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"google.golang.org/grpc/codes"
)

var (
	// ErrImpersonationChained is an operator asking to impersonate somebody while already
	// acting through an impersonation. The caller of that request is the subject, and an
	// impersonation the subject could start is one their operator could launder through them.
	ErrImpersonationChained = platformerrors.New("an impersonation cannot start another")

	// ErrNoSubject is an impersonation that names nobody to act as.
	ErrNoSubject = platformerrors.New("an impersonation must name a subject")
)

// Impersonator mints the token an operator acts as somebody else through —
// signin.Service.IssueImpersonationToken.
type Impersonator interface {
	IssueImpersonationToken(
		ctx context.Context,
		operatorScope tenancy.Scope,
		operatorID string,
		scope tenancy.Scope,
		subjectID, accountID string,
	) (*signin.SignIn, error)
}

var _ Impersonator = (*signin.Service)(nil)

// ImpersonateUser mints a token for the subject the request names that the calling operator
// acts through.
//
// The operator is the authenticated caller and nothing else. platform's door takes an
// operator's ID on trust, which is why platform exposes no RPC for it: a request field naming
// the operator would let anybody who reached this method name whoever they liked. Who may
// impersonate at all is authentication.NewImpersonationPolicy's to say, which the sign-in
// service asks; the permission table refuses everybody else before they get here.
//
// The token is the subject's — their account, their permissions, their rows — and names the
// operator beside them, which the interceptor carries on the session and the audit log records.
// It lives fifteen minutes and has no refresh token.
func (s *serviceImpl) ImpersonateUser(ctx context.Context, request *internalopssvc.ImpersonateUserRequest) (*internalopssvc.ImpersonateUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if sessionContextData.ImpersonatorID != "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(ErrImpersonationChained, logger, span, codes.PermissionDenied, "impersonating a user")
	}

	operatorID := sessionContextData.GetUserID()
	subjectID := strings.TrimSpace(request.GetSubjectId())
	logger = logger.WithValue(identitykeys.ImpersonatorIDKey, operatorID).WithValue(platformkeys.UserIDKey, subjectID)

	if subjectID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(ErrNoSubject, logger, span, codes.InvalidArgument, "impersonating a user")
	}

	// One directory for operators and customers alike, so the operator's scope and the
	// subject's are the same one.
	signedIn, err := s.impersonator.IssueImpersonationToken(ctx,
		tenancy.Global(), operatorID,
		tenancy.Global(), subjectID, strings.TrimSpace(request.GetAccountId()),
	)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "impersonating a user")
	}

	x := &internalopssvc.ImpersonateUserResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Token:     signedIn.Token,
		ExpiresAt: grpcconverters.ConvertTimeToPBTimestamp(signedIn.ExpiresAt),
	}

	if signedIn.Principal != nil {
		x.AccountId = signedIn.Principal.ActiveAccountID
	}

	return x, nil
}
