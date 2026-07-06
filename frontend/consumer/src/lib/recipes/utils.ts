import type {
  RecipeCreationRequestInput,
  RecipeStepCreationRequestInput,
  RecipeStepProductCreationRequestInput,
} from '@dinnerdonebetter/api-client/mealplanning/mealplanning_service_types';
import type {
  Recipe,
  RecipeStep,
  RecipeStepIngredient,
  RecipeStepInstrument,
  RecipeStepProduct,
  RecipeStepVessel,
} from '@dinnerdonebetter/api-client/mealplanning/mealplanning_messages';
import { RecipeStepProductType } from '@dinnerdonebetter/api-client/mealplanning/mealplanning_messages';
import { round } from '@primandproper/numbers';

export const englishListFormatter = new Intl.ListFormat('en');

export function cleanFloat(float: number): number {
  return round(float, 2);
}

/**
 * Returns true if the element references a product from an earlier step.
 */
export function stepElementIsProduct(x: RecipeStepIngredient | RecipeStepInstrument | RecipeStepVessel): boolean {
  return Boolean((x as { recipeStepProductId?: string }).recipeStepProductId);
}

/**
 * Returns 1-based step index containing the product with the given id, or -1.
 */
export function getRecipeStepIndexByProductID(recipe: Recipe, id: string): number {
  let retVal = -1;
  (recipe.steps ?? []).forEach((step: RecipeStep, stepIndex: number) => {
    if ((step.products ?? []).some((p: RecipeStepProduct) => p.id === id)) {
      retVal = stepIndex + 1;
    }
  });
  return retVal;
}

/**
 * Returns 1-based step index for the step with the given id, or -1.
 */
export function getRecipeStepIndexByStepID(recipe: Recipe, id: string): number {
  const idx = (recipe.steps ?? []).findIndex((step: RecipeStep) => step.id === id);
  return idx >= 0 ? idx + 1 : -1;
}

export interface RecipeStepProductSuggestion {
  product: RecipeStepProductCreationRequestInput & { name: string };
  stepIndex: number;
  productIndex: number;
}

export interface RecipeStepVesselSuggestion {
  vessel: { name: string; id?: string };
  stepIndex: number;
  productIndex: number;
}

export interface RecipeStepInstrumentSuggestion {
  product: RecipeStepProductCreationRequestInput & { name: string; id?: string };
  stepIndex: number;
  productIndex: number;
}

/**
 * Returns products from earlier steps that can be used as ingredients in the current step.
 * Products with type 'ingredient' from steps 0..stepIndex-1.
 */
export function determineAvailableRecipeStepProducts(
  recipe: RecipeCreationRequestInput,
  stepIndex: number,
): RecipeStepProductSuggestion[] {
  const results: RecipeStepProductSuggestion[] = [];
  for (let i = 0; i < stepIndex && i < (recipe.steps?.length ?? 0); i++) {
    const step = recipe.steps![i] as RecipeStepCreationRequestInput;
    const products = step.products ?? [];
    for (let j = 0; j < products.length; j++) {
      const product = products[j] as RecipeStepProductCreationRequestInput & { name: string };
      if (product.type === 0) {
        // RECIPE_STEP_PRODUCT_TYPE_INGREDIENT
        results.push({
          product: { ...product, name: product.name ?? `product ${j + 1}` },
          stepIndex: i,
          productIndex: j,
        });
      }
    }
  }
  return results;
}

/**
 * Returns vessels from earlier steps that can be used in the current step.
 * Products with type 'vessel' from steps 0..stepIndex-1.
 */
export function determineAvailableRecipeStepVessels(
  recipe: RecipeCreationRequestInput,
  stepIndex: number,
): RecipeStepVesselSuggestion[] {
  const results: RecipeStepVesselSuggestion[] = [];
  for (let i = 0; i < stepIndex && i < (recipe.steps?.length ?? 0); i++) {
    const step = recipe.steps![i] as RecipeStepCreationRequestInput;
    const products = step.products ?? [];
    for (let j = 0; j < products.length; j++) {
      const product = products[j] as RecipeStepProductCreationRequestInput & { name: string };
      if (product.type === 2) {
        // RECIPE_STEP_PRODUCT_TYPE_VESSEL
        results.push({
          vessel: { name: product.name ?? `vessel ${j + 1}` },
          stepIndex: i,
          productIndex: j,
        });
      }
    }
  }
  return results;
}

/**
 * Returns instruments from earlier steps that can be used in the current step.
 * Products with type 'instrument' from steps 0..stepIndex-1.
 */
