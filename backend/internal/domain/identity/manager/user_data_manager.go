package manager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v7/database"
	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/identifiers"
	"github.com/primandproper/platform-go/v7/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	"github.com/primandproper/platform-go/v7/observability"
	platformkeys "github.com/primandproper/platform-go/v7/observability/keys"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	"github.com/primandproper/platform-go/v7/qrcodes"
	"github.com/primandproper/platform-go/v7/random"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	o11yName = "identity_data_manager"

	minimumPasswordEntropy = 60
	totpSecretSize         = 64
)

var (
	// ErrInvalidIDProvided indicates a required ID was passed in empty.
	ErrInvalidIDProvided = platformerrors.New("required ID was empty")
	// ErrNilInputProvided indicates that a required parameter was nil.
	ErrNilInputProvided = platformerrors.New("nil input provided")
	// ErrEmptyInputProvided indicates that the required input was empty.
	ErrEmptyInputProvided = platformerrors.New("empty input provided")
)

type (
	manager struct {
		tracer               tracing.Tracer
		logger               logging.Logger
		dataChangesPublisher messagequeue.Publisher
		identityRepo         identity.Repository
		secretGenerator      random.Generator
		authenticator        authentication.Hasher
		userSearchIndex      indexing.UserTextSearcher
		qrCodeBuilder        qrcodes.Builder
	}
)

func NewIdentityDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	publisherProvider messagequeue.PublisherProvider,
	identityRepo identity.Repository,
	secretGenerator random.Generator,
	authenticator authentication.Hasher,
	userSearchIndex indexing.UserTextSearcher,
	qrCodeBuilder qrcodes.Builder,
	cfg *msgconfig.QueuesConfig,
) (IdentityDataManager, error) {
	publisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("setting up data changes publisher: %w", err)
	}

	return &manager{
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:               logging.NewNamedLogger(logger, o11yName),
		identityRepo:         identityRepo,
		dataChangesPublisher: publisher,
		secretGenerator:      secretGenerator,
		authenticator:        authenticator,
		userSearchIndex:      userSearchIndex,
		qrCodeBuilder:        qrCodeBuilder,
	}, nil
}

func (m *manager) AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountInvitationID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	invitation, err := m.identityRepo.GetAccountInvitationByTokenAndID(ctx, input.Token, accountInvitationID)
	if errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "account invitation not found")
	} else if err != nil {
		return observability.PrepareError(err, span, "retrieving invitation")
	}

	if err = m.identityRepo.AcceptAccountInvitation(ctx, accountID, accountInvitationID, input.Token, input.Note); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "accepting account invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationAcceptedServiceEventType, map[string]any{
		identitykeys.AccountInvitationIDKey:  accountInvitationID,
		identitykeys.DestinationAccountIDKey: invitation.DestinationAccount.ID,
	}))

	return nil
}

func (m *manager) RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountInvitationID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	// note, this is where you would call input.ValidateWithContext, if that currently had any effect.

	invitation, err := m.identityRepo.GetAccountInvitationByTokenAndID(ctx, input.Token, accountInvitationID)
	if errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "account invitation not found")
	} else if err != nil {
		return observability.PrepareError(err, span, "retrieving invitation")
	}

	if err = m.identityRepo.RejectAccountInvitation(ctx, accountID, invitation.ID, input.Note); err != nil {
		return observability.PrepareError(err, span, "rejecting invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationRejectedServiceEventType, map[string]any{
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}))

	return nil
}

func (m *manager) CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountInvitationID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	if err := m.identityRepo.CancelAccountInvitation(ctx, accountID, accountInvitationID, note); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "canceling account invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationCanceledServiceEventType, map[string]any{
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}))

	return nil
}

func (m *manager) ArchiveAccount(ctx context.Context, accountID, ownerID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || ownerID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
		identitykeys.UserIDKey:    ownerID,
	}, span, m.logger)

	if err := m.identityRepo.ArchiveAccount(ctx, accountID, ownerID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "ArchiveAccount")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountArchivedServiceEventType, map[string]any{
		identitykeys.AccountIDKey: accountID,
		identitykeys.UserIDKey:    ownerID,
	}))

	return nil
}

func (m *manager) ArchiveUserMembership(ctx context.Context, userID, accountID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || userID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
		identitykeys.UserIDKey:    userID,
	}, span, m.logger)

	if err := m.identityRepo.RemoveUserFromAccount(ctx, userID, accountID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "RemoveUserFromAccount")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountMemberRemovedServiceEventType, map[string]any{
		identitykeys.AccountIDKey: accountID,
		identitykeys.UserIDKey:    userID,
	}))

	return nil
}

