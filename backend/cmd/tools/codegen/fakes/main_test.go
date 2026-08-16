package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

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

			rendered, err := renderDomain(d, dir, newEnumIndex(backendRoot))
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
		ArchivedAt       *time.Time
		Nickname         *string
		ID               string
		BelongsToAccount string
		CreatedByUser    string
		Name             string
		Children         []*string
		Ratio            float64
		Quantity         uint16
		Optional         bool
	}

	d := &defaulter{
		builders:   map[reflect.Type]string{},
		domainPath: "example.com/domain",
		enums:      newEnumIndex(backendRoot),
	}
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
		// An optional field is filled, not left nil: a nil optional makes every
		// assertion about it pass whether or not the code under test copies it.
		{field: "Nickname", want: "pointer.To(buildUniqueString())"},
		// Except a tombstone, whose nil is the field's meaning and not the absence of one.
		{field: "ArchivedAt", want: "nil"},
		// A collection of something the domain does not declare has no faithful shape
		// to invent. A collection of entities is filled; see TestDeriveChildren.
		{field: "Children", want: "nil"},
	} {
		T.Run(tc.field, func(t *testing.T) {
			t.Parallel()

			field, found := typ.FieldByName(tc.field)
			require.True(t, found)

			expression, ok, err := d.expr(typ, &field)
			require.NoError(t, err)
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

		expression, ok, err := d.expr(typ, &field)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "BuildFakeTime()", expression)
	})

	T.Run("an enumerated field offers every member", func(t *testing.T) {
		t.Parallel()

		// A status, a kind, a shape: the valid values are constants, so they are read
		// out of the domain's source rather than guessed. Offering all of them is what
		// stops a switch that handles one member from passing every test.
		type withStatus struct {
			_      struct{}
			Status mealplanning.MealPlanStatus
		}

		field, found := reflect.TypeFor[withStatus]().FieldByName("Status")
		require.True(t, found)

		enumerated := &defaulter{
			builders:   map[reflect.Type]string{},
			domainPath: domainImportPath("mealplanning"),
			enums:      newEnumIndex(backendRoot),
		}

		expression, ok, err := enumerated.expr(reflect.TypeFor[withStatus](), &field)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "pickOne(types.MealPlanStatusAwaitingVotes, types.MealPlanStatusFinalized)", expression)
	})

	T.Run("a named string with no constants has no default", func(t *testing.T) {
		t.Parallel()

		// Nothing says what this type is allowed to hold, so the declaration has to.
		type status string

		type withStatus struct {
			_      struct{}
			Status status
		}

		field, found := reflect.TypeFor[withStatus]().FieldByName("Status")
		require.True(t, found)

		_, ok, err := d.expr(reflect.TypeFor[withStatus](), &field)
		require.NoError(t, err)
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
			enums:      newEnumIndex(backendRoot),
		}

		field, found := reflect.TypeFor[parent]().FieldByName("Child")
		require.True(t, found)

		expression, ok, err := withBuilder.expr(reflect.TypeFor[parent](), &field)
		require.NoError(t, err)
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

		_, ok, err := d.expr(reflect.TypeFor[parent](), &field)
		require.NoError(t, err)
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

func TestDeriveChildren(T *testing.T) {
	T.Parallel()

	type Child struct {
		_             struct{}
		ID            string
		BelongsToItem string
	}

	type Item struct {
		_        struct{}
		ID       string
		Children []*Child
	}

	defaults := &defaulter{
		builders: map[reflect.Type]string{
			reflect.TypeFor[Item]():  "BuildFakeItem",
			reflect.TypeFor[Child](): "BuildFakeChild",
		},
		domainPath: "example.com/domain",
		enums:      newEnumIndex(backendRoot),
	}

	T.Run("a collection of entities is filled and wired to its parent", func(t *testing.T) {
		t.Parallel()

		// This is the whole reason the declarations had a Locals section: the same loop,
		// written once per parent, differing only in the type names the generator can see.
		p := &entityPlan{typ: reflect.TypeFor[Item](), builderName: "BuildFakeItem"}

		children, err := planChildren([]entityPlan{*p}, defaults)
		require.NoError(t, err)

		derived, err := derive(p, defaults, children, map[string]struct{}{})
		require.NoError(t, err)

		assert.Equal(t, "children", derived.exprs["Children"])
		assert.Equal(t, "itemID", derived.exprs["ID"], "the parent has to name its own ID for the children to carry it")
		require.Len(t, derived.locals, 2)
		assert.Equal(t, "itemID := BuildFakeID()", derived.locals[0])
		assert.Contains(t, derived.locals[1], "child.BelongsToItem = itemID")
	})

	T.Run("an override wins over derivation", func(t *testing.T) {
		t.Parallel()

		p := &entityPlan{typ: reflect.TypeFor[Item](), builderName: "BuildFakeItem"}

		children, err := planChildren([]entityPlan{*p}, defaults)
		require.NoError(t, err)

		derived, err := derive(p, defaults, children, map[string]struct{}{"Children": {}})
		require.NoError(t, err)

		assert.Empty(t, derived.locals, "nothing to build once the declaration says what the collection holds")
		assert.NotContains(t, derived.exprs, "Children")
	})

	T.Run("a self-referential collection is left nil", func(t *testing.T) {
		t.Parallel()

		// Recipe.AssociatedRecipes is the real one. Filling it would build recipes until
		// the stack ran out, so nil — a shape the field genuinely has — is what it gets.
		type Recursive struct {
			_       struct{}
			ID      string
			Related []*Recursive
		}

		recursiveDefaults := &defaulter{
			builders:   map[reflect.Type]string{reflect.TypeFor[Recursive](): "BuildFakeRecursive"},
			domainPath: "example.com/domain",
			enums:      newEnumIndex(backendRoot),
		}

		p := &entityPlan{typ: reflect.TypeFor[Recursive](), builderName: "BuildFakeRecursive"}

		children, err := planChildren([]entityPlan{*p}, recursiveDefaults)
		require.NoError(t, err)

		derived, err := derive(p, recursiveDefaults, children, map[string]struct{}{})
		require.NoError(t, err)

		assert.NotContains(t, derived.exprs, "Related")
	})
}

func TestValidationDerivedEnums(T *testing.T) {
	T.Parallel()

	// Almost every enumerated field in this repository is a plain `string`, so the only thing
	// that says which strings it may hold is the rule the type declares about itself. Reading
	// that rule is what keeps a generated fake from being one the domain rejects.
	T.Run("a field is faked as a value its own validation admits", func(t *testing.T) {
		t.Parallel()

		index := newEnumIndex(backendRoot)

		values, err := index.permitted(domainImportPath("mealplanning"), "ValidVesselCreationRequestInput", "Shape")
		require.NoError(t, err)

		assert.Contains(t, values, "VesselShapeCone")
		assert.Contains(t, values, "VesselShapeOther")
	})

	T.Run("an entity borrows the rule from its creation input", func(t *testing.T) {
		t.Parallel()

		// A ValidVessel declares no validation about itself — it is what comes back out
		// of the database. The input that put it there is where the rule lives.
		index := newEnumIndex(backendRoot)

		values, err := index.permitted(domainImportPath("mealplanning"), "ValidVessel", "Shape")
		require.NoError(t, err)

		assert.Contains(t, values, "VesselShapeCone")
	})

	T.Run("a type with no rule for a field says nothing", func(t *testing.T) {
		t.Parallel()

		index := newEnumIndex(backendRoot)

		values, err := index.permitted(domainImportPath("mealplanning"), "ValidVessel", "Name")
		require.NoError(t, err)
		assert.Empty(t, values)
	})
}
