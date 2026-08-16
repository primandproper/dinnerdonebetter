package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// backendRoot is where this tool runs from in anger — four directories up from
// cmd/tools/codegen/fakes.
const backendRoot = "../../../.."

func TestFakesAreUpToDate(T *testing.T) {
	T.Parallel()

	// A stale fake is a field that silently stays at its zero value in every test that uses
	// it, which is exactly the failure mode nobody notices: the tests pass, and the thing they
	// were meant to cover was never populated. This is the check that turns forgetting
	// `make fakes` into a failing build.
	T.Run("generated files match the declarations", func(t *testing.T) {
		t.Parallel()

		for _, d := range domains {
			dir := filepath.Join(backendRoot, packageDir(d.name))

			rendered, err := renderDomain(d, dir)
			require.NoError(t, err, d.name)

			onDisk, err := os.ReadFile(filepath.Join(dir, outputName))
			require.NoError(t, err, d.name)

			assert.Equal(t, string(rendered), string(onDisk),
				"%s fakes are stale; run `make fakes`", d.name)
		}
	})
}

func TestDomainsAreComplete(T *testing.T) {
	T.Parallel()

	// The registry in main.go is hand-maintained, so a domain that grows a fakes package and
	// is not added to it generates nothing and says nothing. This is what says it.
	T.Run("every fakes package has a domain entry", func(t *testing.T) {
		t.Parallel()

		registered := map[string]bool{}
		for _, d := range domains {
			registered[d.name] = true
		}

		entries, err := os.ReadDir(filepath.Join(backendRoot, "internal", "domain"))
		require.NoError(t, err)

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			if _, err = os.Stat(filepath.Join(backendRoot, packageDir(entry.Name()))); err != nil {
				continue
			}

			assert.True(t, registered[entry.Name()],
				"internal/domain/%s has a fakes package but no entry in domains", entry.Name())
		}
	})
}

func TestDeclarations(T *testing.T) {
	T.Parallel()

	T.Run("every override names a real field", func(t *testing.T) {
		t.Parallel()

		// renderEntity refuses one, but only for entities it renders: an override on a
		// bespoke entity is never looked at, and would rot unnoticed until somebody
		// un-bespoke'd it.
		for _, d := range domains {
			for _, e := range d.decl.Entities {
				typ := reflect.TypeOf(e.Type)
				require.NotNil(t, typ, d.name)

				for _, f := range e.Fake.Fields {
					_, found := typ.FieldByName(f.Name)
					assert.True(t, found, "%s: %s.%s is not a field", d.name, typ.Name(), f.Name)
				}
			}
		}
	})

	T.Run("bespoke builders say why", func(t *testing.T) {
		t.Parallel()

		for _, d := range domains {
			for _, e := range d.decl.Entities {
				if !e.Fake.Bespoke {
					continue
				}

				assert.NotEmpty(t, e.Fake.BespokeWhy,
					"%s: %s opts out of generation without saying why", d.name, reflect.TypeOf(e.Type).Name())
			}
		}
	})

	T.Run("no entity is declared twice", func(t *testing.T) {
		t.Parallel()

		for _, d := range domains {
			seen := map[reflect.Type]bool{}

			for _, e := range d.decl.Entities {
				typ := reflect.TypeOf(e.Type)
				assert.False(t, seen[typ], "%s declares %s twice", d.name, typ.Name())
				seen[typ] = true
			}
		}
	})
}

