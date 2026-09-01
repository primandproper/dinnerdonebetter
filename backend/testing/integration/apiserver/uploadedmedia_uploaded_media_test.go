package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediafakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	"github.com/primandproper/platform-go/v13/uploads/registry"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkUploadedMediaEquality(t *testing.T, expected, actual *registry.Object) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected UploadedMedia to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected UploadedMedia to have CreatedAt")

	assert.Equal(t, expected.Key, actual.Key, "expected UploadedMedia Key")
	assert.Equal(t, expected.ContentType, actual.ContentType, "expected UploadedMedia ContentType")
	assert.NotEmpty(t, actual.OwnerID, "expected UploadedMedia to have OwnerID")
}

// creationInputFor renders a fake object as the registration request's input. The
// owner and the scope are deliberately absent: both come from the session, and a
// caller who could name either could register an object as somebody else's.
func creationInputFor(object *registry.Object) *uploadedmediasvc.UploadedMediaCreationRequestInput {
	return &uploadedmediasvc.UploadedMediaCreationRequestInput{
		ObjectKey:     object.Key,
		ContentType:   object.ContentType,
		SizeBytes:     object.Size,
		BelongsToType: object.BelongsTo.Type,
		BelongsToId:   object.BelongsTo.ID,
	}
}

func createUploadedMediaForTest(t *testing.T, testClient client.Client) *registry.Object {
	t.Helper()
	ctx := t.Context()

	exampleUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()

	createdUploadedMedia, err := testClient.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{Input: creationInputFor(exampleUploadedMedia)})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCUploadedMediaToUploadedMedia(createdUploadedMedia.Created)
	checkUploadedMediaEquality(t, exampleUploadedMedia, converted)

	retrievedUploadedMedia, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, retrievedUploadedMedia)

	uploadedMediaItem := grpcconverters.ConvertGRPCUploadedMediaToUploadedMedia(retrievedUploadedMedia.Result)
	checkUploadedMediaEquality(t, converted, uploadedMediaItem)

	return uploadedMediaItem
}

func TestUploadedMedia_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUploadedMediaForTest(t, testClient)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "uploaded_media", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		// An empty key names no object and an empty content type is not one this
		// deployment stores; either alone is enough to refuse the registration.
		_, err := testClient.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{},
		})
		assert.Error(t, err)
	})

	T.Run("a key already registered is a conflict", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// The key is unique within a scope, archived rows included: two rows for one
		// object is the drift the registry exists to prevent. The client is told the key
		// is spoken for rather than that the server broke, because the remedy is a new key.
		_, testClient := createUserAndClientForTest(t)
		created := createUploadedMediaForTest(t, testClient)

		_, err := testClient.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: creationInputFor(created),
		})
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})
}

func TestUploadedMedia_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient)

		retrieved, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{})
		assert.Error(t, err)
	})

	T.Run("cannot access other user's media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create uploaded media as user 1
		_, testClient1 := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient1)

		// Try to access as user 2
		_, testClient2 := createUserAndClientForTest(t)
		retrieved, err := testClient2.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		require.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestUploadedMedia_ReadingWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdUploadedMedia := []*registry.Object{}
		ids := []string{}
		for range exampleQuantity {
			created := createUploadedMediaForTest(t, testClient)
			createdUploadedMedia = append(createdUploadedMedia, created)
			ids = append(ids, created.ID)
		}

		results, err := testClient.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Results, len(createdUploadedMedia))
	})

	T.Run("filters out other users' media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create media as user 1
		_, testClient1 := createUserAndClientForTest(t)
		media1 := createUploadedMediaForTest(t, testClient1)

		// Create media as user 2
		_, testClient2 := createUserAndClientForTest(t)
		media2 := createUploadedMediaForTest(t, testClient2)

		// User 2 tries to fetch both IDs
		results, err := testClient2.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{media1.ID, media2.ID},
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		// Should only get their own media
		assert.Len(t, results.Results, 1)
		assert.Equal(t, media2.ID, results.Results[0].Id)
	})

	T.Run("empty IDs list", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{},
		})
		require.Error(t, err)
		assert.Nil(t, results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{Ids: []string{"id1", "id2"}})
		assert.Error(t, err)
	})
}

