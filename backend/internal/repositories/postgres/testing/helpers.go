package testing

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"

	encryptioncfg "github.com/primandproper/platform-go/v4/cryptography/encryption/config"
	"github.com/primandproper/platform-go/v4/database"
	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/identifiers"
	"github.com/primandproper/platform-go/v4/retry"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var RunContainerTests = strings.ToLower(os.Getenv("RUN_CONTAINER_TESTS")) != "false" // on by default

func MustHashStringToNumber(s string) uint64 {
	h := fnv.New64a()

	if _, err := h.Write([]byte(s)); err != nil {
		panic(err)
	}

	return h.Sum64()
}

func HashStringToNumberForTest(t *testing.T, s string) uint64 {
	t.Helper()
	h := fnv.New64a()

	_, err := h.Write([]byte(s))
	require.NoError(t, err)

	return h.Sum64()
}

func reverseString(input string) string {
	runes := []rune(input)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func splitReverseConcat(input string) string {
	length := len(input)
	halfLength := length / 2

	firstHalf := input[:halfLength]
	secondHalf := input[halfLength:]

	reversedFirstHalf := reverseString(firstHalf)
	reversedSecondHalf := reverseString(secondHalf)

	return reversedSecondHalf + reversedFirstHalf
}

const (
	defaultPostgresImage = "postgres:17"
)

// userFromGetUserByIDRow converts a GetUserByIDRow to User. Used by CreateUserForTest.
func userFromGetUserByIDRow(row *generated.GetUserByIDRow) *identity.User {
	var avatar *uploadedmedia.UploadedMedia
	if row.AvatarID.Valid && row.AvatarStoragePath.Valid && row.AvatarMimeType.Valid && row.AvatarCreatedByUser.Valid && row.AvatarCreatedAt.Valid {
		avatar = &uploadedmedia.UploadedMedia{
			ID:            row.AvatarID.String,
			StoragePath:   row.AvatarStoragePath.String,
			MimeType:      string(row.AvatarMimeType.UploadedMediaMimeType),
			CreatedAt:     row.AvatarCreatedAt.Time,
			LastUpdatedAt: database.TimePointerFromNullTime(row.AvatarLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(row.AvatarArchivedAt),
			CreatedByUser: row.AvatarCreatedByUser.String,
		}
	}
	return &identity.User{
		CreatedAt:                  row.CreatedAt,
		PasswordLastChangedAt:      database.TimePointerFromNullTime(row.PasswordLastChangedAt),
		LastUpdatedAt:              database.TimePointerFromNullTime(row.LastUpdatedAt),
		LastAcceptedTermsOfService: database.TimePointerFromNullTime(row.LastAcceptedTermsOfService),
		LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(row.LastAcceptedPrivacyPolicy),
		TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(row.TwoFactorSecretVerifiedAt),
		Birthday:                   database.TimePointerFromNullTime(row.Birthday),
		ArchivedAt:                 database.TimePointerFromNullTime(row.ArchivedAt),
		AccountStatusExplanation:   row.UserAccountStatusExplanation,
		TwoFactorSecret:            row.TwoFactorSecret,
		HashedPassword:             row.HashedPassword,
		ID:                         row.ID,
		AccountStatus:              row.UserAccountStatus,
		Username:                   row.Username,
		FirstName:                  row.FirstName,
		LastName:                   row.LastName,
		EmailAddress:               row.EmailAddress,
		EmailAddressVerifiedAt:     database.TimePointerFromNullTime(row.EmailAddressVerifiedAt),
		Avatar:                     avatar,
		RequiresPasswordChange:     row.RequiresPasswordChange,
	}
}

func BuildDatabaseContainerForTest(t *testing.T) (*postgres.PostgresContainer, *sql.DB, *databasecfg.Config) {
	t.Helper()

	container, db, dbConfig, err := BuildDatabaseContainer(t.Context(), t.Name())
	require.NoError(t, err)

	return container, db, dbConfig
}

func BuildDatabaseContainer(ctx context.Context, dbName string) (*postgres.PostgresContainer, *sql.DB, *databasecfg.Config, error) {
	dbUsername := fmt.Sprintf("%d", MustHashStringToNumber(dbName))

	var container *postgres.PostgresContainer
	policy := retry.NewExponentialBackoffPolicy(retry.Config{MaxAttempts: 5, InitialDelay: 1, UseJitter: false})
	if err := policy.Execute(ctx, func(ctx context.Context) error {
		var containerErr error
		container, containerErr = postgres.Run(
			ctx,
			defaultPostgresImage,
			postgres.WithDatabase(splitReverseConcat(dbUsername)),
			postgres.WithUsername(dbUsername),
			postgres.WithPassword(reverseString(dbUsername)),
			testcontainers.WithWaitStrategyAndDeadline(2*time.Minute, wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
		)
		return containerErr
	}); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if container == nil {
		return nil, nil, nil, fmt.Errorf("container %s not found", dbName)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to postgres container: %w", err)
	}

	dbConfig := &databasecfg.Config{
		RunMigrations:            false,
		Encryption:               encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
		OAuth2TokenEncryptionKey: "blahblahblahblahblahblahblahblah",
	}
	if err = dbConfig.LoadConnectionDetailsFromURL(connStr); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to postgres container: %w", err)
	}
	// LoadConnectionDetailsFromURL only populates ReadConnection; copy it to
	// WriteConnection so NewDatabaseClient can open both handles.
	dbConfig.WriteConnection = dbConfig.ReadConnection

	db, err := dbConfig.ConnectToWriteDatabase()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to postgres container: %w", err)
	}

	return container, db, dbConfig, nil
}

func CreateUserForTest(t *testing.T, exampleUser *identity.User, db *sql.DB) *identity.User {
	t.Helper()

	ctx := t.Context()

	// create
	if exampleUser == nil {
		exampleUser = fakes.BuildFakeUser()
	}
	exampleUser.TwoFactorSecretVerifiedAt = nil

	dbc := generated.New()

	err := dbc.CreateUser(ctx, db, &generated.CreateUserParams{
		ID:                            exampleUser.ID,
		Username:                      exampleUser.Username,
		EmailAddress:                  exampleUser.EmailAddress,
		HashedPassword:                exampleUser.HashedPassword,
		RequiresPasswordChange:        exampleUser.RequiresPasswordChange,
		TwoFactorSecret:               exampleUser.TwoFactorSecret,
		TwoFactorSecretVerifiedAt:     database.NullTimeFromTimePointer(exampleUser.TwoFactorSecretVerifiedAt),
		UserAccountStatus:             exampleUser.AccountStatus,
		UserAccountStatusExplanation:  exampleUser.AccountStatusExplanation,
		Birthday:                      database.NullTimeFromTimePointer(exampleUser.Birthday),
		EmailAddressVerificationToken: database.NullStringFromString("token"),
		FirstName:                     exampleUser.FirstName,
		LastName:                      exampleUser.LastName,
	})
	require.NoError(t, err)

	// Assign default service_user role.
	require.NoError(t, dbc.AssignRoleToUser(ctx, db, &generated.AssignRoleToUserParams{
		ID:        identifiers.New(),
		UserID:    exampleUser.ID,
		RoleID:    authorization.ServiceUserRoleID,
		AccountID: sql.NullString{},
	}))

	dbCreated, err := dbc.GetUserByID(ctx, db, exampleUser.ID)
	require.NoError(t, err)

	created := userFromGetUserByIDRow(dbCreated)
	exampleUser.CreatedAt = created.CreatedAt
	exampleUser.Birthday = created.Birthday
	exampleUser.TwoFactorSecretVerifiedAt = created.TwoFactorSecretVerifiedAt
	assert.Equal(t, exampleUser, created)

	return created
}

func CreateAccountForTest(t *testing.T, exampleAccount *identity.Account, userID string, db *sql.DB) *identity.Account {
	t.Helper()

	// create
	if exampleAccount == nil {
		exampleAccount = fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = userID
	}
	exampleAccount.PaymentProcessorCustomerID = ""
	exampleAccount.Members = nil

	ctx := t.Context()
	dbc := generated.New()

	require.NoError(t, dbc.CreateAccount(ctx, db, &generated.CreateAccountParams{
		ID:                exampleAccount.ID,
		Name:              exampleAccount.Name,
		BillingStatus:     exampleAccount.BillingStatus,
		ContactPhone:      exampleAccount.ContactPhone,
		BelongsToUser:     userID,
		AddressLine1:      exampleAccount.AddressLine1,
		AddressLine2:      exampleAccount.AddressLine2,
		City:              exampleAccount.City,
		State:             exampleAccount.State,
		ZipCode:           exampleAccount.ZipCode,
		Country:           exampleAccount.Country,
		Latitude:          database.NullStringFromFloat64Pointer(exampleAccount.Latitude),
		Longitude:         database.NullStringFromFloat64Pointer(exampleAccount.Longitude),
		WebhookHmacSecret: exampleAccount.WebhookEncryptionKey,
	}))

	require.NoError(t, dbc.CreateAccountUserMembershipForNewUser(ctx, db, &generated.CreateAccountUserMembershipForNewUserParams{
		ID:               identifiers.New(),
		BelongsToAccount: exampleAccount.ID,
		BelongsToUser:    userID,
		DefaultAccount:   true,
	}))

	// Account owners get account_admin role.
	require.NoError(t, dbc.AssignRoleToUser(ctx, db, &generated.AssignRoleToUserParams{
		ID:        identifiers.New(),
		UserID:    userID,
		RoleID:    authorization.AccountAdminRoleID,
		AccountID: sql.NullString{String: exampleAccount.ID, Valid: true},
	}))

	dbCreated, err := dbc.GetAccountsForUser(ctx, db, &generated.GetAccountsForUserParams{
		BelongsToUser: userID,
	})
	require.NoError(t, err)
	require.Len(t, dbCreated, 1)

	created := &identity.Account{
		CreatedAt:                  dbCreated[0].CreatedAt,
		SubscriptionPlanID:         database.StringPointerFromNullString(dbCreated[0].SubscriptionPlanID),
		LastUpdatedAt:              database.TimePointerFromNullTime(dbCreated[0].LastUpdatedAt),
		ArchivedAt:                 database.TimePointerFromNullTime(dbCreated[0].ArchivedAt),
		Longitude:                  database.Float64PointerFromNullString(dbCreated[0].Longitude),
		Latitude:                   database.Float64PointerFromNullString(dbCreated[0].Latitude),
		State:                      dbCreated[0].State,
		ContactPhone:               dbCreated[0].ContactPhone,
		City:                       dbCreated[0].City,
		AddressLine1:               dbCreated[0].AddressLine1,
		ZipCode:                    dbCreated[0].ZipCode,
		Country:                    dbCreated[0].Country,
		BillingStatus:              dbCreated[0].BillingStatus,
		AddressLine2:               dbCreated[0].AddressLine2,
		PaymentProcessorCustomerID: dbCreated[0].PaymentProcessorCustomerID,
		BelongsToUser:              dbCreated[0].BelongsToUser,
		ID:                         dbCreated[0].ID,
		Name:                       dbCreated[0].Name,
		WebhookEncryptionKey:       dbCreated[0].WebhookHmacSecret,
	}

	exampleAccount.CreatedAt = created.CreatedAt
	exampleAccount.WebhookEncryptionKey = created.WebhookEncryptionKey
	assert.Equal(t, exampleAccount, created)

	return created
}

// PaginationTestConfig contains the configuration for testing cursor-based pagination.
type PaginationTestConfig[T any] struct {
	CreateItem  func(ctx context.Context, i int) *T
	FetchPage   func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error)
	GetID       func(item *T) string
	CleanupItem func(ctx context.Context, item *T) error
	ItemName    string
	TotalItems  int
	PageSize    int
	// ExpectedTotalCount is the total number of items expected in the result (TotalCount/FilteredCount).
	// Use when the table has pre-existing rows from migrations. If 0, TotalItems is used.
	ExpectedTotalCount int
}

