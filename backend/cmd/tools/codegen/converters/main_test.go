package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// backendRoot is where this tool runs from in anger — four directories up from
// cmd/tools/codegen/converters.
const backendRoot = "../../../.."

func TestGeneratedConvertersAreUpToDate(T *testing.T) {
	T.Parallel()

	// This is the check that turns forgetting `make converters` into a failing build. It
	// matters more than staleness usually does: the generated converters are what carries an
	// entity between its shapes, so a declaration that has been edited and not regenerated is
	// a field that silently stops being copied.
	T.Run("every domain's generated file matches its declarations", func(t *testing.T) {
		t.Parallel()

		index, err := buildStructIndex(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)
		require.NotEmpty(t, registered)

		for _, declarations := range registered {
			rendered, renderErr := renderDomain(index, declarations)
			require.NoError(t, renderErr, "rendering %s", declarations.Domain)

			onDisk, readErr := os.ReadFile(filepath.Join(backendRoot, generatedPath(declarations.Domain)))
			require.NoError(t, readErr)

			assert.Equal(t, string(rendered), string(onDisk),
				"%s converters are stale; run `make converters`", declarations.Domain)
		}
	})

	T.Run("every declared conversion resolves against the real domain types", func(t *testing.T) {
		t.Parallel()

		index, err := buildStructIndex(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)

		for _, declarations := range registered {
			resolver := &planner{index: index, domain: declarations.Domain, pkg: declarations.Domain}
			for _, conversion := range declarations.Conversions {
				_, planErr := resolver.Plan(conversion)
				assert.NoError(t, planErr, "%s", conversion.Name)
			}
		}
	})

	T.Run("every declared reason says something", func(t *testing.T) {
		t.Parallel()

		// A rule whose reason is empty is a rule nobody has to justify, and the point of
		// requiring one is that the justification is where the domain knowledge lives.
		for _, declarations := range registered {
			for _, conversion := range declarations.Conversions {
				for field, rule := range conversion.Fields {
					if rule.kind != ruleSkip && rule.kind != ruleExpr && rule.kind != ruleDeref {
						continue
					}

					assert.NotEmpty(t, rule.why, "%s: %s", conversion.Name, field)
				}
			}
		}
	})
}

func TestBuildStructIndex(T *testing.T) {
	T.Parallel()

	T.Run("indexes the domain packages", func(t *testing.T) {
		t.Parallel()

		index, err := buildStructIndex(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)

		webhook, err := index.Lookup("webhooks", "Webhook")
		require.NoError(t, err)

		// The sentinel field every domain type opens with is not something a conversion
		// assigns, and indexing it would make every destination unfillable.
		for _, field := range webhook.Fields {
			assert.NotEqual(t, "_", field.Name)
		}

		configs, ok := webhook.Field("TriggerConfigs")
		require.True(t, ok)
		assert.Equal(t, "[]*WebhookTriggerConfig", configs.Type)
	})

	T.Run("a domain's subpackages are not indexed as the domain", func(t *testing.T) {
		t.Parallel()

		index, err := buildStructIndex(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)

		// catalog, converters, fakes and mock all sit under webhooks, and none of them
		// declares a type a conversion reads or writes.
		assert.False(t, index.Declares("webhooks", "Catalog"))
	})

	T.Run("unknown lookups are errors rather than empty structs", func(t *testing.T) {
		t.Parallel()

		index, err := buildStructIndex(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)

		_, err = index.Lookup("webhooks", "NotAType")
		require.Error(t, err)

		_, err = index.Lookup("notadomain", "Webhook")
		require.Error(t, err)
	})
}

// fixtureIndex is a two-type domain, small enough that a test can say what it expects of a plan
// without the reader having to look a real domain type up.
func fixtureIndex() *structIndex {
	source := newStructTypeFromFields("Source",
		structField{Name: "ID", Type: "string"},
		structField{Name: "Name", Type: "string"},
		structField{Name: "Optional", Type: "*string"},
		structField{Name: "Nested", Type: "*Nested"},
		structField{Name: "Children", Type: "[]*Child"},
	)
	destination := newStructTypeFromFields("Destination",
		structField{Name: "ID", Type: "string"},
		structField{Name: "Name", Type: "string"},
		structField{Name: "NamePointer", Type: "*string"},
		structField{Name: "Optional", Type: "*string"},
		structField{Name: "NestedID", Type: "*string"},
		structField{Name: "RequiredNestedID", Type: "string"},
		structField{Name: "Children", Type: "[]*ChildInput"},
		structField{Name: "Unfillable", Type: "string"},
	)

	return &structIndex{byDomain: map[string]map[string]*structType{
		"fixture": {
			"Source":      source,
			"Destination": destination,
			"Nested":      newStructTypeFromFields("Nested", structField{Name: "ID", Type: "string"}),
			"Child":       newStructTypeFromFields("Child", structField{Name: "ID", Type: "string"}),
			"ChildInput":  newStructTypeFromFields("ChildInput", structField{Name: "ID", Type: "string"}),
		},
	}}
}