func TestUploadedMedia_ListingForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		createdUploadedMedia := []*registry.Object{}
		for range exampleQuantity {
			createdUploadedMedia = append(createdUploadedMedia, createUploadedMediaForTest(t, testClient))
		}

		results, err := testClient.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdUploadedMedia))
	})

	T.Run("cannot access other user's media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user1, testClient1 := createUserAndClientForTest(t)
		createUploadedMediaForTest(t, testClient1)

		_, testClient2 := createUserAndClientForTest(t)

		// User 2 tries to list user 1's media
		results, err := testClient2.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: user1.ID,
		})
		require.Error(t, err)
		assert.Nil(t, results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{})
		assert.Error(t, err)
	})
}

func TestUploadedMedia_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient)

		archived, err := testClient.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: createdUploadedMedia.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, archived)

		// Verify it's been archived (should not be retrievable)
		retrieved, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		require.Error(t, err)
		assert.Nil(t, retrieved)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "uploaded_media", RelevantID: createdUploadedMedia.ID},
		})
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		archived, err := testClient.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, archived)
	})

	T.Run("cannot archive other user's media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create media as user 1
		_, testClient1 := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient1)

		// Try to archive as user 2
		_, testClient2 := createUserAndClientForTest(t)

		archived, err := testClient2.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: createdUploadedMedia.ID,
		})
		require.Error(t, err)
		assert.Nil(t, archived)

		// Verify it's still accessible to user 1
		retrieved, err := testClient1.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{})
		assert.Error(t, err)
	})
}

const uploadedMediaUploadChunkSize = 32 * 1024

func TestUploadedMedia_Upload(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		fileData := []byte("fake image data for integration test")
		filename := "test-image.jpg"
		contentType := uploadedmedia.MimeTypeImageJPEG

		stream, err := testClient.Upload(ctx)
		require.NoError(t, err)

		// First message: metadata
		err = stream.Send(&uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: &uploadedmediasvc.UploadMetadata{
					ObjectName:  filename,
					ContentType: contentType,
				},
			},
		})
		require.NoError(t, err)

		// Stream chunks
		for offset := 0; offset < len(fileData); offset += uploadedMediaUploadChunkSize {
			end := min(offset+uploadedMediaUploadChunkSize, len(fileData))
			chunk := fileData[offset:end]
			err = stream.Send(&uploadedmediasvc.UploadRequest{
				Payload: &uploadedmediasvc.UploadRequest_Chunk{Chunk: chunk},
			})
			require.NoError(t, err)
		}

		resp, err := stream.CloseAndRecv()
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.ObjectUrl)
		assert.Equal(t, int64(len(fileData)), resp.SizeBytes)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		stream, err := c.Upload(ctx)
		require.NoError(t, err)

		err = stream.Send(&uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: &uploadedmediasvc.UploadMetadata{
					ObjectName:  "test.jpg",
					ContentType: uploadedmedia.MimeTypeImageJPEG,
				},
			},
		})
		requireStreamSend(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})

	T.Run("missing metadata", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		stream, err := testClient.Upload(ctx)
		require.NoError(t, err)

		// Send chunk without metadata first
		err = stream.Send(&uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{Chunk: []byte("data")},
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})

	T.Run("missing object name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		stream, err := testClient.Upload(ctx)
		require.NoError(t, err)

		err = stream.Send(&uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: &uploadedmediasvc.UploadMetadata{
					ObjectName:  "",
					ContentType: uploadedmedia.MimeTypeImageJPEG,
				},
			},
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})

	T.Run("missing content type", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		stream, err := testClient.Upload(ctx)
		require.NoError(t, err)

		err = stream.Send(&uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: &uploadedmediasvc.UploadMetadata{
					ObjectName:  "test.jpg",
					ContentType: "",
				},
			},
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})

	T.Run("no file data", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		stream, err := testClient.Upload(ctx)
		require.NoError(t, err)

		// Send metadata only, no chunks
		err = stream.Send(&uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: &uploadedmediasvc.UploadMetadata{
					ObjectName:  "test.jpg",
					ContentType: uploadedmedia.MimeTypeImageJPEG,
				},
			},
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})
}

func TestUploadedMedia_Pagination(T *testing.T) {
	T.Parallel()

	T.Run("respects limit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		// Create more media than the limit
		for range 10 {
			createUploadedMediaForTest(t, testClient)
		}

		results, err := testClient.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: user.ID,
			Filter: &filteringpb.QueryFilter{
				MaxResponseSize: new(uint32(5)),
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.LessOrEqual(t, len(results.Results), 5)
	})
}
