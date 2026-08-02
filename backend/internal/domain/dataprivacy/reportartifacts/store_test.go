package reportartifacts

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v9/cryptography/encryption"
	"github.com/primandproper/platform-go/v9/cryptography/encryption/aes"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	"github.com/primandproper/platform-go/v9/uploads"
	uploadsmock "github.com/primandproper/platform-go/v9/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildStoreForTest(t *testing.T, uploadManager uploads.UploadManager) (Store, encryption.EncryptorDecryptor) {
	t.Helper()

	encDec, err := aes.NewEncryptorDecryptor([]byte(testutils.Example32ByteKey))
	require.NoError(t, err)

	return NewStore(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), uploadManager, encDec), encDec
}

func TestStore_Save(T *testing.T) {
	T.Parallel()

	T.Run("writes ciphertext under the encrypted extension", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		reportID := identifiers.New()
		report := []byte(`{"identity":{"user":{"id":"whoever"}}}`)

		var savedPath string
		var savedBytes []byte
		uploadManager := &uploadsmock.UploadManagerMock{
			SaveFunc: func(_ context.Context, path string, r io.Reader, _ ...uploads.SaveOption) error {
				savedPath = path
				var err error
				savedBytes, err = io.ReadAll(r)
				return err
			},
		}

		s, encDec := buildStoreForTest(t, uploadManager)

		require.NoError(t, s.Save(ctx, reportID, report))
		assert.Equal(t, reportID+artifactExtension, savedPath)

		// The whole point: what lands in the bucket must not be the plaintext.
		assert.NotEqual(t, report, savedBytes)
		assert.False(t, bytes.Contains(savedBytes, []byte("whoever")))

		plaintext, err := encDec.Decrypt(ctx, string(savedBytes))
		require.NoError(t, err)
		assert.Equal(t, string(report), plaintext)
	})

	T.Run("with empty report ID", func(t *testing.T) {
		t.Parallel()

		s, _ := buildStoreForTest(t, &uploadsmock.UploadManagerMock{})

		err := s.Save(t.Context(), "", []byte("{}"))
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with error saving", func(t *testing.T) {
		t.Parallel()

		expectedError := platformerrors.New("blah")
		uploadManager := &uploadsmock.UploadManagerMock{
			SaveFunc: func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error {
				return expectedError
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		assert.Error(t, s.Save(t.Context(), identifiers.New(), []byte("{}")))
	})
}

func TestStore_Open(T *testing.T) {
	T.Parallel()

	T.Run("round trips what Save wrote", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		reportID := identifiers.New()
		report := []byte(`{"identity":{"user":{"id":"whoever"}}}`)

		bucket := map[string][]byte{}
		uploadManager := &uploadsmock.UploadManagerMock{
			SaveFunc: func(_ context.Context, path string, r io.Reader, _ ...uploads.SaveOption) error {
				b, err := io.ReadAll(r)
				bucket[path] = b
				return err
			},
			OpenFunc: func(_ context.Context, path string) (io.ReadCloser, error) {
				b, ok := bucket[path]
				if !ok {
					return nil, platformerrors.New("not found")
				}
				return io.NopCloser(bytes.NewReader(b)), nil
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		require.NoError(t, s.Save(ctx, reportID, report))

		actual, err := s.Open(ctx, reportID)
		require.NoError(t, err)
		assert.Equal(t, report, actual)
	})

	T.Run("with empty report ID", func(t *testing.T) {
		t.Parallel()

		s, _ := buildStoreForTest(t, &uploadsmock.UploadManagerMock{})

		actual, err := s.Open(t.Context(), "")
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with error reading", func(t *testing.T) {
		t.Parallel()

		uploadManager := &uploadsmock.UploadManagerMock{
			OpenFunc: func(_ context.Context, _ string) (io.ReadCloser, error) {
				return nil, platformerrors.New("blah")
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		actual, err := s.Open(t.Context(), identifiers.New())
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("refuses an artifact it cannot decrypt", func(t *testing.T) {
		t.Parallel()

		// A plaintext object sitting where ciphertext belongs must not be handed back as if
		// it were a report.
		uploadManager := &uploadsmock.UploadManagerMock{
			OpenFunc: func(_ context.Context, _ string) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte(`{"identity":{}}`))), nil
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		actual, err := s.Open(t.Context(), identifiers.New())
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestStore_Delete(T *testing.T) {
	T.Parallel()

	T.Run("reaps both the encrypted and the legacy plaintext object", func(t *testing.T) {
		t.Parallel()

		reportID := identifiers.New()
		present := map[string]bool{
			reportID + artifactExtension:       true,
			reportID + legacyArtifactExtension: true,
		}

		var deleted []string
		uploadManager := &uploadsmock.UploadManagerMock{
			ExistsFunc: func(_ context.Context, path string) (bool, error) { return present[path], nil },
			DeleteFunc: func(_ context.Context, path string) error {
				deleted = append(deleted, path)
				return nil
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		require.NoError(t, s.Delete(t.Context(), reportID))
		assert.ElementsMatch(t, []string{reportID + artifactExtension, reportID + legacyArtifactExtension}, deleted)
	})

	T.Run("is a no-op when the artifact is already gone", func(t *testing.T) {
		t.Parallel()

		uploadManager := &uploadsmock.UploadManagerMock{
			ExistsFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
			DeleteFunc: func(_ context.Context, _ string) error {
				t.Error("Delete should not be called for an absent object")
				return nil
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		assert.NoError(t, s.Delete(t.Context(), identifiers.New()))
	})

	T.Run("with empty report ID", func(t *testing.T) {
		t.Parallel()

		// Guards against deleting the object literally named ".json.enc".
		uploadManager := &uploadsmock.UploadManagerMock{
			DeleteFunc: func(_ context.Context, _ string) error {
				t.Error("Delete should not be called without a report ID")
				return nil
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		assert.ErrorIs(t, s.Delete(t.Context(), ""), platformerrors.ErrInvalidIDProvided)
	})

	T.Run("reports a failure on one path without skipping the other", func(t *testing.T) {
		t.Parallel()

		reportID := identifiers.New()

		var deleted []string
		uploadManager := &uploadsmock.UploadManagerMock{
			ExistsFunc: func(_ context.Context, _ string) (bool, error) { return true, nil },
			DeleteFunc: func(_ context.Context, path string) error {
				if path == reportID+artifactExtension {
					return platformerrors.New("blah")
				}
				deleted = append(deleted, path)
				return nil
			},
		}

		s, _ := buildStoreForTest(t, uploadManager)

		assert.Error(t, s.Delete(t.Context(), reportID))
		assert.Equal(t, []string{reportID + legacyArtifactExtension}, deleted)
	})
}
