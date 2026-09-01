package grpc

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediafakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/metering"
	meteringmock "github.com/primandproper/platform-go/v13/metering/mock"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"
	"github.com/primandproper/platform-go/v13/uploads"
	mockuploads "github.com/primandproper/platform-go/v13/uploads/mock"
	"github.com/primandproper/platform-go/v13/uploads/registry"
	registrymock "github.com/primandproper/platform-go/v13/uploads/registry/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	testAccountID = identifiers.New()
	testUserID    = identifiers.New()
)

func buildTestService(t *testing.T) (*serviceImpl, *registrymock.StoreMock, *mockuploads.UploadManagerMock) {
	t.Helper()

	service, store, uploadManager, _ := buildTestServiceWithRecorder(t)

	return service, store, uploadManager
}

// buildTestServiceWithRecorder is buildTestService with the usage recorder handed back, for the
// tests that care what got metered.
//
// The recorder accepts everything by default. Upload counts bytes after the file and its row are
// already committed, so a recorder that failed would be testing the swallow rather than the
// count — see recordUploadUsage.
func buildTestServiceWithRecorder(t *testing.T) (*serviceImpl, *registrymock.StoreMock, *mockuploads.UploadManagerMock, *meteringmock.RecorderMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	uploadsRegistry := &registrymock.StoreMock{}
	uploadManager := &mockuploads.UploadManagerMock{}
	usageRecorder := &meteringmock.RecorderMock{
		RecordFunc: func(context.Context, ...metering.Usage) error { return nil },
	}

	service := &serviceImpl{
		tracer:        tracer,
		logger:        logger,
		registry:      uploadsRegistry,
		uploadManager: uploadManager,
		usageRecorder: usageRecorder,
	}

	return service, uploadsRegistry, uploadManager, usageRecorder
}

// ownedByTestUser is a fake object whose owner is the user every session context here reports,
// which is what the handlers' ownership checks compare against.
func ownedByTestUser() *registry.Object {
	object := uploadedmediafakes.BuildFakeUploadedMedia()
	object.OwnerID = testUserID

	return object
}

// mockUploadStream is a fake upload stream. Recv yields queued messages in order and then
// returns recvErr (io.EOF for a clean end of stream), which models a stream far more directly
// than a call-expectation mock does.
type mockUploadStream struct {
	ctx             context.Context
	recvErr         error
	sendAndCloseErr error
	recvQueue       []*uploadedmediasvc.UploadRequest
	closedWith      []*uploadedmediasvc.UploadResponse
	recvIndex       int
}

func (m *mockUploadStream) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}

	return m.ctx
}

func (m *mockUploadStream) SendMsg(any) error { return nil }

func (m *mockUploadStream) RecvMsg(any) error { return nil }

func (m *mockUploadStream) Recv() (*uploadedmediasvc.UploadRequest, error) {
	if m.recvIndex < len(m.recvQueue) {
		req := m.recvQueue[m.recvIndex]
		m.recvIndex++

		return req, nil
	}

	if m.recvErr != nil {
		return nil, m.recvErr
	}

	return nil, io.EOF
}

func (m *mockUploadStream) SendAndClose(response *uploadedmediasvc.UploadResponse) error {
	m.closedWith = append(m.closedWith, response)

	return m.sendAndCloseErr
}

func (m *mockUploadStream) SendHeader(metadata.MD) error { return nil }

func (m *mockUploadStream) SetHeader(metadata.MD) error { return nil }

func (m *mockUploadStream) SetTrailer(metadata.MD) {}

func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: testAccountID,
		Requester:       sessions.RequesterInfo{UserID: testUserID},
	})
}

