package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/identifiers"
	"github.com/primandproper/platform-go/v10/observability"
	platformkeys "github.com/primandproper/platform-go/v10/observability/keys"
	"github.com/primandproper/platform-go/v10/observability/tracing"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	resourceTypeUsers = "users"

	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	postgresDuplicateEntryErrorCode = "23505"
)

var (
	_ identity.UserDataManager = (*repository)(nil)
)

func avatarFromRow(id, storagePath sql.NullString, mimeType generated.NullUploadedMediaMimeType, createdAt, lastUpdatedAt, archivedAt sql.NullTime, createdByUser sql.NullString) *uploadedmedia.UploadedMedia {
	if !id.Valid || !storagePath.Valid || !mimeType.Valid || !createdByUser.Valid {
		return nil
	}
	return &uploadedmedia.UploadedMedia{
		ID:            id.String,
		StoragePath:   storagePath.String,
		MimeType:      string(mimeType.UploadedMediaMimeType),
		CreatedAt:     createdAt.Time,
		LastUpdatedAt: database.TimePointerFromNullTime(lastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(archivedAt),
		CreatedByUser: createdByUser.String,
	}
}

// GetUser fetches a user.
func (r *repository) GetUser(ctx context.Context, userID string) (*identity.User, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	result, err := r.generatedQuerier.GetUserByID(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting user")
	}

	u := &identity.User{
		CreatedAt:                  result.CreatedAt,
		PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
		LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
		LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
		Birthday:                   database.TimePointerFromNullTime(result.Birthday),
		ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
		AccountStatusExplanation:   result.UserAccountStatusExplanation,
		TwoFactorSecret:            result.TwoFactorSecret,
		HashedPassword:             result.HashedPassword,
		ID:                         result.ID,
		AccountStatus:              result.UserAccountStatus,
		Username:                   result.Username,
		FirstName:                  result.FirstName,
		LastName:                   result.LastName,
		EmailAddress:               result.EmailAddress,
		EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
		Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
		RequiresPasswordChange:     result.RequiresPasswordChange,
	}

	return u, nil
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified 2FA secret.
func (r *repository) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID string) (*identity.User, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	result, err := r.generatedQuerier.GetUserWithUnverifiedTwoFactor(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting user with unverified two factor")
	}

	u := &identity.User{
		CreatedAt:                  result.CreatedAt,
		PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
		LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
		LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
		Birthday:                   database.TimePointerFromNullTime(result.Birthday),
		ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
		AccountStatusExplanation:   result.UserAccountStatusExplanation,
		TwoFactorSecret:            result.TwoFactorSecret,
		HashedPassword:             result.HashedPassword,
		ID:                         result.ID,
		AccountStatus:              result.UserAccountStatus,
		Username:                   result.Username,
		FirstName:                  result.FirstName,
		LastName:                   result.LastName,
		EmailAddress:               result.EmailAddress,
		EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
		Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
		RequiresPasswordChange:     result.RequiresPasswordChange,
	}

	return u, nil
}

// GetUserByUsername fetches a user by their username.
func (r *repository) GetUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if username == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.UsernameKey, username)

	result, err := r.generatedQuerier.GetUserByUsername(ctx, r.readDB, username)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting user by username")
	}

	u := &identity.User{
		CreatedAt:                  result.CreatedAt,
		PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
		LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
		LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
		Birthday:                   database.TimePointerFromNullTime(result.Birthday),
		ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
		AccountStatusExplanation:   result.UserAccountStatusExplanation,
		TwoFactorSecret:            result.TwoFactorSecret,
		HashedPassword:             result.HashedPassword,
		ID:                         result.ID,
		AccountStatus:              result.UserAccountStatus,
		Username:                   result.Username,
		FirstName:                  result.FirstName,
		LastName:                   result.LastName,
		EmailAddress:               result.EmailAddress,
		EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
		Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
		RequiresPasswordChange:     result.RequiresPasswordChange,
	}

	return u, nil
}

