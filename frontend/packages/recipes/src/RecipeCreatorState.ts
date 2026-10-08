import { MealComponentType, RecipeStepProductType } from './client-enums';
import type {
  RecipeCreationRequestInput,
  RecipeStepCreationRequestInput,
  RecipeStepIngredientCreationRequestInput,
  RecipeStepInstrumentCreationRequestInput,
  RecipeStepProductCreationRequestInput,
  RecipeStepVesselCreationRequestInput,
  RecipePrepTaskWithinRecipeCreationRequestInput,
  RecipePrepTaskStepWithinRecipeCreationRequestInput,
  ValidPreparation,
  ValidIngredient,
  ValidMeasurementUnit,
  ValidIngredientState,
  ValidPreparationInstrument,
  ValidPreparationVessel,
  ValidIngredientMeasurementUnit,
} from './client-types';

function createEmptyStep(index: number): RecipeStepCreationRequestInput {
  return {
    preparationId: '',
    explicitInstructions: '',
    notes: '',
    conditionExpression: '',
    instruments: [createEmptyInstrument()],
    vessels: [],
    ingredients: [createEmptyIngredient()],
    products: [createEmptyProduct(0)],
    completionConditions: [],
    index,
    optional: false,
    startTimerAutomatically: false,
  };
}

function createEmptyIngredient(): RecipeStepIngredientCreationRequestInput {
  return {
    name: '',
    ingredientNotes: '',
    quantityNotes: '',
    minQuantity: 1,
    optionIndex: 0,
    optional: false,
    toTaste: false,
  };
}

function createEmptyInstrument(): RecipeStepInstrumentCreationRequestInput {
  return {
    name: '',
    notes: '',
    minQuantity: 1,
    optionIndex: 0,
    optional: false,
    preferenceRank: 0,
  };
}

function createEmptyVessel(): RecipeStepVesselCreationRequestInput {
  return {
    name: '',
    notes: '',
    minQuantity: 1,
    vesselPreposition: 'in',
    unavailableAfterStep: false,
    optionIndex: 0,
  };
}

function createEmptyProduct(index: number): RecipeStepProductCreationRequestInput {
  return {
    name: '',
    type: RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_INGREDIENT,
    quantityNotes: '',
    storageInstructions: '',
    index,
    compostable: false,
    isLiquid: false,
    isWaste: false,
  };
}

function createEmptyPrepTask(): RecipePrepTaskWithinRecipeCreationRequestInput {
  return {
    name: '',
    description: '',
    storageType: '',
    explicitStorageInstructions: '',
    notes: '',
    belongsToRecipe: '',
    minTimeBufferBeforeRecipeInSeconds: 0,
    recipeSteps: [],
    optional: false,
  };
}

export interface StepHelper {
  show: boolean;
  preparationQuery: string;
  preparationSuggestions: ValidPreparation[];
  selectedPreparation: ValidPreparation | null;
  ingredientQueries: string[];
  ingredientSuggestions: (ValidIngredient & { validIngredientPreparationId?: string })[][];
  selectedIngredients: (
    | (ValidIngredient & { validIngredientPreparationId?: string; validIngredientMeasurementUnitId?: string })
    | null
  )[];
  ingredientMeasurementUnitSuggestions: ValidIngredientMeasurementUnit[][];
  selectedMeasurementUnits: (ValidIngredientMeasurementUnit | null)[];
  instrumentSuggestions: ValidPreparationInstrument[];
  selectedInstruments: (ValidPreparationInstrument | null)[];
  vesselQueries: string[];
  vesselSuggestions: ValidPreparationVessel[][];
  selectedVessels: (ValidPreparationVessel | null)[];
  productMeasurementUnitSuggestions: ValidMeasurementUnit[][];
  selectedProductMeasurementUnits: (ValidMeasurementUnit | null)[];
  ingredientIsProduct: boolean[];
  instrumentIsProduct: boolean[];
  vesselIsProduct: boolean[];
  ingredientIsRanged: boolean[];
  instrumentIsRanged: boolean[];
  vesselIsRanged: boolean[];
  productIsRanged: boolean[];
  productIsNamedManually: boolean[];
  completionConditionQueries: string[];
  completionConditionSuggestions: ValidIngredientState[][];
}

