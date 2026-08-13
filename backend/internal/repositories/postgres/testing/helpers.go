package testing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"

	encryptioncfg "github.com/primandproper/platform-go/v10/cryptography/encryption/config"
	"github.com/primandproper/platform-go/v10/database"
	databasecfg "github.com/primandproper/platform-go/v10/database/config"
	mockdatabase "github.com/primandproper/platform-go/v10/database/mock"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/identifiers"
	"github.com/primandproper/platform-go/v10/testutils/containers"
	"github.com/primandproper/platform-go/v10/testutils/containers/pgtest"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

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

	// driverName is the database/sql driver these helpers open pools with. The pgx
	// stdlib driver is registered by this package's blank import.
	driverName = "pgx"

	// testOAuth2TokenEncryptionKey is a throwaway key for containers that never
	// outlive the test that started them.
	testOAuth2TokenEncryptionKey = "blahblahblahblahblahblahblahblah" /* #nosec G101 */
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

// startupDeadline bounds how long a container has to become ready. It is applied
// to each sub-strategy individually as well as to the wait as a whole — see
// waitStrategy for why the latter alone is not enough.
const startupDeadline = 2 * time.Minute

// waitStrategy is the readiness check every postgres container here is started with.
//
// The log line alone is not enough: postgres announces readiness inside the container
// before Docker's host-side port forward reliably accepts connections, and this suite
// starts containers many-at-a-time from parallel subtests. Since pgtest pings the pool as
// soon as the container reports ready, a log-only strategy loses that race under load and
// the ping fails with "connection refused". Waiting on the mapped port as well closes it.
//
// Every sub-strategy carries its own startup timeout, and the deadline on the
// enclosing ForAll does not override them: each one falls back to testcontainers'
// 60s default unless told otherwise. Under a loaded Docker daemon that is the
// timeout that actually fires, so both are pinned to startupDeadline explicitly.
// A timeout here is also effectively fatal — containers.StartWithRetry runs on a
// retry policy that treats a wrapped context.DeadlineExceeded as terminal, so a
// container that trips this wait gets no second attempt.
func waitStrategy() testcontainers.ContainerCustomizer {
	return testcontainers.WithWaitStrategyAndDeadline(startupDeadline, wait.ForAll(
		wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(startupDeadline),
		wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(startupDeadline),
	))
}

// credentialsFor derives the database name, username and password a container is
// provisioned with from a caller-supplied name, so concurrently running suites do not
// share identifiers.
func credentialsFor(name string) (dbName, username, password string) {
	username = fmt.Sprintf("%d", MustHashStringToNumber(name))

	return splitReverseConcat(username), username, reverseString(username)
}

// databaseConfigForConnectionString renders a container's DSN as the database config the
// repositories expect.
func databaseConfigForConnectionString(connectionString string) (*dbcfg.Config, error) {
	dbConfig := &dbcfg.Config{
		Config: databasecfg.Config{
			RunMigrations: false,
		},
		Encryption:               encryptioncfg.Config{Provider: encryptioncfg.ProviderAES, CurrentKeyID: "v1"},
		OAuth2TokenEncryptionKey: testOAuth2TokenEncryptionKey,
	}

	if err := dbConfig.LoadConnectionDetailsFromURL(connectionString); err != nil {
		return nil, fmt.Errorf("failed to load connection details from postgres container: %w", err)
	}
	// LoadConnectionDetailsFromURL only populates ReadConnection; copy it to
	// WriteConnection so NewDatabaseClient can open both handles.
	dbConfig.WriteConnection = dbConfig.ReadConnection

	return dbConfig, nil
}

