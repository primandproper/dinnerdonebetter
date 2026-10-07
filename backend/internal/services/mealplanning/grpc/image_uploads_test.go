package grpc

import (
	"testing"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	"github.com/primandproper/primitives-go/v2/fake"

	"github.com/stretchr/testify/assert"
)

func TestValidObjectName(T *testing.T) {
	T.Parallel()

	T.Run("admits one path segment", func(t *testing.T) {
		t.Parallel()

		assert.True(t, validObjectName(fake.BuildFakeID()+".png"))
	})

	T.Run("refuses an empty name", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName(""))
	})

	T.Run("refuses the names of a directory", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName("."))
		assert.False(t, validObjectName(".."))
	})

	T.Run("refuses a name that walks out of its prefix", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName("../../"+fake.BuildFakeID()+"/"+fake.BuildFakeID()+".png"))
	})

	T.Run("refuses a name carrying either separator", func(t *testing.T) {
		t.Parallel()

		assert.False(t, validObjectName(fake.BuildFakeID()+"/"+fake.BuildFakeID()+".png"))
		assert.False(t, validObjectName(fake.BuildFakeID()+`\`+fake.BuildFakeID()+".png"))
	})
}

func TestBelongsToAgrees(T *testing.T) {
	T.Parallel()

	subject := mediaregistry.Subject{Type: recipeSubjectType, ID: fake.BuildFakeID()}

	T.Run("admits an absent belongs_to", func(t *testing.T) {
		t.Parallel()

		assert.True(t, belongsToAgrees(mediaregistry.Subject{}, subject))
	})

	T.Run("admits the subject the call uploads to", func(t *testing.T) {
		t.Parallel()

		assert.True(t, belongsToAgrees(mediaregistry.Subject{Type: subject.Type, ID: subject.ID}, subject))
	})

	T.Run("refuses another subject of the same type", func(t *testing.T) {
		t.Parallel()

		assert.False(t, belongsToAgrees(mediaregistry.Subject{Type: subject.Type, ID: fake.BuildFakeID()}, subject))
	})

	T.Run("refuses the same id under another type", func(t *testing.T) {
		t.Parallel()

		assert.False(t, belongsToAgrees(mediaregistry.Subject{Type: mealSubjectType, ID: subject.ID}, subject))
	})

	T.Run("refuses a subject naming only half of one", func(t *testing.T) {
		t.Parallel()

		assert.False(t, belongsToAgrees(mediaregistry.Subject{ID: subject.ID}, subject))
		assert.False(t, belongsToAgrees(mediaregistry.Subject{Type: subject.Type}, subject))
	})
}
