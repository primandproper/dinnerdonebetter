// Package grocerylistpreparation builds grocery lists from meal plans.
package grocerylistpreparation

//go:generate go tool github.com/matryer/moq -out grocery_list_creator_mock.go -pkg grocerylistpreparation -rm -fmt goimports . GroceryListCreator:GroceryListCreatorMock