func TestDefaults(T *testing.T) {
	T.Parallel()

	type example struct {
		_ struct{}

		CreatedAt        any
		ArchivedAt       *string
		ID               string
		BelongsToAccount string
		CreatedByUser    string
		Name             string
		Children         []*string
		Ratio            float64
		Quantity         uint16
		Optional         bool
	}

	d := &defaulter{builders: map[reflect.Type]string{}, domainPath: "example.com/domain"}
	typ := reflect.TypeFor[example]()

	for _, tc := range []struct {
		field string
		want  string
	}{
		// A string field's name is the only thing separating an identifier from prose,
		// and a lorem-ipsum sentence in an ID column fails every foreign key it meets.
		{field: "ID", want: "BuildFakeID()"},
		{field: "BelongsToAccount", want: "BuildFakeID()"},
		{field: "CreatedByUser", want: "BuildFakeID()"},
		{field: "Name", want: "buildUniqueString()"},
		{field: "Optional", want: "fake.Bool()"},
		{field: "Quantity", want: "uint16(buildFakeNumber())"},
		{field: "Ratio", want: "buildFakeNumber()"},
		{field: "ArchivedAt", want: "nil"},
		{field: "Children", want: "nil"},
	} {
		T.Run(tc.field, func(t *testing.T) {
			t.Parallel()

			field, found := typ.FieldByName(tc.field)
			require.True(t, found)

			expression, ok := d.expr(&field)
			require.True(t, ok)
			assert.Equal(t, tc.want, expression)
		})
	}

	T.Run("time fields get a faked time", func(t *testing.T) {
		t.Parallel()

		type withTime struct {
			_         struct{}
			CreatedAt time.Time
		}

		field, found := reflect.TypeFor[withTime]().FieldByName("CreatedAt")
		require.True(t, found)

		expression, ok := d.expr(&field)
		require.True(t, ok)
		assert.Equal(t, "BuildFakeTime()", expression)
	})

	T.Run("a named string has no default", func(t *testing.T) {
		t.Parallel()

		// A status, a kind, a shape: the valid values are constants, and a random string
		// is not one of them. Refusing here is what forces the declaration to say which
		// constant, instead of producing a fake that only exercises the rejection path.
		type status string

		type withStatus struct {
			_      struct{}
			Status status
		}

		field, found := reflect.TypeFor[withStatus]().FieldByName("Status")
		require.True(t, found)

		_, ok := d.expr(&field)
		assert.False(t, ok)
	})

	T.Run("a struct with a builder gets it", func(t *testing.T) {
		t.Parallel()

		type child struct {
			_ struct{}
		}

		type parent struct {
			_     struct{}
			Child child
		}

		withBuilder := &defaulter{
			builders:   map[reflect.Type]string{reflect.TypeFor[child](): "BuildFakeChild"},
			domainPath: "example.com/domain",
		}

		field, found := reflect.TypeFor[parent]().FieldByName("Child")
		require.True(t, found)

		expression, ok := withBuilder.expr(&field)
		require.True(t, ok)
		assert.Equal(t, "*BuildFakeChild()", expression)
	})

	T.Run("a struct without one does not", func(t *testing.T) {
		t.Parallel()

		type child struct {
			_ struct{}
		}

		type parent struct {
			_     struct{}
			Child child
		}

		field, found := reflect.TypeFor[parent]().FieldByName("Child")
		require.True(t, found)

		_, ok := d.expr(&field)
		assert.False(t, ok)
	})
}

func TestRenderHelpers(T *testing.T) {
	T.Parallel()

	T.Run("emits only what is named", func(t *testing.T) {
		t.Parallel()

		rendered := renderHelpers(map[string]struct{}{"BuildFakeID": {}})

		assert.Contains(t, rendered, "func BuildFakeID() string")
		assert.Contains(t, rendered, "func init()", "the seed is unconditional")
		assert.NotContains(t, rendered, "func buildFakePassword")
	})

	T.Run("pulls in what a wanted helper needs", func(t *testing.T) {
		t.Parallel()

		// Nothing in a package need mention buildFakeNumber for it to be required: the
		// min/max helper that does is enough, and a package missing it does not compile.
		rendered := renderHelpers(map[string]struct{}{"BuildFakeUint16WithOptionalMax": {}})

		assert.Contains(t, rendered, "func buildFakeNumber() float64")
	})
}

func TestRenderImports(T *testing.T) {
	T.Parallel()

	T.Run("groups match gci's sections", func(t *testing.T) {
		t.Parallel()

		rendered := renderImports("webhooks", "time.Now(); types.Webhook{}; filtering.Pagination{}; fake.Bool()")

		lines := []string{}
		for line := range strings.SplitSeq(rendered, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" && trimmed != "import (" && trimmed != ")" {
				lines = append(lines, trimmed)
			}
		}

		assert.Equal(t, []string{
			`"time"`,
			`types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"`,
			`"github.com/primandproper/platform-go/v10/filtering"`,
			`fake "github.com/brianvoe/gofakeit/v7"`,
		}, lines)
	})

	T.Run("ignores qualifiers inside comments", func(t *testing.T) {
		t.Parallel()

		// Override comments quote code freely, and an import added because a comment
		// mentioned a package is an import nothing uses and a file that does not compile.
		rendered := renderImports("webhooks", "// see time.Now\nfake.Bool()")

		assert.NotContains(t, rendered, `"time"`)
		assert.Contains(t, rendered, "gofakeit")
	})
}

func TestWrap(T *testing.T) {
	T.Parallel()

	T.Run("breaks on words", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"one two", "three"}, wrap("one two three", 8))
	})

	T.Run("keeps paragraph breaks", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"one", "", "two"}, wrap("one\n\ntwo", 40))
	})
}

func TestPlan(T *testing.T) {
	T.Parallel()

	T.Run("refuses a pointer", func(t *testing.T) {
		t.Parallel()

		// `Type: &Webhook{}` is the natural typo, and reflecting over it would report a
		// pointer with no fields — a builder that sets nothing, silently.
		_, err := plan(entitydecl.Domain{Entities: []entitydecl.Entity{{Type: &struct{}{}}}})
		assert.Error(t, err)
	})

	T.Run("refuses a missing type", func(t *testing.T) {
		t.Parallel()

		_, err := plan(entitydecl.Domain{Entities: []entitydecl.Entity{{}}})
		assert.Error(t, err)
	})
}