func (m *manager) ArchiveUser(ctx context.Context, userID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.ArchiveUser(ctx, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "ArchiveUser")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserArchivedServiceEventType, map[string]any{
		identitykeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) CreateAccount(ctx context.Context, input *identity.AccountCreationRequestInput) (*identity.Account, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := m.logger.WithSpan(span)

	created, err := m.identityRepo.CreateAccount(ctx, converters.ConvertAccountCreationInputToAccountDatabaseCreationInput(input))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating Account")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountCreatedServiceEventType, map[string]any{
		identitykeys.AccountIDKey: created.ID,
	}))

	return created, nil
}

func (m *manager) CreateAccountInvitation(ctx context.Context, userID, accountID string, input *identity.AccountInvitationCreationRequestInput) (*identity.AccountInvitation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey:    userID,
		identitykeys.AccountIDKey: accountID,
	}, span, m.logger)

	token, err := m.secretGenerator.GenerateBase64EncodedString(ctx, 64)
	if err != nil {
		return nil, observability.PrepareError(err, span, "generating account invitation token")
	}

	convertedInput := converters.ConvertAccountInvitationCreationInputToAccountInvitationDatabaseCreationInput(userID, accountID, token, input)

	created, err := m.identityRepo.CreateAccountInvitation(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating account invitation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountInvitationCreatedServiceEventType, map[string]any{
		identitykeys.AccountInvitationIDKey:  created.ID,
		identitykeys.UserIDKey:               userID,
		identitykeys.DestinationAccountIDKey: accountID,
	}))

	return created, nil
}

func (m *manager) CreateUser(ctx context.Context, input *identity.UserRegistrationInput) (*identity.UserCreationResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UsernameKey: input.Username,
	}, span, m.logger)

	input.Username = strings.TrimSpace(input.Username)
	tracing.AttachToSpan(span, identitykeys.UsernameKey, input.Username)
	input.EmailAddress = strings.TrimSpace(strings.ToLower(input.EmailAddress))
	tracing.AttachToSpan(span, identitykeys.UserEmailAddressKey, input.EmailAddress)
	input.Password = strings.TrimSpace(input.Password)

	logger = logger.WithValues(map[string]any{
		identitykeys.UsernameKey:            input.Username,
		identitykeys.UserEmailAddressKey:    input.EmailAddress,
		identitykeys.AccountInvitationIDKey: input.InvitationID,
	})

	if err := input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(platformkeys.ValidationErrorKey, err).Debug("provided dbInput was invalid")
		return nil, observability.PrepareError(err, span, "invalid user creation dbInput provided")
	}

	// ensure the password is not garbage-tier
	if err := passwordvalidator.Validate(strings.TrimSpace(input.Password), minimumPasswordEntropy); err != nil {
		logger.WithValue("password_validation_error", err).Debug("weak password provided to user creation route")
		return nil, observability.PrepareAndLogError(err, logger, span, "weak password provided for user creation")
	}

	var invitation *identity.AccountInvitation
	if input.InvitationID != "" && input.InvitationToken != "" {
		invite, err := m.identityRepo.GetAccountInvitationByTokenAndID(ctx, input.InvitationToken, input.InvitationID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, observability.PrepareAndLogError(err, logger, span, "no account invitation found")
		} else if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "getting account invitation by token and ID")
		}

		invitation = invite
		logger.Debug("retrieved account invitation")
	}

	logger.Debug("completed invitation check")

	// hash the password
	hp, err := m.authenticator.HashPassword(ctx, strings.TrimSpace(input.Password))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "hashing user creation password")
	}

	// generate a two-factor secret.
	tfs, err := m.secretGenerator.GenerateBase32EncodedString(ctx, totpSecretSize)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "generating two factor secret")
	}

	dbInput := &identity.UserDatabaseCreationInput{
		ID:              identifiers.New(),
		Username:        input.Username,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		EmailAddress:    input.EmailAddress,
		HashedPassword:  hp,
		TwoFactorSecret: tfs,
		InvitationToken: input.InvitationToken,
		Birthday:        input.Birthday,
		AccountName:     input.AccountName,
	}

	if invitation != nil {
		logger.Debug("supplementing user creation dbInput with invitation data")
		dbInput.DestinationAccountID = invitation.DestinationAccount.ID
		dbInput.InvitationToken = invitation.Token
	}

	// create the user.
	user, err := m.identityRepo.CreateUser(ctx, dbInput)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating user")
		if errors.Is(err, database.ErrUserAlreadyExists) {
			return nil, observability.PrepareAndLogError(err, logger, span, "user already exists")
		}
		return nil, observability.PrepareAndLogError(err, logger, span, "creating user in database")
	}

	logger.Debug("user created")

	defaultAccountID, err := m.identityRepo.GetDefaultAccountIDForUser(ctx, user.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching default account ID for user")
	}

	emailVerificationToken, err := m.identityRepo.GetEmailAddressVerificationTokenForUser(ctx, user.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching email verification token for user")
	}

	// notify the relevant parties.
	tracing.AttachToSpan(span, identitykeys.UserIDKey, user.ID)

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserSignedUpServiceEventType, map[string]any{
		identitykeys.AccountIDKey:                  defaultAccountID,
		identitykeys.UserIDKey:                     user.ID,
		identitykeys.UserEmailVerificationTokenKey: emailVerificationToken,
	}))

	twoFactorQRCode, qrCodeErr := m.qrCodeBuilder.BuildQRCode(ctx, user.Username, user.TwoFactorSecret)
	if qrCodeErr != nil {
		// the QR code is a convenience rendering of the secret, which is still returned; don't fail the signup over it.
		observability.AcknowledgeError(qrCodeErr, logger, span, "building two factor QR code")
	}

	// UserCreationResponse is a struct we can use to notify the user of their two factor secret, but ideally just this once and then never again.
	ucr := &identity.UserCreationResponse{
		CreatedUserID:   user.ID,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		EmailAddress:    user.EmailAddress,
		CreatedAt:       user.CreatedAt,
		TwoFactorSecret: user.TwoFactorSecret,
		Birthday:        user.Birthday,
		TwoFactorQRCode: twoFactorQRCode,
	}

	return ucr, nil
}

