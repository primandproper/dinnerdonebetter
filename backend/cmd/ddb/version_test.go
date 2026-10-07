package main

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/primandproper/primitives-go/v2/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stampedPackages finds every package an -X flag in a build file stamps, and every package
// build.sh's VERSION_PKG names.
var stampedPackages = regexp.MustCompile(`-X\s+([\w./-]+)\.\w+=|VERSION_PKG="([\w./-]+)"`)

// TestBuildScriptStampsTheLinkedVersionPackage pins the -X flags the deployed binaries are built
// with to the version package `ddb version` reads.
//
// The linker ignores an -X that names a package not in the build, so a flag left pointing at a
// package the binary stopped importing is no error anywhere: the binary just reports "unknown".
// That is how every deployed binary came to report no version after the version package moved
// from platform-go to primitives-go.
func TestBuildScriptStampsTheLinkedVersionPackage(T *testing.T) {
	T.Parallel()

	linked := reflect.TypeFor[version.Info]().PkgPath()

	for _, file := range []string{
		filepath.Join("..", "..", "scripts", "build.sh"),
		filepath.Join("..", "..", "skaffold.yaml"),
	} {
		T.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(file)
			require.NoError(t, err)

			matches := stampedPackages.FindAllStringSubmatch(string(content), -1)
			require.NotEmpty(t, matches, "%s stamps no version package", file)

			for _, match := range matches {
				stamped := match[1] + match[2]
				assert.Equal(t, linked, stamped, "%s stamps a version package the binary does not link", file)
			}
		})
	}
}
