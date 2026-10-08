package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratedGRPCConvertersAreUpToDate(T *testing.T) {
	T.Parallel()

	// The gRPC axis's half of the staleness check: a message or a domain type that has grown a
	// field without `make converters` is a field that silently stops crossing the wire.
	T.Run("the generated file matches the messages and the domain types", func(t *testing.T) {
		t.Parallel()

		rendered, err := generateGRPC(backendRoot, realIndex(t))
		require.NoError(t, err)

		onDisk, err := os.ReadFile(filepath.Join(backendRoot, grpcGeneratedPath()))
		require.NoError(t, err)

		assert.Equal(t, string(rendered), string(onDisk), "gRPC converters are stale; run `make converters`")
	})
}

func TestGRPCDeclarations(T *testing.T) {
	T.Parallel()

	declaredNames := func() map[string]grpcDeclaration {
		names := map[string]grpcDeclaration{}
		for _, declared := range grpcConversionList() {
			names[declared.Name()] = declared
		}

		return names
	}

	T.Run("no conversion is declared twice", func(t *testing.T) {
		t.Parallel()

		seen := map[string]struct{}{}
		for _, declared := range grpcConversionList() {
			_, duplicate := seen[declared.Name()]
			assert.False(t, duplicate, "%s is declared twice", declared.Name())
			seen[declared.Name()] = struct{}{}
		}
	})

	// An exception that names nothing is worse than no exception: the field it was meant to
	// answer quietly goes back to whatever the derivation makes of it.
	T.Run("every exception names a conversion that is generated", func(t *testing.T) {
		t.Parallel()

		names := declaredNames()

		for name := range grpcFieldExceptions {
			_, declared := names[name]
			_, manual := grpcHandWritten[name]

			assert.True(t, declared && !manual, "%s has field exceptions but is not generated", name)
		}
	})

	T.Run("every hand-written entry names a declared conversion", func(t *testing.T) {
		t.Parallel()

		names := declaredNames()

		for name := range grpcHandWritten {
			_, declared := names[name]
			assert.True(t, declared, "%s is listed as hand-written but not declared", name)
		}
	})

	T.Run("a hand-written conversion is actually hand-written", func(t *testing.T) {
		t.Parallel()

		source, err := os.ReadFile(filepath.Join(backendRoot, grpcConvertersDir, "converters_manual.go"))
		require.NoError(t, err)

		for name := range grpcHandWritten {
			assert.Contains(t, string(source), "\nfunc "+name+"(", "%s is not generated and not written by hand either", name)
		}
	})

	T.Run("every reason says something", func(t *testing.T) {
		t.Parallel()

		for name, fields := range grpcFieldExceptions {
			for field, rule := range fields {
				assert.NotEmpty(t, rule.why, "%s: %s", name, field)
			}
		}

		for name, why := range grpcHandWritten {
			assert.NotEmpty(t, why, name)
		}
	})
}

func TestGRPCDeclarationNames(T *testing.T) {
	T.Parallel()

	T.Run("names the message side with a GRPC prefix in both directions", func(t *testing.T) {
		t.Parallel()

		declarations := both("Widget")

		assert.Equal(t, "ConvertWidgetToGRPCWidget", declarations[0].Name())
		assert.Equal(t, "ConvertGRPCWidgetToWidget", declarations[1].Name())
		assert.Equal(t, "ConvertWidgetToGRPCWidgetSummary", toGRPCAs("Widget", "WidgetSummary")[0].Name())
		assert.Equal(t, "ConvertGRPCWidgetSummaryToWidget", fromGRPCAs("WidgetSummary", "Widget")[0].Name())
	})
}

func TestQualifyType(T *testing.T) {
	T.Parallel()

	T.Run("qualifies the package's own types and nothing else", func(t *testing.T) {
		t.Parallel()

		for input, expected := range map[string]string{
			"string":                 "string",
			"*Recipe":                "*pkg.Recipe",
			"[]*RecipeStep":          "[]*pkg.RecipeStep",
			"time.Time":              "time.Time",
			"*timestamppb.Timestamp": "*timestamppb.Timestamp",
			"map[string]Recipe":      "map[string]pkg.Recipe",
		} {
			assert.Equal(t, expected, qualifyType(input, "pkg"), input)
		}
	})
}

