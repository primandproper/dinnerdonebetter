package entitlements

import (
	"testing"

	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	platformentitlements "github.com/primandproper/platform-go/v13/entitlements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatures(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		features := Features()
		require.Len(t, features, 1)

		assert.Equal(t, UploadedMediaBytesFeature, features[0].Key)
		assert.Equal(t, platformentitlements.KindQuota, features[0].Kind)
		assert.Equal(t, appmetering.UploadedMediaBytesMeter, features[0].Meter)
	})

	T.Run("every feature registers", func(t *testing.T) {
		t.Parallel()

		// RegisterFeature is where a key that is not a plain identifier, a kind that is not
		// one of the two, and a quota feature naming no meter are all caught. None of those
		// is reachable from configuration, so this is the only place they can be caught at
		// all.
		catalog := platformentitlements.NewCatalog()

		features := Features()
		for idx := range features {
			require.NoError(t, catalog.RegisterFeature(features[idx]), "feature %q", features[idx].Key)
		}

		assert.Equal(t, []string{UploadedMediaBytesFeature}, catalog.FeatureKeys())
	})

	T.Run("every quota feature names a registered meter", func(t *testing.T) {
		t.Parallel()

		// A quota feature whose meter is unregistered presents as a Check that errors for
		// one feature and works for the rest. NewQuotaSource runs this same check at
		// startup; running it here means the failure is a test rather than a boot.
		registry, err := appmetering.NewRegistry()
		require.NoError(t, err)

		assert.NoError(t, shippedCatalog(t).ValidateMeters(registry))
	})
}
