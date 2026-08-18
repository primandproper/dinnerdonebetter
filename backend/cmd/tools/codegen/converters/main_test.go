package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// backendRoot is where this tool runs from in anger — four directories up from
// cmd/tools/codegen/converters.
const backendRoot = "../../../.."

func realIndex(t *testing.T) *structIndex {
	t.Helper()

	index, err := buildStructIndex(filepath.Join(backendRoot, domainDir))
	require.NoError(t, err)

	return index
}

func realDomains(t *testing.T) []string {
	t.Helper()

	domains, err := domainsWithConverters(filepath.Join(backendRoot, domainDir))
	require.NoError(t, err)
	require.NotEmpty(t, domains)

	return domains
}

func TestGeneratedConvertersAreUpToDate(T *testing.T) {
	T.Parallel()

	// This is the check that turns forgetting `make converters` into a failing build. It
	// matters more than staleness usually does: these functions are what carries an entity
	// between its shapes, so a domain type that has grown a field and not been regenerated is
	// a field that silently stops being copied.
	T.Run("every domain's generated file matches its types", func(t *testing.T) {
		t.Parallel()

		index := realIndex(t)

		for _, domain := range realDomains(t) {
			rendered, err := renderDomain(index, enumerate(index, domain))
			require.NoError(t, err, "rendering %s", domain)

			onDisk, readErr := os.ReadFile(filepath.Join(backendRoot, generatedPath(domain)))
			require.NoError(t, readErr)

			assert.Equal(t, string(rendered), string(onDisk),
				"%s converters are stale; run `make converters`", domain)
		}
	})
}

func TestExceptions(T *testing.T) {
	T.Parallel()

	// An exception that names nothing is worse than no exception: the field it was meant to
	// answer quietly goes back to whatever the derivation makes of it, and the entry sits
	// there reading as though it still applies.
	T.Run("every exception names a conversion that is generated", func(t *testing.T) {
		t.Parallel()

		index := realIndex(t)

		generated := map[string]struct{}{}
		for _, domain := range realDomains(t) {
			for _, conversion := range enumerate(index, domain).Conversions {
				generated[conversion.Name] = struct{}{}
			}
		}

		for name := range fieldExceptions {
			_, ok := generated[name]
			assert.True(t, ok, "%s has field exceptions but is not generated", name)
		}
	})

	T.Run("every hand-written entry names a conversion the rules would have produced", func(t *testing.T) {
		t.Parallel()

		index := realIndex(t)

		// handWritten suppresses generation, so an entry naming a conversion the shape rules
		// never produce suppresses nothing and only misleads.
		producible := map[string]struct{}{}

		for _, domain := range realDomains(t) {
			for _, typeName := range index.TypeNames(domain) {
				entityName, isEntity := entityOf(typeName)
				if !isEntity {
					continue
				}

				for _, shape := range conversionShapes {
					from := shape.From.typeName(entityName)
					to := shape.To.typeName(entityName)
					if index.Declares(domain, from) && index.Declares(domain, to) {
						producible[conversionName(from, to)] = struct{}{}
					}
				}
			}
		}

		for name := range handWritten {
			_, ok := producible[name]
			assert.True(t, ok, "%s is listed as hand-written but the shape rules do not produce it", name)
		}
	})

	T.Run("every reason says something", func(t *testing.T) {
		t.Parallel()

		for name, fields := range fieldExceptions {
			for field, rule := range fields {
				assert.NotEmpty(t, rule.why, "%s: %s", name, field)
			}
		}

		for name, why := range handWritten {
			assert.NotEmpty(t, why, name)
		}
	})

	T.Run("a hand-written conversion is actually hand-written", func(t *testing.T) {
		t.Parallel()

		// The generator declining to emit a conversion is only half of it: the function still
		// has to exist, or every call site breaks.
		manual := map[string]struct{}{}

		for _, domain := range realDomains(t) {
			path := filepath.Join(backendRoot, domainDir, domain, "converters", "converters_manual.go")

			source, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			for _, line := range strings.Split(string(source), "\n") {
				name, isFunc := strings.CutPrefix(line, "func ")
				if !isFunc {
					continue
				}

				if index := strings.IndexByte(name, '('); index > 0 {
					manual[name[:index]] = struct{}{}
				}
			}
		}

		for name := range handWritten {
			_, ok := manual[name]
			assert.True(t, ok, "%s is not generated and not written by hand either", name)
		}
	})
}