// fixtureGRPCIndex is a domain type and a message small enough to state expectations about.
func fixtureGRPCIndex() *grpcIndex {
	domainTypes := map[string]*structType{
		"Thing": newStructTypeFromFields("Thing",
			structField{Name: "ID", Type: "string"},
			structField{Name: "CreatedAt", Type: "time.Time"},
			structField{Name: "ArchivedAt", Type: "*time.Time"},
			structField{Name: "Count", Type: "uint16"},
			structField{Name: "Kind", Type: "string"},
			structField{Name: "OptionalKind", Type: "*string"},
			structField{Name: "Owner", Type: "Part"},
			structField{Name: "Maybe", Type: "*Part"},
			structField{Name: "Parts", Type: "[]*Part"},
			structField{Name: "Note", Type: "*string"},
			structField{Name: "BelongsToUser", Type: "string"},
		),
		"Part": newStructTypeFromFields("Part", structField{Name: "ID", Type: "string"}),
	}

	messageTypes := map[string]*structType{
		"Thing": newStructTypeFromFields("Thing",
			structField{Name: "Id", Type: "string"},
			structField{Name: "CreatedAt", Type: "*timestamppb.Timestamp"},
			structField{Name: "ArchivedAt", Type: "*timestamppb.Timestamp"},
			structField{Name: "Count", Type: "uint32"},
			structField{Name: "Kind", Type: "Kind"},
			structField{Name: "OptionalKind", Type: "*Kind"},
			structField{Name: "OwnerId", Type: "string"},
			structField{Name: "Owner", Type: "*Part"},
			structField{Name: "Maybe", Type: "*Part"},
			structField{Name: "Parts", Type: "[]*Part"},
			structField{Name: "Note", Type: "*string"},
			structField{Name: "ByUser", Type: "string"},
		),
		"Part": newStructTypeFromFields("Part", structField{Name: "Id", Type: "string"}),
	}

	index := &grpcIndex{
		domain:   map[string]*structType{},
		messages: map[string]*structType{},
		enums:    map[string]struct{}{grpcMessagesQualifier + ".Kind": {}},
	}

	for name, declared := range domainTypes {
		index.domain[name] = qualifyStruct(declared, grpcDomain)
	}

	for name, declared := range messageTypes {
		index.messages[name] = qualifyStruct(declared, grpcMessagesQualifier)
	}

	return index
}

func fixtureGRPCPlanner() *grpcPlanner {
	var declarations []grpcDeclaration
	declarations = append(declarations, both("Thing")...)
	declarations = append(declarations, both("Part")...)

	return newGRPCPlanner(fixtureGRPCIndex(), declarations)
}

