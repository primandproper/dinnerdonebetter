package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/keys"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"
	waitlists "github.com/primandproper/platform-go/v13/waitlists"

	"google.golang.org/grpc/codes"
)

// The two refusals this service makes on its own, before the store is reached.
//
// They are here rather than in internal/services/waitlists/errors because they
// are not the store's: platform deliberately does not decide who may administer
// a list, and both of these are that decision.
var (
	// errNotAuthorizedForWaitlistSignup is a caller reaching for a signup that is
	// not theirs.
	errNotAuthorizedForWaitlistSignup = platformerrors.New("not authorized to access waitlist signup")
	// errNotAuthorizedToWorkTheQueue is a caller trying to move somebody through
	// the lifecycle. Inviting and converting are the operator's, not the
	// signatory's.
	errNotAuthorizedToWorkTheQueue = platformerrors.New("not authorized to work the waitlist queue")
	// errInputRequired is a request whose input message is absent.
	errInputRequired = platformerrors.New("input is required")
)

// requester is what almost every method here needs first: the session, and the
// account to name in the response.
//
// The scope is not among them. Every list and every signup this application
// keeps is global, so the scope is a constant — ddbwaitlists.Scope — and who a
// signup belongs to is decided against the session's user instead.
func (s *serviceImpl) requester(ctx context.Context, span tracing.Span) (*sessions.ContextData, error) {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Unauthenticated, "fetching session context data")
	}

	tracing.AttachToSpan(span, identitykeys.UserIDKey, sessionContextData.GetUserID())
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	return sessionContextData, nil
}

// pageRequest is the session plus the page a list method was asked for.
func (s *serviceImpl) pageRequest(ctx context.Context, span tracing.Span, protoFilter *filteringpb.QueryFilter) (*sessions.ContextData, *filtering.QueryFilter, error) {
	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, nil, err
	}

	filter, err := filteringgrpc.FromProto(protoFilter)
	if err != nil {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	return sessionContextData, filter, nil
}

// responseDetails is the envelope every response carries.
func responseDetails(span tracing.Span, sessionContextData *sessions.ContextData) *types.ResponseDetails {
	return &types.ResponseDetails{
		TraceId:          span.SpanContext().TraceID().String(),
		CurrentAccountId: sessionContextData.GetActiveAccountID(),
	}
}

// convertLists renders a page of waitlists for the wire.
func convertLists(page *filtering.QueryFilteredResult[waitlists.List]) []*waitlistssvc.Waitlist {
	results := make([]*waitlistssvc.Waitlist, 0, len(page.Data))
	for _, list := range page.Data {
		results = append(results, converters.ConvertWaitlistToGRPCWaitlist(list))
	}

	return results
}

func (s *serviceImpl) CreateWaitlist(ctx context.Context, request *waitlistssvc.CreateWaitlistRequest) (*waitlistssvc.CreateWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span)

	list := converters.ConvertGRPCWaitlistCreationRequestInputToWaitlist(request.GetInput())
	if list == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "input is required")
	}

	// The store refuses a list with no name and one with no closing time — see
	// internal/services/waitlists/errors for how each refusal reaches the client.
	// There is no second validation here, because a second one is one that can
	// disagree.
	created, err := s.waitlists.CreateList(ctx, ddbwaitlists.Scope(), list)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating waitlist")
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, created.ID)

	return &waitlistssvc.CreateWaitlistResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Created:         converters.ConvertWaitlistToGRPCWaitlist(created),
	}, nil
}

func (s *serviceImpl) GetWaitlist(ctx context.Context, request *waitlistssvc.GetWaitlistRequest) (*waitlistssvc.GetWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.GetWaitlistId())

	list, err := s.waitlists.GetList(ctx, ddbwaitlists.Scope(), request.GetWaitlistId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching waitlist")
	}

	return &waitlistssvc.GetWaitlistResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertWaitlistToGRPCWaitlist(list),
	}, nil
}

// GetWaitlists pages the whole catalog, open and closed alike. It is the
// administrative read; GetOpenWaitlists is the one a signup form offers.
func (s *serviceImpl) GetWaitlists(ctx context.Context, request *waitlistssvc.GetWaitlistsRequest) (*waitlistssvc.GetWaitlistsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.waitlists.ListLists(ctx, ddbwaitlists.Scope(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching waitlists")
	}

	return &waitlistssvc.GetWaitlistsResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         convertLists(page),
	}, nil
}

