// Package recipeanalysis derives step graphs and timing information from recipes.
package recipeanalysis

//go:generate go tool github.com/matryer/moq -out recipe_analyzer_mock.go -pkg recipeanalysis -rm -fmt goimports . RecipeAnalyzer:RecipeAnalyzerMock
