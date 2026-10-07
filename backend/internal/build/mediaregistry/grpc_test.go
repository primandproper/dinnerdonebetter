package mediaregistry

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/platform-go/v15/mediaregistry"
	mediaregistrygrpc "github.com/primandproper/platform-go/v15/mediaregistry/grpc"
	mediaregistryhttp "github.com/primandproper/platform-go/v15/mediaregistry/http"
	"github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb"
	registrymock "github.com/primandproper/platform-go/v15/mediaregistry/mock"
	"github.com/primandproper/platform-go/v15/metering"
	meteringmock "github.com/primandproper/platform-go/v15/metering/mock"
	"github.com/primandproper/primitives-go/v2/database"
	databasemock "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"
	uploadsmock "github.com/primandproper/primitives-go/v2/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func sessionFor(t *testing.T, userID, accountID string) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester:       sessions.RequesterInfo{UserID: userID},
		ActiveAccountID: accountID,
	})
}

func TestCaller(T *testing.T) {
	T.Parallel()

	T.Run("is the signed-in user, in the global directory", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()

		actual, err := Caller(sessionFor(t, userID, fake.BuildFakeID()))
		require.NoError(t, err)

		assert.Equal(t, userID, actual.PrincipalID)
		assert.True(t, actual.Scope.IsGlobal())
	})

	T.Run("refuses a request with no session", func(t *testing.T) {
		t.Parallel()

		_, err := Caller(t.Context())
		assert.ErrorIs(t, err, callers.ErrNoPrincipal)
	})
}

func TestObjectKey(T *testing.T) {
	T.Parallel()

	T.Run("lays an upload out under its uploader", func(t *testing.T) {
		t.Parallel()

		userID, objectID, name := fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID()+".png"
		caller := mediaregistryhttp.Caller{PrincipalID: userID, Scope: tenancy.Global()}

		assert.Equal(t, userID+"/"+objectID+"/"+name, ObjectKey(caller, objectID, name))
	})

	T.Run("escapes a principal into one segment", func(t *testing.T) {
		t.Parallel()

		first, second, objectID, name := fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID()+".png"
		caller := mediaregistryhttp.Caller{PrincipalID: first + "/" + second, Scope: tenancy.Global()}

		assert.Equal(t, first+"%2F"+second+"/"+objectID+"/"+name, ObjectKey(caller, objectID, name))
	})
}

// The record policy the server is built with is platform's default over ObjectKey, so the
// part of the bucket a caller may register keys under is the part ObjectKey lays their
// uploads out in, and nothing else.
func TestObjectKey_RecordPolicy(T *testing.T) {
	T.Parallel()

	policy := mediaregistrygrpc.KeysUnderPrefix(ObjectKey)

	T.Run("admits a key under the caller's own prefix", func(t *testing.T) {
		t.Parallel()

		caller := mediaregistryhttp.Caller{PrincipalID: fake.BuildFakeID(), Scope: tenancy.Global()}

		admitted, err := policy(t.Context(), caller, ObjectKey(caller, fake.BuildFakeID(), fake.BuildFakeID()+".png"))
		require.NoError(t, err)
		assert.True(t, admitted)
	})

	T.Run("refuses a key under somebody else's prefix", func(t *testing.T) {
		t.Parallel()

		caller := mediaregistryhttp.Caller{PrincipalID: fake.BuildFakeID(), Scope: tenancy.Global()}
		victim := mediaregistryhttp.Caller{PrincipalID: fake.BuildFakeID(), Scope: tenancy.Global()}

		admitted, err := policy(t.Context(), caller, ObjectKey(victim, fake.BuildFakeID(), fake.BuildFakeID()+".png"))
		require.NoError(t, err)
		assert.False(t, admitted)
	})

	T.Run("refuses a key that walks out of the caller's prefix", func(t *testing.T) {
		t.Parallel()

		callerID, victimID := fake.BuildFakeID(), fake.BuildFakeID()
		caller := mediaregistryhttp.Caller{PrincipalID: callerID, Scope: tenancy.Global()}

		admitted, err := policy(t.Context(), caller, callerID+"/../"+victimID+"/"+fake.BuildFakeID()+".png")
		require.NoError(t, err)
		assert.False(t, admitted)
	})
}

