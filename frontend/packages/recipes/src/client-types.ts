/**
 * Client-safe types for recipe creation. NO imports from mealplanning or generated code—
 * those pull in @grpc/grpc-js (Node-only) and break in the browser.
 */

import type { MealComponentType, RecipeStepProductType } from './client-enums';

// API response types (from search endpoints)
export interface ValidPreparation {
  id: string;
  name?: string;
  label?: string;
  [key: string]: unknown;
}
export interface ValidIngredient {
  id: string;
  name?: string;
  [key: string]: unknown;
}
export interface ValidIngredientPreparation {
  id: string;
  [key: string]: unknown;
}
export interface ValidIngredientMeasurementUnit {
  id: string;
  measurementUnit?: { id?: string; name?: string };
  [key: string]: unknown;
}
export interface ValidPreparationInstrument {
  id: string;
  [key: string]: unknown;
}
export interface ValidPreparationVessel {
  id: string;
  [key: string]: unknown;
}
export interface ValidMeasurementUnit {
  id: string;
  [key: string]: unknown;
}
export interface ValidIngredientState {
  id: string;
  [key: string]: unknown;
}
export interface ValidVessel {
  id: string;
  [key: string]: unknown;
}

// Recipe creation request types
export interface RecipeStepCompletionConditionCreationRequestInput {
  ingredientStateId: string;
  belongsToRecipeStep: string;
  notes: string;
  ingredients: number[];
  optional: boolean;
}

export interface RecipeStepIngredientCreationRequestInput {
  productOfRecipeStepIndex?: number;
  productOfRecipeStepProductIndex?: number;
  vesselIndex?: number;
  productPercentageToUse?: number;
  recipeStepProductRecipeId?: string;
  recipeStepProductRecipeSlug?: string;
  ingredientNotes: string;
  name: string;
  quantityNotes: string;
  minQuantity: number;
  maxQuantity?: number;
  optionIndex: number;
  optional: boolean;
  toTaste: boolean;
  validIngredientPreparationId?: string;
  validIngredientMeasurementUnitId?: string;
  index?: number;
}

export interface RecipeStepInstrumentCreationRequestInput {
  recipeStepProductId?: string;
  productOfRecipeStepIndex?: number;
  productOfRecipeStepProductIndex?: number;
  name: string;
  notes: string;
  minQuantity: number;
  maxQuantity?: number;
  optionIndex: number;
  optional: boolean;
  preferenceRank: number;
}

export interface RecipeStepProductCreationRequestInput {
  minStorageTemperatureInCelsius?: number;
  maxStorageTemperatureInCelsius?: number;
  minStorageDurationInSeconds?: number;
  maxStorageDurationInSeconds?: number;
  minMeasurementQuantity?: number;
  maxMeasurementQuantity?: number;
  minItemQuantity?: number;
  maxItemQuantity?: number;
  measurementUnitId?: string;
  containedInVesselIndex?: number;
  quantityNotes: string;
  name: string;
  storageInstructions: string;
  type: RecipeStepProductType;
  index: number;
  compostable: boolean;
  isLiquid: boolean;
  isWaste: boolean;
}

export interface RecipeStepVesselCreationRequestInput {
  recipeStepProductId?: string;
  productOfRecipeStepIndex?: number;
  productOfRecipeStepProductIndex?: number;
  minQuantity: number;
  maxQuantity?: number;
  name: string;
  notes: string;
  vesselPreposition: string;
  unavailableAfterStep: boolean;
  validPreparationVesselId?: string;
  index?: number;
  optionIndex: number;
}

export interface RecipeStepCreationRequestInput {
  minEstimatedTimeInSeconds?: number;
  maxEstimatedTimeInSeconds?: number;
  minTemperatureInCelsius?: number;
  maxTemperatureInCelsius?: number;
  preparationId: string;
  notes: string;
  conditionExpression: string;
  explicitInstructions: string;
  instruments: RecipeStepInstrumentCreationRequestInput[];
  vessels: RecipeStepVesselCreationRequestInput[];
  products: RecipeStepProductCreationRequestInput[];
  ingredients: RecipeStepIngredientCreationRequestInput[];
  completionConditions: RecipeStepCompletionConditionCreationRequestInput[];
  index: number;
  optional: boolean;
  startTimerAutomatically: boolean;
}

export interface RecipeMediaCreationRequestInput {
  belongsToRecipe: string;
  belongsToRecipeStep: string;
  mimeType: string;
  internalPath: string;
  externalPath: string;
  index: number;
}

export interface RecipePrepTaskStepWithinRecipeCreationRequestInput {
  belongsToRecipeStepIndex: number;
  satisfiesRecipeStep: boolean;
}

export interface RecipePrepTaskWithinRecipeCreationRequestInput {
  minStorageTemperatureInCelsius?: number;
  maxStorageTemperatureInCelsius?: number;
  minTimeBufferBeforeRecipeInSeconds: number;
  maxTimeBufferBeforeRecipeInSeconds?: number;
  storageType: string;
  name: string;
  description: string;
  explicitStorageInstructions: string;
  notes: string;
  belongsToRecipe: string;
  recipeSteps: RecipePrepTaskStepWithinRecipeCreationRequestInput[];
  optional: boolean;
}

export interface RecipeCreationRequestInput {
  inspiredByRecipeId?: string;
  name: string;
  source: string;
  sourceIsbn: string;
  description: string;
  pluralPortionName: string;
  portionName: string;
  slug: string;
  yieldsComponentType: MealComponentType;
  minEstimatedPortions: number;
  maxEstimatedPortions?: number;
  prepTasks: RecipePrepTaskWithinRecipeCreationRequestInput[];
  steps: RecipeStepCreationRequestInput[];
  alsoCreateMeal: boolean;
  eligibleForMeals: boolean;
  media: RecipeMediaCreationRequestInput[];
}
