package grocerylistpreparation

import (
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterGroceryListCreator registers the grocery list creator with the injector.
func RegisterGroceryListCreator(i do.Injector) {
	do.Provide[GroceryListCreator](i, func(i do.Injector) (GroceryListCreator, error) {
		return NewGroceryListCreator(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
		), nil
	})
}
