package mealplanning

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spineMethods are the repository's ways onto platform's recording spine; see client.go. A write
// that reaches one of them announces itself on its own transaction.
var spineMethods = map[string]struct{}{
	"withEvent":  {},
	"withRecord": {},
	"record":     {},
	"emit":       {},
	"emitIndex":  {},
}

// unannouncedWrites are the exported writes that deliberately reach no part of the spine, each
// with the reason. An entry here is a decision, so it is held to still being needed: one whose
// method has stopped writing, or has started announcing, fails the sweep until it is removed.
var unannouncedWrites = map[string]string{
	"MarkMealPlanTaskNotificationSent": "delivery bookkeeping the notification worker keeps so it sends once; " +
		"the notification is what the user receives, and the flag describes nothing a subscriber or an investigation reads",
	"ClearMealPlanTaskNotificationSentForEvent": "resets that bookkeeping when a meal plan event is rescheduled; " +
		"the reschedule is UpdateMealPlanEvent or SwapMealPlanEvents, which record and announce it themselves",
	"AttachMealPlanFinalizationSaga": "claims a plan for one finalization saga so no other replica starts a second; " +
		"the claim is the saga's bookkeeping, and what the saga changes is announced by its steps",
	"MarkMealsAsIndexed":                 markedAsIndexed,
	"MarkRecipesAsIndexed":               markedAsIndexed,
	"MarkValidIngredientStatesAsIndexed": markedAsIndexed,
	"MarkValidIngredientsAsIndexed":      markedAsIndexed,
	"MarkValidInstrumentsAsIndexed":      markedAsIndexed,
	"MarkValidMeasurementUnitsAsIndexed": markedAsIndexed,
	"MarkValidPreparationsAsIndexed":     markedAsIndexed,
	"MarkValidVesselsAsIndexed":          markedAsIndexed,
}

// markedAsIndexed is why the indexer's own stamps go unannounced.
const markedAsIndexed = "the search indexer's record of having indexed these rows; the rows did not change, " +
	"and an announcement would be derived into another index event for the rows it just indexed"

// TestEveryWriteIsRecorded holds the claim on repository: every exported method that changes a
// row reaches the recording spine, or is named in unannouncedWrites with a reason.
//
// It reads the package's source rather than running it, because what is being checked is which
// path a write takes, not what one run of it happened to do. A method writes if it, or a
// repository method it reaches, calls a generated query whose SQL is an INSERT, UPDATE or DELETE.
// It announces if it, or a method it reaches, calls one of spineMethods.
func TestEveryWriteIsRecorded(T *testing.T) {
	T.Parallel()

	T.Run("every exported write reaches the spine or is excused", func(t *testing.T) {
		t.Parallel()

		s := sweepRepository(t)

		// A sweep that found nothing to check would pass for the wrong reason.
		require.Contains(t, s.writes, "CreateMealList", "the sweep no longer recognizes a write it should")
		require.Contains(t, s.announces, "CreateMealList", "the sweep no longer recognizes an announcement it should")

		var unannounced []string
		for name := range s.writes {
			if !ast.IsExported(name) {
				continue
			}

			if _, ok := s.announces[name]; ok {
				continue
			}

			if _, excused := unannouncedWrites[name]; excused {
				continue
			}

			unannounced = append(unannounced, name)
		}
		slices.Sort(unannounced)

		assert.Empty(t, unannounced, "these writes announce nothing: route them through withEvent or withRecord, or add them to unannouncedWrites with the reason")
	})

	T.Run("every excused write still needs its excuse", func(t *testing.T) {
		t.Parallel()

		s := sweepRepository(t)

		for name, reason := range unannouncedWrites {
			assert.NotEmpty(t, reason, "%s is excused without a reason", name)
			assert.Contains(t, s.methods, name, "%s is excused but is not a repository method", name)
			assert.Contains(t, s.writes, name, "%s is excused but no longer writes", name)
			assert.NotContains(t, s.announces, name, "%s is excused but now announces itself", name)
		}
	})
}

// sweptMethod is what the sweep reads out of one repository method's body, closures included.
type sweptMethod struct {
	// calls are the repository methods it names on its receiver.
	calls []string
	// queries are the generated queries it calls.
	queries []string
}

// sweep is the package's repository methods, and which of them write and which announce.
type sweep struct {
	methods   map[string]*sweptMethod
	writes    map[string]struct{}
	announces map[string]struct{}
}