// GetOpenWaitlists pages the lists still taking signups.
//
// It is a separate read rather than a filter applied to GetWaitlists' results,
// because a page filtered after the fact is a page whose size the caller cannot
// rely on.
func (s *serviceImpl) GetOpenWaitlists(ctx context.Context, request *waitlistssvc.GetOpenWaitlistsRequest) (*waitlistssvc.GetOpenWaitlistsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.waitlists.ListOpenLists(ctx, ddbwaitlists.Scope(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching open waitlists")
	}

	return &waitlistssvc.GetOpenWaitlistsResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         convertLists(page),
	}, nil
}

// UpdateWaitlist rewrites the list's name, description and closing time.
//
// The update is applied to the row as read rather than built from the request,
// because the store's update takes a whole list: a value assembled from the
// three optional fields alone would blank whichever two the client left out.
func (s *serviceImpl) UpdateWaitlist(ctx context.Context, request *waitlistssvc.UpdateWaitlistRequest) (*waitlistssvc.UpdateWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.GetWaitlistId())

	list, err := s.waitlists.GetList(ctx, ddbwaitlists.Scope(), request.GetWaitlistId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching waitlist for update")
	}

	converters.ApplyGRPCWaitlistUpdateRequestInput(list, request.GetInput())

	if err = s.waitlists.UpdateList(ctx, ddbwaitlists.Scope(), list); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating waitlist")
	}

	return &waitlistssvc.UpdateWaitlistResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Updated:         converters.ConvertWaitlistToGRPCWaitlist(list),
	}, nil
}

// ArchiveWaitlist retires a list. The signups against it are left alone and stay
// readable, because archiving is not erasure.
func (s *serviceImpl) ArchiveWaitlist(ctx context.Context, request *waitlistssvc.ArchiveWaitlistRequest) (*waitlistssvc.ArchiveWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.GetWaitlistId())

	if err = s.waitlists.ArchiveList(ctx, ddbwaitlists.Scope(), request.GetWaitlistId()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving waitlist")
	}

	return &waitlistssvc.ArchiveWaitlistResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// WaitlistIsOpen answers whether one list is still taking signups.
//
// The boundary is List.OpenAt rather than a comparison spelled here, which is
// what keeps this answer and the one ListOpenLists pages by from disagreeing:
// archived counts as closed, and the closing instant itself is closed.
func (s *serviceImpl) WaitlistIsOpen(ctx context.Context, request *waitlistssvc.WaitlistIsOpenRequest) (*waitlistssvc.WaitlistIsOpenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.GetWaitlistId())

	list, err := s.waitlists.GetList(ctx, ddbwaitlists.Scope(), request.GetWaitlistId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching waitlist")
	}

	return &waitlistssvc.WaitlistIsOpenResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		IsOpen:          list.OpenAt(s.clock.Now()),
	}, nil
}

// JoinWaitlist adds the caller to a list.
//
// The address is the session's, not the request's. It is what the list exists to
// write to and what a withdrawal later suppresses, so a signup that could name
// its own contact would be a signup anybody could make on anybody's behalf — and
// a suppression anybody could evade.
func (s *serviceImpl) JoinWaitlist(ctx context.Context, request *waitlistssvc.JoinWaitlistRequest) (*waitlistssvc.JoinWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.GetWaitlistId())

	signup := converters.ConvertGRPCWaitlistSignupCreationRequestInputToSignup(
		request.GetInput(),
		sessionContextData.GetEmailAddress(),
		ddbwaitlists.SubjectFor(sessionContextData.GetUserID()),
	)
	if signup == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "input is required")
	}

	// The store refuses a closed or missing list, an address already on it, and
	// an address that has withdrawn from it. The last is the obligation this
	// package was adopted for, and it reaches the client as PermissionDenied
	// rather than as a conflict — see internal/services/waitlists/errors.
	created, err := s.waitlists.Join(ctx, ddbwaitlists.Scope(), request.GetWaitlistId(), signup)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "joining waitlist")
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, created.ID)

	return &waitlistssvc.JoinWaitlistResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Created:         converters.ConvertWaitlistSignupToGRPCWaitlistSignup(created),
	}, nil
}