// uploadStreamFor builds a stream carrying one metadata message and the given chunks.
func uploadStreamFor(t *testing.T, contentType, objectName string, chunks ...[]byte) *mockUploadStream {
	t.Helper()

	stream := &mockUploadStream{
		ctx: buildSessionContextForTest(t),
		recvQueue: []*uploadedmediasvc.UploadRequest{
			{Payload: &uploadedmediasvc.UploadRequest_Metadata{Metadata: &uploadedmediasvc.UploadMetadata{
				Bucket:      "test-bucket",
				ObjectName:  objectName,
				ContentType: contentType,
			}}},
		},
	}

	for _, chunk := range chunks {
		stream.recvQueue = append(stream.recvQueue, &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{Chunk: chunk},
		})
	}

	return stream
}

func TestServiceImpl_CreateUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, uploadsRegistry, _ := buildTestService(t)

		fakeObject := ownedByTestUser()

		var recorded *registry.Object
		uploadsRegistry.RecordObjectFunc = func(_ context.Context, object *registry.Object) error {
			recorded = object

			return nil
		}

		response, err := service.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{
				ObjectKey:   fakeObject.Key,
				ContentType: fakeObject.ContentType,
				SizeBytes:   fakeObject.Size,
			},
		})

		require.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeObject.Key, response.Created.ObjectKey)
		assert.Equal(t, fakeObject.ContentType, response.Created.ContentType)

		// The owner is the session's user rather than anything the request carried: it is
		// the whole of what a later read's permission check consults.
		require.NotNil(t, recorded)
		assert.Equal(t, testUserID, recorded.OwnerID)
		assert.Equal(t, uploadedmedia.Scope(), recorded.Scope)
		assert.NotEmpty(t, recorded.ID)

		assert.Len(t, uploadsRegistry.RecordObjectCalls(), 1)
	})

	t.Run("attached to a subject", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, uploadsRegistry, _ := buildTestService(t)

		subject := registry.Subject{Type: "recipe", ID: identifiers.New()}

		var recorded *registry.Object
		uploadsRegistry.RecordObjectFunc = func(_ context.Context, object *registry.Object) error {
			recorded = object

			return nil
		}

		_, err := service.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{
				ObjectKey:     "recipes/whatever.png",
				ContentType:   uploadedmedia.MimeTypeImagePNG,
				BelongsToType: subject.Type,
				BelongsToId:   subject.ID,
			},
		})

		require.NoError(t, err)
		require.NotNil(t, recorded)
		assert.Equal(t, subject, recorded.BelongsTo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		response, err := service.CreateUploadedMedia(t.Context(), &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{},
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("without input", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		response, err := service.CreateUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.CreateUploadedMediaRequest{})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, uploadsRegistry.RecordObjectCalls())
	})

	t.Run("with unsupported content type", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		response, err := service.CreateUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{
				ObjectKey:   "whatever.pdf",
				ContentType: "application/pdf",
			},
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, uploadsRegistry.RecordObjectCalls())
	})

	t.Run("with no object key", func(t *testing.T) {
		t.Parallel()

		// A row with no key names no object, which the registry's own validation refuses.
		service, uploadsRegistry, _ := buildTestService(t)

		response, err := service.CreateUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{
				ContentType: uploadedmedia.MimeTypeImagePNG,
			},
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, uploadsRegistry.RecordObjectCalls())
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		uploadsRegistry.RecordObjectFunc = func(context.Context, *registry.Object) error {
			return errors.New("registry error")
		}

		response, err := service.CreateUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{
				ObjectKey:   "whatever.png",
				ContentType: uploadedmedia.MimeTypeImagePNG,
			},
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Len(t, uploadsRegistry.RecordObjectCalls(), 1)
	})

	t.Run("a taken key reads as a conflict", func(t *testing.T) {
		t.Parallel()

		// The client's remedy is a new key, which is a different thing to be told than
		// "the server broke" — see internal/services/uploadedmedia/errors.
		service, uploadsRegistry, _ := buildTestService(t)

		uploadsRegistry.RecordObjectFunc = func(context.Context, *registry.Object) error {
			return registry.ErrObjectKeyTaken
		}

		_, err := service.CreateUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{
				ObjectKey:   "whatever.png",
				ContentType: uploadedmedia.MimeTypeImagePNG,
			},
		})

		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})
}