func sweepRepository(t *testing.T) *sweep {
	t.Helper()

	mutating := mutatingQueries(t)
	methods := repositoryMethods(t)

	return &sweep{
		methods: methods,
		writes: reachingAny(methods, func(m *sweptMethod) bool {
			return slices.ContainsFunc(m.queries, func(query string) bool {
				_, ok := mutating[query]
				return ok
			})
		}),
		announces: reachingAny(methods, func(m *sweptMethod) bool {
			return slices.ContainsFunc(m.calls, func(callee string) bool {
				_, ok := spineMethods[callee]
				return ok
			})
		}),
	}
}

// repositoryMethods parses this package's non-test source and returns every method on *repository.
func repositoryMethods(t *testing.T) map[string]*sweptMethod {
	t.Helper()

	out := map[string]*sweptMethod{}
	for _, file := range parseGoFiles(t, ".") {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !isRepositoryReceiver(fn.Recv) {
				continue
			}

			receiver := fn.Recv.List[0].Names[0].Name
			method := &sweptMethod{}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				selector, isSelector := node.(*ast.SelectorExpr)
				if !isSelector {
					return true
				}

				switch x := selector.X.(type) {
				case *ast.Ident:
					if x.Name == receiver {
						method.calls = append(method.calls, selector.Sel.Name)
					}
				case *ast.SelectorExpr:
					if ident, isIdent := x.X.(*ast.Ident); isIdent && ident.Name == receiver && x.Sel.Name == "generatedQuerier" {
						method.queries = append(method.queries, selector.Sel.Name)
					}
				}

				return true
			})

			out[fn.Name.Name] = method
		}
	}

	return out
}

func isRepositoryReceiver(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) != 1 || len(recv.List[0].Names) != 1 {
		return false
	}

	star, ok := recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	ident, ok := star.X.(*ast.Ident)

	return ok && ident.Name == "repository"
}

// reachingAny returns the methods for which direct holds, of the method itself or of any
// repository method it reaches.
func reachingAny(methods map[string]*sweptMethod, direct func(*sweptMethod) bool) map[string]struct{} {
	out := map[string]struct{}{}

	for name := range methods {
		seen := map[string]struct{}{}
		queue := []string{name}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if _, visited := seen[current]; visited {
				continue
			}
			seen[current] = struct{}{}

			method, known := methods[current]
			if !known {
				continue
			}

			if direct(method) {
				out[name] = struct{}{}
				break
			}

			queue = append(queue, method.calls...)
		}
	}

	return out
}

var (
	// queryHeader is the line sqlc opens every generated statement with.
	queryHeader = regexp.MustCompile(`\A-- name: (\w+) :\w+\n`)
	// mutatingStatement matches a statement that changes rows: one that is an INSERT, UPDATE or
	// DELETE, or a WITH whose body contains one.
	mutatingStatement = regexp.MustCompile(`(?is)\A\s*(?:INSERT|UPDATE|DELETE)\b|\A\s*WITH\b.*\b(?:INSERT|UPDATE|DELETE)\b`)
)

// mutatingQueries returns the name of every generated query whose SQL changes rows.
func mutatingQueries(t *testing.T) map[string]struct{} {
	t.Helper()

	out := map[string]struct{}{}
	statements := 0

	for _, file := range parseGoFiles(t, "generated") {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}

			for _, spec := range gen.Specs {
				for _, value := range spec.(*ast.ValueSpec).Values {
					literal, isLiteral := value.(*ast.BasicLit)
					if !isLiteral || literal.Kind != token.STRING {
						continue
					}

					sql, err := strconv.Unquote(literal.Value)
					require.NoError(t, err)

					header := queryHeader.FindStringSubmatch(sql)
					if header == nil {
						continue
					}
					statements++

					if mutatingStatement.MatchString(sql[len(header[0]):]) {
						out[header[1]] = struct{}{}
					}
				}
			}
		}
	}

	require.NotZero(t, statements, "no generated statements were found to classify")
	require.NotEmpty(t, out, "no generated statement was classified as a write")

	return out
}

// parseGoFiles parses every non-test Go file directly in dir.
func parseGoFiles(t *testing.T, dir string) []*ast.File {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	fset := token.NewFileSet()
	var files []*ast.File

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, parseErr := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		require.NoError(t, parseErr)

		files = append(files, file)
	}

	require.NotEmpty(t, files, "no Go files in %s", dir)

	return files
}