// TestCursorBasedPagination is a generic test function for cursor-based pagination.
// It creates items, fetches them using cursor-based pagination, and verifies:
//   - All items are retrieved exactly once
//   - Items are returned in ascending ID order
//   - Pagination counts are accurate
//   - The expected number of pages is fetched
func TestCursorBasedPagination[T any](t *testing.T, ctx context.Context, config PaginationTestConfig[T]) {
	t.Helper()

	expectedTotal := config.TotalItems
	if config.ExpectedTotalCount > 0 {
		expectedTotal = config.ExpectedTotalCount
	}

	// Calculate expected pages
	expectedPages := (expectedTotal + config.PageSize - 1) / config.PageSize
	if expectedTotal%config.PageSize == 0 {
		// For evenly divisible cases, we still get one empty page at the end
		expectedPages = expectedTotal / config.PageSize
	}

	// Create test items
	createdItems := make([]*T, 0, config.TotalItems)

	for i := 0; i < config.TotalItems; i++ {
		item := config.CreateItem(ctx, i)
		createdItems = append(createdItems, item)
	}

	// Track all items we retrieve via pagination
	allPaginatedItems := []*T{}
	var cursor *string // Start with no cursor for the first page
	pageCount := 0

	// Paginate through all results
	for {
		pageCount++
		filter := &filtering.QueryFilter{
			MaxResponseSize: filtering.DefaultQueryFilter().MaxResponseSize,
			Cursor:          cursor,
		}
		// Override the default page size with our test page size
		customPageSize := uint8(config.PageSize)
		filter.MaxResponseSize = &customPageSize

		result, err := config.FetchPage(ctx, filter)
		require.NoError(t, err, "failed to fetch page %d", pageCount)
		require.NotNil(t, result, "result should not be nil for page %d", pageCount)

		// If this page is empty, we've gone past the end (cursor-based pagination characteristic)
		if len(result.Data) == 0 {
			break
		}

		// Verify we got the expected number of results (full pages should be evenly sized)
		if len(result.Data) == config.PageSize {
			assert.Len(t, result.Data, config.PageSize, "page %d should contain exactly %d %ss", pageCount, config.PageSize, config.ItemName)
		}

		// Verify counts are accurate when there's data
		assert.Equal(t, uint64(expectedTotal), result.TotalCount, "total count should be %d", expectedTotal)
		assert.Equal(t, uint64(expectedTotal), result.FilteredCount, "filtered count should be %d", expectedTotal)

		// Add results to our collection
		allPaginatedItems = append(allPaginatedItems, result.Data...)

		// If we got fewer results than the page size, we're on the last page
		if len(result.Data) < config.PageSize {
			break
		}

		// Use the last ID from this page as the cursor for the next page
		if len(result.Data) > 0 {
			lastID := config.GetID(result.Data[len(result.Data)-1])
			cursor = &lastID
		} else {
			break
		}

		// Safety check to prevent infinite loops
		assert.False(t, pageCount > expectedTotal+5, "Too many pages fetched, possible infinite loop")
	}

	// With cursor-based pagination: when the last page is partial we break on it (pageCount = expectedPages);
	// when all pages are full we need one more request to get empty (pageCount = expectedPages+1)
	expectedPageRequests := expectedPages
	if expectedTotal%config.PageSize == 0 {
		expectedPageRequests = expectedPages + 1
	}
	assert.Equal(t, expectedPageRequests, pageCount, "should have made %d requests", expectedPageRequests)

	// Verify we got all items
	assert.Len(t, allPaginatedItems, expectedTotal, "should have retrieved all %d %ss via pagination", expectedTotal, config.ItemName)

	// Verify no duplicates - create a map of IDs
	seenIDs := make(map[string]bool)
	for _, item := range allPaginatedItems {
		id := config.GetID(item)
		assert.False(t, seenIDs[id], "Duplicate %s ID found: %s", config.ItemName, id)
		seenIDs[id] = true
	}

	// Verify all created items were retrieved
	for _, created := range createdItems {
		id := config.GetID(created)
		assert.True(t, seenIDs[config.GetID(created)], "Created %s %s was not retrieved via pagination", config.ItemName, id)
	}

	// Verify items are returned in ascending ID order (cursor-based pagination requirement)
	for i := 1; i < len(allPaginatedItems); i++ {
		prevID := config.GetID(allPaginatedItems[i-1])
		currID := config.GetID(allPaginatedItems[i])
		assert.True(t, prevID < currID,
			"%ss should be ordered by ID ascending: %s should be < %s (position %d and %d)",
			config.ItemName, prevID, currID, i-1, i)
	}

	// Cleanup all created items
	for _, item := range createdItems {
		if config.CleanupItem != nil {
			err := config.CleanupItem(ctx, item)
			assert.NoError(t, err, "failed to cleanup %s %s", config.ItemName, config.GetID(item))
		}
	}
}
