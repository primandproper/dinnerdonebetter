package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashClientSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		digest := HashClientSecret("example-secret")

		assert.NotEmpty(t, digest)
		assert.NotEqual(t, "example-secret", digest)
		assert.Len(t, digest, 64)
		assert.Equal(t, digest, HashClientSecret("example-secret"))
	})
}

func TestClientSecretMatches(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		stored := HashClientSecret("example-secret")

		assert.True(t, ClientSecretMatches(stored, "example-secret"))
	})

	T.Run("with wrong secret", func(t *testing.T) {
		t.Parallel()

		stored := HashClientSecret("example-secret")

		assert.False(t, ClientSecretMatches(stored, "some-other-secret"))
	})

	T.Run("with plaintext accidentally stored", func(t *testing.T) {
		t.Parallel()

		// a stored plaintext secret must never match itself; only digests are valid stored forms.
		assert.False(t, ClientSecretMatches("example-secret", "example-secret"))
	})
}