// BuildDatabaseContainerForTest stands up a postgres container for the calling test and
// returns an open connection to it along with the config describing it. Both the pool
// and the container are torn down when the test ends, so callers say what they want to
// do with a database and nothing about how it is stood up or shut down.
//
// The container is gated on RUN_CONTAINER_TESTS=true (see containers.SkipIfNotRunning),
// which pgtest.Run enforces on the caller's behalf.
func BuildDatabaseContainerForTest(t *testing.T) (*sql.DB, *dbcfg.Config) {
	t.Helper()

	dbName, username, password := credentialsFor(t.Name())

	var (
		db       *sql.DB
		dbConfig *dbcfg.Config
	)

	pgtest.Run(t,
		func(_ context.Context, pg *pgtest.Instance) {
			var err error
			dbConfig, err = databaseConfigForConnectionString(pg.ConnectionString)
			require.NoError(t, err)

			db = pg.DB
		},
		pgtest.WithImage(defaultPostgresImage),
		pgtest.WithCredentials(dbName, username, password),
		pgtest.WithCustomizers(waitStrategy()),
	)

	return db, dbConfig
}

// BuildDatabaseContainer stands up a postgres container outside of a test, for callers
// like localdev that need a database for the lifetime of a process rather than the
// lifetime of a test. Termination is the caller's responsibility; tests should use
// BuildDatabaseContainerForTest, which handles it.
//
// Extra customizers are applied after the defaults, so a caller can override them —
// RunTestsWithSharedDatabase uses this to raise the server's connection ceiling.
func BuildDatabaseContainer(ctx context.Context, dbName string, customizers ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, *sql.DB, *dbcfg.Config, error) {
	name, username, password := credentialsFor(dbName)

	options := append([]testcontainers.ContainerCustomizer{
		postgres.WithDatabase(name),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		waitStrategy(),
	}, customizers...)

	container, err := containers.StartWithRetry(ctx, func(ctx context.Context) (*postgres.PostgresContainer, error) {
		return postgres.Run(ctx, defaultPostgresImage, options...)
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to postgres container: %w", err)
	}

	dbConfig, err := databaseConfigForConnectionString(connStr)
	if err != nil {
		return nil, nil, nil, err
	}

	// Opened here rather than through the database config, which stopped handing out raw
	// *sql.DB handles in platform-go v10 — the client is the supported surface now. Tests
	// want the handle itself, to seed rows and assert against them without going through a
	// repository, and they want it uninstrumented: a container that lives for one test has
	// nothing to export spans to.
	db, err := sql.Open("pgx", dbConfig.GetWriteConnectionString())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to postgres container: %w", err)
	}

	return container, db, dbConfig, nil
}

func CreateUserForTest(t *testing.T, exampleUser *identity.User, db database.SQLQueryExecutor) *identity.User {
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

func CreateAccountForTest(t *testing.T, exampleAccount *identity.Account, userID string, db database.SQLQueryExecutor) *identity.Account {
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
		customPageSize := uint16(config.PageSize)
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
		assert.LessOrEqual(t, pageCount, expectedTotal+5, "Too many pages fetched, possible infinite loop")
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
		assert.Less(t, prevID, currID,
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

// NewSQLMockDatabaseClient wraps a *sql.DB — typically one produced by sqlmock — in a
// database.Client so repositories under test can exercise Reader, Writer, and
// WithTransaction against it. Transactions run through database.RunInTransaction, so
// begin/commit/rollback behave exactly as they do in production and a mock's
// ExpectBegin/ExpectRollback expectations still apply.
func NewSQLMockDatabaseClient(db *sql.DB) database.Client {
	return &mockdatabase.ClientMock{
		ReaderFunc:      func() database.SQLQueryExecutor { return db },
		WriterFunc:      func() database.SQLQueryExecutor { return db },
		CurrentTimeFunc: time.Now,
		CloseFunc:       db.Close,
		WithTransactionFunc: func(ctx context.Context, fn func(tx database.SQLQueryExecutor) error) error {
			return database.RunInTransaction(ctx, db, rollbackTestTransaction, fn)
		},
	}
}

// rollbackTestTransaction mirrors the production rollback hook for tests. A transaction
// that has already been committed or rolled back reports sql.ErrTxDone, which is not a
// failure worth surfacing; anything else is logged so a mock's unmet ExpectRollback does
// not vanish silently.
func rollbackTestTransaction(_ context.Context, tx database.SQLQueryExecutorAndTransactionManager) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.Printf("rolling back test transaction: %v", err)
	}
}
