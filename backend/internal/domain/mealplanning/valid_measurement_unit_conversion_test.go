package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// conversionEpsilon is the relative tolerance for asserting on converted measurement
// quantities. A conversion is a float32 multiply or divide by an arbitrary modifier
// followed by a rounding pass, so the result can land a few ULPs off an expectation
// that happens to be exactly representable today. This is looser than the tolerance
// the grocery list tests use, because a conversion divides by a modifier that is
// itself inexact, whereas aggregation only adds and scales. It is still far tighter
// than the rounding granularity of any of these conversions, so a wrong modifier or
// a wrong direction still fails.
const conversionEpsilon = 1e-5

func TestValidMeasurementUnitConversion_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			OnlyForIngredient: &ValidIngredient{},
		}
		input := &ValidMeasurementUnitConversionUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestValidMeasurementUnitConversionCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversionCreationRequestInput{
			From:     "from",
			To:       "to",
			Modifier: 1.0,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversionCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidMeasurementUnitConversionDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversionDatabaseCreationInput{
			ID:       t.Name(),
			From:     "from",
			To:       "to",
			Modifier: 1.0,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversionDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidMeasurementUnitConversionUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversionUpdateRequestInput{
			From:     new("from"),
			To:       new("to"),
			Modifier: new(float32(1.0)),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversionUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidMeasurementUnitConversion_ConvertFromToTo(T *testing.T) {
	T.Parallel()

	T.Run("standard conversion with cups to milliliters", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 2 cups to milliliters (default 2 decimal places)
		result := x.ConvertFromToTo(2.0)
		expected := float32(473.18)

		assert.InDelta(t, expected, result, 0.01)
	})

	T.Run("conversion with custom precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 2 cups to milliliters with 3 decimal places
		result := x.ConvertFromToTo(2.0, 3)
		expected := float32(473.176)

		assert.InDelta(t, expected, result, 0.001)
	})

	T.Run("conversion with zero precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 2 cups to milliliters with 0 decimal places (whole number)
		result := x.ConvertFromToTo(2.0, 0)
		expected := float32(473.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with high precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 2 cups to milliliters with 5 decimal places
		result := x.ConvertFromToTo(2.0, 5)
		expected := float32(473.176)

		assert.InDelta(t, expected, result, 0.00001)
	})

	T.Run("conversion with modifier of 1", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 1.0,
		}

		result := x.ConvertFromToTo(10.0)
		expected := float32(10.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with zero value", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 2.5,
		}

		result := x.ConvertFromToTo(0.0)
		expected := float32(0.0)

		assert.InDelta(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with fractional modifier", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 0.5, // e.g., 1 tablespoon = 0.5 fluid ounces
		}

		result := x.ConvertFromToTo(4.0)
		expected := float32(2.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with large values", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 1000.0,
		}

		result := x.ConvertFromToTo(5.0)
		expected := float32(5000.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})
}

func TestValidMeasurementUnitConversion_ConvertToToFrom(T *testing.T) {
	T.Parallel()

	T.Run("standard conversion with milliliters to cups", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 473.176 milliliters to cups (default 2 decimal places)
		result := x.ConvertToToFrom(473.176)
		expected := float32(2.0)

		assert.InDelta(t, expected, result, 0.01)
	})

	T.Run("conversion with custom precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 500 milliliters to cups with 3 decimal places
		result := x.ConvertToToFrom(500.0, 3)
		expected := float32(2.113)

		assert.InDelta(t, expected, result, 0.001)
	})

	T.Run("conversion with zero precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 500 milliliters to cups with 0 decimal places
		result := x.ConvertToToFrom(500.0, 0)
		expected := float32(2.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with high precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		// Convert 500 milliliters to cups with 5 decimal places
		result := x.ConvertToToFrom(500.0, 5)
		expected := float32(2.11338)

		assert.InDelta(t, expected, result, 0.00001)
	})

	T.Run("conversion with modifier of 1", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 1.0,
		}

		result := x.ConvertToToFrom(10.0)
		expected := float32(10.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with zero value", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 2.5,
		}

		result := x.ConvertToToFrom(0.0)
		expected := float32(0.0)

		assert.InDelta(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with fractional modifier", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 0.5, // e.g., 1 tablespoon = 0.5 fluid ounces
		}

		result := x.ConvertToToFrom(2.0)
		expected := float32(4.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})

	T.Run("conversion with large values", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 1000.0, // e.g., 1 kilogram = 1000 grams
		}

		result := x.ConvertToToFrom(5000.0)
		expected := float32(5.0)

		assert.InEpsilon(t, expected, result, conversionEpsilon)
	})
}

func TestValidMeasurementUnitConversion_RoundTripConversion(T *testing.T) {
	T.Parallel()

	T.Run("round trip conversion maintains value with default precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		original := float32(5.0)
		converted := x.ConvertFromToTo(original)
		roundTrip := x.ConvertToToFrom(converted)

		assert.InDelta(t, original, roundTrip, 0.01)
	})

	T.Run("round trip conversion with high precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 236.588,
		}

		original := float32(5.0)
		converted := x.ConvertFromToTo(original, 4)
		roundTrip := x.ConvertToToFrom(converted, 4)

		assert.InDelta(t, original, roundTrip, 0.0001)
	})

	T.Run("reverse round trip conversion maintains value", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 28.3495,
		}

		original := float32(100.0)
		// Use higher precision to minimize rounding errors in round trip
		converted := x.ConvertToToFrom(original, 4)
		roundTrip := x.ConvertFromToTo(converted, 4)

		assert.InDelta(t, original, roundTrip, 0.01)
	})

	T.Run("round trip with fractional modifier", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 0.333333,
		}

		original := float32(3.0)
		converted := x.ConvertFromToTo(original, 3)
		roundTrip := x.ConvertToToFrom(converted, 3)

		assert.InDelta(t, original, roundTrip, 0.01)
	})

	T.Run("round trip with zero precision", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitConversion{
			Modifier: 2.0,
		}

		original := float32(10.0)
		converted := x.ConvertFromToTo(original, 0)
		roundTrip := x.ConvertToToFrom(converted, 0)

		assert.InEpsilon(t, original, roundTrip, conversionEpsilon)
	})
}