func (s *serviceImpl) GetWaitlistSignup(ctx context.Context, request *waitlistssvc.GetWaitlistSignupRequest) (*waitlistssvc.GetWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, signup, err := s.ownSignup(ctx, span, request.GetWaitlistId(), request.GetWaitlistSignupId())
	if err != nil {
		return nil, err
	}

	return &waitlistssvc.GetWaitlistSignupResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup),
	}, nil
}

// GetWaitlistSignupsForWaitlist pages one list's signups.
//
// It is service-admin-only because it exposes every signatory's address, which
// is exactly the read a member of the list must not have.
func (s *serviceImpl) GetWaitlistSignupsForWaitlist(ctx context.Context, request *waitlistssvc.GetWaitlistSignupsForWaitlistRequest) (*waitlistssvc.GetWaitlistSignupsForWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.GetWaitlistId())

	if !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForWaitlistSignup, logger, span, codes.PermissionDenied, "not authorized to list waitlist signups")
	}

	page, err := s.waitlists.ListSignups(ctx, ddbwaitlists.Scope(), request.GetWaitlistId(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching waitlist signups")
	}

	results := make([]*waitlistssvc.WaitlistSignup, 0, len(page.Data))
	for _, signup := range page.Data {
		results = append(results, converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup))
	}

	return &waitlistssvc.GetWaitlistSignupsForWaitlistResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

// UpdateWaitlistSignup rewrites the note against a signup, and moves nobody. It
// deliberately leaves the lifecycle stamp alone: a typo fixed in a note must not
// reschedule the reminder somebody's invitation started.
func (s *serviceImpl) UpdateWaitlistSignup(ctx context.Context, request *waitlistssvc.UpdateWaitlistSignupRequest) (*waitlistssvc.UpdateWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, signup, err := s.ownSignup(ctx, span, request.GetWaitlistId(), request.GetWaitlistSignupId())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, signup.ID)

	notes := signup.Notes
	if request.GetInput() != nil && request.GetInput().Notes != nil {
		notes = request.GetInput().GetNotes()
	}

	if err = s.waitlists.UpdateSignupNotes(ctx, ddbwaitlists.Scope(), request.GetWaitlistId(), signup.ID, notes); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating waitlist signup")
	}

	signup.Notes = notes

	return &waitlistssvc.UpdateWaitlistSignupResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Updated:         converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup),
	}, nil
}

// InviteWaitlistSignup lets somebody in. It records that they were let in;
// what reaches them is the notifications side's to deliver, off the stamp this
// leaves behind.
func (s *serviceImpl) InviteWaitlistSignup(ctx context.Context, request *waitlistssvc.InviteWaitlistSignupRequest) (*waitlistssvc.InviteWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, signup, err := s.workedSignup(ctx, span, request.GetWaitlistId(), request.GetWaitlistSignupId(), s.waitlists.Invite, "inviting waitlist signup")
	if err != nil {
		return nil, err
	}

	return &waitlistssvc.InviteWaitlistSignupResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Updated:         converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup),
	}, nil
}

// ConvertWaitlistSignup marks an invitation taken up.
func (s *serviceImpl) ConvertWaitlistSignup(ctx context.Context, request *waitlistssvc.ConvertWaitlistSignupRequest) (*waitlistssvc.ConvertWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, signup, err := s.workedSignup(ctx, span, request.GetWaitlistId(), request.GetWaitlistSignupId(), s.waitlists.Convert, "converting waitlist signup")
	if err != nil {
		return nil, err
	}

	return &waitlistssvc.ConvertWaitlistSignupResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Updated:         converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup),
	}, nil
}