func newStructTypeFromFields(name string, fields ...structField) *structType {
	declared := &structType{Name: name, byName: map[string]structField{}}
	for _, field := range fields {
		declared.add(field.Name, field.Type)
	}

	return declared
}

func fixturePlanner() *planner {
	return &planner{index: fixtureIndex(), domain: "fixture", pkg: "fixture"}
}

// fixtureConversion is a conversion that fills every destination field, so a test can remove one
// rule and see only the failure it is about.
func fixtureConversion() *Conversion {
	return &Conversion{
		Name: "ConvertSourceToDestination",
		From: Param{Name: "src", Type: "Source"},
		To:   "Destination",
		Fields: map[string]Rule{
			"NamePointer":      Ref("Name"),
			"NestedID":         OptionalNestedID("Nested"),
			"RequiredNestedID": NestedID("Nested"),
			"Children":         MapSlice("ConvertChildToChildInput"),
			"Unfillable":       Skip("nothing on the source names it"),
		},
	}
}

func TestPlan(T *testing.T) {
	T.Parallel()

	T.Run("copies same-named fields and takes the address of a value for a pointer", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixturePlanner().Plan(fixtureConversion())
		require.NoError(t, err)

		assert.Equal(t, "src.Name", expressionFor(resolved, "Name"))
		assert.Equal(t, "&src.Name", expressionFor(resolved, "NamePointer"))
		assert.Equal(t, "src.Optional", expressionFor(resolved, "Optional"))
	})

	T.Run("a destination field with no counterpart and no rule is refused", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		delete(conversion.Fields, "Unfillable")

		_, err := fixturePlanner().Plan(conversion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unfillable")
	})

	T.Run("a type the generator cannot adapt is refused rather than guessed at", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		delete(conversion.Fields, "Children")

		_, err := fixturePlanner().Plan(conversion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Children")
	})

	T.Run("a rule for a field the destination does not have is refused", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		conversion.Fields["Renamed"] = Skip("this field was renamed away")

		_, err := fixturePlanner().Plan(conversion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Renamed")
	})

	T.Run("a nested ID is guarded only where the destination can hold the absence", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixturePlanner().Plan(fixtureConversion())
		require.NoError(t, err)

		assert.Equal(t, "src.Nested.ID", expressionFor(resolved, "RequiredNestedID"))
		assert.Equal(t, "nestedID", expressionFor(resolved, "NestedID"))
		require.NotEmpty(t, resolved.Prelude)
		assert.Contains(t, resolved.Prelude[0], "if src.Nested != nil {")
	})

	T.Run("a guarded nested ID needs somewhere to put the absence", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		conversion.Fields["RequiredNestedID"] = OptionalNestedID("Nested")

		_, err := fixturePlanner().Plan(conversion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pointer destination")
	})

	T.Run("a mapped slice is allocated with room for its source", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixturePlanner().Plan(fixtureConversion())
		require.NoError(t, err)

		assert.Contains(t, resolved.Prelude[1], "make([]*fixture.ChildInput, 0, len(src.Children))")
		assert.Contains(t, resolved.Prelude[1], "ConvertChildToChildInput(item)")
	})

	T.Run("a nil-when-empty slice is declared rather than allocated", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		conversion.Fields["Children"] = MapSlice("ConvertChildToChildInput", NilWhenEmpty())

		resolved, err := fixturePlanner().Plan(conversion)
		require.NoError(t, err)
		assert.Contains(t, resolved.Prelude[1], "var children []*fixture.ChildInput")
	})

	T.Run("children stamped with a minted parent ID read the local it was minted into", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		conversion.Fields["ID"] = NewID()
		conversion.Fields["Children"] = MapSlice("ConvertChildToChildInput", BelongsTo("BelongsToParent"))

		resolved, err := fixturePlanner().Plan(conversion)
		require.NoError(t, err)

		// The ID has to exist before the children are built, so it is hoisted out of the
		// literal and into a local that both can read.
		assert.True(t, resolved.UsesIdentifiers)
		assert.Equal(t, "destinationID := identifiers.New()", resolved.Prelude[0])
		assert.Equal(t, "destinationID", expressionFor(resolved, "ID"))
		assert.Contains(t, resolved.Prelude[2], "converted.BelongsToParent = destinationID")
	})

	T.Run("children stamped with a copied parent ID read the copy, not a fresh one", func(t *testing.T) {
		t.Parallel()

		// The failure this guards against is silent: hoisting a minted ID here would give
		// the children an identifier the parent does not have, and every child row would
		// point at nothing.
		conversion := fixtureConversion()
		conversion.Fields["Children"] = MapSlice("ConvertChildToChildInput", BelongsTo("BelongsToParent"))

		resolved, err := fixturePlanner().Plan(conversion)
		require.NoError(t, err)

		assert.False(t, resolved.UsesIdentifiers)
		assert.Equal(t, "src.ID", expressionFor(resolved, "ID"))
		assert.Contains(t, resolved.Prelude[1], "converted.BelongsToParent = src.ID")
	})

	T.Run("children stamped with an ID that is neither copied nor minted are refused", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		conversion.Fields["ID"] = Expr("someHelper(src)", "computed somewhere else")
		conversion.Fields["Children"] = MapSlice("ConvertChildToChildInput", BelongsTo("BelongsToParent"))

		_, err := fixturePlanner().Plan(conversion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "neither copied nor minted")
	})

	T.Run("a minted ID nothing else reads stays inside the literal", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion()
		conversion.Fields["ID"] = NewID()

		resolved, err := fixturePlanner().Plan(conversion)
		require.NoError(t, err)

		assert.True(t, resolved.UsesIdentifiers)
		assert.Equal(t, "identifiers.New()", expressionFor(resolved, "ID"))
	})

	T.Run("a skipped field is carried through so its reason can be rendered", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixturePlanner().Plan(fixtureConversion())
		require.NoError(t, err)

		for _, assigned := range resolved.Assignments {
			if assigned.Field != "Unfillable" {
				continue
			}

			assert.True(t, assigned.Skipped)
			assert.Equal(t, "nothing on the source names it", assigned.Why)

			return
		}

		t.Fatal("the skipped field was dropped from the plan")
	})
}