func (m *manager) GetAccount(ctx context.Context, accountID string) (*identity.Account, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
	}, span, m.logger)

	account, err := m.identityRepo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account")
	}

	return account, nil
}

func (m *manager) GetAccountInvitation(ctx context.Context, accountID, accountInvitationID string) (*identity.AccountInvitation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || accountInvitationID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey:           accountID,
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}, span, m.logger)

	invitation, err := m.identityRepo.GetAccountInvitationByAccountAndID(ctx, accountID, accountInvitationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting invitation")
	}

	return invitation, nil
}

func (m *manager) GetAccounts(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	accounts, err := m.identityRepo.GetAccounts(ctx, userID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching accounts")
	}

	return accounts, nil
}

func (m *manager) GetReceivedAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	invites, err := m.identityRepo.GetPendingAccountInvitationsForUser(ctx, userID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching invites")
	}

	return invites, nil
}

func (m *manager) GetSentAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	invites, err := m.identityRepo.GetPendingAccountInvitationsFromUser(ctx, userID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching invites")
	}

	return invites, nil
}

func (m *manager) GetUser(ctx context.Context, userID string) (*identity.User, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	user, err := m.identityRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting user")
	}

	return user, nil
}

func (m *manager) GetUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.identityRepo.GetUserByEmail(ctx, email)
}

func (m *manager) GetUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.identityRepo.GetUserByUsername(ctx, username)
}

func (m *manager) GetAdminUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.identityRepo.GetAdminUserByUsername(ctx, username)
}

func (m *manager) GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.identityRepo.GetDefaultAccountIDForUser(ctx, userID)
}

func (m *manager) BuildSessionContextDataForUser(ctx context.Context, userID, activeAccountID string) (*sessions.ContextData, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.identityRepo.BuildSessionContextDataForUser(ctx, userID, activeAccountID)
}

func (m *manager) GetUsers(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span)

	users, err := m.identityRepo.GetUsers(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching users")
	}

	return users, nil
}

func (m *manager) GetUsersForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span)

	users, err := m.identityRepo.GetUsersForAccount(ctx, accountID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching users")
	}

	return users, nil
}

func (m *manager) SearchForUsers(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if query == "" {
		return nil, platformerrors.New("query cannot be empty")
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := observability.ObserveValues(map[string]any{
		platformkeys.UseDatabaseKey: !useSearchService,
	}, span, m.logger)

	if !useSearchService {
		users, err := m.identityRepo.SearchForUsersByUsername(ctx, query, filter)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, observability.PrepareError(err, span, "no users found")
			}
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for users")
		}

		return users, nil
	} else {
		uss, err := m.userSearchIndex.Search(ctx, query)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for users")
		}

		// The search index has no notion of cursor pagination, so the best we can do is
		// honor the filter's page size and report the index's full hit count as the total.
		totalHits := uint64(len(uss))

		userIDs := []string{}
		for _, us := range uss {
			userIDs = append(userIDs, us.ID)
		}

		if filter.MaxResponseSize != nil && len(userIDs) > int(*filter.MaxResponseSize) {
			userIDs = userIDs[:*filter.MaxResponseSize]
		}

		users, err := m.identityRepo.GetUsersWithIDs(ctx, userIDs)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for users")
		}

		result := filtering.NewQueryFilteredResult(users, uint64(len(users)), totalHits, func(u *identity.User) string {
			return u.ID
		}, filter)

		return result, nil
	}
}

