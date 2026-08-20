package fakes

import (
	"fmt"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeMedia builds a faked piece of recipe media.
func BuildFakeRecipeMedia() *types.RecipeMedia {
	media := fake.BuildFakeRecord[types.RecipeMedia]()

	// A MIME type the uploads path recognizes, and a path that ends in an extension
	// matching it — the two are read together when the file is served.
	media.MimeType = gofakeit.FileMimeType()
	media.InternalPath = fmt.Sprintf("%s.%s", fake.BuildFakePassword(), gofakeit.FileExtension())

	// Empty until the media has been published somewhere it can be fetched from.
	media.ExternalPath = ""

	return media
}

// BuildFakeRecipeMediaList builds a faked RecipeMediaList.
func BuildFakeRecipeMediaList() *filtering.QueryFilteredResult[types.RecipeMedia] {
	return fake.BuildFakePage(BuildFakeRecipeMedia)
}

// BuildFakeRecipeMediaUpdateRequestInput builds a faked RecipeMediaUpdateRequestInput from a piece of recipe media.
func BuildFakeRecipeMediaUpdateRequestInput() *types.RecipeMediaUpdateRequestInput {
	media := BuildFakeRecipeMedia()

	return &types.RecipeMediaUpdateRequestInput{
		BelongsToRecipe:     media.BelongsToRecipe,
		BelongsToRecipeStep: media.BelongsToRecipeStep,
		MimeType:            &media.MimeType,
		InternalPath:        &media.InternalPath,
		ExternalPath:        &media.ExternalPath,
	}
}

// BuildFakeRecipeMediaCreationRequestInput builds a faked RecipeMediaCreationRequestInput.
func BuildFakeRecipeMediaCreationRequestInput() *types.RecipeMediaCreationRequestInput {
	media := BuildFakeRecipeMedia()

	return converters.ConvertRecipeMediaToRecipeMediaCreationRequestInput(media)
}
