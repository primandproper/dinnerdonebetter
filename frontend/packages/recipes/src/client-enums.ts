/**
 * Client-safe enum constants for recipe creation.
 * Do NOT import from mealplanning_messages or mealplanning_service_types in client code—
 * they transitively import platform's mediaregistry messages, which pull in @grpc/grpc-js (Node-only).
 */
export const RecipeStepProductType = {
  RECIPE_STEP_PRODUCT_TYPE_INGREDIENT: 0,
  RECIPE_STEP_PRODUCT_TYPE_INSTRUMENT: 1,
  RECIPE_STEP_PRODUCT_TYPE_VESSEL: 2,
  UNRECOGNIZED: -1,
} as const;

export type RecipeStepProductType = (typeof RecipeStepProductType)[keyof typeof RecipeStepProductType];

export const MealComponentType = {
  MEAL_COMPONENT_TYPE_UNSPECIFIED: 0,
  MEAL_COMPONENT_TYPE_AMUSE_BOUCHE: 1,
  MEAL_COMPONENT_TYPE_APPETIZER: 2,
  MEAL_COMPONENT_TYPE_SOUP: 3,
  MEAL_COMPONENT_TYPE_MAIN: 4,
  MEAL_COMPONENT_TYPE_SALAD: 5,
  MEAL_COMPONENT_TYPE_BEVERAGE: 6,
  MEAL_COMPONENT_TYPE_SIDE: 7,
  MEAL_COMPONENT_TYPE_DESSERT: 8,
  UNRECOGNIZED: -1,
} as const;

export type MealComponentType = (typeof MealComponentType)[keyof typeof MealComponentType];