func TestRenderDomain(T *testing.T) {
	T.Parallel()

	T.Run("two conversions of the same name are refused", func(t *testing.T) {
		t.Parallel()

		_, err := renderDomain(fixtureIndex(), &domainConversions{
			Domain:      "fixture",
			Conversions: []*Conversion{fixtureConversion(), fixtureConversion()},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "declared twice")
	})

	T.Run("a domain with nothing declared is refused rather than emptied", func(t *testing.T) {
		t.Parallel()

		_, err := renderDomain(fixtureIndex(), &domainConversions{Domain: "fixture"})
		require.Error(t, err)
	})

	T.Run("the rendered file compiles as Go and carries its reasons", func(t *testing.T) {
		t.Parallel()

		rendered, err := renderDomain(fixtureIndex(), &domainConversions{
			Domain:      "fixture",
			Conversions: []*Conversion{fixtureConversion()},
		})
		require.NoError(t, err)

		source := string(rendered)
		assert.Contains(t, source, "// Code generated by cmd/tools/codegen/converters. DO NOT EDIT.")
		assert.Contains(t, source, "func ConvertSourceToDestination(src *fixture.Source) *fixture.Destination {")
		assert.Contains(t, source, "// Unfillable is left unset. nothing on the source names it")
	})
}

func expressionFor(resolved *plan, field string) string {
	for _, assigned := range resolved.Assignments {
		if assigned.Field == field {
			return assigned.Expr
		}
	}

	return ""
}

func TestLowerFirstWord(T *testing.T) {
	T.Parallel()

	T.Run("keeps an initialism whole", func(t *testing.T) {
		t.Parallel()

		for input, expected := range map[string]string{
			"TriggerConfigs": "triggerConfigs",
			"ID":             "id",
			"URLPath":        "urlPath",
			"Name":           "name",
			"":               "",
			"lower":          "lower",
		} {
			assert.Equal(t, expected, lowerFirstWord(input), input)
		}
	})
}

func TestEntityLocalName(T *testing.T) {
	T.Parallel()

	T.Run("drops the input shape's suffix", func(t *testing.T) {
		t.Parallel()

		for input, expected := range map[string]string{
			"WebhookDatabaseCreationInput": "webhook",
			"MealPlanCreationRequestInput": "mealPlan",
			"RecipeUpdateRequestInput":     "recipe",
			"Webhook":                      "webhook",
		} {
			assert.Equal(t, expected, entityLocalName(input), input)
		}
	})
}

func TestWrap(T *testing.T) {
	T.Parallel()

	T.Run("breaks on word boundaries within the width", func(t *testing.T) {
		t.Parallel()

		lines := wrap("one two three four five", 9)
		assert.Equal(t, []string{"one two", "three", "four five"}, lines)
	})

	T.Run("an empty reason renders no comment at all", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, wrap("", 80))
	})

	T.Run("a word longer than the width gets its own line", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"a", "supercalifragilistic"}, wrap("a supercalifragilistic", 5))
	})
}

func TestArticle(T *testing.T) {
	T.Parallel()

	T.Run("agrees with the type name's first letter", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "a Webhook", article("Webhook"))
		assert.Equal(t, "an OAuth2Client", article("OAuth2Client"))
		assert.Empty(t, article(""))
	})
}