// WithdrawFromWaitlist takes somebody off a list at their own request.
//
// It is the opt-out, and it is not ArchiveWaitlistSignup: archiving hides the
// row and keeps the address, so the next signup from that address succeeds.
// Withdrawing blanks the address and keeps a digest of it, so the next signup is
// refused. Anybody clicking "unsubscribe" wants this one.
//
// The owner may withdraw, and so may a service admin — somebody who has asked to
// come off a list by another channel has still asked.
//
// It is a one-way door for the owner in a second sense: the withdrawal blanks
// the subject reference, so a second attempt at the same signup no longer looks
// like theirs and is refused by the ownership check rather than by the store's
// ErrAlreadyWithdrawn. That is the anonymization working. What tells somebody
// they are already off a list is JoinWaitlist's refusal, which names the
// withdrawal.
func (s *serviceImpl) WithdrawFromWaitlist(ctx context.Context, request *waitlistssvc.WithdrawFromWaitlistRequest) (*waitlistssvc.WithdrawFromWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, signup, err := s.ownSignup(ctx, span, request.GetWaitlistId(), request.GetWaitlistSignupId())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, signup.ID)

	if err = s.waitlists.Withdraw(ctx, ddbwaitlists.Scope(), request.GetWaitlistId(), signup.ID); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "withdrawing from waitlist")
	}

	withdrawn, err := s.waitlists.GetSignup(ctx, ddbwaitlists.Scope(), request.GetWaitlistId(), signup.ID)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching withdrawn waitlist signup")
	}

	return &waitlistssvc.WithdrawFromWaitlistResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Updated:         converters.ConvertWaitlistSignupToGRPCWaitlistSignup(withdrawn),
	}, nil
}

// ArchiveWaitlistSignup retires a signup administratively.
//
// It suppresses nothing: the address is still stored and the uniqueness still
// covers the row, so the next attempt from it is refused as a duplicate rather
// than honored as an opt-out. Somebody asking to come off a list wants
// WithdrawFromWaitlist.
func (s *serviceImpl) ArchiveWaitlistSignup(ctx context.Context, request *waitlistssvc.ArchiveWaitlistSignupRequest) (*waitlistssvc.ArchiveWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, signup, err := s.ownSignup(ctx, span, request.GetWaitlistId(), request.GetWaitlistSignupId())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, signup.ID)

	if err = s.waitlists.ArchiveSignup(ctx, ddbwaitlists.Scope(), request.GetWaitlistId(), signup.ID); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving waitlist signup")
	}

	return &waitlistssvc.ArchiveWaitlistSignupResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// ownSignup reads a signup and refuses it to anybody but its owner and a service
// admin.
//
// Signups are user-owned, which the scope cannot express here: every list is
// global, so what makes a signup somebody's is its subject. The check is
// therefore made after the read rather than by it — and the read is the store's,
// so a signup on another list, or one that does not exist, is a not-found before
// this is reached.
func (s *serviceImpl) ownSignup(ctx context.Context, span tracing.Span, listID, signupID string) (*sessions.ContextData, *waitlists.Signup, error) {
	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, nil, err
	}

	logger := s.logger.WithSpan(span).
		WithValue(waitlistkeys.WaitlistIDKey, listID).
		WithValue(waitlistkeys.WaitlistSignupIDKey, signupID)

	signup, err := s.waitlists.GetSignup(ctx, ddbwaitlists.Scope(), listID, signupID)
	if err != nil {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching waitlist signup")
	}

	if signup.Subject.ID != sessionContextData.GetUserID() && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForWaitlistSignup, logger, span, codes.PermissionDenied, "not authorized to access waitlist signup")
	}

	return sessionContextData, signup, nil
}

// workedSignup is Invite and Convert: a service admin moves somebody through the
// queue, and the signup as it now stands is read back.
//
// Only a service admin may. Being on a list does not entitle somebody to invite
// themselves off it, which is the whole difference between a queue and a form.
//
// The read back is a second round trip and is worth it: the guard decided the
// status, and a caller rendering the queue needs the stamp the move left, which
// only the row has.
func (s *serviceImpl) workedSignup(
	ctx context.Context,
	span tracing.Span,
	listID, signupID string,
	move func(ctx context.Context, scope tenancy.Scope, listID, signupID string) error,
	description string,
) (*sessions.ContextData, *waitlists.Signup, error) {
	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, nil, err
	}

	logger := s.logger.WithSpan(span).
		WithValue(waitlistkeys.WaitlistIDKey, listID).
		WithValue(waitlistkeys.WaitlistSignupIDKey, signupID)

	if !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedToWorkTheQueue, logger, span, codes.PermissionDenied, "not authorized to work the waitlist queue")
	}

	if err = move(ctx, ddbwaitlists.Scope(), listID, signupID); err != nil {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "%s", description)
	}

	signup, err := s.waitlists.GetSignup(ctx, ddbwaitlists.Scope(), listID, signupID)
	if err != nil {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching waitlist signup after %s", description)
	}

	return sessionContextData, signup, nil
}