func (m *manager) SetDefaultAccount(ctx context.Context, userID, accountID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey:    userID,
		identitykeys.AccountIDKey: accountID,
	}, span, m.logger)

	// mark household as default in database.
	if err := m.identityRepo.MarkAccountAsUserDefault(ctx, userID, accountID); err != nil {
		return observability.PrepareError(err, span, "marking default account as user")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountSetAsDefaultServiceEventType, map[string]any{
		identitykeys.AccountIDKey: accountID,
		identitykeys.UserIDKey:    userID,
	}))

	return nil
}

func (m *manager) TransferAccountOwnership(ctx context.Context, accountID string, input *identity.AccountOwnershipTransferInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
	}, span, m.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "")
	}

	// transfer ownership of household in database.
	if err := m.identityRepo.TransferAccountOwnership(ctx, accountID, input); err != nil {
		return observability.PrepareError(err, span, "")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountOwnershipTransferredServiceEventType, map[string]any{
		identitykeys.AccountIDKey: accountID,
	}))

	return nil
}

func (m *manager) UpdateAccount(ctx context.Context, accountID string, input *identity.AccountUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
	}, span, m.logger)
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "validating account update")
	}

	// fetch the account from the database.
	account, err := m.identityRepo.GetAccount(ctx, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "no account found")
	} else if err != nil {
		return observability.PrepareError(err, span, "fetching account")
	}

	account.Update(input)

	// update the account in the database.
	if err = m.identityRepo.UpdateAccount(ctx, account); err != nil {
		return observability.PrepareError(err, span, "updating account")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountUpdatedServiceEventType, map[string]any{
		identitykeys.AccountIDKey: accountID,
	}))

	return nil
}

func (m *manager) UpdateAccountBillingFields(ctx context.Context, accountID string, billingStatus, subscriptionPlanID, paymentProcessorCustomerID *string, lastPaymentProviderSyncOccurredAt *time.Time) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return ErrInvalidIDProvided
	}

	return m.identityRepo.UpdateAccountBillingFields(ctx, accountID, billingStatus, subscriptionPlanID, paymentProcessorCustomerID, lastPaymentProviderSyncOccurredAt)
}

func (m *manager) UpdateAccountMemberPermissions(ctx context.Context, accountID, userID string, input *identity.ModifyUserPermissionsInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	if err := m.identityRepo.ModifyUserPermissions(ctx, accountID, userID, input); err != nil {
		return observability.PrepareError(err, span, "modifying user permissions")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.AccountMembershipPermissionsUpdatedServiceEventType, map[string]any{
		identitykeys.AccountIDKey: accountID,
	}))

	return nil
}

func (m *manager) UpdateUserDetails(ctx context.Context, userID string, input *identity.UserDetailsUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "invalid input attached to request")
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserDetails(ctx, userID, converters.ConvertUserDetailsUpdateRequestInputToUserDetailsUpdateInput(input)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user details")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserDetailsChangedEventType, map[string]any{
		identitykeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) UpdateUserEmailAddress(ctx context.Context, userID, newEmail string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserEmailAddress(ctx, userID, newEmail); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user email address")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.EmailAddressChangedEventType, map[string]any{
		identitykeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) UpdateUserUsername(ctx context.Context, userID, newUsername string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	if newUsername == "" {
		return ErrEmptyInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserUsername(ctx, userID, newUsername); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user username")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UsernameChangedEventType, map[string]any{
		identitykeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) SetUserAvatar(ctx context.Context, userID, uploadedMediaID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrInvalidIDProvided
	}

	if uploadedMediaID == "" {
		return ErrEmptyInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.SetUserAvatar(ctx, userID, uploadedMediaID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "setting user avatar")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserAvatarChangedEventType, map[string]any{
		identitykeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) AdminUpdateUserStatus(ctx context.Context, input *identity.UserAccountStatusUpdateInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: input.TargetUserID,
		platformkeys.ReasonKey: input.Reason,
	}, span, m.logger)

	if err := m.identityRepo.UpdateUserAccountStatus(ctx, input.TargetUserID, input); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user account status")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserStatusChangedServiceEventType, map[string]any{
		identitykeys.UserIDKey: input.TargetUserID,
	}))

	return nil
}

func (m *manager) AdminSetPasswordChangeRequired(ctx context.Context, userID string, requiresChange bool) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return ErrEmptyInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	if err := m.identityRepo.SetUserRequiresPasswordChange(ctx, userID, requiresChange); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "setting user requires password change")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, identity.UserPasswordChangeRequiredServiceEventType, map[string]any{
		identitykeys.UserIDKey: userID,
	}))

	return nil
}

func (m *manager) UserRequiresPasswordChange(ctx context.Context, userID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return false, ErrEmptyInputProvided
	}

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: userID,
	}, span, m.logger)

	result, err := m.identityRepo.UserRequiresPasswordChange(ctx, userID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking if user requires password change")
	}

	return result, nil
}
