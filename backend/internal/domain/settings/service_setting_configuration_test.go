package settings

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceSettingConfiguration_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ServiceSettingConfiguration{}
		input := &ServiceSettingConfigurationUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestServiceSettingConfigurationCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationCreationRequestInput{
			Value:            t.Name(),
			ServiceSettingID: t.Name(),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})

	T.Run("with invalid struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationCreationRequestInput{}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})
}

func TestServiceSettingConfigurationDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationDatabaseCreationInput{
			ID:            t.Name(),
			Value:         t.Name(),
			BelongsToUser: t.Name(),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})

	T.Run("with invalid struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationDatabaseCreationInput{}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})
}

func TestServiceSettingConfigurationUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationUpdateRequestInput{
			Value:            new(t.Name()),
			ServiceSettingID: new(t.Name()),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})

	T.Run("with invalid struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationUpdateRequestInput{}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})
}