func TestUploadMeter(T *testing.T) {
	T.Parallel()

	T.Run("counts the upload's bytes against the caller's active account", func(t *testing.T) {
		t.Parallel()

		accountID := fake.BuildFakeID()
		object := &mediaregistry.Object{ID: fake.BuildFakeID(), Size: 1234, ContentType: uploadedmedia.MimeTypeImagePNG}

		db := &databasemock.ClientMock{
			WithTransactionFunc: func(ctx context.Context, fn func(database.Tx) error) error {
				return fn(nil)
			},
		}

		var recorded []metering.Usage
		recorder := &meteringmock.RecorderMock{
			RecordFunc: func(_ context.Context, _ database.Tx, scope tenancy.Scope, u ...metering.Usage) error {
				assert.True(t, scope.IsGlobal())
				recorded = append(recorded, u...)

				return nil
			},
		}

		caller := mediaregistryhttp.Caller{PrincipalID: fake.BuildFakeID(), Scope: tenancy.Global()}
		require.NoError(t, uploadMeter(db, recorder)(sessionFor(t, caller.PrincipalID, accountID), caller, object))

		require.Len(t, recorded, 1)
		assert.Equal(t, accountID, recorded[0].Subject)
		assert.Equal(t, appmetering.UploadedMediaBytesMeter, recorded[0].Meter)
		assert.Equal(t, object.Size, recorded[0].Quantity)
		assert.Equal(t, object.ID, recorded[0].IdempotencyKey)
		assert.Equal(t, object.ContentType, recorded[0].Dimensions["mime_type"])
	})

	T.Run("hands a failure back to the surface", func(t *testing.T) {
		t.Parallel()

		expected := errors.New(fake.BuildFakeID())
		db := &databasemock.ClientMock{
			WithTransactionFunc: func(context.Context, func(database.Tx) error) error {
				return expected
			},
		}

		caller := mediaregistryhttp.Caller{PrincipalID: fake.BuildFakeID(), Scope: tenancy.Global()}
		err := uploadMeter(db, &meteringmock.RecorderMock{})(sessionFor(t, caller.PrincipalID, fake.BuildFakeID()), caller, &mediaregistry.Object{})
		assert.ErrorIs(t, err, expected)
	})
}

func TestPermissions(T *testing.T) {
	T.Parallel()

	T.Run("declares every method the surface serves", func(t *testing.T) {
		t.Parallel()

		actual := Permissions()

		for _, method := range mediaregistrypb.MediaRegistryService_ServiceDesc.Methods {
			assert.Contains(t, actual, "/"+mediaregistrypb.MediaRegistryService_ServiceDesc.ServiceName+"/"+method.MethodName)
		}

		for _, stream := range mediaregistrypb.MediaRegistryService_ServiceDesc.Streams {
			assert.Contains(t, actual, "/"+mediaregistrypb.MediaRegistryService_ServiceDesc.ServiceName+"/"+stream.StreamName)
		}
	})
}

// uploadStream is the server's half of an UploadObject call, delivering msgs and then io.EOF.
type uploadStream struct {
	grpc.ServerStream

	ctx  context.Context
	msgs []*mediaregistrypb.UploadObjectRequest
}

func (s *uploadStream) Context() context.Context { return s.ctx }

func (s *uploadStream) Recv() (*mediaregistrypb.UploadObjectRequest, error) {
	if len(s.msgs) == 0 {
		return nil, io.EOF
	}

	msg := s.msgs[0]
	s.msgs = s.msgs[1:]

	return msg, nil
}

func (s *uploadStream) SendAndClose(*mediaregistrypb.UploadObjectResponse) error { return nil }

func TestNewServer_UploadObject(T *testing.T) {
	T.Parallel()

	// Every dependency is a mock with nothing set, so a refusal that reached the bucket, the
	// registry or the database would panic rather than pass.
	buildServer := func(t *testing.T) *mediaregistrygrpc.Server {
		t.Helper()

		server, err := NewServer(
			&registrymock.StoreMock{},
			&databasemock.ClientMock{},
			&uploadsmock.UploadManagerMock{},
			&meteringmock.RecorderMock{},
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			nil,
		)
		require.NoError(t, err)

		return server
	}

	upload := func(t *testing.T, name, contentType string) error {
		t.Helper()

		return buildServer(t).UploadObject(&uploadStream{
			ctx: sessionFor(t, fake.BuildFakeID(), fake.BuildFakeID()),
			msgs: []*mediaregistrypb.UploadObjectRequest{
				{Part: &mediaregistrypb.UploadObjectRequest_Header{Header: &mediaregistrypb.UploadObjectHeader{
					Name:        name,
					ContentType: contentType,
				}}},
				{Part: &mediaregistrypb.UploadObjectRequest_Chunk{Chunk: []byte(fake.BuildFakeID())}},
			},
		})
	}

	T.Run("refuses a name that walks out of the caller's prefix", func(t *testing.T) {
		t.Parallel()

		err := upload(t, "../../"+fake.BuildFakeID()+"/"+fake.BuildFakeID()+".png", uploadedmedia.MimeTypeImagePNG)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("refuses a name carrying a separator", func(t *testing.T) {
		t.Parallel()

		err := upload(t, fake.BuildFakeID()+"/"+fake.BuildFakeID()+".png", uploadedmedia.MimeTypeImagePNG)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("refuses a name that is a directory", func(t *testing.T) {
		t.Parallel()

		err := upload(t, "..", uploadedmedia.MimeTypeImagePNG)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("refuses a content type this application does not store", func(t *testing.T) {
		t.Parallel()

		err := upload(t, fake.BuildFakeID()+".svg", "image/svg+xml")
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("refuses an attachment to somebody else as an absence", func(t *testing.T) {
		t.Parallel()

		err := buildServer(t).UploadObject(&uploadStream{
			ctx: sessionFor(t, fake.BuildFakeID(), fake.BuildFakeID()),
			msgs: []*mediaregistrypb.UploadObjectRequest{
				{Part: &mediaregistrypb.UploadObjectRequest_Header{Header: &mediaregistrypb.UploadObjectHeader{
					Name:        fake.BuildFakeID() + ".png",
					ContentType: uploadedmedia.MimeTypeImagePNG,
					BelongsTo:   &mediaregistrypb.Subject{Type: mediaregistrygrpc.UserSubjectType, Id: fake.BuildFakeID()},
				}}},
			},
		})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}