function createStepHelper(): StepHelper {
  return {
    show: true,
    preparationQuery: '',
    preparationSuggestions: [],
    selectedPreparation: null,
    ingredientQueries: [''],
    ingredientSuggestions: [[]],
    selectedIngredients: [null],
    ingredientMeasurementUnitSuggestions: [[]],
    selectedMeasurementUnits: [null],
    instrumentSuggestions: [],
    selectedInstruments: [null],
    vesselQueries: [''],
    vesselSuggestions: [[]],
    selectedVessels: [null],
    productMeasurementUnitSuggestions: [[]],
    selectedProductMeasurementUnits: [null],
    ingredientIsProduct: [false],
    instrumentIsProduct: [false],
    vesselIsProduct: [false],
    ingredientIsRanged: [false],
    instrumentIsRanged: [false],
    vesselIsRanged: [false],
    productIsRanged: [false],
    productIsNamedManually: [false],
    completionConditionQueries: [],
    completionConditionSuggestions: [],
  };
}

export type RecipeCreatorState = ReturnType<typeof createRecipeCreatorState>;

export function createRecipeCreatorState() {
  const state = {
    submissionError: null as string | null,
    recipe: {
      name: '',
      slug: '',
      source: '',
      sourceIsbn: '',
      description: '',
      portionName: 'portion',
      pluralPortionName: 'portions',
      yieldsComponentType: MealComponentType.MEAL_COMPONENT_TYPE_MAIN,
      minEstimatedPortions: 1,
      prepTasks: [],
      steps: [createEmptyStep(0), createEmptyStep(1)],
      alsoCreateMeal: false,
      eligibleForMeals: true,
      media: [],
    } as RecipeCreationRequestInput,
    stepHelpers: [createStepHelper(), createStepHelper()] as StepHelper[],

    addStep() {
      const idx = this.recipe.steps.length;
      this.recipe = {
        ...this.recipe,
        steps: [...this.recipe.steps, createEmptyStep(idx)],
      };
      this.stepHelpers = [...this.stepHelpers, createStepHelper()];
    },

    removeStep(stepIndex: number) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.filter((_, i) => i !== stepIndex),
      };
      this.stepHelpers = this.stepHelpers.filter((_, i) => i !== stepIndex);
      this.recipe.steps = this.recipe.steps.map((s, i) => ({ ...s, index: i }));
    },

    addIngredientToStep(stepIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientQueries: [...h.ingredientQueries, ''],
              ingredientSuggestions: [...h.ingredientSuggestions, []],
              selectedIngredients: [...h.selectedIngredients, null],
              ingredientMeasurementUnitSuggestions: [...h.ingredientMeasurementUnitSuggestions, []],
              selectedMeasurementUnits: [...h.selectedMeasurementUnits, null],
              ingredientIsProduct: [...h.ingredientIsProduct, false],
              ingredientIsRanged: [...h.ingredientIsRanged, false],
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, ingredients: [...s.ingredients, createEmptyIngredient()] } : s,
        ),
      };
    },

    removeIngredientFromStep(stepIndex: number, ingredientIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientQueries: h.ingredientQueries.filter((_, j) => j !== ingredientIndex),
              ingredientSuggestions: h.ingredientSuggestions.filter((_, j) => j !== ingredientIndex),
              selectedIngredients: h.selectedIngredients.filter((_, j) => j !== ingredientIndex),
              ingredientMeasurementUnitSuggestions: h.ingredientMeasurementUnitSuggestions.filter(
                (_, j) => j !== ingredientIndex,
              ),
              selectedMeasurementUnits: h.selectedMeasurementUnits.filter((_, j) => j !== ingredientIndex),
              ingredientIsProduct: h.ingredientIsProduct.filter((_, j) => j !== ingredientIndex),
              ingredientIsRanged: h.ingredientIsRanged.filter((_, j) => j !== ingredientIndex),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, ingredients: s.ingredients.filter((_, j) => j !== ingredientIndex) } : s,
        ),
      };
    },

    addInstrumentToStep(stepIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedInstruments: [...h.selectedInstruments, null],
              instrumentIsProduct: [...h.instrumentIsProduct, false],
              instrumentIsRanged: [...h.instrumentIsRanged, false],
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, instruments: [...s.instruments, createEmptyInstrument()] } : s,
        ),
      };
    },

    removeInstrumentFromStep(stepIndex: number, instrumentIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedInstruments: h.selectedInstruments.filter((_, j) => j !== instrumentIndex),
              instrumentIsProduct: h.instrumentIsProduct.filter((_, j) => j !== instrumentIndex),
              instrumentIsRanged: h.instrumentIsRanged.filter((_, j) => j !== instrumentIndex),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, instruments: s.instruments.filter((_, j) => j !== instrumentIndex) } : s,
        ),
      };
    },

    addVesselToStep(stepIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              vesselQueries: [...h.vesselQueries, ''],
              vesselSuggestions: [...h.vesselSuggestions, []],
              selectedVessels: [...h.selectedVessels, null],
              vesselIsProduct: [...h.vesselIsProduct, false],
              vesselIsRanged: [...h.vesselIsRanged, false],
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, vessels: [...(s.vessels ?? []), createEmptyVessel()] } : s,
        ),
      };
    },

    removeVesselFromStep(stepIndex: number, vesselIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              vesselQueries: h.vesselQueries.filter((_, j) => j !== vesselIndex),
              vesselSuggestions: h.vesselSuggestions.filter((_, j) => j !== vesselIndex),
              selectedVessels: h.selectedVessels.filter((_, j) => j !== vesselIndex),
              vesselIsProduct: h.vesselIsProduct.filter((_, j) => j !== vesselIndex),
              vesselIsRanged: h.vesselIsRanged.filter((_, j) => j !== vesselIndex),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, vessels: (s.vessels ?? []).filter((_, j) => j !== vesselIndex) } : s,
        ),
      };
    },

    addProductToStep(stepIndex: number) {
      const step = this.recipe.steps[stepIndex];
      const productIndex = step.products.length;
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              productMeasurementUnitSuggestions: [...h.productMeasurementUnitSuggestions, []],
              selectedProductMeasurementUnits: [...h.selectedProductMeasurementUnits, null],
              productIsRanged: [...h.productIsRanged, false],
              productIsNamedManually: [...h.productIsNamedManually, false],
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, products: [...s.products, createEmptyProduct(productIndex)] } : s,
        ),
      };
    },

    removeProductFromStep(stepIndex: number, productIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              productMeasurementUnitSuggestions: h.productMeasurementUnitSuggestions.filter(
                (_, j) => j !== productIndex,
              ),
              selectedProductMeasurementUnits: h.selectedProductMeasurementUnits.filter((_, j) => j !== productIndex),
              productIsRanged: h.productIsRanged.filter((_, j) => j !== productIndex),
              productIsNamedManually: h.productIsNamedManually.filter((_, j) => j !== productIndex),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex ? { ...s, products: s.products.filter((_, j) => j !== productIndex) } : s,
        ),
      };
    },

    setPreparation(stepIndex: number, preparation: ValidPreparation | null) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedPreparation: preparation,
              preparationQuery: preparation?.name ?? '',
              preparationSuggestions: [],
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) => (i === stepIndex ? { ...s, preparationId: preparation?.id ?? '' } : s)),
      };
      if (preparation) {
        // Reset step instruments/ingredients/vessels when preparation changes
        this.recipe = {
          ...this.recipe,
          steps: this.recipe.steps.map((s, i) =>
            i === stepIndex
              ? {
                  ...s,
                  instruments: [createEmptyInstrument()],
                  ingredients: [createEmptyIngredient()],
                  vessels: [],
                  products: [createEmptyProduct(0)],
                  completionConditions: [],
                }
              : s,
          ),
        };
        const freshHelper = createStepHelper();
        freshHelper.selectedPreparation = preparation;
        freshHelper.preparationQuery = preparation.name ?? '';
        this.stepHelpers = this.stepHelpers.map((h, i) => (i === stepIndex ? freshHelper : h));
      }
    },

    setPreparationSuggestions(stepIndex: number, suggestions: ValidPreparation[]) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex ? { ...h, preparationSuggestions: suggestions } : h,
      );
    },

    setPreparationQuery(stepIndex: number, query: string) {
      this.stepHelpers = this.stepHelpers.map((h, i) => (i === stepIndex ? { ...h, preparationQuery: query } : h));
    },

    setInstrumentSuggestions(stepIndex: number, suggestions: ValidPreparationInstrument[]) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex ? { ...h, instrumentSuggestions: suggestions } : h,
      );
    },

    setInstrument(stepIndex: number, instrumentIndex: number, vpi: ValidPreparationInstrument | null) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedInstruments: h.selectedInstruments.map((sel, j) => (j === instrumentIndex ? vpi : sel)),
            }
          : h,
      );
      if (vpi) {
        this.recipe = {
          ...this.recipe,
          steps: this.recipe.steps.map((s, i) =>
            i === stepIndex
              ? {
                  ...s,
                  instruments: s.instruments.map((inst, j) =>
                    j === instrumentIndex
                      ? {
                          ...inst,
                          name: (vpi as { instrument?: { name?: string } }).instrument?.name ?? '',
                          validPreparationInstrumentId: vpi.id ?? '',
                          productOfRecipeStepIndex: undefined,
                          productOfRecipeStepProductIndex: undefined,
                        }
                      : inst,
                  ),
                }
              : s,
          ),
        };
      }
    },

    clearInstrumentProductRef(stepIndex: number, instrumentIndex: number) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                instruments: s.instruments.map((inst, j) =>
                  j === instrumentIndex
                    ? {
                        ...inst,
                        name: '',
                        validPreparationInstrumentId: undefined,
                        productOfRecipeStepIndex: undefined,
                        productOfRecipeStepProductIndex: undefined,
                      }
                    : inst,
                ),
              }
            : s,
        ),
      };
    },

    setInstrumentFromProduct(
      stepIndex: number,
      instrumentIndex: number,
      fromStepIndex: number,
      fromProductIndex: number,
    ) {
      const fromStep = this.recipe.steps[fromStepIndex];
      const product = fromStep?.products?.[fromProductIndex];
      const name = product && 'name' in product ? (product.name as string) : '';
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedInstruments: h.selectedInstruments.map((sel, j) => (j === instrumentIndex ? null : sel)),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                instruments: s.instruments.map((inst, j) =>
                  j === instrumentIndex
                    ? {
                        ...inst,
                        name,
                        validPreparationInstrumentId: undefined,
                        productOfRecipeStepIndex: fromStepIndex,
                        productOfRecipeStepProductIndex: fromProductIndex,
                      }
                    : inst,
                ),
              }
            : s,
        ),
      };
    },

    setIngredient(
      stepIndex: number,
      ingredientIndex: number,
      ingredient: (ValidIngredient & { validIngredientPreparationId?: string }) | null,
      measurementUnit: ValidIngredientMeasurementUnit | null,
    ) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedIngredients: h.selectedIngredients.map((sel, j) => (j === ingredientIndex ? ingredient : sel)),
              selectedMeasurementUnits: h.selectedMeasurementUnits.map((sel, j) =>
                j === ingredientIndex ? measurementUnit : sel,
              ),
              ingredientQueries: h.ingredientQueries.map((q, j) =>
                j === ingredientIndex ? (ingredient?.name ?? '') : q,
              ),
              ingredientSuggestions: h.ingredientSuggestions.map((sug, j) => (j === ingredientIndex ? [] : sug)),
            }
          : h,
      );
      if (ingredient) {
        this.recipe = {
          ...this.recipe,
          steps: this.recipe.steps.map((s, i) =>
            i === stepIndex
              ? {
                  ...s,
                  ingredients: s.ingredients.map((ing, j) =>
                    j === ingredientIndex
                      ? {
                          ...ing,
                          name: ingredient.name ?? '',
                          validIngredientPreparationId: ingredient.validIngredientPreparationId,
                          validIngredientMeasurementUnitId: measurementUnit?.id,
                          productOfRecipeStepIndex: undefined,
                          productOfRecipeStepProductIndex: undefined,
                        }
                      : ing,
                  ),
                }
              : s,
          ),
        };
      }
    },

    clearIngredientProductRef(stepIndex: number, ingredientIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientQueries: h.ingredientQueries.map((q, j) => (j === ingredientIndex ? '' : q)),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                ingredients: s.ingredients.map((ing, j) =>
                  j === ingredientIndex
                    ? {
                        ...ing,
                        name: '',
                        productOfRecipeStepIndex: undefined,
                        productOfRecipeStepProductIndex: undefined,
                        recipeStepProductRecipeId: undefined,
                        recipeStepProductRecipeSlug: undefined,
                      }
                    : ing,
                ),
              }
            : s,
        ),
      };
    },

    setIngredientRecipeRef(
      stepIndex: number,
      ingredientIndex: number,
      recipeId: string,
      recipeName: string,
      slug?: string,
    ) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientQueries: h.ingredientQueries.map((q, j) => (j === ingredientIndex ? recipeName : q)),
              selectedIngredients: h.selectedIngredients.map((sel, j) => (j === ingredientIndex ? null : sel)),
              selectedMeasurementUnits: h.selectedMeasurementUnits.map((sel, j) =>
                j === ingredientIndex ? null : sel,
              ),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                ingredients: s.ingredients.map((ing, j) =>
                  j === ingredientIndex
                    ? {
                        ...ing,
                        name: recipeName,
                        recipeStepProductRecipeId: recipeId,
                        recipeStepProductRecipeSlug: slug,
                        productOfRecipeStepIndex: undefined,
                        productOfRecipeStepProductIndex: undefined,
                      }
                    : ing,
                ),
              }
            : s,
        ),
      };
    },

    clearIngredientRecipeRef(stepIndex: number, ingredientIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientQueries: h.ingredientQueries.map((q, j) => (j === ingredientIndex ? '' : q)),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                ingredients: s.ingredients.map((ing, j) =>
                  j === ingredientIndex
                    ? {
                        ...ing,
                        name: '',
                        recipeStepProductRecipeId: undefined,
                        recipeStepProductRecipeSlug: undefined,
                      }
                    : ing,
                ),
              }
            : s,
        ),
      };
    },

    setIngredientFromProduct(
      stepIndex: number,
      ingredientIndex: number,
      fromStepIndex: number,
      fromProductIndex: number,
    ) {
      const fromStep = this.recipe.steps[fromStepIndex];
      const product = fromStep?.products?.[fromProductIndex];
      const name = product && 'name' in product ? (product.name as string) : '';
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedIngredients: h.selectedIngredients.map((sel, j) => (j === ingredientIndex ? null : sel)),
              selectedMeasurementUnits: h.selectedMeasurementUnits.map((sel, j) =>
                j === ingredientIndex ? null : sel,
              ),
              ingredientQueries: h.ingredientQueries.map((q, j) => (j === ingredientIndex ? name : q)),
              ingredientSuggestions: h.ingredientSuggestions.map((sug, j) => (j === ingredientIndex ? [] : sug)),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                ingredients: s.ingredients.map((ing, j) =>
                  j === ingredientIndex
                    ? {
                        ...ing,
                        name,
                        recipeStepProductRecipeId: undefined,
                        recipeStepProductRecipeSlug: undefined,
                        validIngredientPreparationId: undefined,
                        validIngredientMeasurementUnitId: undefined,
                        productOfRecipeStepIndex: fromStepIndex,
                        productOfRecipeStepProductIndex: fromProductIndex,
                      }
                    : ing,
                ),
              }
            : s,
        ),
      };
    },

    setIngredientSuggestions(stepIndex: number, ingredientIndex: number, suggestions: ValidIngredient[]) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientSuggestions: h.ingredientSuggestions.map((sug, j) =>
                j === ingredientIndex ? suggestions : sug,
              ),
            }
          : h,
      );
    },

    setIngredientQuery(stepIndex: number, ingredientIndex: number, query: string) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientQueries: h.ingredientQueries.map((q, j) => (j === ingredientIndex ? query : q)),
            }
          : h,
      );
    },

    setIngredientMeasurementUnitSuggestions(
      stepIndex: number,
      ingredientIndex: number,
      suggestions: ValidIngredientMeasurementUnit[],
    ) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientMeasurementUnitSuggestions: h.ingredientMeasurementUnitSuggestions.map((sug, j) =>
                j === ingredientIndex ? suggestions : sug,
              ),
            }
          : h,
      );
    },

    setIngredientMeasurementUnit(
      stepIndex: number,
      ingredientIndex: number,
      vimu: ValidIngredientMeasurementUnit | null,
    ) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedMeasurementUnits: h.selectedMeasurementUnits.map((sel, j) =>
                j === ingredientIndex ? vimu : sel,
              ),
            }
          : h,
      );
      if (vimu) {
        this.recipe = {
          ...this.recipe,
          steps: this.recipe.steps.map((s, i) =>
            i === stepIndex
              ? {
                  ...s,
                  ingredients: s.ingredients.map((ing, j) =>
                    j === ingredientIndex ? { ...ing, validIngredientMeasurementUnitId: vimu.id } : ing,
                  ),
                }
              : s,
          ),
        };
      }
    },

    setVessel(stepIndex: number, vesselIndex: number, vpv: ValidPreparationVessel | null) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedVessels: h.selectedVessels.map((sel, j) => (j === vesselIndex ? vpv : sel)),
              vesselQueries: h.vesselQueries.map((q, j) =>
                j === vesselIndex ? ((vpv as { vessel?: { name?: string } })?.vessel?.name ?? '') : q,
              ),
            }
          : h,
      );
      if (vpv) {
        this.recipe = {
          ...this.recipe,
          steps: this.recipe.steps.map((s, i) =>
            i === stepIndex
              ? {
                  ...s,
                  vessels: (s.vessels ?? []).map((v, j) =>
                    j === vesselIndex
                      ? {
                          ...v,
                          name: (vpv as { vessel?: { name?: string } }).vessel?.name ?? '',
                          validPreparationVesselId: vpv.id,
                          productOfRecipeStepIndex: undefined,
                          productOfRecipeStepProductIndex: undefined,
                        }
                      : v,
                  ),
                }
              : s,
          ),
        };
      }
    },

    clearVesselProductRef(stepIndex: number, vesselIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              vesselQueries: h.vesselQueries.map((q, j) => (j === vesselIndex ? '' : q)),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                vessels: (s.vessels ?? []).map((v, j) =>
                  j === vesselIndex
                    ? {
                        ...v,
                        name: '',
                        validPreparationVesselId: undefined,
                        productOfRecipeStepIndex: undefined,
                        productOfRecipeStepProductIndex: undefined,
                      }
                    : v,
                ),
              }
            : s,
        ),
      };
    },

    setVesselFromProduct(stepIndex: number, vesselIndex: number, fromStepIndex: number, fromProductIndex: number) {
      const fromStep = this.recipe.steps[fromStepIndex];
      const product = fromStep?.products?.[fromProductIndex];
      const name = product && 'name' in product ? (product.name as string) : '';
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              selectedVessels: h.selectedVessels.map((sel, j) => (j === vesselIndex ? null : sel)),
              vesselQueries: h.vesselQueries.map((q, j) => (j === vesselIndex ? name : q)),
            }
          : h,
      );
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                vessels: (s.vessels ?? []).map((v, j) =>
                  j === vesselIndex
                    ? {
                        ...v,
                        name,
                        validPreparationVesselId: undefined,
                        productOfRecipeStepIndex: fromStepIndex,
                        productOfRecipeStepProductIndex: fromProductIndex,
                      }
                    : v,
                ),
              }
            : s,
        ),
      };
    },

    setVesselSuggestions(stepIndex: number, vesselIndex: number, suggestions: ValidPreparationVessel[]) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              vesselSuggestions: h.vesselSuggestions.map((sug, j) => (j === vesselIndex ? suggestions : sug)),
            }
          : h,
      );
    },

    setVesselQuery(stepIndex: number, vesselIndex: number, query: string) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              vesselQueries: h.vesselQueries.map((q, j) => (j === vesselIndex ? query : q)),
            }
          : h,
      );
    },

    updateRecipeField<K extends keyof RecipeCreationRequestInput>(field: K, value: RecipeCreationRequestInput[K]) {
      this.recipe = { ...this.recipe, [field]: value };
    },

    addPrepTask() {
      this.recipe = {
        ...this.recipe,
        prepTasks: [...(this.recipe.prepTasks ?? []), createEmptyPrepTask()],
      };
    },

    removePrepTask(taskIndex: number) {
      this.recipe = {
        ...this.recipe,
        prepTasks: (this.recipe.prepTasks ?? []).filter((_, i) => i !== taskIndex),
      };
    },

    updatePrepTaskField<K extends keyof RecipePrepTaskWithinRecipeCreationRequestInput>(
      taskIndex: number,
      field: K,
      value: RecipePrepTaskWithinRecipeCreationRequestInput[K],
    ) {
      const tasks = [...(this.recipe.prepTasks ?? [])];
      tasks[taskIndex] = { ...tasks[taskIndex]!, [field]: value };
      this.recipe = { ...this.recipe, prepTasks: tasks };
    },

    setPrepTaskRecipeSteps(taskIndex: number, recipeSteps: RecipePrepTaskStepWithinRecipeCreationRequestInput[]) {
      this.updatePrepTaskField(taskIndex, 'recipeSteps', recipeSteps);
    },

    setIngredientOptionIndex(stepIndex: number, ingredientIndex: number, optionIndex: number) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                ingredients: s.ingredients.map((ing, j) => (j === ingredientIndex ? { ...ing, optionIndex } : ing)),
              }
            : s,
        ),
      };
    },

    setInstrumentOptionIndex(stepIndex: number, instrumentIndex: number, optionIndex: number) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                instruments: s.instruments.map((inst, j) => (j === instrumentIndex ? { ...inst, optionIndex } : inst)),
              }
            : s,
        ),
      };
    },

    setVesselOptionIndex(stepIndex: number, vesselIndex: number, optionIndex: number) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                vessels: (s.vessels ?? []).map((v, j) => (j === vesselIndex ? { ...v, optionIndex } : v)),
              }
            : s,
        ),
      };
    },

    updateStepField(stepIndex: number, field: keyof RecipeStepCreationRequestInput, value: unknown) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) => (i === stepIndex ? { ...s, [field]: value } : s)),
      };
    },

    updateIngredientQuantity(stepIndex: number, ingredientIndex: number, min: number, max?: number) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                ingredients: s.ingredients.map((ing, j) =>
                  j === ingredientIndex ? { ...ing, minQuantity: min, maxQuantity: max } : ing,
                ),
              }
            : s,
        ),
      };
    },

    updateProduct(stepIndex: number, productIndex: number, updates: Partial<RecipeStepProductCreationRequestInput>) {
      this.recipe = {
        ...this.recipe,
        steps: this.recipe.steps.map((s, i) =>
          i === stepIndex
            ? {
                ...s,
                products: s.products.map((p, j) => (j === productIndex ? { ...p, ...updates } : p)),
              }
            : s,
        ),
      };
    },

    setProductMeasurementUnitSuggestions(stepIndex: number, productIndex: number, units: ValidMeasurementUnit[]) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              productMeasurementUnitSuggestions: h.productMeasurementUnitSuggestions.map((unitsForProd, j) =>
                j === productIndex ? units : unitsForProd,
              ),
            }
          : h,
      );
    },

    toggleIngredientProduct(stepIndex: number, ingredientIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              ingredientIsProduct: h.ingredientIsProduct.map((v, j) => (j === ingredientIndex ? !v : v)),
            }
          : h,
      );
    },

    toggleInstrumentProduct(stepIndex: number, instrumentIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              instrumentIsProduct: h.instrumentIsProduct.map((v, j) => (j === instrumentIndex ? !v : v)),
            }
          : h,
      );
    },

    toggleVesselProduct(stepIndex: number, vesselIndex: number) {
      this.stepHelpers = this.stepHelpers.map((h, i) =>
        i === stepIndex
          ? {
              ...h,
              vesselIsProduct: h.vesselIsProduct.map((v, j) => (j === vesselIndex ? !v : v)),
            }
          : h,
      );
    },

    setSubmissionError(error: string | null) {
      this.submissionError = error;
    },

    dumpState(): { recipe: RecipeCreationRequestInput; stepHelpers: StepHelper[] } {
      return JSON.parse(JSON.stringify({ recipe: this.recipe, stepHelpers: this.stepHelpers }));
    },

    loadState(dump: { recipe: RecipeCreationRequestInput; stepHelpers: StepHelper[] }) {
      const { recipe, stepHelpers } = dump;
      const stepCount = recipe.steps?.length ?? 0;
      let helpers = stepHelpers ?? [];
      if (helpers.length < stepCount) {
        helpers = [...helpers, ...Array.from({ length: stepCount - helpers.length }, () => createStepHelper())];
      } else if (helpers.length > stepCount) {
        helpers = helpers.slice(0, stepCount);
      }
      this.recipe = JSON.parse(JSON.stringify(recipe));
      this.stepHelpers = JSON.parse(JSON.stringify(helpers));
    },
  };

  return state;
}
