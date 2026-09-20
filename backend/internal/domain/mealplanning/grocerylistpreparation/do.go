package grocerylistpreparation

import (
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterGroceryListCreator registers the grocery list creator with the injector.
func RegisterGroceryListCreator(i do.Injector) {
	do.Provide[GroceryListCreator](i, func(i do.Injector) (GroceryListCreator, error) {
		return NewGroceryListCreator(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
		), nil
	})
}
