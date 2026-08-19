package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_internalKubernetesEndpoint(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := "service.namespace.svc.cluster.local:1234"
		actual := internalKubernetesEndpoint("service", "namespace", 1234)

		assert.Equal(t, expected, actual)
	})
}