func TestServiceImpl_GetUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		fakeObject := ownedByTestUser()

		var readScope tenancy.Scope
		uploadsRegistry.GetObjectFunc = func(_ context.Context, scope tenancy.Scope, _ string) (*registry.Object, error) {
			readScope = scope

			return fakeObject, nil
		}

		response, err := service.GetUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: fakeObject.ID,
		})

		require.NoError(t, err)
		require.NotNil(t, response.Result)
		assert.Equal(t, fakeObject.ID, response.Result.Id)
		assert.Equal(t, fakeObject.Key, response.Result.ObjectKey)
		assert.Equal(t, uploadedmedia.Scope(), readScope)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		response, err := service.GetUploadedMedia(t.Context(), &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: identifiers.New(),
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("belonging to another user", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		fakeObject := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeObject.OwnerID = identifiers.New()

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return fakeObject, nil
		}

		response, err := service.GetUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: fakeObject.ID,
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("an absent object reads as not found", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return nil, registry.ErrObjectNotFound
		}

		response, err := service.GetUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: identifiers.New(),
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return nil, errors.New("registry error")
		}

		response, err := service.GetUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: identifiers.New(),
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestServiceImpl_GetUploadedMediaWithIDs(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		first, second := ownedByTestUser(), ownedByTestUser()
		byID := map[string]*registry.Object{first.ID: first, second.ID: second}

		uploadsRegistry.GetObjectFunc = func(_ context.Context, _ tenancy.Scope, objectID string) (*registry.Object, error) {
			object, ok := byID[objectID]
			if !ok {
				return nil, registry.ErrObjectNotFound
			}

			return object, nil
		}

		response, err := service.GetUploadedMediaWithIDs(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{first.ID, second.ID},
		})

		require.NoError(t, err)
		require.Len(t, response.Results, 2)
		assert.Equal(t, first.ID, response.Results[0].Id)
		assert.Equal(t, second.ID, response.Results[1].Id)
	})

	t.Run("another user's media is left out", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		mine := ownedByTestUser()
		theirs := uploadedmediafakes.BuildFakeUploadedMedia()
		theirs.OwnerID = identifiers.New()
		byID := map[string]*registry.Object{mine.ID: mine, theirs.ID: theirs}

		uploadsRegistry.GetObjectFunc = func(_ context.Context, _ tenancy.Scope, objectID string) (*registry.Object, error) {
			return byID[objectID], nil
		}

		response, err := service.GetUploadedMediaWithIDs(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{mine.ID, theirs.ID},
		})

		require.NoError(t, err)
		require.Len(t, response.Results, 1)
		assert.Equal(t, mine.ID, response.Results[0].Id)
	})

	t.Run("an absent id is skipped rather than failing the read", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		mine := ownedByTestUser()

		uploadsRegistry.GetObjectFunc = func(_ context.Context, _ tenancy.Scope, objectID string) (*registry.Object, error) {
			if objectID == mine.ID {
				return mine, nil
			}

			return nil, registry.ErrObjectNotFound
		}

		response, err := service.GetUploadedMediaWithIDs(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{identifiers.New(), mine.ID},
		})

		require.NoError(t, err)
		require.Len(t, response.Results, 1)
		assert.Equal(t, mine.ID, response.Results[0].Id)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		response, err := service.GetUploadedMediaWithIDs(t.Context(), &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{identifiers.New()},
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("without IDs", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		response, err := service.GetUploadedMediaWithIDs(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaWithIDsRequest{})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Empty(t, uploadsRegistry.GetObjectCalls())
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return nil, errors.New("registry error")
		}

		response, err := service.GetUploadedMediaWithIDs(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{identifiers.New()},
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestServiceImpl_GetUploadedMediaForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		page := uploadedmediafakes.BuildFakeUploadedMediaList()

		var listedOwner string
		uploadsRegistry.ListObjectsByOwnerFunc = func(_ context.Context, _ tenancy.Scope, ownerID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[registry.Object], error) {
			listedOwner = ownerID

			return page, nil
		}

		response, err := service.GetUploadedMediaForUser(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: testUserID,
			Filter: &filteringpb.QueryFilter{},
		})

		require.NoError(t, err)
		assert.Len(t, response.Results, len(page.Data))
		assert.NotNil(t, response.Pagination)
		assert.Equal(t, testUserID, listedOwner)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		response, err := service.GetUploadedMediaForUser(t.Context(), &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: testUserID,
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("asking for another user's media", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		response, err := service.GetUploadedMediaForUser(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: identifiers.New(),
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.Empty(t, uploadsRegistry.ListObjectsByOwnerCalls())
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		uploadsRegistry.ListObjectsByOwnerFunc = func(context.Context, tenancy.Scope, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[registry.Object], error) {
			return nil, errors.New("registry error")
		}

		response, err := service.GetUploadedMediaForUser(buildSessionContextForTest(t), &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: testUserID,
			Filter: &filteringpb.QueryFilter{},
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestServiceImpl_ArchiveUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		fakeObject := ownedByTestUser()

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return fakeObject, nil
		}
		uploadsRegistry.ArchiveObjectFunc = func(context.Context, tenancy.Scope, string) error { return nil }

		response, err := service.ArchiveUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeObject.ID,
		})

		require.NoError(t, err)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, uploadsRegistry.ArchiveObjectCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		response, err := service.ArchiveUploadedMedia(t.Context(), &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: identifiers.New(),
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("belonging to another user", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		fakeObject := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeObject.OwnerID = identifiers.New()

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return fakeObject, nil
		}

		response, err := service.ArchiveUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeObject.ID,
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.Empty(t, uploadsRegistry.ArchiveObjectCalls())
	})

	t.Run("an absent object reads as not found", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return nil, registry.ErrObjectNotFound
		}

		response, err := service.ArchiveUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: identifiers.New(),
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, _ := buildTestService(t)

		fakeObject := ownedByTestUser()

		uploadsRegistry.GetObjectFunc = func(context.Context, tenancy.Scope, string) (*registry.Object, error) {
			return fakeObject, nil
		}
		uploadsRegistry.ArchiveObjectFunc = func(context.Context, tenancy.Scope, string) error {
			return errors.New("registry error")
		}

		response, err := service.ArchiveUploadedMedia(buildSessionContextForTest(t), &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeObject.ID,
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestServiceImpl_Upload(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, mockUploadMgr, usageRecorder := buildTestServiceWithRecorder(t)

		chunk1 := []byte("test file content part 1")
		chunk2 := []byte("test file content part 2")
		stream := uploadStreamFor(t, uploadedmedia.MimeTypeImagePNG, "test-file.png", chunk1, chunk2)

		var savedKey string
		mockUploadMgr.SaveFunc = func(_ context.Context, key string, r io.Reader, _ ...uploads.SaveOption) error {
			savedKey = key
			_, err := io.Copy(io.Discard, r)

			return err
		}

		var recorded *registry.Object
		uploadsRegistry.RecordObjectFunc = func(_ context.Context, object *registry.Object) error {
			recorded = object

			return nil
		}

		require.NoError(t, service.Upload(stream))
		assert.Len(t, uploadsRegistry.RecordObjectCalls(), 1)

		// The row's key is where the bytes went, and its size is what actually went past
		// rather than what the chunks claimed.
		require.NotNil(t, recorded)
		assert.Equal(t, savedKey, recorded.Key)
		assert.Equal(t, testUserID, recorded.OwnerID)
		assert.Equal(t, uploadedmedia.Scope(), recorded.Scope)
		assert.Equal(t, int64(len(chunk1)+len(chunk2)), recorded.Size)

		require.Len(t, stream.closedWith, 1)
		assert.Equal(t, recorded.Key, stream.closedWith[0].ObjectUrl)
		assert.Equal(t, recorded.Size, stream.closedWith[0].SizeBytes)

		// The bytes are counted against the account, keyed by the row that was created —
		// not by anything request-scoped, because a retried upload stores a second object
		// and is genuinely a second charge.
		require.Len(t, usageRecorder.RecordCalls(), 1)
		require.Len(t, usageRecorder.RecordCalls()[0].U, 1)

		usage := usageRecorder.RecordCalls()[0].U[0]
		assert.Equal(t, appmetering.UploadedMediaBytesMeter, usage.Meter)
		assert.Equal(t, testAccountID, usage.Subject)
		assert.Equal(t, recorded.ID, usage.IdempotencyKey)
		assert.Equal(t, int64(len(chunk1)+len(chunk2)), usage.Quantity)
		assert.Equal(t, uploadedmedia.MimeTypeImagePNG, usage.Dimensions["mime_type"])
	})

	t.Run("a failed usage record does not fail the upload", func(t *testing.T) {
		t.Parallel()

		// The file is in the bucket and its row is in the registry before the meter is
		// touched, so failing here would tell the client an upload did not happen that did.
		service, uploadsRegistry, mockUploadMgr, usageRecorder := buildTestServiceWithRecorder(t)

		stream := uploadStreamFor(t, uploadedmedia.MimeTypeImagePNG, "test-file.png", []byte("test file content"))

		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, r io.Reader, _ ...uploads.SaveOption) error {
			_, err := io.Copy(io.Discard, r)

			return err
		}
		uploadsRegistry.RecordObjectFunc = func(context.Context, *registry.Object) error { return nil }
		usageRecorder.RecordFunc = func(context.Context, ...metering.Usage) error {
			return platformerrors.New("blah")
		}

		require.NoError(t, service.Upload(stream))
		assert.Len(t, usageRecorder.RecordCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		err := service.Upload(&mockUploadStream{ctx: t.Context()})

		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("missing metadata in first message", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		stream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
			recvQueue: []*uploadedmediasvc.UploadRequest{
				{Payload: &uploadedmediasvc.UploadRequest_Chunk{Chunk: []byte("some data")}},
			},
		}

		err := service.Upload(stream)

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing object_name", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		err := service.Upload(uploadStreamFor(t, uploadedmedia.MimeTypeImagePNG, ""))

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing content_type", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		err := service.Upload(uploadStreamFor(t, "", "test-file.png"))

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("unsupported MIME type", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		err := service.Upload(uploadStreamFor(t, "application/pdf", "test-file.pdf"))

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("file too large", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		err := service.Upload(uploadStreamFor(t, uploadedmedia.MimeTypeImagePNG, "large-file.png", make([]byte, maxUploadSize+1)))

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("no file data", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		err := service.Upload(uploadStreamFor(t, uploadedmedia.MimeTypeImagePNG, "empty-file.png"))

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("upload manager error", func(t *testing.T) {
		t.Parallel()

		// The bytes go first, so a storage failure means nothing was registered — the row
		// that would have promised them is never written.
		service, uploadsRegistry, mockUploadMgr := buildTestService(t)

		mockUploadMgr.SaveFunc = func(context.Context, string, io.Reader, ...uploads.SaveOption) error {
			return errors.New("storage error")
		}

		err := service.Upload(uploadStreamFor(t, uploadedmedia.MimeTypeImagePNG, "test-file.png", []byte("test file content")))

		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Empty(t, uploadsRegistry.RecordObjectCalls())
	})

	t.Run("registry error", func(t *testing.T) {
		t.Parallel()

		service, uploadsRegistry, mockUploadMgr := buildTestService(t)

		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, r io.Reader, _ ...uploads.SaveOption) error {
			_, err := io.Copy(io.Discard, r)

			return err
		}
		uploadsRegistry.RecordObjectFunc = func(context.Context, *registry.Object) error {
			return errors.New("database error")
		}

		err := service.Upload(uploadStreamFor(t, uploadedmedia.MimeTypeImagePNG, "test-file.png", []byte("test file content")))

		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Len(t, uploadsRegistry.RecordObjectCalls(), 1)
	})
}
