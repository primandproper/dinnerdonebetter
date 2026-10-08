package registration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// marker is what every composition-root site that names this domain carries.
const marker = "// Domain: mealplanning"

// domainRoots are the import paths that are this domain's: the three roots, and the generated
// gRPC surface, which is the domain's by content.
var domainRoots = []string{
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning",
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning",
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning",
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning",
}

// censusRoots are the directories the census holds to the marker: the composition root, its
// config, the binaries and the dev-only processes, the client library, and the generic
// consumers that merge each domain's entry into one list — the data change handler, the index
// event rules, the MCP server, and the analytics allowlist. Every file under them that imports
// a domain package must say so.
var censusRoots = []string{
	"internal/build",
	"internal/config",
	"internal/domain/analytics",
	"internal/functions",
	"internal/indexevents",
	"internal/localdev",
	"internal/mcpserver",
	"internal/mcptools",
	"internal/searchindexes",
	"cmd/ddb",
	"cmd/tools",
	"pkg/client",
}

// TestDomainMarkerCensus holds the composition root to the marker.
//
// The point of the marker is that grepping for it is the complete list of what to edit when the
// domain is swapped. A site that names the domain without carrying it is a site that list
// misses, and the only time anyone finds out is while swapping.
func TestDomainMarkerCensus(T *testing.T) {
	T.Parallel()

	T.Run("every composition-root file that imports the domain carries the marker", func(t *testing.T) {
		t.Parallel()

		moduleRoot := findModuleRoot(t)
		fset := token.NewFileSet()

		var unmarked []string
		marked := 0

		for _, root := range censusRoots {
			err := filepath.WalkDir(filepath.Join(moduleRoot, root), func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}

				if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}

				file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly|parser.ParseComments)
				if parseErr != nil {
					return parseErr
				}

				if !importsDomain(file.Imports) {
					return nil
				}

				contents, readErr := os.ReadFile(path)
				if readErr != nil {
					return readErr
				}

				relative, relErr := filepath.Rel(moduleRoot, path)
				if relErr != nil {
					return relErr
				}

				if strings.Contains(string(contents), marker) {
					marked++
				} else {
					unmarked = append(unmarked, relative)
				}

				return nil
			})
			require.NoError(t, err)
		}

		assert.Empty(t, unmarked, "these files name the domain and do not carry %q", marker)
		assert.NotZero(t, marked, "the census found no marked site, so it is looking in the wrong place")
	})
}

// importsDomain reports whether any of imports names a package under a domain root.
func importsDomain(imports []*ast.ImportSpec) bool {
	for _, imp := range imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		for _, root := range domainRoots {
			if path == root || strings.HasPrefix(path, root+"/") {
				return true
			}
		}
	}

	return false
}

// findModuleRoot walks up from the test's directory to the go.mod.
func findModuleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "no go.mod above %s", dir)

		dir = parent
	}
}