func TestGRPCDerivation(T *testing.T) {
	T.Parallel()

	toMessage := toGRPC("Thing")[0]
	fromMessage := fromGRPC("Thing")[0]

	T.Run("carries every field across without being told to", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixtureGRPCPlanner().Plan(toMessage, map[string]Rule{
			"ByUser": From("BelongsToUser", "renamed"),
		})
		require.NoError(t, err)

		prelude := strings.Join(resolved.Prelude, "\n")

		// An initialism is matched regardless of case, because that is how protoc-gen-go
		// spells one.
		assert.Equal(t, "x.ID", expressionFor(resolved, "Id"))
		assert.Equal(t, "converters.ConvertTimeToPBTimestamp(x.CreatedAt)", expressionFor(resolved, "CreatedAt"))
		assert.Equal(t, "converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt)", expressionFor(resolved, "ArchivedAt"))
		assert.Equal(t, "uint32(x.Count)", expressionFor(resolved, "Count"))
		assert.Equal(t, "ConvertStringToKind(x.Kind)", expressionFor(resolved, "Kind"))
		assert.Equal(t, "x.Owner.ID", expressionFor(resolved, "OwnerId"))
		assert.Equal(t, "ConvertPartToGRPCPart(&x.Owner)", expressionFor(resolved, "Owner"))
		assert.Equal(t, "x.Note", expressionFor(resolved, "Note"))
		assert.Equal(t, "x.BelongsToUser", expressionFor(resolved, "ByUser"))

		// An optional enum is converted behind a nil check into a new pointer.
		assert.Equal(t, "optionalKind", expressionFor(resolved, "OptionalKind"))
		assert.Contains(t, prelude, "optionalKind = new(ConvertStringToKind(*x.OptionalKind))")

		// A nested message held by pointer on both sides is guarded.
		assert.Equal(t, "maybe", expressionFor(resolved, "Maybe"))
		assert.Contains(t, prelude, "if x.Maybe != nil {\nmaybe = ConvertPartToGRPCPart(x.Maybe)\n}")

		// A collection starts nil, so that an empty one is null on the wire.
		assert.Equal(t, "parts", expressionFor(resolved, "Parts"))
		assert.Contains(t, prelude, "var parts []*mealplanningsvc.Part\nfor _, item := range x.Parts {\nparts = append(parts, ConvertPartToGRPCPart(item))\n}")
	})

	T.Run("reads a nested message the domain holds by value straight through", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixtureGRPCPlanner().Plan(fromMessage, map[string]Rule{
			"BelongsToUser": From("ByUser", "renamed"),
		})
		require.NoError(t, err)

		assert.Equal(t, "*ConvertGRPCPartToPart(x.Owner)", expressionFor(resolved, "Owner"))
		assert.Equal(t, "uint16(x.Count)", expressionFor(resolved, "Count"))
		assert.Equal(t, "ConvertKindToString(x.Kind)", expressionFor(resolved, "Kind"))
		assert.Equal(t, "converters.ConvertPBTimestampToTime(x.CreatedAt)", expressionFor(resolved, "CreatedAt"))
	})

	T.Run("the exceptions preserve what the derivation would have corrected", func(t *testing.T) {
		t.Parallel()

		resolved, err := fixtureGRPCPlanner().Plan(toMessage, map[string]Rule{
			"ByUser": From("BelongsToUser", "renamed"),
			"Maybe":  Unguarded("read straight through"),
			"Parts":  EmptySlice("empty, not null"),
			"Note":   Detached("a copy, not the same pointer"),
			"Count":  Skip("left unset"),
		})
		require.NoError(t, err)

		prelude := strings.Join(resolved.Prelude, "\n")

		assert.Equal(t, "ConvertPartToGRPCPart(x.Maybe)", expressionFor(resolved, "Maybe"))
		assert.Contains(t, prelude, "parts := []*mealplanningsvc.Part{}")
		assert.Contains(t, prelude, "if x.Note != nil {\nnote = new(*x.Note)\n}")

		for _, assigned := range resolved.Assignments {
			if assigned.Field == "Count" {
				assert.True(t, assigned.Skipped)
				assert.Equal(t, "left unset", assigned.Why)
			}
		}
	})

	T.Run("a field nothing answers fails the build rather than going quiet", func(t *testing.T) {
		t.Parallel()

		_, err := fixtureGRPCPlanner().Plan(toMessage, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ByUser")
	})

	T.Run("an exception for a field the destination lacks is refused", func(t *testing.T) {
		t.Parallel()

		_, err := fixtureGRPCPlanner().Plan(toMessage, map[string]Rule{
			"ByUser":  From("BelongsToUser", "renamed"),
			"Renamed": Skip("this field was renamed away"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Renamed")
	})

	T.Run("an exception that does not fit its field is refused", func(t *testing.T) {
		t.Parallel()

		_, err := fixtureGRPCPlanner().Plan(toMessage, map[string]Rule{
			"ByUser": From("BelongsToUser", "renamed"),
			"Count":  EmptySlice("Count is not a collection"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Count")
	})
}

func TestLocalName(T *testing.T) {
	T.Parallel()

	T.Run("never collides with a keyword or a name the function already uses", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "typeValue", localName("Type"))
		assert.Equal(t, "itemValue", localName("Item"))
		assert.Equal(t, "xValue", localName("X"))
		assert.Equal(t, "purchasedMeasurementUnit", localName("PurchasedMeasurementUnit"))
	})
}
