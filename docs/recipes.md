# Recipe Object Documentation

This document describes all the fields of the `Recipe` object and their purposes within the meal planning system.

## If you're new here

The `Recipe` object is the central entity in the meal planning system. It represents a complete cooking recipe that can be used for meal planning, grocery list generation, and cooking guidance. Key concepts to understand:

- **Recipe vs. Ingredient Distinction**: Recipes are multi-step cooking processes, while ingredients are raw materials. The system prevents single-step "recipes" (like "diced onion") but allows proper preparation recipes (like "caramelized onions").
- **Meal Component Types**: Recipes are categorized by what type of meal component they produce (appetizer, main course, dessert, etc.) to help with meal planning.
- **Recipe Relationships**: Recipes can be cloned/inspired by other recipes, maintaining provenance while allowing customization.
- **Scaling Support**: The system supports dynamic scaling of recipes through the frontend interface.
- **Preparation vs. Cooking**: Recipes separate prep tasks (advance preparation) from cooking steps (immediate cooking process).

## URL and Routing Fields

### `Slug` (string)

- **Purpose**: URL-friendly version of the recipe name
- **Usage**: Intended for SEO-friendly routing and frontend interfaces
- **Example**: "chocolate-chip-cookies", "beef-stir-fry"
- **Note**: Currently not implemented - no way to fetch recipes via slug. This field is subject to either enhancement or deprecation, but not really acceptable as-is.

## Metadata Fields

### `Description` (string)

- **Purpose**: Detailed description of the recipe
- **Usage**: Provides context about the dish, its origins, or special characteristics
- **Example**: "A classic American cookie recipe with crispy edges and chewy centers"

### `Source` (string)

- **Purpose**: Attribution or origin of the recipe
- **Usage**: Credits the original creator, cookbook, or website
- **Example**: "Grandma's Cookbook", "Food Network", "AllRecipes.com"

### `CreatedByUser` (string)

- **Purpose**: ID of the user who created this recipe
- **Usage**: Used for ownership tracking and user-specific recipe management

## Timing and Lifecycle Fields

### `CreatedAt` (time.Time)

- **Purpose**: Timestamp when the recipe was first created
- **Usage**: Audit trail, sorting, and display purposes

### `LastUpdatedAt` (*time.Time)

- **Purpose**: Timestamp of the most recent modification
- **Usage**: Tracks when recipe was last edited, can be null if never updated

### `ArchivedAt` (*time.Time)

- **Purpose**: Timestamp when the recipe was archived/deleted
- **Usage**: Soft delete functionality, can be null if recipe is active

## Portion and Yield Fields

### `MinEstimatedPortions` / `MaxEstimatedPortions`

- **Purpose**: Range of how many portions/servings the recipe yields
- **Structure**:
  - `MinEstimatedPortions` (float32): Minimum number of portions (required)
  - `MaxEstimatedPortions` (*float32): Maximum number of portions (optional)
- **Usage**: Helps users understand recipe scale and plan for different group sizes
- **Logic**: If `MaxEstimatedPortions` is null, serves exactly `MinEstimatedPortions` people. If `MaxEstimatedPortions` is present, serves `Min` to `Max` people
- **Example**: Min 4, Max 6 (serves 4-6 people) or Min 4, Max null (serves exactly 4 people)

### `PortionName` (string)

- **Purpose**: Singular name for what the recipe yields
- **Usage**: Used in UI to describe individual servings
- **Example**: "serving", "cookie", "slice", "bowl"

### `PluralPortionName` (string)

- **Purpose**: Plural name for what the recipe yields
- **Usage**: Used in UI when referring to multiple servings
- **Example**: "servings", "cookies", "slices", "bowls"

## Recipe Classification Fields

### `YieldsComponentType` (string)

- **Purpose**: Categorizes what type of meal component this recipe produces
- **Valid Values**:
  - `"unspecified"` - No specific category
  - `"amuse-bouche"` - Small appetizer
  - `"appetizer"` - Starter course
  - `"soup"` - Soup course
  - `"main"` - Main course
  - `"salad"` - Salad course
  - `"beverage"` - Drink
  - `"side"` - Side dish
  - `"dessert"` - Dessert course