export function determinePreparedInstrumentOptions(
  recipe: RecipeCreationRequestInput,
  stepIndex: number,
): RecipeStepInstrumentSuggestion[] {
  const results: RecipeStepInstrumentSuggestion[] = [];
  for (let i = 0; i < stepIndex && i < (recipe.steps?.length ?? 0); i++) {
    const step = recipe.steps![i] as RecipeStepCreationRequestInput;
    const products = step.products ?? [];
    for (let j = 0; j < products.length; j++) {
      const product = products[j] as RecipeStepProductCreationRequestInput & { name: string };
      if (product.type === 1) {
        // RECIPE_STEP_PRODUCT_TYPE_INSTRUMENT
        results.push({
          product: { ...product, name: product.name ?? `instrument ${j + 1}` },
          stepIndex: i,
          productIndex: j,
        });
      }
    }
  }
  return results;
}

/**
 * Build human-readable step text from vessels, instruments, and ingredients.
 */
export function buildRecipeStepText(recipe: Recipe, recipeStep: RecipeStep, recipeScale = 1): string {
  const vesselList = englishListFormatter.format(
    (recipeStep.vessels ?? []).map((x: RecipeStepVessel) => {
      const elementIsProduct = stepElementIsProduct(x);
      const min = x.minQuantity ?? 1;
      const max = x.maxQuantity ?? -1;
      return (
        (min === 1
          ? `${x.vesselPreposition ? `${x.vesselPreposition} ` : ''}${elementIsProduct ? 'the' : 'a'} ${x.vessel?.name ?? x.name}`
          : `${min}${max > min ? ` to ${max}` : ''} ${x.vessel?.pluralName ?? x.name}`) +
        (elementIsProduct && x.recipeStepProductId
          ? ` from step #${getRecipeStepIndexByProductID(recipe, x.recipeStepProductId)}`
          : '')
      );
    }),
  );

  const instrumentList = englishListFormatter.format(
    (recipeStep.instruments ?? []).map((x: RecipeStepInstrument) => {
      const elementIsProduct = stepElementIsProduct(x);
      const min = x.minQuantity ?? 1;
      const max = x.maxQuantity ?? -1;
      return (
        (min === 1
          ? `${elementIsProduct ? 'the' : 'a'} ${x.instrument?.name ?? x.name}`
          : `${min}${max > min ? ` to ${max}` : ''} ${x.instrument?.pluralName ?? x.name}`) +
        (elementIsProduct && x.recipeStepProductId
          ? ` from step #${getRecipeStepIndexByProductID(recipe, x.recipeStepProductId)}`
          : '')
      );
    }),
  );

  const allInstrumentsShouldBeExcludedFromSummaries = (recipeStep.instruments ?? []).every(
    (x: RecipeStepInstrument) => !x.instrument || x.instrument.displayInSummaryLists,
  );
  const intro = allInstrumentsShouldBeExcludedFromSummaries ? `Using ${instrumentList}, ` : '';

  const ingredientList = englishListFormatter.format(
    (recipeStep.ingredients ?? []).map((x: RecipeStepIngredient) => {
      const scaleFactor = (x as { scaleFactor?: number }).scaleFactor ?? 1;
      const effectiveScale = recipeScale * (scaleFactor > 0 ? scaleFactor : 1);
      const elementIsProduct = stepElementIsProduct(x);
      const mu = x.measurementUnit;
      let measurementUnit =
        cleanFloat((x.minQuantity ?? 0) * effectiveScale) === 1 ? (mu?.name ?? '') : (mu?.pluralName ?? '');
      measurementUnit = ['unit', 'units'].includes(measurementUnit) ? '' : measurementUnit;

      const min = x.minQuantity ?? 1;
      const max = x.maxQuantity ?? -1;
      const intro = elementIsProduct
        ? ''
        : `${cleanFloat(min * effectiveScale)}${max > min ? ` to ${cleanFloat((max ?? 0) * effectiveScale)} ` : ''} ${measurementUnit}`;

      const name =
        cleanFloat(min * effectiveScale) === 1 ? (x.ingredient?.name ?? x.name) : (x.ingredient?.pluralName ?? x.name);

      return (
        `${intro} ${elementIsProduct ? 'the' : ''} ${name}` +
        (elementIsProduct && x.recipeStepProductId
          ? ` from step #${getRecipeStepIndexByProductID(recipe, x.recipeStepProductId)}`
          : '')
      );
    }),
  );

  const ingredientProducts: RecipeStepProduct[] = [];
  const instrumentProducts: RecipeStepProduct[] = [];
  const vesselProducts: RecipeStepProduct[] = [];
  for (const p of recipeStep.products ?? []) {
    if (p.type === RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_INGREDIENT) ingredientProducts.push(p);
    else if (p.type === RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_INSTRUMENT) instrumentProducts.push(p);
    else if (p.type === RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_VESSEL) vesselProducts.push(p);
  }

  const productList = englishListFormatter.format(
    [
      ingredientProducts.length === 0
        ? ''
        : `the ${ingredientProducts.length === 1 ? 'ingredient' : 'ingredients'} ${englishListFormatter.format(ingredientProducts.map((x) => x.name))}`,
      instrumentProducts.length === 0
        ? ''
        : englishListFormatter.format(
            instrumentProducts.map(
              (x) => `${(x.minItemQuantity ?? x.minMeasurementQuantity ?? 1) === 1 ? 'a' : 'the'} ${x.name}`,
            ),
          ),
      vesselProducts.length === 0
        ? ''
        : englishListFormatter.format(
            vesselProducts.map(
              (x) => `${(x.minItemQuantity ?? x.minMeasurementQuantity ?? 1) === 1 ? 'a' : 'the'} ${x.name}`,
            ),
          ),
    ].filter((s) => s.length > 0),
  );

  const prepName = recipeStep.preparation?.name ?? 'prepare';
  const output = (
    recipeStep.explicitInstructions || `${intro}${prepName} ${ingredientList} ${vesselList} to yield ${productList}.`
  ).trim();

  return output.charAt(0).toUpperCase() + output.slice(1);
}

