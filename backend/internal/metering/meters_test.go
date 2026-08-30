package metering

import (
	"testing"

	platformmetering "github.com/primandproper/platform-go/v13/metering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		registry, err := NewRegistry()
		require.NoError(t, err)
		require.NotNil(t, registry)

		assert.Equal(t, []string{UploadedMediaBytesMeter}, registry.MeterNames())
	})

	T.Run("every meter carries a quota", func(t *testing.T) {
		t.Parallel()

		// Enforcer.Check errors on a meter with no quota rather than reading the absence as
		// permission, so a meter registered without one is a meter nothing can ever check.
		registry, err := NewRegistry()
		require.NoError(t, err)

		assert.Equal(t, registry.MeterNames(), registry.QuotaMeters())
	})

	T.Run("every quota is unlimited", func(t *testing.T) {
		t.Parallel()

		registry, err := NewRegistry()
		require.NoError(t, err)

		for _, name := range registry.MeterNames() {
			quota, ok := registry.Quota(name)
			require.True(t, ok, "meter %q has no quota", name)

			assert.Equal(t, platformmetering.BehaviorAllowOverage, quota.Behavior, "meter %q", name)
			assert.Equal(t, unlimited, quota.Limit, "meter %q", name)
		}
	})

	T.Run("every meter is bucketed by the calendar month", func(t *testing.T) {
		t.Parallel()

		// The recorder and the enforcer are wired without a period resolver, which leaves
		// them on the calendar resolver — and that one refuses PeriodBillingPeriod outright.
		// A meter declared with it here would fail at startup rather than at registration.
		registry, err := NewRegistry()
		require.NoError(t, err)

		for _, name := range registry.MeterNames() {
			meter, ok := registry.Meter(name)
			require.True(t, ok)

			assert.Equal(t, platformmetering.PeriodMonth, meter.Period, "meter %q", name)
		}
	})

	T.Run("meters and quotas agree on period", func(t *testing.T) {
		t.Parallel()

		// RegisterQuota already refuses a mismatch, so NewRegistry returning no error is
		// the assertion. This says so out loud because the failure it guards against —
		// a quota over a window the meter does not bucket by — is a table scan on the
		// read path rather than a visible error.
		registry, err := NewRegistry()
		require.NoError(t, err)

		for _, name := range registry.QuotaMeters() {
			meter, _ := registry.Meter(name)
			quota, _ := registry.Quota(name)

			assert.Equal(t, meter.Period, quota.Period, "meter %q", name)
		}
	})
}