- **Usage**: Helps organize recipes in meal planning and determines how they can be used in meals
- **Note**: These values are defined as constants in [meal.go](../internal/domain/mealplanning/meal.go) and are used throughout the meal planning system

### `EligibleForMeals` (bool)

- **Purpose**: Whether this recipe can be included in new meals
- **Use Case**: Intended as a soft deletion mechanism for recipes currently in use. Allows admins to prevent new usage of a recipe while keeping existing meals functional, giving time to replace the recipe in existing meal plans
- **Note**: Recipes for individual ingredients should be marked as ineligible for meals. **This functionality is largely unimplemented** - the system doesn't currently enforce this flag when creating new meals

## Quality and Approval Fields

### `SealOfApproval` (bool)

- **Purpose**: Indicates if the recipe has been reviewed and approved by service operators
- **Usage**: Intended for highlighting "winner" recipes to new users
- **Note**: Currently unused - this field was added prematurely for future functionality

## Recipe Inspiration

### `InspiredByRecipeID` (*string)

- **Purpose**: References another recipe that inspired this one
- **Usage**: Enables easy recipe cloning and modification while maintaining provenance
- **Note**: If null or blank, the recipe is considered "original". Used when users want to make minor adjustments to existing recipes (e.g., reducing garlic in a soup recipe)

## Recipe Relationships

### `AssociatedRecipes` ([]*[Recipe](../internal/domain/mealplanning/recipe.go))

- **Purpose**: Contains recipes that are referenced as components within this recipe's steps
- **Usage**: Automatically populated when fetching a recipe from the database. Provides complete recipe information for recipes that use other recipes as ingredients
- **How It Works**:
  - When a recipe step ingredient references a product from another recipe via `RecipeStepProductRecipeID`, that other recipe is considered an "associated recipe"
  - The system recursively loads nested associated recipes (if recipe A uses recipe B, and recipe B uses recipe C, then recipe A will have both B and C in its `AssociatedRecipes`)
  - **Flattening**: All associated recipes are returned in a **flat list** at the root level. If you have a recipe chain (e.g., Salad uses Croutons, which uses Infused Olive Oil), fetching the Salad recipe will return both Croutons and Infused Olive Oil in a single flat list
  - **Empty AssociatedRecipes**: Each recipe in the `AssociatedRecipes` array will have an **empty `AssociatedRecipes` array** - the flattening process clears nested associations to prevent deep nesting
  - Duplicates are automatically filtered out
  - The recipe itself is never included in its own `AssociatedRecipes`
- **Example Use Cases**:
  - A "Caesar Salad" recipe might use "Caesar Dressing" as an ingredient, so "Caesar Dressing" would appear in the salad's `AssociatedRecipes`
  - A "Chicken Parmesan" recipe might use "Marinara Sauce" as an ingredient, so "Marinara Sauce" would appear in the chicken's `AssociatedRecipes`
  - A "Salad" recipe uses "Croutons", which uses "Infused Olive Oil". Fetching the Salad recipe returns both Croutons and Infused Olive Oil in a flat list, and both will have empty `AssociatedRecipes` arrays