// GetAdminUserByUsername fetches a user by their username.
func (r *repository) GetAdminUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if username == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.UsernameKey, username)

	result, err := r.generatedQuerier.GetAdminUserByUsername(ctx, r.readDB, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, observability.PrepareError(err, span, "getting admin user by username")
	}

	u := &identity.User{
		CreatedAt:                  result.CreatedAt,
		PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
		LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
		LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
		Birthday:                   database.TimePointerFromNullTime(result.Birthday),
		ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
		AccountStatusExplanation:   result.UserAccountStatusExplanation,
		TwoFactorSecret:            result.TwoFactorSecret,
		HashedPassword:             result.HashedPassword,
		ID:                         result.ID,
		AccountStatus:              result.UserAccountStatus,
		Username:                   result.Username,
		FirstName:                  result.FirstName,
		LastName:                   result.LastName,
		EmailAddress:               result.EmailAddress,
		EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
		Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
		RequiresPasswordChange:     result.RequiresPasswordChange,
	}

	return u, nil
}

// GetUserByEmail fetches a user by their email.
func (r *repository) GetUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if email == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserEmailAddressKey, email)

	result, err := r.generatedQuerier.GetUserByEmail(ctx, r.readDB, email)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting user by email")
	}

	u := &identity.User{
		CreatedAt:                  result.CreatedAt,
		PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
		LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
		LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
		Birthday:                   database.TimePointerFromNullTime(result.Birthday),
		ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
		AccountStatusExplanation:   result.UserAccountStatusExplanation,
		TwoFactorSecret:            result.TwoFactorSecret,
		HashedPassword:             result.HashedPassword,
		ID:                         result.ID,
		AccountStatus:              result.UserAccountStatus,
		Username:                   result.Username,
		FirstName:                  result.FirstName,
		LastName:                   result.LastName,
		EmailAddress:               result.EmailAddress,
		EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
		Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
		RequiresPasswordChange:     result.RequiresPasswordChange,
	}

	return u, nil
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (r *repository) SearchForUsersByUsername(ctx context.Context, usernameQuery string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span)

	if usernameQuery == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, usernameQuery)
	logger = logger.WithValue(identitykeys.UsernameKey, usernameQuery)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	filter.AttachToLogger(logger)

	results, err := r.generatedQuerier.SearchUsersByUsername(ctx, r.readDB, &generated.SearchUsersByUsernameParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Username:        usernameQuery,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "querying database for users")
	}

	var (
		users                     = []*identity.User{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)

		users = append(users, &identity.User{
			CreatedAt:                  result.CreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
			Birthday:                   database.TimePointerFromNullTime(result.Birthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
			AccountStatusExplanation:   result.UserAccountStatusExplanation,
			TwoFactorSecret:            result.TwoFactorSecret,
			HashedPassword:             result.HashedPassword,
			ID:                         result.ID,
			AccountStatus:              result.UserAccountStatus,
			Username:                   result.Username,
			FirstName:                  result.FirstName,
			LastName:                   result.LastName,
			EmailAddress:               result.EmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
			Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
			RequiresPasswordChange:     result.RequiresPasswordChange,
		})
	}

	if len(users) == 0 {
		return nil, sql.ErrNoRows
	}

	x := filtering.NewQueryFilteredResult(users, filteredCount, totalCount, func(t *identity.User) string {
		return t.ID
	}, filter)

	return x, nil
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (r *repository) GetUsers(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	filter.AttachToLogger(logger) // TODO: is assignment necessary here? if not, make consistent

	results, err := r.generatedQuerier.GetUsers(ctx, r.readDB, &generated.GetUsersParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "scanning user")
	}

	var (
		data                      = []*identity.User{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		u := &identity.User{
			CreatedAt:                  result.CreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
			Birthday:                   database.TimePointerFromNullTime(result.Birthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
			AccountStatusExplanation:   result.UserAccountStatusExplanation,
			TwoFactorSecret:            result.TwoFactorSecret,
			HashedPassword:             result.HashedPassword,
			ID:                         result.ID,
			AccountStatus:              result.UserAccountStatus,
			Username:                   result.Username,
			FirstName:                  result.FirstName,
			LastName:                   result.LastName,
			EmailAddress:               result.EmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
			Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
			RequiresPasswordChange:     result.RequiresPasswordChange,
		}

		data = append(data, u)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *identity.User) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetUsersForAccount fetches a list of users from the database that meet a particular filter.
func (r *repository) GetUsersForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[identity.User], err error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	filter.AttachToLogger(logger)

	x = &filtering.QueryFilteredResult[identity.User]{
		Pagination: filter.ToPagination(),
	}

	results, err := r.generatedQuerier.GetUsersForAccount(ctx, r.readDB, &generated.GetUsersForAccountParams{
		BelongsToAccount: accountID,
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "scanning user")
	}

	for _, result := range results {
		u := &identity.User{
			CreatedAt:                  result.CreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
			Birthday:                   database.TimePointerFromNullTime(result.Birthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
			AccountStatusExplanation:   result.UserAccountStatusExplanation,
			TwoFactorSecret:            result.TwoFactorSecret,
			HashedPassword:             result.HashedPassword,
			ID:                         result.ID,
			AccountStatus:              result.UserAccountStatus,
			Username:                   result.Username,
			FirstName:                  result.FirstName,
			LastName:                   result.LastName,
			EmailAddress:               result.EmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
			Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
			RequiresPasswordChange:     result.RequiresPasswordChange,
		}

		x.Data = append(x.Data, u)
		x.FilteredCount = uint64(result.FilteredCount)
		x.TotalCount = uint64(result.TotalCount)
	}

	return x, nil
}

// GetUsersWithIDs fetches a list of users from the database that meet a particular filter.
func (r *repository) GetUsersWithIDs(ctx context.Context, ids []string) (x []*identity.User, err error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithValue("user_id_count", len(ids))

	results, err := r.generatedQuerier.GetUsersWithIDs(ctx, r.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "scanning user")
	}

	for _, result := range results {
		u := &identity.User{
			CreatedAt:                  result.CreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
			Birthday:                   database.TimePointerFromNullTime(result.Birthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
			AccountStatusExplanation:   result.UserAccountStatusExplanation,
			TwoFactorSecret:            result.TwoFactorSecret,
			HashedPassword:             result.HashedPassword,
			ID:                         result.ID,
			AccountStatus:              result.UserAccountStatus,
			Username:                   result.Username,
			FirstName:                  result.FirstName,
			LastName:                   result.LastName,
			EmailAddress:               result.EmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
			Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
			RequiresPasswordChange:     result.RequiresPasswordChange,
		}

		x = append(x, u)
	}

	return x, nil
}

// ScanUserIDsForReindex returns up to limit IDs sorting strictly after `after`, in ascending byte order.
//
// It is the source half of a search reindex: searchsync.Reindexer walks this to find every
// document that should exist, and prunes the index of anything it does not name. It replaces
// the "IDs that need indexing" sampler platform-go v10 removed, which asked a different and
// weaker question — which rows look stale — and could only ever be probabilistically right,
// because a row the sampler had not reached was a row the index was wrong about.
func (r *repository) ScanUserIDsForReindex(ctx context.Context, after string, limit int) ([]string, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	results, err := r.generatedQuerier.ScanUserIDsForReindex(ctx, r.readDB, &generated.ScanUserIDsForReindexParams{
		Cursor:      after,
		ResultLimit: limit,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing users reindex scan query")
	}

	return results, nil
}

// MarkUserAsIndexed updates a particular user's last_indexed_at value.
func (r *repository) MarkUserAsIndexed(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if _, err := r.generatedQuerier.UpdateUserLastIndexedAt(ctx, r.writeDB, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking user as indexed")
	}

	logger.Info("user marked as indexed")

	return nil
}

// CreateUser creates a user. TODO: this should return an account as well.
func (r *repository) CreateUser(ctx context.Context, input *identity.UserDatabaseCreationInput) (*identity.User, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	tracing.AttachToSpan(span, identitykeys.UsernameKey, input.Username)
	logger := r.logger.WithValues(map[string]any{
		identitykeys.UsernameKey:         input.Username,
		identitykeys.UserEmailAddressKey: input.EmailAddress,
		"destination_account":            input.DestinationAccountID,
	})

	// begin user creation transaction
	var user *identity.User
	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		token, err := r.secretGenerator.GenerateBase64EncodedString(ctx, 32)
		if err != nil {
			return observability.PrepareError(err, span, "generating email verification token")
		}

		if err = r.generatedQuerier.CreateUser(ctx, tx, &generated.CreateUserParams{
			ID:                            input.ID,
			FirstName:                     input.FirstName,
			LastName:                      input.LastName,
			Username:                      input.Username,
			EmailAddress:                  input.EmailAddress,
			HashedPassword:                input.HashedPassword,
			TwoFactorSecret:               input.TwoFactorSecret,
			UserAccountStatus:             string(identity.UnverifiedAccountStatus),
			Birthday:                      database.NullTimeFromTimePointer(input.Birthday),
			EmailAddressVerificationToken: database.NullStringFromString(token),
			TwoFactorSecretVerifiedAt:     sql.NullTime{},
			UserAccountStatusExplanation:  "",
			RequiresPasswordChange:        false,
		}); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == postgresDuplicateEntryErrorCode {
					return database.ErrUserAlreadyExists
				}
			}

			return observability.PrepareError(err, span, "creating user")
		}

		hasValidInvite := input.InvitationToken != "" && input.DestinationAccountID != ""

		user = &identity.User{
			ID:              input.ID,
			FirstName:       input.FirstName,
			LastName:        input.LastName,
			Username:        input.Username,
			EmailAddress:    input.EmailAddress,
			HashedPassword:  input.HashedPassword,
			TwoFactorSecret: input.TwoFactorSecret,
			AccountStatus:   string(identity.UnverifiedAccountStatus),
			Birthday:        input.Birthday,
			CreatedAt:       r.CurrentTime(),
		}
		logger = logger.WithValue(identitykeys.UserIDKey, user.ID)
		tracing.AttachToSpan(span, identitykeys.UserIDKey, user.ID)

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    input.ID,
			EventType:     audit.AuditLogEventTypeCreated,
			BelongsToUser: input.ID,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// Assign the default service_user role.
		if err = r.generatedQuerier.AssignRoleToUser(ctx, tx, &generated.AssignRoleToUserParams{
			ID:        identifiers.New(),
			UserID:    user.ID,
			RoleID:    authorization.ServiceUserRoleID,
			AccountID: sql.NullString{},
		}); err != nil {
			return observability.PrepareError(err, span, "assigning service role to user")
		}

		if strings.TrimSpace(input.AccountName) == "" {
			input.AccountName = fmt.Sprintf("%s's cool account", input.Username)
		}

		account, err := r.createAccountForUser(ctx, tx, hasValidInvite, input.AccountName, user.ID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating account for new user")
		}
		logger = logger.WithValue(identitykeys.AccountIDKey, account.ID)
		logger.Debug("account created")

		if hasValidInvite {
			if err = r.acceptInvitationForUser(ctx, tx, input); err != nil {
				return observability.PrepareAndLogError(err, logger, span, "accepting account invitations")
			}
			logger.Debug("accepted invitation and joined account for user")
		}

		if err = r.attachInvitationsToUser(ctx, tx, user.EmailAddress, user.ID); err != nil {
			logger = logger.WithValue("email_address", user.EmailAddress).WithValue("user_id", user.ID)
			return observability.PrepareAndLogError(err, logger, span, "attaching existing invitations to new user")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.UserSignedUpServiceEventType, "", map[string]any{
			identitykeys.AccountIDKey:                  account.ID,
			identitykeys.UserIDKey:                     input.ID,
			identitykeys.UserEmailVerificationTokenKey: token,
		}, events.WithIndexUpsert(identityindexing.IndexTypeUsers, input.ID)); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Debug("user and account created")

	return user, nil
}

func (r *repository) createAccountForUser(ctx context.Context, querier database.SQLQueryExecutor, hasValidInvite bool, accountName, userID string) (*identity.Account, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	// standard registration: we need to create the account
	accountID := identifiers.New()
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	hn := accountName
	if accountName == "" {
		hn = fmt.Sprintf("%s_default", userID)
	}

	accountCreationInput := &identity.AccountDatabaseCreationInput{
		ID:            accountID,
		Name:          hn,
		BelongsToUser: userID,
	}

	// create the account.
	if err := r.generatedQuerier.CreateAccount(ctx, querier, &generated.CreateAccountParams{
		City:              accountCreationInput.City,
		Name:              accountCreationInput.Name,
		BillingStatus:     identity.UnpaidAccountBillingStatus,
		ContactPhone:      accountCreationInput.ContactPhone,
		AddressLine1:      accountCreationInput.AddressLine1,
		AddressLine2:      accountCreationInput.AddressLine2,
		ID:                accountCreationInput.ID,
		State:             accountCreationInput.State,
		ZipCode:           accountCreationInput.ZipCode,
		Country:           accountCreationInput.Country,
		BelongsToUser:     accountCreationInput.BelongsToUser,
		Latitude:          database.NullStringFromFloat64Pointer(accountCreationInput.Latitude),
		Longitude:         database.NullStringFromFloat64Pointer(accountCreationInput.Longitude),
		WebhookHmacSecret: accountCreationInput.WebhookEncryptionKey,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating account")
	}

	if err := r.auditLogEntryRepo.Record(ctx, querier, &audit.AuditLogEntry{
		BelongsToAccount: &accountCreationInput.ID,
		ResourceType:     resourceTypeAccounts,
		RelevantID:       accountCreationInput.ID,
		EventType:        audit.AuditLogEventTypeCreated,
		BelongsToUser:    accountCreationInput.BelongsToUser,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	accountMembershipID := identifiers.New()
	if err := r.generatedQuerier.CreateAccountUserMembershipForNewUser(ctx, querier, &generated.CreateAccountUserMembershipForNewUserParams{
		ID:               accountMembershipID,
		BelongsToUser:    userID,
		BelongsToAccount: accountID,
		DefaultAccount:   !hasValidInvite,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "writing account user membership")
	}

	// Account owners get account_admin role.
	if err := r.generatedQuerier.AssignRoleToUser(ctx, querier, &generated.AssignRoleToUserParams{
		ID:        identifiers.New(),
		UserID:    userID,
		RoleID:    authorization.AccountAdminRoleID,
		AccountID: sql.NullString{String: accountID, Valid: true},
	}); err != nil {
		return nil, observability.PrepareError(err, span, "assigning account role to user")
	}

	if err := r.auditLogEntryRepo.Record(ctx, querier, &audit.AuditLogEntry{
		BelongsToAccount: &accountCreationInput.ID,
		ResourceType:     resourceTypeAccountUserMemberships,
		RelevantID:       accountMembershipID,
		EventType:        audit.AuditLogEventTypeCreated,
		BelongsToUser:    accountCreationInput.BelongsToUser,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	account := &identity.Account{
		CreatedAt:            r.CurrentTime(),
		Longitude:            accountCreationInput.Longitude,
		Latitude:             accountCreationInput.Latitude,
		State:                accountCreationInput.State,
		ContactPhone:         accountCreationInput.ContactPhone,
		City:                 accountCreationInput.City,
		AddressLine1:         accountCreationInput.AddressLine1,
		ZipCode:              accountCreationInput.ZipCode,
		Country:              accountCreationInput.Country,
		BillingStatus:        identity.UnpaidAccountBillingStatus,
		AddressLine2:         accountCreationInput.AddressLine2,
		BelongsToUser:        accountCreationInput.BelongsToUser,
		ID:                   accountCreationInput.ID,
		Name:                 accountCreationInput.Name,
		WebhookEncryptionKey: accountCreationInput.WebhookEncryptionKey,
		Members:              nil,
	}

	return account, nil
}

// UpdateUserUsername updates a user's username.
func (r *repository) UpdateUserUsername(ctx context.Context, userID, newUsername string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if newUsername == "" {
		return platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(identitykeys.UsernameKey, newUsername)
	tracing.AttachToSpan(span, identitykeys.UsernameKey, newUsername)

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		user, err := r.GetUser(ctx, userID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching user")
		}

		if _, err = r.generatedQuerier.UpdateUserUsername(ctx, tx, &generated.UpdateUserUsernameParams{
			Username: newUsername,
			ID:       userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating username")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"username": {
					Old: user.Username,
					New: newUsername,
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.UsernameChangedEventType, "", map[string]any{
			identitykeys.UserIDKey: userID,
		}, events.WithIndexUpsert(identityindexing.IndexTypeUsers, userID)); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("username updated")

	return nil
}

// UpdateUserEmailAddress updates a user's username.
func (r *repository) UpdateUserEmailAddress(ctx context.Context, userID, newEmailAddress string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := r.logger.WithValue(identitykeys.UserEmailAddressKey, newEmailAddress).WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if newEmailAddress == "" {
		return platformerrors.ErrEmptyInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserEmailAddressKey, newEmailAddress)

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		user, err := r.GetUser(ctx, userID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching user")
		}

		if _, err = r.generatedQuerier.UpdateUserEmailAddress(ctx, tx, &generated.UpdateUserEmailAddressParams{
			EmailAddress: newEmailAddress,
			ID:           userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user email address")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"email_address": {
					Old: user.EmailAddress,
					New: newEmailAddress,
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.EmailAddressChangedEventType, "", map[string]any{
			identitykeys.UserIDKey: userID,
		}, events.WithIndexUpsert(identityindexing.IndexTypeUsers, userID)); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user email address updated")

	return nil
}

// UpdateUserDetails updates a user's username.
func (r *repository) UpdateUserDetails(ctx context.Context, userID string, input *identity.UserDetailsDatabaseUpdateInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrEmptyInputProvided
	}

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		user, err := r.GetUser(ctx, userID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching user")
		}

		if _, err = r.generatedQuerier.UpdateUserDetails(ctx, tx, &generated.UpdateUserDetailsParams{
			FirstName: input.FirstName,
			LastName:  input.LastName,
			Birthday:  database.NullTimeFromTime(input.Birthday),
			ID:        userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user details")
		}

		changes := map[string]audit.Change{}
		if input.FirstName != user.FirstName {
			changes["first_name"] = audit.Change{New: input.FirstName, Old: user.FirstName}
		}

		if input.LastName != user.LastName {
			changes["last_name"] = audit.Change{New: input.LastName, Old: user.LastName}
		}

		if input.Birthday.Format(time.Kitchen) != user.Birthday.Format(time.Kitchen) {
			changes["birthday"] = audit.Change{New: input.Birthday.Format(time.Kitchen), Old: user.Birthday.Format(time.Kitchen)}
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes:       changes,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.UserDetailsChangedEventType, "", map[string]any{
			identitykeys.UserIDKey: userID,
		}, events.WithIndexUpsert(identityindexing.IndexTypeUsers, userID)); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user details updated")

	return nil
}

// SetUserAvatar sets a user's avatar to the given uploaded media.
func (r *repository) SetUserAvatar(ctx context.Context, userID, uploadedMediaID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if uploadedMediaID == "" {
		return platformerrors.ErrEmptyInputProvided
	}

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.ArchiveUserAvatar(ctx, tx, userID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving previous user avatar")
		}

		if err = r.generatedQuerier.CreateUserAvatar(ctx, tx, &generated.CreateUserAvatarParams{
			ID:              identifiers.New(),
			BelongsToUser:   userID,
			UploadedMediaID: uploadedMediaID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating user avatar")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"avatar": {},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.UserAvatarChangedEventType, "", map[string]any{
			identitykeys.UserIDKey: userID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user avatar updated")

	return nil
}

// UpdateUserPassword updates a user's passwords hash in the database.
func (r *repository) UpdateUserPassword(ctx context.Context, userID, newHash string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if newHash == "" {
		return platformerrors.ErrEmptyInputProvided
	}

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if _, err = r.generatedQuerier.UpdateUserPassword(ctx, tx, &generated.UpdateUserPasswordParams{
			HashedPassword: newHash,
			ID:             userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user password")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"password": {},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user password updated")

	return nil
}

// UpdateUserTwoFactorSecret marks a user's two factor secret as validated.
func (r *repository) UpdateUserTwoFactorSecret(ctx context.Context, userID, newSecret string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if newSecret == "" {
		return platformerrors.ErrEmptyInputProvided
	}

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if _, err = r.generatedQuerier.UpdateUserTwoFactorSecret(ctx, tx, &generated.UpdateUserTwoFactorSecretParams{
			TwoFactorSecret: newSecret,
			ID:              userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user 2FA secret")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"two_factor_secret": {},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user two factor secret updated")

	return nil
}

// MarkUserTwoFactorSecretAsVerified marks a user's two factor secret as validated.
func (r *repository) MarkUserTwoFactorSecretAsVerified(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.MarkTwoFactorSecretAsVerified(ctx, tx, userID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "writing verified two factor status to database")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"two_factor_secret": {
					Old: "unverified",
					New: "verified",
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user two factor secret verified")

	return nil
}

// MarkUserTwoFactorSecretAsUnverified marks a user's two factor secret as unverified.
func (r *repository) MarkUserTwoFactorSecretAsUnverified(ctx context.Context, userID, newSecret string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if newSecret == "" {
		return platformerrors.ErrEmptyInputProvided
	}

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.MarkTwoFactorSecretAsUnverified(ctx, tx, &generated.MarkTwoFactorSecretAsUnverifiedParams{
			TwoFactorSecret: newSecret,
			ID:              userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "writing verified two factor status to database")
		}

		// Both entries in one call: they share a scope, so recording them together
		// pays one chain-head lookup and one INSERT rather than two of each, while
		// holding a lock on this user's chain row for half as long.
		if err = r.auditLogEntryRepo.Record(ctx, tx,
			&audit.AuditLogEntry{
				ResourceType:  resourceTypeUsers,
				RelevantID:    userID,
				EventType:     audit.AuditLogEventTypeArchived,
				BelongsToUser: userID,
			},
			&audit.AuditLogEntry{
				ResourceType:  resourceTypeUsers,
				RelevantID:    userID,
				EventType:     audit.AuditLogEventTypeCreated,
				BelongsToUser: userID,
				Changes: map[string]audit.Change{
					"two_factor_secret": {
						Old: "verified",
						New: "unverified",
					},
				},
			},
		); err != nil {
			return observability.PrepareError(err, span, "creating audit log entries")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user two factor secret unverified")

	return nil
}

// ArchiveUser archives a user.
func (r *repository) ArchiveUser(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	// begin archive user transaction
	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		changed, err := r.generatedQuerier.ArchiveUser(ctx, tx, userID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving user")
		}

		if changed == 0 {
			return sql.ErrNoRows
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeArchived,
			BelongsToUser: userID,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		if _, err = r.generatedQuerier.ArchiveUserMemberships(ctx, tx, userID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving user account memberships")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.UserArchivedServiceEventType, "", map[string]any{
			identitykeys.UserIDKey: userID,
		}, events.WithIndexDelete(identityindexing.IndexTypeUsers, userID)); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *repository) GetEmailAddressVerificationTokenForUser(ctx context.Context, userID string) (string, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return "", platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	result, err := r.generatedQuerier.GetEmailVerificationTokenByUserID(ctx, r.readDB, userID)
	if err != nil {
		return "", observability.PrepareError(err, span, "getting user by email address verification token")
	}

	return result.String, nil
}

func (r *repository) GetUserByEmailAddressVerificationToken(ctx context.Context, token string) (*identity.User, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if token == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	result, err := r.generatedQuerier.GetUserByEmailAddressVerificationToken(ctx, r.readDB, database.NullStringFromString(token))
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting user by email address verification token")
	}

	u := &identity.User{
		CreatedAt:                  result.CreatedAt,
		PasswordLastChangedAt:      database.TimePointerFromNullTime(result.PasswordLastChangedAt),
		LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
		LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.TwoFactorSecretVerifiedAt),
		Avatar:                     avatarFromRow(result.AvatarID, result.AvatarStoragePath, result.AvatarMimeType, result.AvatarCreatedAt, result.AvatarLastUpdatedAt, result.AvatarArchivedAt, result.AvatarCreatedByUser),
		Birthday:                   database.TimePointerFromNullTime(result.Birthday),
		ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
		AccountStatusExplanation:   result.UserAccountStatusExplanation,
		TwoFactorSecret:            result.TwoFactorSecret,
		HashedPassword:             result.HashedPassword,
		ID:                         result.ID,
		AccountStatus:              result.UserAccountStatus,
		Username:                   result.Username,
		FirstName:                  result.FirstName,
		LastName:                   result.LastName,
		EmailAddress:               result.EmailAddress,
		EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.EmailAddressVerifiedAt),
		RequiresPasswordChange:     result.RequiresPasswordChange,
	}

	return u, nil
}

func (r *repository) MarkUserEmailAddressAsVerified(ctx context.Context, userID, token string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	if token == "" {
		return platformerrors.ErrEmptyInputProvided
	}

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.MarkEmailAddressAsVerified(ctx, tx, &generated.MarkEmailAddressAsVerifiedParams{
			ID:                            userID,
			EmailAddressVerificationToken: database.NullStringFromString(token),
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return err
			}

			return observability.PrepareAndLogError(err, logger, span, "writing verified email address status to database")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"email_address_verification": {
					Old: "unverified",
					New: "verified",
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		if _, err = r.generatedQuerier.SetUserAccountStatus(ctx, tx, &generated.SetUserAccountStatusParams{
			UserAccountStatus:            string(identity.GoodStandingUserAccountStatus),
			UserAccountStatusExplanation: "verified email address",
			ID:                           userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user account status")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *repository) MarkUserEmailAddressAsUnverified(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.MarkEmailAddressAsUnverified(ctx, tx, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return err
			}

			return observability.PrepareAndLogError(err, logger, span, "writing email address verification status to database")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUsers,
			RelevantID:    userID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: userID,
			Changes: map[string]audit.Change{
				"email_address_verification": {
					Old: "verified",
					New: "unverified",
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		if _, err = r.generatedQuerier.SetUserAccountStatus(ctx, tx, &generated.SetUserAccountStatusParams{
			UserAccountStatus:            string(identity.UnverifiedAccountStatus),
			UserAccountStatusExplanation: "unverified email address",
			ID:                           userID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user account status")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *repository) UpdateUserAccountStatus(ctx context.Context, userID string, input *identity.UserAccountStatusUpdateInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	return r.withEvent(ctx, logger, identity.UserStatusChangedServiceEventType, "", map[string]any{
		identitykeys.UserIDKey: userID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsChanged, err := r.generatedQuerier.SetUserAccountStatus(ctx, tx, &generated.SetUserAccountStatusParams{
			UserAccountStatus:            input.NewStatus,
			UserAccountStatusExplanation: input.Reason,
			ID:                           input.TargetUserID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "user status update")
		}

		if rowsChanged == 0 {
			return sql.ErrNoRows
		}

		logger.Info("user account status updated")

		return nil
	})
}

func (r *repository) SetUserRequiresPasswordChange(ctx context.Context, userID string, requiresChange bool) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	return r.withEvent(ctx, logger, identity.UserPasswordChangeRequiredServiceEventType, "", map[string]any{
		identitykeys.UserIDKey: userID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsChanged, err := r.generatedQuerier.SetUserRequiresPasswordChange(ctx, tx, &generated.SetUserRequiresPasswordChangeParams{
			RequiresPasswordChange: requiresChange,
			ID:                     userID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "setting user requires password change")
		}

		if rowsChanged == 0 {
			return sql.ErrNoRows
		}

		logger.Info("user requires password change updated")

		return nil
	})
}

func (r *repository) UserRequiresPasswordChange(ctx context.Context, userID string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	result, err := r.generatedQuerier.GetUserRequiresPasswordChange(ctx, r.readDB, userID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking if user requires password change")
	}

	return result, nil
}