/**
 * Collect unique vessels from recipes for grocery/prep summaries.
 */
export function determineVesselsForRecipes(recipes: Recipe[]): RecipeStepVessel[] {
  const allVessels = recipes.flatMap((recipe) =>
    (recipe.steps ?? []).flatMap((step) =>
      (step.vessels ?? []).filter(
        (v: RecipeStepVessel) => (v.vessel && v.vessel.displayInSummaryLists) || !!v.recipeStepProductId,
      ),
    ),
  );

  const uniqueVessels: Record<string, RecipeStepVessel> = {};
  for (const vessel of allVessels) {
    if (vessel.vessel) {
      const id = vessel.vessel.id;
      const existing = uniqueVessels[id];
      if (existing) {
        uniqueVessels[id] = {
          ...existing,
          minQuantity: (existing.minQuantity ?? 0) + (vessel.minQuantity ?? 0),
          maxQuantity:
            vessel.maxQuantity !== undefined ? (existing.maxQuantity ?? 0) + vessel.maxQuantity : existing.maxQuantity,
        };
      } else {
        uniqueVessels[id] = { ...vessel };
      }
    }
  }
  return Object.values(uniqueVessels);
}

export interface MealRecipeInput {
  scale: number;
  recipe: Recipe;
}

/**
 * Aggregate ingredients across recipes with scaling.
 */
export function determineAllIngredientsForRecipes(input: MealRecipeInput[]): RecipeStepIngredient[] {
  const allIngredients = input.flatMap((x) =>
    (x.recipe.steps ?? []).flatMap((step) =>
      (step.ingredients ?? [])
        .filter((i) => i.ingredient)
        .map((y) => ({
          ...y,
          minQuantity: (y.minQuantity ?? 0) * x.scale,
          maxQuantity: y.maxQuantity !== undefined ? y.maxQuantity * x.scale : undefined,
        })),
    ),
  );

  const uniqueIngredients: Record<string, RecipeStepIngredient> = {};
  for (const ing of allIngredients) {
    if (ing.ingredient) {
      const id = ing.ingredient.id;
      const existing = uniqueIngredients[id];
      if (existing) {
        uniqueIngredients[id] = {
          ...existing,
          minQuantity: (existing.minQuantity ?? 0) + (ing.minQuantity ?? 0),
          maxQuantity:
            ing.maxQuantity !== undefined ? (existing.maxQuantity ?? 0) + ing.maxQuantity : existing.maxQuantity,
        };
      } else {
        uniqueIngredients[id] = { ...ing };
      }
    }
  }
  return Object.values(uniqueIngredients);
}

/**
 * Collect unique instruments and vessels for prep summaries.
 */
export function determineAllInstrumentsForRecipes(recipes: Recipe[]): (RecipeStepInstrument | RecipeStepVessel)[] {
  const unique: Record<string, RecipeStepInstrument | RecipeStepVessel> = {};

  for (const recipe of recipes) {
    for (const step of recipe.steps ?? []) {
      for (const instrument of step.instruments ?? []) {
        if (instrument.instrument?.displayInSummaryLists) {
          unique[instrument.instrument.id] = instrument;
        }
      }
      for (const vessel of step.vessels ?? []) {
        if (vessel.vessel?.displayInSummaryLists) {
          unique[vessel.vessel.id] = vessel;
        }
      }
    }
  }
  return Object.values(unique);
}