- **Note**: This field is read-only and automatically populated by the repository layer. It cannot be set directly when creating or updating recipes. The relationship is established by setting `RecipeStepProductRecipeID` on recipe step ingredients, and a relationship that would lead back to the recipe is refused on write (see [Validation Rules](#validation-rules))

## Content Fields

### `Steps` ([]*[RecipeStep](../internal/domain/mealplanning/recipe_step.go))

- **Purpose**: Ordered list of cooking steps for the recipe
- **Usage**: Contains the actual cooking instructions and process
- **Note**: Must have at least 2 steps to prevent overly granular "recipes" (like "diced onion") while allowing proper prepared ingredient recipes (like "caramelized onions")
- **Step Requirements**: Each step must have at least one instrument OR vessel (not necessarily both)
  - **Vessels**: Identify what the preparation happens in (cutting board for chopping, pot for boiling)
  - **Instruments**: Tools used for the preparation (knife for chopping, spoon for stirring)
  - **Special Cases**: Some steps may only have instruments (cleaning a knife) or only vessels (preheating a pan)

### `PrepTasks` ([]*[RecipePrepTask](../internal/domain/mealplanning/recipe_prep_tasks.go))

- **Purpose**: Preparation tasks that can be done ahead of time
- **Usage**: Helps with meal planning by identifying advance preparation opportunities
- **Note**: These are separate from the main cooking steps

### `Media` ([]*[RecipeMedia](../internal/domain/mealplanning/recipe_media.go))

- **Purpose**: Images, videos, or other media associated with the recipe
- **Usage**: Can be attached to both the recipe level and individual steps
- **Note**: Currently a nascent concept. Planned to abstract to a general `UploadedMedia` type supporting images and videos. Recipe-level media might include finished product photos or full video tutorials, while step-level media provides preparation guidance

## Recipe Step Products

Recipe steps produce **products** - the outputs of each step that can be used in subsequent steps or as final outputs. Products can be either **discrete** (countable items like patties, cookies, or slices) or **continuous** (bulk quantities like sauce, liquid, or powder). Products can also be **vessels** or **instruments** that are passed between steps.

### Discrete Products

Discrete products represent countable items where the count should scale independently from the per-item measurement.

**Fields:**

- `MinItemQuantity` / `MaxItemQuantity` (optional float32): The count of discrete items (e.g., 4 patties, 12 cookies, 8 slices)
- `MinMeasurementQuantity` / `MaxMeasurementQuantity` (optional float32): The weight/volume measurement **per item** (e.g., 4 ounces per patty)
- `MeasurementUnit` (ValidMeasurementUnit): The unit for the per-item measurement (e.g. ounce)

**Example - Discrete Product:**

```json
{
  "name": "beef patties",
  "type": "ingredient",
  "minItemQuantity": 4,
  "minMeasurementQuantity": 4,
  "measurementUnitId": "ounce"
}
```

This represents "4 patties, each 4 ounces" (16 ounces total).

**When to Use:**

- Items that should scale by count (patties, cookies, slices, pieces)
- When the per-item size should remain constant when scaling
- Example: A recipe that divides 16 ounces of meat into 4 patties of 4 ounces each

### Continuous Products

Continuous products represent bulk quantities where the total amount scales proportionally.

**Fields:**

- `MinItemQuantity` / `MaxItemQuantity` (optional float32): Both unset/null - indicates continuous product
- `MinMeasurementQuantity` / `MaxMeasurementQuantity` (optional float32): The **total** weight/volume quantity (e.g., 16 ounces of sauce)
- `MeasurementUnit` (ValidMeasurementUnit): The unit for the total measurement

**Example - Continuous Product:**

```json
{
  "name": "sauce",
  "type": "ingredient",
  "minMeasurementQuantity": 16,
  "measurementUnitId": "ounce"
}
```

This represents "16 ounces of sauce" (total quantity). Note: `minItemQuantity` and `maxItemQuantity` are both omitted, signalling a continuous product.

**When to Use:**

- Bulk quantities (sauces, liquids, powders, mixtures)
- When the total quantity should scale proportionally
- Example: A recipe that produces 2 cups of sauce

### Determining Product Type

A product is **discrete** if `MinItemQuantity` or `MaxItemQuantity` is set (not null). A product is **continuous** if both `MinItemQuantity` and `MaxItemQuantity` are null/omitted.

### Vessel Product Chaining

When a vessel (like a saucepan, pot, or skillet) is used throughout multiple steps of a recipe, it should be **chained** through the steps rather than having one step's vessel product used in multiple subsequent steps.

**Pattern:**

1. The first step that uses the vessel should reference it via `ValidPreparationVesselID` and output it as a vessel product
2. Each subsequent step that uses the same vessel should:
   - Consume the vessel product from the immediately previous step (using `ProductOfRecipeStepIndex` and `ProductOfRecipeStepProductIndex`)
   - Output a new vessel product for the next step to use

**Example - Simple White Rice Recipe:**

```text
Step 1 (Simmer): 
  - Uses ValidPreparationVesselID for "2-quart saucepan"
  - Outputs: "2-quart saucepan" (vessel product, index 1)

Step 2 (Stir):
  - Uses vessel product from Step 1
  - Outputs: "2-quart saucepan" (vessel product, index 1)

Step 3 (Cover):
  - Uses vessel product from Step 2
  - Outputs: "2-quart saucepan" (vessel product, index 1)

Step 4 (Remove from heat):
  - Uses vessel product from Step 3
  - Outputs: "2-quart saucepan" (vessel product, index 1)

Step 5 (Rest):
  - Uses vessel product from Step 4
  - Outputs: "2-quart saucepan" (vessel product, index 1)

Step 6 (Fluff):
  - Uses vessel product from Step 5
  - (No output needed - final step)
```

This creates a linear chain where each step's vessel product is used in exactly one subsequent step, making the recipe flow clear and preventing confusion about vessel state.

**When to Use Multiple References:**

A vessel product should **rarely** be used in more than one subsequent step. This pattern would only be appropriate in exceptional cases such as:

- **Parallel Processing**: When a vessel's contents are divided and used in multiple parallel steps (e.g., a pot of stock where some is used for soup and some for sauce in separate steps)
- **Temporary Storage**: When a vessel serves as temporary storage while other work happens, then is used again later (e.g., a baking sheet that holds prepped ingredients, then is used again for baking)
- **Shared Container**: When multiple steps need to access the same container simultaneously (e.g., a shared mixing bowl used by multiple people)

In practice, these cases are extremely rare. The vast majority of recipes should use the chaining pattern described above.

## Option Groups (Alternative Ingredients, Instruments, and Vessels)

Recipe steps can include **option groups** - sets of alternative items where any one can be used. This is useful when a recipe allows substitutions, such as using butter or margarine, or using either a stand mixer or hand mixer.

### How Option Groups Work

Option groups are defined using two key fields on recipe step ingredients, instruments, and vessels:

- **`Index`** (uint16): Identifies the position of this item within the step. Items with the same `Index` belong to the same option group.
- **`OptionIndex`** (uint16): The position within the option group. `OptionIndex: 0` is the primary/default option, `OptionIndex: 1` is the first alternative, etc.

### Option Group Structure

Consider a recipe step that allows either butter or margarine:

```text
Step 1: Mix ingredients
  Ingredients:
    - Butter     (Index: 0, OptionIndex: 0)  ─┐
    - Margarine  (Index: 0, OptionIndex: 1)  ─┴── Option Group at Index 0
    - Flour      (Index: 1, OptionIndex: 0)  ─── Regular ingredient (no alternatives)
    - Sugar      (Index: 2, OptionIndex: 0)  ─── Regular ingredient (no alternatives)
```

In this example:

- **Butter and Margarine** share `Index: 0`, making them alternatives for the same ingredient slot
- **Butter** has `OptionIndex: 0`, making it the default choice
- **Margarine** has `OptionIndex: 1`, making it the first alternative
- **Flour and Sugar** are regular ingredients with no alternatives (each has a unique `Index`)

### Creating Option Groups

When creating a recipe, define option groups by setting the same `Index` for alternative items:

```json
{
  "steps": [
    {
      "preparationId": "prep-mix-id",
      "index": 0,
      "ingredients": [
        {
          "name": "butter",
          "index": 0,
          "optionIndex": 0,
          "validIngredientPreparationId": "vip-butter-id",
          "validIngredientMeasurementUnitId": "vimu-butter-tbsp-id",
          "quantity": { "min": 4 }
        },
        {
          "name": "margarine",
          "index": 0,
          "optionIndex": 1,
          "validIngredientPreparationId": "vip-margarine-id",
          "validIngredientMeasurementUnitId": "vimu-margarine-tbsp-id",
          "quantity": { "min": 4 }
        },
        {
          "name": "flour",
          "index": 1,
          "optionIndex": 0,
          "validIngredientPreparationId": "vip-flour-id",
          "validIngredientMeasurementUnitId": "vimu-flour-cup-id",
          "quantity": { "min": 2 }
        }
      ]
    }
  ]
}
```

### Option Groups for Instruments and Vessels

The same pattern applies to instruments and vessels:

**Instrument Option Group Example:**

```json
{
  "instruments": [
    {
      "name": "stand mixer",
      "index": 0,
      "optionIndex": 0,
      "validPreparationInstrumentId": "vpi-standmixer-id"
    },
    {
      "name": "hand mixer",
      "index": 0,
      "optionIndex": 1,
      "validPreparationInstrumentId": "vpi-handmixer-id"
    }
  ]
}
```

**Vessel Option Group Example:**

```json
{
  "vessels": [
    {
      "name": "Dutch oven",
      "index": 0,
      "optionIndex": 0,
      "validPreparationVesselId": "vpv-dutchoven-id"
    },
    {
      "name": "slow cooker",
      "index": 0,
      "optionIndex": 1,
      "validPreparationVesselId": "vpv-slowcooker-id"
    }
  ]
}
```

### Unique Constraint

Each combination of `(recipe_step_id, index, option_index)` must be unique. This constraint is enforced at the database level for ingredients, instruments, and vessels.

### Integration with Meal Planning

When recipes with option groups are included in meal plans, users can specify which alternative they prefer using **selections**. See [Meal Planning - Recipe Option Selections](meal_planning.md#recipe-option-selections) for details on:

- How selections are specified during meal plan creation
- How selections affect grocery list generation
- Default behavior when no selection is made (uses `OptionIndex: 0`)

### Best Practices

1. **Primary Option First**: Always set the most common/recommended option as `OptionIndex: 0`
2. **Similar Quantities**: Alternatives in the same option group should typically require similar quantities
3. **Appropriate Substitutions**: Only group items that are true substitutes (e.g., don't group "butter" with "olive oil" unless the recipe works equally well with both)
4. **Document Differences**: Use the `Notes` field on ingredients to explain when one alternative might be preferred over another

### When to Use Option Groups vs. Clone the Recipe

Option groups are intended for **simple substitutions that don't change how you prepare the recipe**. The cooking process should remain essentially the same regardless of which option is selected.

**Good candidates for option groups:**

- Sugar or honey (both sweeteners, same technique)
- Fish sauce or soy sauce (both umami sources, same usage)
- Butter or margarine (same role in the recipe)
- Stand mixer or hand mixer (same mixing action)
- Dutch oven or slow cooker (for recipes where either works identically)

**Not appropriate for option groups:**

- Deep fried or baked (completely different techniques and steps)
- Fresh pasta or dried pasta (different cooking times and possibly different steps)
- Bone-in or boneless chicken (may require different cooking times or techniques)
- Any substitution that would add, remove, or significantly modify recipe steps

**Rule of thumb**: If selecting a different option would require you to add a step, skip a step, or significantly change the instructions for a step, that's not an option group - that's a different recipe. Use the [recipe cloning workflow](#recipe-cloning-workflow) to create a variant instead.

For example, if a recipe says "you can use canned beans (skip step 1) or dried beans (soak overnight in step 1)", these should be two separate recipes, not options within the same recipe. The preparation process is fundamentally different.

## Recipe Scaling

The system supports dynamic recipe scaling through the frontend interface. Users can adjust the scale of a recipe (e.g., 2x, 0.5x) which automatically multiplies ingredient quantities in all steps. For example, if a step calls for 1 clove of garlic and the recipe is scaled to 2x, it will display 2 cloves of garlic.

### Scaling Behavior

**Ingredients:**

- Ingredient quantities are multiplied by the scale factor
- Example: 1 clove of garlic at 2x scale = 2 cloves of garlic

**Products - Discrete:**

- `ItemQuantity` (count) is multiplied by the scale factor
- `MeasurementQuantity` (per-item measurement) **remains constant**
- Example: 4 patties (4 oz each) at 2x scale = 8 patties (still 4 oz each)
- This ensures that when scaling a recipe, you get more items of the same size, not larger items

**Products - Continuous:**

- `MeasurementQuantity` (total quantity) is multiplied by the scale factor
- Example: 16 ounces of sauce at 2x scale = 32 ounces of sauce
- This is the traditional scaling behavior for bulk quantities

### Scaling Examples

**Discrete Product Scaling:**

```text
Original: 4 patties, each 4 ounces (16 oz total)
Scale 2x:  8 patties, each 4 ounces (32 oz total)
           ↑ count doubles    ↑ per-item stays same
```

**Continuous Product Scaling:**

```text
Original: 16 ounces of sauce
Scale 2x:  32 ounces of sauce
           ↑ total quantity doubles
```

**Future Enhancement**: The system may support step-level scaling modifiers, allowing certain ingredients to scale at different rates (e.g., garlic scaling at 75% of the overall scale).

## Search and Discovery

Recipe search is currently implemented through a `RecipeSearchSubset` type that provides name-based searching. This is a basic implementation that may be enhanced in the future.

## Additional Notes

- **Recipe Relationships**: Recipes can be assigned as components into meals, but meal planning functionality is documented separately
- **Portion Ranges**: The min/max portion ranges serve as guidelines rather than strict requirements, reflecting the belief that cooking is inherently imprecise
- **Step Validation**: Each recipe step has its own validation logic enforced through typing (e.g., Preparation field is required)

## Common Patterns and Usage

### Creating a New Recipe

1. Set basic metadata (name, description, source)
2. Define portion information (estimated portions, portion names)
3. Set meal component type and eligibility
4. Add at least 2 cooking steps with ingredients and instruments and/or vessels
5. Optionally add prep tasks for advance preparation
6. Add media attachments if available

#### Bridge Table ID Requirements

When creating recipe steps, you must use **bridge table IDs** to specify ingredients, instruments, and vessels. These bridge tables define which combinations of entities are valid together.

**For Recipe Step Ingredients:**

- `ValidIngredientPreparationID` (required) - References a `ValidIngredientPreparation` that defines which ingredient can be used with which preparation method
- `ValidIngredientMeasurementUnitID` (required) - References a `ValidIngredientMeasurementUnit` that defines which measurement unit can be used with which ingredient

**For Recipe Step Instruments:**

- `ValidPreparationInstrumentID` (required) - References a `ValidPreparationInstrument` that defines which instrument can be used with which preparation method

**For Recipe Step Vessels:**

- `ValidPreparationVesselID` (required) - References a `ValidPreparationVessel` that defines which vessel can be used with which preparation method

**Exception - Recipe Step Products:**
When an ingredient, instrument, or vessel is the **output of a previous recipe step** (a "recipe step product"), you don't need bridge table IDs. Instead, set `ProductOfRecipeStepIndex` and `ProductOfRecipeStepProductIndex` to reference the previous step's output. This is common for multi-step recipes where one step's output becomes another step's input (e.g., "soaked beans" from step 1 used in step 2).

#### Example Request Format

```json
{
  "name": "Sopa de Frijol",
  "slug": "sopa-de-frijol",
  "yieldsComponentType": "main",
  "portionName": "serving",
  "pluralPortionName": "servings",
  "estimatedPortions": { "min": 4 },
  "steps": [
    {
      "preparationId": "prep-soak-id",
      "notes": "Soak the beans overnight",
      "index": 0,
      "ingredients": [
        {
          "name": "pinto beans",
          "validIngredientPreparationId": "vip-pinto-soak-id",
          "validIngredientMeasurementUnitId": "vimu-pinto-grams-id",
          "quantity": { "min": 500 }
        }
      ],
      "instruments": [
        {
          "name": "large bowl",
          "validPreparationInstrumentId": "vpi-bowl-soak-id"
        }
      ],
      "vessels": [
        {
          "name": "container",
          "validPreparationVesselId": "vpv-container-soak-id"
        }
      ],
      "products": [
        {
          "name": "soaked pinto beans",
          "type": "ingredient",
          "measurementUnitId": "grams-id",
          "measurementQuantity": { "min": 1000 }
        }
      ]
    },
    {
      "preparationId": "prep-cook-id",
      "notes": "Cook the beans",
      "index": 1,
      "ingredients": [
        {
          "name": "soaked pinto beans",
          "productOfRecipeStepIndex": 0,
          "productOfRecipeStepProductIndex": 0,
          "quantity": { "min": 1000 }
        }
      ],
      "instruments": [
        {
          "name": "pot",
          "validPreparationInstrumentId": "vpi-pot-cook-id"
        }
      ],
      "products": [
        {
          "name": "cooked beans",
          "type": "ingredient",
          "measurementUnitId": "grams-id",
          "measurementQuantity": { "min": 1000 }
        }
      ]
    }
  ]
}
```

Note in the example above:

- Step 0's ingredient uses `validIngredientPreparationId` and `validIngredientMeasurementUnitId`
- Step 1's ingredient uses `productOfRecipeStepIndex` and `productOfRecipeStepProductIndex` (referencing step 0's product)

#### Example: Discrete Product (Cheeseburger Patties)

This example shows how to create a discrete product where the count scales independently from the per-item measurement:

```json
{
  "name": "Cheeseburgers",
  "steps": [
    {
      "preparationId": "prep-shape-id",
      "notes": "Shape the meat into patties",
      "index": 0,
      "ingredients": [
        {
          "name": "ground beef",
          "validIngredientPreparationId": "vip-beef-shape-id",
          "validIngredientMeasurementUnitId": "vimu-beef-ounce-id",
          "quantity": { "min": 16 }
        }
      ],
      "instruments": [
        {
          "name": "hands",
          "validPreparationInstrumentId": "vpi-hands-shape-id"
        }
      ],
      "products": [
        {
          "name": "beef patties",
          "type": "ingredient",
          "measurementUnitId": "ounce",
          "itemQuantity": { "min": 4 },
          "measurementQuantity": { "min": 4 }
        }
      ]
    }
  ]
}
```

This creates 4 patties, each 4 ounces (16 ounces total). When the recipe is scaled 2x:

- `itemQuantity` becomes 8 (8 patties)
- `measurementQuantity` stays 4 (still 4 oz per patty)
- Total meat needed: 32 ounces (for 8 patties of 4 oz each)

**Key Points:**

- `itemQuantity` specifies the count of discrete items (4 patties)
- `measurementQuantity` specifies the per-item measurement (4 ounces per patty)
- When scaling, the count multiplies but the per-item size stays constant

### Recipe Cloning Workflow

1. User finds a recipe they like
2. System creates a copy with `InspiredByRecipeID` pointing to original
3. `CreatedByUser` field is changed to the new user
4. All non-foreign-key IDs are regenerated (step IDs, product IDs, etc.)
5. Preparation IDs remain the same (they reference shared preparation methods)
6. User can modify any aspect of the cloned recipe
7. Original recipe remains unchanged

### Validation Rules

- **Minimum Steps**: Recipes must have at least 2 steps
- **Required Fields**: Name, slug, estimated portions, portion names, and yields component type are required
- **Step Requirements**: Each step must have at least one instrument OR vessel (not necessarily both), and a preparation method
- **Component Type**: Must be one of the predefined meal component types
- **Bridge Table ID Requirements**: See below
- **Products flow one way**: Within a recipe, steps must form a DAG (`recipeanalysis`'
  `ValidateRecipeCreationRequestInputIsDAG`). Across recipes, a recipe may not draw a product —
  through `RecipeStepProductRecipeID` — from a recipe that, somewhere down the line, draws on it
  (`mealplanning.CheckRecipeDependencies`). The cross-recipe half is checked on recipe creation,
  cloning, and every recipe step ingredient creation or update that names another recipe.

### Where the Rules Live

Every rule above is the meal planning manager's to enforce
(`internal/domain/mealplanning/managers`). The manager validates the request's shape, then — in
`checkRecipeForCreation`, which creation and cloning share — the rules that need a read: the
bridge rows (through `recipevalidator`) and the cross-recipe dependencies. The repository writes
what it is handed and decides nothing. A refusal wraps `mealplanning.ErrInvalidRecipeInput`, which
reaches a gRPC client as `InvalidArgument`.

### Who May Change a Recipe

Anyone may read a recipe; only its author may change it or anything it is made of. The manager
methods that write to a recipe or its children take the caller's user ID as `ownerID`:

- `UpdateRecipe` and `ArchiveRecipe` bind it in the statement itself (`created_by_user = $n`), so
  a recipe somebody else wrote matches no row.
- Writes to a recipe's steps, prep tasks, and the ingredients, instruments, vessels, products and
  completion conditions of its steps ask the database first (`RecipeIsOwnedBy`, and for anything
  beneath a step, that the step is the recipe's).

Either way a caller who does not own the recipe gets `sql.ErrNoRows`, which reaches a gRPC client
as `NotFound`. The gRPC handlers read the caller from the session and pass it down; they do not
check ownership themselves.

### Bridge Table Validation

The system validates bridge table IDs during recipe creation to ensure data integrity:

**Validation Checks:**

1. **Existence**: The bridge table entry must exist
2. **Preparation Matching**: The bridge table entry's preparation must match the step's preparation
3. **Ingredient Matching**: For `ValidIngredientMeasurementUnit`, the ingredient must match the one from `ValidIngredientPreparation`

**Error Messages:**
Validation errors follow this format:

```text
step {stepIndex} ingredient {ingredientIndex}: {specific error}
step {stepIndex} instrument {instrumentIndex}: {specific error}
step {stepIndex} vessel {vesselIndex}: {specific error}
```

**Example Errors:**

- `step 0 ingredient 0: ValidIngredientPreparation "abc123" not found`
- `step 0 ingredient 0: ValidIngredientPreparation "abc123" is for preparation "chop", but step uses preparation "soak"`
- `step 0 ingredient 0: ValidIngredientMeasurementUnit "xyz789" is for ingredient "flour", but ingredient "sugar" was specified`
- `step 0 instrument 0: ValidPreparationInstrument "def456" is for preparation "chop", but step uses preparation "soak"`

**Recipe Step Products (No Validation):**
Ingredients, instruments, or vessels that come from previous recipe steps (identified by `ProductOfRecipeStepIndex` or `RecipeStepProductID`) skip bridge table validation. Their validity was established when the original product was created.

### Data Flow

- Recipes are created through the API using [`RecipeCreationRequestInput`](../internal/domain/mealplanning/recipe.go)
- Database operations use [`RecipeDatabaseCreationInput`](../internal/domain/mealplanning/recipe.go)
- Updates use [`RecipeUpdateRequestInput`](../internal/domain/mealplanning/recipe.go)
- The system supports soft deletes via the `ArchivedAt` field

## Common Gotchas and Edge Cases

### Recipe vs. Ingredient Confusion

- **Problem**: New developers might try to create single-step "recipes" for basic ingredients
- **Solution**: The 2-step minimum prevents this. Use the ingredient system for raw materials
- **Example**: Don't create a recipe for "1 cup flour" - that's an ingredient. But "caramelized onions" is a valid recipe

### Slug Implementation Status

- **Problem**: Slug field exists but isn't functional
- **Solution**: Don't rely on slug-based routing until it's implemented
- **Workaround**: Use recipe ID for all routing needs
