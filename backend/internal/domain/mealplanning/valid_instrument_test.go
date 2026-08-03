package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidInstrument_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidInstrument{}

		input := &ValidInstrumentUpdateRequestInput{}
		require.NoError(t, fake.Struct(&input))
		input.UsableForStorage = new(true)
		input.DisplayInSummaryLists = new(true)
		input.IncludeInGeneratedInstructions = new(true)

		x.Update(input)
	})
}

func TestValidInstrumentCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidInstrumentCreationRequestInput{
			Name:        t.Name(),
			Description: t.Name(),
			IconPath:    t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidInstrumentCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidInstrumentDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidInstrumentDatabaseCreationInput{
			ID:   t.Name(),
			Name: t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidInstrumentDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidInstrumentUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidInstrumentUpdateRequestInput{
			Name:        new(t.Name()),
			Description: new(t.Name()),
			IconPath:    new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidInstrumentUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
