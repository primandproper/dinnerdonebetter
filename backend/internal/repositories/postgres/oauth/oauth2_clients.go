package oauth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	oauthkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	resourceTypeOAuth2Clients = "oauth2_clients"
)

var (
	_ types.OAuth2ClientDataManager = (*repository)(nil)
)

// GetOAuth2ClientByClientID gets an OAuth2 client from the database.
func (q *repository) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*types.OAuth2Client, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if clientID == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientClientIDKey, clientID)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientClientIDKey, clientID)

	result, err := q.generatedQuerier.GetOAuth2ClientByClientID(ctx, q.readDB, clientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching oauth2 client")
	}

	client := &types.OAuth2Client{
		CreatedAt:    result.CreatedAt,
		ArchivedAt:   database.TimePointerFromNullTime(result.ArchivedAt),
		Name:         result.Name,
		Description:  result.Description,
		ClientID:     result.ClientID,
		ID:           result.ID,
		ClientSecret: result.ClientSecret,
	}

	return client, nil
}

// GetOAuth2ClientByDatabaseID gets an OAuth2 client from the database.
func (q *repository) GetOAuth2ClientByDatabaseID(ctx context.Context, clientID string) (*types.OAuth2Client, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if clientID == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientClientIDKey, clientID)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientClientIDKey, clientID)

	result, err := q.generatedQuerier.GetOAuth2ClientByDatabaseID(ctx, q.readDB, clientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching oauth2 client")
	}

	client := &types.OAuth2Client{
		CreatedAt:    result.CreatedAt,
		ArchivedAt:   database.TimePointerFromNullTime(result.ArchivedAt),
		Name:         result.Name,
		Description:  result.Description,
		ClientID:     result.ClientID,
		ID:           result.ID,
		ClientSecret: result.ClientSecret,
	}

	return client, nil
}

// GetOAuth2Clients gets a list of OAuth2 clients.
func (q *repository) GetOAuth2Clients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.OAuth2Client], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetOAuth2Clients(ctx, q.readDB, &generated.GetOAuth2ClientsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching oauth2 clients")
	}

	var (
		data                      = []*types.OAuth2Client{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.OAuth2Client{
			CreatedAt:    result.CreatedAt,
			ArchivedAt:   database.TimePointerFromNullTime(result.ArchivedAt),
			Name:         result.Name,
			Description:  result.Description,
			ClientID:     result.ClientID,
			ID:           result.ID,
			ClientSecret: result.ClientSecret,
		})
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.OAuth2Client) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateOAuth2Client creates an OAuth2 client.
func (q *repository) CreateOAuth2Client(ctx context.Context, input *types.OAuth2ClientDatabaseCreationInput) (*types.OAuth2Client, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]any{
		oauthkeys.OAuth2ClientClientIDKey: input.ClientID,
	})

	var err error
	if err = q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if writeErr := q.generatedQuerier.CreateOAuth2Client(ctx, tx, &generated.CreateOAuth2ClientParams{
			ID:           input.ID,
			Description:  input.Description,
			Name:         input.Name,
			ClientID:     input.ClientID,
			ClientSecret: input.ClientSecret,
		}); writeErr != nil {
			return observability.PrepareError(writeErr, span, "creating OAuth2 client")
		}

		tracing.AttachToSpan(span, oauthkeys.OAuth2ClientClientIDKey, input.ID)

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			ResourceType: resourceTypeOAuth2Clients,
			ResourceID:   input.ID,
			EventType:    audit.EventCreated,
			Actor:        audit.SystemActor(),
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	client := &types.OAuth2Client{
		ID:           input.ID,
		Name:         input.Name,
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		CreatedAt:    q.CurrentTime(),
	}

	logger.Info("OAuth2 client created")

	return client, nil
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (q *repository) ArchiveOAuth2Client(ctx context.Context, clientID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == "" {
		return platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientClientIDKey, clientID)
	logger := q.logger.WithValue(oauthkeys.OAuth2ClientIDKey, clientID)

	// The archival and the entry describing it go in one transaction. Record locks
	// the scope's chain-head row for the rest of the caller's transaction, and
	// against the connection pool there is no rest of the transaction — the lock
	// lapses before the INSERT it was taken for, so two concurrent writers compute
	// the same chain position and the unique index rejects the second.
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveOAuth2Client(ctx, tx, clientID)
		if archiveErr != nil {
			return archiveErr
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return q.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			ResourceType: resourceTypeOAuth2Clients,
			ResourceID:   clientID,
			EventType:    audit.EventArchived,
			Actor:        audit.SystemActor(),
		})
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}

		return observability.PrepareAndLogError(err, logger, span, "archiving OAuth2 client")
	}

	return nil
}