func TestEntityOf(T *testing.T) {
	T.Parallel()

	T.Run("separates an entity from its shapes", func(t *testing.T) {
		t.Parallel()

		for typeName, expected := range map[string]string{
			"Webhook":                      "Webhook",
			"WebhookCreationRequestInput":  "Webhook",
			"WebhookDatabaseCreationInput": "Webhook",
			"WebhookUpdateRequestInput":    "Webhook",
			"WebhookTriggerConfig":         "WebhookTriggerConfig",
		} {
			entityName, isEntity := entityOf(typeName)
			assert.Equal(t, expected, entityName, typeName)
			assert.Equal(t, typeName == expected, isEntity, typeName)
		}
	})
}

func TestConversionName(T *testing.T) {
	T.Parallel()

	T.Run("is the two type names", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t,
			"ConvertWebhookToWebhookCreationRequestInput",
			conversionName("Webhook", "WebhookCreationRequestInput"),
		)
	})
}

// fixtureIndex is a small domain a test can state expectations about without looking a real domain
// type up.
func fixtureIndex() *structIndex {
	return &structIndex{byDomain: map[string]map[string]*structType{
		"fixture": {
			"Thing": newStructTypeFromFields("Thing",
				structField{Name: "ID", Type: "string"},
				structField{Name: "Name", Type: "string"},
				structField{Name: "Optional", Type: "*string"},
				structField{Name: "Nested", Type: "*Nested"},
				structField{Name: "Owner", Type: "Nested"},
				structField{Name: "Parts", Type: "[]*Part"},
			),
			"ThingCreationRequestInput": newStructTypeFromFields("ThingCreationRequestInput",
				structField{Name: "Name", Type: "string"},
				structField{Name: "Parts", Type: "[]*PartCreationRequestInput"},
			),
			"ThingDatabaseCreationInput": newStructTypeFromFields("ThingDatabaseCreationInput",
				structField{Name: "ID", Type: "string"},
				structField{Name: "Name", Type: "string"},
				structField{Name: "NestedID", Type: "*string"},
				structField{Name: "OwnerID", Type: "string"},
				structField{Name: "Parts", Type: "[]*PartDatabaseCreationInput"},
			),
			"ThingUpdateRequestInput": newStructTypeFromFields("ThingUpdateRequestInput",
				structField{Name: "Name", Type: "*string"},
			),
			"Nested":                   newStructTypeFromFields("Nested", structField{Name: "ID", Type: "string"}),
			"Part":                     newStructTypeFromFields("Part", structField{Name: "ID", Type: "string"}),
			"PartCreationRequestInput": newStructTypeFromFields("PartCreationRequestInput", structField{Name: "ID", Type: "string"}),
			"PartDatabaseCreationInput": newStructTypeFromFields("PartDatabaseCreationInput",
				structField{Name: "ID", Type: "string"}),
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
	return &planner{index: fixtureIndex(), domain: "fixture"}
}

func fixtureConversion(from, to string) *Conversion {
	return &Conversion{Entity: "Thing", From: from, To: to, Name: conversionName(from, to)}
}

func TestEnumerate(T *testing.T) {
	T.Parallel()

	T.Run("produces every shape the entity has types for", func(t *testing.T) {
		t.Parallel()

		enumerated := enumerate(fixtureIndex(), "fixture")

		var names []string
		for _, conversion := range enumerated.Conversions {
			names = append(names, conversion.Name)
		}

		// Thing has no UpdateRequestInput, so no conversion to one is produced — and
		// nothing had to say so.
		// Part has no UpdateRequestInput, so no conversion to one is produced — and nothing
		// had to say so. Thing has one, so it gets one.
		assert.Equal(t, []string{
			"ConvertPartToPartCreationRequestInput",
			"ConvertPartToPartDatabaseCreationInput",
			"ConvertPartCreationRequestInputToPartDatabaseCreationInput",
			"ConvertThingToThingCreationRequestInput",
			"ConvertThingToThingDatabaseCreationInput",
			"ConvertThingToThingUpdateRequestInput",
			"ConvertThingCreationRequestInputToThingDatabaseCreationInput",
		}, names)
	})
}

func TestDerivation(T *testing.T) {
	T.Parallel()

	T.Run("copies, takes addresses, and maps collections without being told to", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixturePlanner().Plan(fixtureConversion("Thing", "ThingDatabaseCreationInput"))
		require.NoError(t, err)

		assert.Equal(t, "x.ID", expressionFor(resolved, "ID"))
		assert.Equal(t, "x.Name", expressionFor(resolved, "Name"))
		assert.Equal(t, "parts", expressionFor(resolved, "Parts"))
		assert.Contains(t, strings.Join(resolved.Prelude, "\n"), "ConvertPartToPartDatabaseCreationInput(item)")

		// A destination that holds a pointer to what the source holds by value takes its
		// address, which is what an update input is made of.
		update, err := fixturePlanner().Plan(fixtureConversion("Thing", "ThingUpdateRequestInput"))
		require.NoError(t, err)
		assert.Equal(t, "&x.Name", expressionFor(update, "Name"))
	})

	T.Run("reduces a nested entity to its identifier, guarding only where absence fits", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixturePlanner().Plan(fixtureConversion("Thing", "ThingDatabaseCreationInput"))
		require.NoError(t, err)

		// Nested is a pointer and NestedID is a pointer, so the read is guarded; Owner is a
		// value and OwnerID is a value, so it is read straight through.
		assert.Equal(t, "nestedID", expressionFor(resolved, "NestedID"))
		assert.Contains(t, strings.Join(resolved.Prelude, "\n"), "if x.Nested != nil {")
		assert.Equal(t, "x.Owner.ID", expressionFor(resolved, "OwnerID"))
	})

	T.Run("mints an identifier a request input cannot supply", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion("ThingCreationRequestInput", "ThingDatabaseCreationInput")
		conversion.Fields = map[string]Rule{
			"NestedID": Skip("the caller resolves this"),
			"OwnerID":  Skip("the caller stamps this"),
		}

		resolved, err := fixturePlanner().Plan(conversion)
		require.NoError(t, err)

		assert.True(t, resolved.UsesIdentifiers)
		assert.Equal(t, "identifiers.New()", expressionFor(resolved, "ID"))
	})

	T.Run("a field nothing answers fails the build rather than going quiet", func(t *testing.T) {
		t.Parallel()

		index := fixtureIndex()
		index.byDomain["fixture"]["ThingDatabaseCreationInput"].add("Unanswerable", "string")

		resolver := &planner{index: index, domain: "fixture"}

		_, err := resolver.Plan(fixtureConversion("Thing", "ThingDatabaseCreationInput"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unanswerable")
	})

	T.Run("every unanswered field is reported, not just the first", func(t *testing.T) {
		t.Parallel()

		index := fixtureIndex()
		index.byDomain["fixture"]["ThingDatabaseCreationInput"].add("FirstMissing", "string")
		index.byDomain["fixture"]["ThingDatabaseCreationInput"].add("SecondMissing", "string")

		resolver := &planner{index: index, domain: "fixture"}

		_, err := resolver.Plan(fixtureConversion("Thing", "ThingDatabaseCreationInput"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "FirstMissing")
		assert.Contains(t, err.Error(), "SecondMissing")
	})

	T.Run("a dereference is never derived", func(t *testing.T) {
		t.Parallel()

		// *string on the source and string on the destination is a read that panics on nil.
		// Deriving it would be the generator making that call on the domain's behalf.
		index := fixtureIndex()
		index.byDomain["fixture"]["ThingDatabaseCreationInput"].add("Optional", "string")

		resolver := &planner{index: index, domain: "fixture"}

		_, err := resolver.Plan(fixtureConversion("Thing", "ThingDatabaseCreationInput"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Optional")
	})

	T.Run("an ambiguous relation is refused rather than picked", func(t *testing.T) {
		t.Parallel()

		// A destination named for the relation's type rather than for the field is
		// resolved by finding the field holding that type — which only works when there
		// is one of them.
		index := fixtureIndex()
		index.byDomain["fixture"]["Thing"] = newStructTypeFromFields("Thing",
			structField{Name: "ID", Type: "string"},
			structField{Name: "Name", Type: "string"},
			structField{Name: "Optional", Type: "*string"},
			structField{Name: "Alpha", Type: "*Nested"},
			structField{Name: "Beta", Type: "*Nested"},
			structField{Name: "Owner", Type: "Nested"},
			structField{Name: "Parts", Type: "[]*Part"},
		)

		resolver := &planner{index: index, domain: "fixture"}

		_, err := resolver.Plan(fixtureConversion("Thing", "ThingDatabaseCreationInput"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NestedID")
	})

	T.Run("an exception for a field the destination lacks is refused", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion("Thing", "ThingDatabaseCreationInput")
		conversion.Fields = map[string]Rule{"Renamed": Skip("this field was renamed away")}

		_, err := fixturePlanner().Plan(conversion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Renamed")
	})

	T.Run("a skipped field is carried through so its reason can be rendered", func(t *testing.T) {
		t.Parallel()

		conversion := fixtureConversion("Thing", "ThingDatabaseCreationInput")
		conversion.Fields = map[string]Rule{"OwnerID": Skip("the caller stamps this")}

		resolved, err := fixturePlanner().Plan(conversion)
		require.NoError(t, err)

		for _, assigned := range resolved.Assignments {
			if assigned.Field != "OwnerID" {
				continue
			}

			assert.True(t, assigned.Skipped)
			assert.Equal(t, "the caller stamps this", assigned.Why)

			return
		}

		t.Fatal("the skipped field was dropped from the plan")
	})
}

func TestRenderDomain(T *testing.T) {
	T.Parallel()

	T.Run("the rendered file names its source x and carries its reasons", func(t *testing.T) {
		t.Parallel()

		index := fixtureIndex()
		enumerated := enumerate(index, "fixture")

		for _, conversion := range enumerated.Conversions {
			switch conversion.Name {
			case "ConvertThingToThingDatabaseCreationInput":
				conversion.Fields = map[string]Rule{"OwnerID": Skip("the caller stamps this")}
			case "ConvertThingCreationRequestInputToThingDatabaseCreationInput":
				conversion.Fields = map[string]Rule{
					"NestedID": Skip("the caller resolves this"),
					"OwnerID":  Skip("the caller stamps this"),
				}
			}
		}

		rendered, err := renderDomain(index, enumerated)
		require.NoError(t, err)

		source := string(rendered)
		assert.Contains(t, source, "// Code generated by cmd/tools/codegen/converters. DO NOT EDIT.")
		assert.Contains(t, source, "func ConvertThingToThingDatabaseCreationInput(x *fixture.Thing) *fixture.ThingDatabaseCreationInput {")
		assert.Contains(t, source, "// OwnerID is left unset. the caller stamps this")
	})

	T.Run("a domain with nothing to convert is refused rather than emptied", func(t *testing.T) {
		t.Parallel()

		_, err := renderDomain(fixtureIndex(), &domainConversions{Domain: "fixture"})
		require.Error(t, err)
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

func TestWrap(T *testing.T) {
	T.Parallel()

	T.Run("breaks on word boundaries within the width", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"one two", "three", "four five"}, wrap("one two three four five", 9))
	})

	T.Run("an empty reason renders no comment at all", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, wrap("", 80))
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
