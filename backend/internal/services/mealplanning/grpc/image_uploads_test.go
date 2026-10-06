package grpc

import (
	"testing"

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
