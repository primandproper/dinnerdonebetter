package reportartifacts

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	encryptioncfg "github.com/primandproper/platform-go/v9/cryptography/encryption/config"
	"github.com/primandproper/platform-go/v9/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	"github.com/primandproper/platform-go/v9/uploads/objectstorage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func storageConfigForTest(t *testing.T) (cfg *objectstorage.Config, root string) {
	t.Helper()

	root = t.TempDir()

	return &objectstorage.Config{
		Provider:         objectstorage.FilesystemProvider,
		BucketName:       "userdata",
		FilesystemConfig: &objectstorage.FilesystemConfig{RootDirectory: root},
	}, root
}

// findArtifactFile locates the object the filesystem provider wrote, without this test having to
// know how that provider lays out a bucket on disk.
func findArtifactFile(t *testing.T, root, name string) string {
	t.Helper()

	var found string
	require.NoError(t, filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == name {
			found = path
		}
		return nil
	}))
	require.NotEmpty(t, found, "no artifact named %s was written under %s", name, root)

	return found
}

func TestProvideStore(T *testing.T) {
	T.Parallel()

	T.Run("round trips an artifact through real object storage", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		storage, root := storageConfigForTest(t)

		s, err := ProvideStore(
			ctx,
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
			storage,
			&encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
			testutils.Example32ByteKey,
		)
		require.NoError(t, err)

		reportID := identifiers.New()
		report := []byte(`{"identity":{"user":{"id":"whoever"}}}`)

		require.NoError(t, s.Save(ctx, reportID, report))

		// What actually landed on disk must not be readable as the report.
		artifactFile := findArtifactFile(t, root, reportID+artifactExtension)
		onDisk, err := os.ReadFile(artifactFile)
		require.NoError(t, err)
		assert.NotContains(t, string(onDisk), "whoever")

		actual, err := s.Open(ctx, reportID)
		require.NoError(t, err)
		assert.Equal(t, report, actual)

		require.NoError(t, s.Delete(ctx, reportID))

		_, err = os.Stat(artifactFile)
		assert.True(t, os.IsNotExist(err), "the artifact must be gone from the bucket")

		// Deleting again is not an error — the reaper guarantees absence, not a deletion event.
		assert.NoError(t, s.Delete(ctx, reportID))
	})

	T.Run("with no encryption key", func(t *testing.T) {
		t.Parallel()

		storage, _ := storageConfigForTest(t)

		// Refusing here is what keeps a misconfigured process from writing plaintext, or from
		// starting up and failing only once a subject asks for their data.
		s, err := ProvideStore(
			t.Context(),
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
			storage,
			&encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
			"",
		)
		assert.Error(t, err)
		assert.Nil(t, s)
	})

	T.Run("with an invalid storage config", func(t *testing.T) {
		t.Parallel()

		s, err := ProvideStore(
			t.Context(),
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
			&objectstorage.Config{},
			&encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
			testutils.Example32ByteKey,
		)
		assert.Error(t, err)
		assert.Nil(t, s)
	})
}
