<script lang="ts">
  import { onMount } from 'svelte';
  import { enhance } from '$app/forms';
  import {
    PageContainer,
    FormField,
    Input,
    Button,
    Alert,
    Link,
    Autocomplete,
    NumberInput,
    Card,
  } from '@dinnerdonebetter/ui';
  import {
    createRecipeCreatorState,
    type RecipeCreatorState,
    type StepHelper,
  } from '@dinnerdonebetter/recipes/RecipeCreatorState';
  import { renderMermaidForRecipeCreationInput } from '@dinnerdonebetter/recipes/recipeMermaid';
  import type {
    ValidPreparation,
    ValidIngredient,
    ValidIngredientPreparation,
    ValidIngredientMeasurementUnit,
    ValidPreparationVessel,
    ValidPreparationInstrument,
    ValidMeasurementUnit,
    RecipePrepTaskStepWithinRecipeCreationRequestInput,
  } from '@dinnerdonebetter/recipes/client-types';
  import { RecipeStepProductType, MealComponentType } from '@dinnerdonebetter/recipes/client-enums';
  import type {
    RecipeStepCreationRequestInput,
    RecipeStepProductCreationRequestInput,
  } from '@dinnerdonebetter/recipes/client-types';
  import mermaid from 'mermaid';

  /** Previous-step product option for use as ingredient/instrument/vessel. */
  type PreviousStepProductOption = {
    stepIndex: number;
    productIndex: number;
    name: string;
    type: number;
  };

  /**
   * Returns products from steps 0..stepIndex-1 that are not yet used as ingredient/instrument/vessel
   * in stepIndex or any later step. Used = has productOfRecipeStepIndex/productOfRecipeStepProductIndex set.
   */
  function getAvailablePreviousProducts(
    steps: RecipeStepCreationRequestInput[],
    stepIndex: number,
  ): PreviousStepProductOption[] {
    // Local computation only; not reactive state — built-in Set is appropriate
    // eslint-disable-next-line svelte/prefer-svelte-reactivity -- not used as reactive state
    const used = new Set<string>();
    for (let k = stepIndex; k < steps.length; k++) {
      const step = steps[k];
      for (const ing of step.ingredients ?? []) {
        if (ing.productOfRecipeStepIndex != null && ing.productOfRecipeStepProductIndex != null) {
          used.add(`${ing.productOfRecipeStepIndex}:${ing.productOfRecipeStepProductIndex}`);
        }
      }
      for (const inst of step.instruments ?? []) {
        if (inst.productOfRecipeStepIndex != null && inst.productOfRecipeStepProductIndex != null) {
          used.add(`${inst.productOfRecipeStepIndex}:${inst.productOfRecipeStepProductIndex}`);
        }
      }
      for (const v of step.vessels ?? []) {
        if (v.productOfRecipeStepIndex != null && v.productOfRecipeStepProductIndex != null) {
          used.add(`${v.productOfRecipeStepIndex}:${v.productOfRecipeStepProductIndex}`);
        }
      }
    }
    const results: PreviousStepProductOption[] = [];
    for (let i = 0; i < stepIndex && i < steps.length; i++) {
      const step = steps[i];
      const products = (step.products ?? []) as (RecipeStepProductCreationRequestInput & { name?: string })[];
      for (let j = 0; j < products.length; j++) {
        const key = `${i}:${j}`;
        if (used.has(key)) continue;
        const p = products[j];
        results.push({
          stepIndex: i,
          productIndex: j,
          name: p?.name ?? `Product ${j + 1}`,
          type: p?.type ?? RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_INGREDIENT,
        });
      }
    }
    return results;
  }

  /** Resolve the display name of a product referenced by step index and product index (for from-product summaries). */
  function getReferencedProductName(
    steps: RecipeStepCreationRequestInput[],
    stepIndex: number | undefined | null,
    productIndex: number | undefined | null,
  ): string {
    if (stepIndex == null || productIndex == null || !steps?.length) return '';
    const step = steps[stepIndex];
    const product = step?.products?.[productIndex] as { name?: string } | undefined;
    return typeof product?.name === 'string' ? product.name : '';
  }

  let { form } = $props();

  // Svelte 5 runes: typings can treat $state as store-based; assert so state is RecipeCreatorState in template.
  // @ts-expect-error - runes vs store typings: state is reactive rune state, not a Svelte store
  const state = $state(createRecipeCreatorState()) as unknown as RecipeCreatorState;

  let mermaidContainer: HTMLDivElement | null = $state(null);
  function useMermaidContainer(node: HTMLDivElement) {
    mermaidContainer = node;
    return {
      destroy() {
        mermaidContainer = null;
      },
    };
  }
  const mermaidCode = $derived(
    renderMermaidForRecipeCreationInput(
      state.recipe,
      state.stepHelpers.map((h: StepHelper) => h.selectedPreparation?.name),
    ),
  );

  /** Serialized recipe for form submission and dumps – derived so it always reflects current state. */
  const recipePayload = $derived(JSON.stringify(state.recipe));

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function debounce<F extends (...args: any[]) => any>(fn: F, ms: number): (...args: Parameters<F>) => void {
    let timer: ReturnType<typeof setTimeout> | null = null;
    return (...args: Parameters<F>) => {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => fn(...args), ms);
    };
  }

  const debouncedPrepSearch = debounce(async (stepIndex: number, query: string) => {
    const results = await fetchPreparations(query);
    state.setPreparationSuggestions(
      stepIndex,
      results.map((r) => ({ id: r.id, name: r.label })),
    );
  }, 300);

  const debouncedIngredientSearch = debounce(async (stepIndex: number, ingIdx: number, query: string) => {
    const prep = state.stepHelpers[stepIndex]?.selectedPreparation;
    if (!prep) return;
    const results = await fetchIngredients(query, prep.id);
    state.setIngredientSuggestions(stepIndex, ingIdx, results);
  }, 300);

  async function fetchVessels(query: string, preparationId: string): Promise<ValidPreparationVessel[]> {
    if (!preparationId) return [];
    const res = await fetch(
      `/api/mealplanning/search-vessels?q=${encodeURIComponent(query)}&preparationId=${encodeURIComponent(preparationId)}`,
    );
    const json = await res.json();
    return json.results ?? [];
  }

  const debouncedVesselSearch = debounce(async (stepIndex: number, vesselIdx: number, query: string) => {
    const prep = state.stepHelpers[stepIndex]?.selectedPreparation;
    if (!prep) return;
    const results = await fetchVessels(query, prep.id);
    state.setVesselSuggestions(stepIndex, vesselIdx, results);
  }, 300);

  async function fetchPreparations(query: string): Promise<{ id: string; label: string }[]> {
    if (query.length < 2) return [];
    const res = await fetch(`/api/mealplanning/search-preparations?q=${encodeURIComponent(query)}`);
    const json = await res.json();
    return (json.results ?? []).map((p: ValidPreparation) => ({ id: p.id, label: p.name }));
  }

  async function fetchIngredients(query: string, preparationId: string): Promise<ValidIngredient[]> {
    if (query.length < 2 || !preparationId) return [];
    const res = await fetch(
      `/api/mealplanning/search-ingredients?q=${encodeURIComponent(query)}&preparationId=${encodeURIComponent(preparationId)}`,
    );
    const json = await res.json();
    return json.results ?? [];
  }

  async function fetchIngredientPreparations(preparationId: string): Promise<ValidIngredientPreparation[]> {
    const res = await fetch(
      `/api/mealplanning/ingredient-preparations?preparationId=${encodeURIComponent(preparationId)}`,
    );
    const json = await res.json();
    return json.results ?? [];
  }

  async function fetchProductMeasurementUnits(query = ''): Promise<ValidMeasurementUnit[]> {
    const res = await fetch(`/api/mealplanning/search-measurement-units?q=${encodeURIComponent(query)}`);
    const json = await res.json();
    return (json.results ?? []) as ValidMeasurementUnit[];
  }

  type RecipeSearchHit = { id: string; name: string; slug: string };
  let recipeSearchBySlot: Record<string, string> = $state({});
  let recipeSuggestionsBySlot: Record<string, RecipeSearchHit[]> = $state({});
  function recipeSlotKey(stepIndex: number, ingIdx: number) {
    return `${stepIndex}-${ingIdx}`;
  }

  /** Keys for optional UI: "stepIndex-type-idx" e.g. "0-instrument-0", "1-ingredient-2". */
  function optionalKey(stepIndex: number, type: 'instrument' | 'vessel' | 'ingredient', idx: number) {
    return `${stepIndex}-${type}-${idx}`;
  }
  let showFromProductByKey: Record<string, boolean> = $state({});
  let showOptionGroupByKey: Record<string, boolean> = $state({});
  let showFromRecipeByKey: Record<string, boolean> = $state({});
  function toggleShowFromProduct(key: string) {
    showFromProductByKey = { ...showFromProductByKey, [key]: !showFromProductByKey[key] };
  }
  function toggleShowOptionGroup(key: string) {
    showOptionGroupByKey = { ...showOptionGroupByKey, [key]: !showOptionGroupByKey[key] };
  }
  function toggleShowFromRecipe(key: string) {
    showFromRecipeByKey = { ...showFromRecipeByKey, [key]: !showFromRecipeByKey[key] };
  }
  async function fetchRecipeSuggestions(stepIndex: number, ingIdx: number, query: string) {
    const key = recipeSlotKey(stepIndex, ingIdx);
    if (query.length < 2) {
      recipeSuggestionsBySlot = { ...recipeSuggestionsBySlot, [key]: [] };
      return;
    }
    const res = await fetch(`/api/mealplanning/search-recipes?q=${encodeURIComponent(query)}`);
    const json = await res.json();
    recipeSuggestionsBySlot = { ...recipeSuggestionsBySlot, [key]: json.results ?? [] };
  }
  const debouncedRecipeSearch = debounce(async (stepIndex: number, ingIdx: number, query: string) => {
    await fetchRecipeSuggestions(stepIndex, ingIdx, query);
  }, 300);

  const productTypeOptions = [
    { value: '0', label: 'Ingredient' },
    { value: '1', label: 'Instrument' },
    { value: '2', label: 'Vessel' },
  ];
  const optionGroupOptions = [
    { value: 0, label: '—' },
    { value: 1, label: 'Option A (OR)' },
    { value: 2, label: 'Option B (OR)' },
    { value: 3, label: 'Option C (OR)' },
    { value: 4, label: 'Option D (OR)' },
  ];

  function slugify(s: string): string {
    return s
      .toLowerCase()
      .trim()
      .replace(/[^\w\s-]/g, '')
      .replace(/[\s_-]+/g, '-')
      .replace(/^-+|-+$/g, '');
  }

  $effect(() => {
    const name = state.recipe.name;
    if (name && !state.recipe.slug) {
      state.updateRecipeField('slug', slugify(name));
    }
  });

  let mounted = false;
  onMount(() => {
    mermaid.initialize({ startOnLoad: false });
    mounted = true;
  });
  $effect(() => {
    if (!mounted || !mermaidContainer) return;
    const code = mermaidCode;
    mermaidContainer.innerHTML = '';
    const pre = document.createElement('pre');
    pre.className = 'mermaid';
    pre.textContent = code;
    mermaidContainer.appendChild(pre);
    mermaid.run({ nodes: [pre], suppressErrors: true }).catch(() => {});
  });

  let debugLoadJson = $state('');
  let debugLoadError: string | null = $state(null);
  let debugCopied = $state(false);
  let debugPanelExpanded = $state(false);

  async function copyState() {
    try {
      const dump = state.dumpState();
      await navigator.clipboard.writeText(JSON.stringify(dump, null, 2));
      debugCopied = true;
      debugLoadError = null;
      setTimeout(() => (debugCopied = false), 2000);
    } catch (e) {
      debugLoadError = e instanceof Error ? e.message : String(e);
    }
  }

  function loadStateFromDebug() {
    debugLoadError = null;
    try {
      const parsed = JSON.parse(debugLoadJson);
      if (!parsed || typeof parsed !== 'object') throw new Error('Invalid dump: expected an object');
      if (!parsed.recipe) throw new Error('Invalid dump: missing recipe');
      state.loadState(parsed);
    } catch (e) {
      debugLoadError = e instanceof Error ? e.message : String(e);
    }
  }
</script>

<PageContainer wide>
  <h1>Create Recipe</h1>
  <p><Link href="/recipes">Back to Recipes</Link></p>

  {#if form?.error}
    <Alert variant="error">{form.error}</Alert>
  {/if}
  {#if state.submissionError}
    <Alert variant="error">{state.submissionError}</Alert>
  {/if}
  {#if state.recipe.steps.length < 2}
    <Alert variant="error">Recipe must have at least 2 steps.</Alert>
  {/if}

  <form method="POST" action="?/default" use:enhance class="recipe-form">
    <input type="hidden" name="recipe" value={recipePayload} data-testid="recipe-payload" />

    <div class="recipe-layout">
      <div class="recipe-layout__left">
        <section class="recipe-section recipe-section--info">
          <h2>Recipe details</h2>
          <FormField id="name" label="Name" required>
            <Input
              id="name"
              bind:value={state.recipe.name}
              placeholder="Recipe name"
              required
              dataTestId="recipe-name"
            />
          </FormField>
          <div class="info-row">
            <div class="info-row__slug">
              <FormField id="slug" label="URL Slug">
                <Input
                  id="slug"
                  bind:value={state.recipe.slug}
                  placeholder="recipe-url-slug"
                  dataTestId="recipe-slug"
                />
              </FormField>
            </div>
            <div class="info-row__portions">
              <FormField id="portions" label="Est. Portions" required>
                <div class="portions-range">
                  <NumberInput
                    id="portions"
                    bind:value={state.recipe.minEstimatedPortions}
                    min={1}
                    required
                    dataTestId="recipe-portions"
                  />
                  <span class="portions-range__sep" aria-hidden="true">–</span>
                  <input
                    type="number"
                    id="portions-max"
                    min={1}
                    placeholder="max"
                    class="number-input"
                    data-testid="recipe-portions-max"
                    value={state.recipe.maxEstimatedPortions ?? ''}
                    oninput={(e) => {
                      const v = (e.currentTarget as HTMLInputElement).value;
                      const n = v === '' ? undefined : parseInt(v, 10);
                      state.recipe.maxEstimatedPortions = n !== undefined && !Number.isNaN(n) ? n : undefined;
                    }}
                  />
                </div>
              </FormField>
            </div>
          </div>
          <FormField id="source" label="Source">
            <Input
              id="source"
              bind:value={state.recipe.source}
              placeholder="Recipe source"
              dataTestId="recipe-source"
            />
          </FormField>
          <FormField id="description" label="Description">
            <textarea
              id="description"
              name="description"
              class="textarea textarea--compact"
              placeholder="Recipe description"
              bind:value={state.recipe.description}
              rows="2"
              data-testid="recipe-description"
            ></textarea>
          </FormField>
          <FormField id="portionNames" label="Portion name">
            <div class="portion-name-row">
              <Input
                id="portionName"
                bind:value={state.recipe.portionName}
                placeholder="portion"
                dataTestId="recipe-portion-name"
              />
              <span class="portion-name-row__sep" aria-hidden="true">/</span>
              <Input
                id="pluralPortionName"
                bind:value={state.recipe.pluralPortionName}
                placeholder="portions"
                dataTestId="recipe-plural-portion-name"
              />
            </div>
          </FormField>
          <FormField id="yieldsComponentType" label="Yields (component type)">
            <select
              class="select-input"
              aria-label="Recipe component type"
              value={state.recipe.yieldsComponentType}
              data-testid="recipe-yields-component-type"
              onchange={(e) => {
                const v = (e.currentTarget as HTMLSelectElement).value;
                const n = parseInt(v, 10);
                if (!Number.isNaN(n)) state.recipe.yieldsComponentType = n as typeof state.recipe.yieldsComponentType;
              }}
            >
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_APPETIZER}>Appetizer</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_SOUP}>Soup</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_MAIN}>Main</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_SALAD}>Salad</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_SIDE}>Side</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_DESSERT}>Dessert</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_BEVERAGE}>Beverage</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_AMUSE_BOUCHE}>Amuse bouche</option>
              <option value={MealComponentType.MEAL_COMPONENT_TYPE_UNSPECIFIED}>Unspecified</option>
            </select>
          </FormField>
          <FormField id="eligibleForMeals" label="">
            <label class="checkbox-label">
              <input
                type="checkbox"
                bind:checked={state.recipe.eligibleForMeals}
                data-testid="recipe-eligible-for-meals"
              />
              <span>Eligible for meal plans</span>
            </label>
          </FormField>
        </section>

        <section class="recipe-section recipe-section--prep-tasks">
          <div class="recipe-section--prep-tasks__header">
            <h2>Prep tasks</h2>
            <p class="recipe-section__hint">
              Advance prep (e.g. make dressing ahead). Optionally link which steps each task satisfies.
            </p>
          </div>
          {#each state.recipe.prepTasks ?? [] as prepTask, taskIdx (taskIdx)}
            <Card title={prepTask.name || `Prep task ${taskIdx + 1}`} collapsible>
              <FormField label="Name">
                <Input
                  value={prepTask.name}
                  placeholder="e.g. Make dressing"
                  dataTestId={'prep-task-' + taskIdx + '-name'}
                  oninput={(e) => state.updatePrepTaskField(taskIdx, 'name', e.currentTarget.value)}
                />
              </FormField>
              <FormField label="Description">
                <textarea
                  class="textarea textarea--compact"
                  placeholder="What to do"
                  rows="2"
                  data-testid={'prep-task-' + taskIdx + '-description'}
                  value={prepTask.description}
                  oninput={(e) =>
                    state.updatePrepTaskField(taskIdx, 'description', (e.currentTarget as HTMLTextAreaElement).value)}
                ></textarea>
              </FormField>
              <FormField label="Storage type">
                <select
                  class="select-input"
                  value={prepTask.storageType}
                  data-testid={'prep-task-' + taskIdx + '-storage-type'}
                  onchange={(e) =>
                    state.updatePrepTaskField(taskIdx, 'storageType', (e.currentTarget as HTMLSelectElement).value)}
                >
                  <option value="">—</option>
                  <option value="AirtightContainer">Airtight container</option>
                  <option value="Covered">Covered</option>
                  <option value="WireRack">Wire rack</option>
                </select>
              </FormField>
              <FormField label="Storage instructions">
                <Input
                  value={prepTask.explicitStorageInstructions}
                  placeholder="e.g. Refrigerate up to 3 days"
                  dataTestId={'prep-task-' + taskIdx + '-storage-instructions'}
                  oninput={(e) =>
                    state.updatePrepTaskField(taskIdx, 'explicitStorageInstructions', e.currentTarget.value)}
                />
              </FormField>
              <FormField label="Notes">
                <Input
                  value={prepTask.notes}
                  placeholder="Optional notes"
                  dataTestId={'prep-task-' + taskIdx + '-notes'}
                  oninput={(e) => state.updatePrepTaskField(taskIdx, 'notes', e.currentTarget.value)}
                />
              </FormField>
              <FormField label="Satisfies steps">
                <div class="prep-task-steps">
                  {#each state.recipe.steps as _, stepIdx (stepIdx)}
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        checked={(prepTask.recipeSteps ?? []).some(
                          (rs: RecipePrepTaskStepWithinRecipeCreationRequestInput) =>
                            rs.belongsToRecipeStepIndex === stepIdx,
                        )}
                        data-testid={'prep-task-' + taskIdx + '-step-' + stepIdx}
                        onchange={(e) => {
                          const checked = (e.currentTarget as HTMLInputElement).checked;
                          const current = prepTask.recipeSteps ?? [];
                          if (checked) {
                            state.setPrepTaskRecipeSteps(taskIdx, [
                              ...current.filter(
                                (rs: RecipePrepTaskStepWithinRecipeCreationRequestInput) =>
                                  rs.belongsToRecipeStepIndex !== stepIdx,
                              ),
                              { belongsToRecipeStepIndex: stepIdx, satisfiesRecipeStep: true },
                            ]);
                          } else {
                            state.setPrepTaskRecipeSteps(
                              taskIdx,
                              current.filter(
                                (rs: RecipePrepTaskStepWithinRecipeCreationRequestInput) =>
                                  rs.belongsToRecipeStepIndex !== stepIdx,
                              ),
                            );
                          }
                        }}
                      />
                      <span>Step {stepIdx + 1}</span>
                    </label>
                  {/each}
                </div>
              </FormField>
              <FormField label="">
                <label class="checkbox-label">
                  <input
                    type="checkbox"
                    checked={prepTask.optional}
                    data-testid={'prep-task-' + taskIdx + '-optional'}
                    onchange={(e) =>
                      state.updatePrepTaskField(taskIdx, 'optional', (e.currentTarget as HTMLInputElement).checked)}
                  />
                  <span>Optional</span>
                </label>
              </FormField>
              <div class="prep-task-actions">
                <Button type="button" variant="default" onclick={() => state.removePrepTask(taskIdx)}>
                  Remove prep task
                </Button>
              </div>
            </Card>
          {/each}
          <Button type="button" variant="default" onclick={() => state.addPrepTask()}>Add prep task</Button>
        </section>

        <section class="recipe-section recipe-section--debug">
          <Card title="Debug: state dump" collapsible bind:expanded={debugPanelExpanded}>
            <div class="debug-actions">
              <Button type="button" variant="default" onclick={copyState}>
                {debugCopied ? 'Copied' : 'Copy state'}
              </Button>
              <Button type="button" variant="default" onclick={loadStateFromDebug}>Load state</Button>
            </div>
            {#if debugLoadError}
              <Alert variant="error" class="debug-error">{debugLoadError}</Alert>
            {/if}
            <FormField id="debug-json" label="Paste JSON to restore">
              <textarea
                id="debug-json"
                class="textarea textarea--compact debug-textarea"
                placeholder="Paste JSON with &quot;recipe&quot; and &quot;stepHelpers&quot; keys"
                bind:value={debugLoadJson}
                rows="4"
                data-testid="debug-state-json"
              ></textarea>
            </FormField>
          </Card>
        </section>

        <section class="recipe-section dag-panel">
          <h2>Step flow</h2>
          <div class="dag-panel__diagram" use:useMermaidContainer></div>
        </section>
      </div>

      <div class="recipe-layout__right">
        <div class="step-entry-scroll">
          <section class="recipe-section recipe-section--steps">
            <h2>Steps</h2>
            <div class="steps-list">
              {#each state.recipe.steps as step, stepIndex (stepIndex)}
                <Card title="Step {stepIndex + 1}" collapsible bind:expanded={state.stepHelpers[stepIndex].show}>
                  <FormField label="Preparation" required>
                    <Autocomplete
                      bind:value={state.stepHelpers[stepIndex].preparationQuery}
                      placeholder="Search preparation (e.g. dice, chop)"
                      dataTestId={'recipe-step-' + stepIndex + '-preparation'}
                      suggestions={state.stepHelpers[stepIndex].preparationSuggestions.map((p: ValidPreparation) => ({
                        id: p.id,
                        label: p.name ?? '',
                      }))}
                      onInput={(query: string) => debouncedPrepSearch(stepIndex, query)}
                      onSelect={async (item: { id: string; label: string }) => {
                        const prep = state.stepHelpers[stepIndex].preparationSuggestions.find(
                          (p: ValidPreparation) => p.id === item.id,
                        ) ?? { id: item.id, name: item.label };
                        state.setPreparation(stepIndex, prep as ValidPreparation);
                        const res = await fetch(
                          `/api/mealplanning/search-instruments?preparationId=${encodeURIComponent(prep.id)}`,
                        );
                        const json = await res.json();
                        state.setInstrumentSuggestions(stepIndex, json.results ?? []);
                      }}
                    />
                  </FormField>

                  {#if state.stepHelpers[stepIndex].selectedPreparation}
                    <div class="step-section">
                      <h3>Instruments</h3>
                      {#each step.instruments as instrument, instIdx (`${stepIndex}-inst-${instIdx}`)}
                        {@const instKey = optionalKey(stepIndex, 'instrument', instIdx)}
                        {@const availableForStepInst = getAvailablePreviousProducts(state.recipe.steps, stepIndex)}
                        {@const availableInstrumentProducts = availableForStepInst.filter(
                          (p) => p.type === RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_INSTRUMENT,
                        )}
                        {@const instrumentFromProduct = instrument.productOfRecipeStepIndex != null}
                        {@const instFromStepNum =
                          instrument.productOfRecipeStepIndex != null ? instrument.productOfRecipeStepIndex + 1 : 0}
                        {@const showInstOptionGroup =
                          showOptionGroupByKey[instKey] || (instrument.optionIndex ?? 0) !== 0}
                        <div class="step-item">
                          <FormField label="Instrument {instIdx + 1}">
                            {#if instrumentFromProduct}
                              <div
                                class="from-product-summary"
                                data-testid={'recipe-step-' +
                                  stepIndex +
                                  '-instrument-' +
                                  instIdx +
                                  '-from-product-summary'}
                              >
                                <span class="from-product-summary__label"
                                  >From Step {instFromStepNum}:
                                  <strong
                                    >{getReferencedProductName(
                                      state.recipe.steps,
                                      instrument.productOfRecipeStepIndex,
                                      instrument.productOfRecipeStepProductIndex,
                                    ) ||
                                      instrument.name ||
                                      'Unnamed product'}</strong
                                  ></span
                                >
                                <div class="from-product-summary__actions">
                                  {#if availableInstrumentProducts.length > 1}
                                    <select
                                      class="select-input select-input--inline"
                                      value={`${instrument.productOfRecipeStepIndex}:${instrument.productOfRecipeStepProductIndex}`}
                                      aria-label="Change which step’s product to use"
                                      onchange={(e) => {
                                        const v = (e.currentTarget as HTMLSelectElement).value;
                                        const [si, pi] = v.split(':').map(Number);
                                        if (!Number.isNaN(si) && !Number.isNaN(pi)) {
                                          state.setInstrumentFromProduct(stepIndex, instIdx, si, pi);
                                        }
                                      }}
                                    >
                                      {#each availableInstrumentProducts as opt (`${opt.stepIndex}:${opt.productIndex}`)}
                                        <option value="{opt.stepIndex}:{opt.productIndex}">
                                          Step {opt.stepIndex + 1}: {opt.name || 'Product'}
                                        </option>
                                      {/each}
                                    </select>
                                  {/if}
                                  <Button
                                    type="button"
                                    variant="default"
                                    onclick={() => state.clearInstrumentProductRef(stepIndex, instIdx)}
                                  >
                                    Clear
                                  </Button>
                                </div>
                              </div>
                            {:else}
                              <Autocomplete
                                value={instrument.name}
                                placeholder="Search instrument"
                                dataTestId={'recipe-step-' + stepIndex + '-instrument-' + instIdx}
                                suggestions={state.stepHelpers[stepIndex].instrumentSuggestions.map(
                                  (vpi: ValidPreparationInstrument) => ({
                                    id: vpi.id,
                                    label: (vpi as { instrument?: { name?: string } }).instrument?.name ?? vpi.id,
                                  }),
                                )}
                                onSelect={(item: { id: string; label: string }) => {
                                  const vpi = state.stepHelpers[stepIndex].instrumentSuggestions.find(
                                    (x: ValidPreparationInstrument) => x.id === item.id,
                                  );
                                  if (vpi) state.setInstrument(stepIndex, instIdx, vpi);
                                }}
                              />
                              {#if availableInstrumentProducts.length > 0}
                                {#if showFromProductByKey[instKey]}
                                  <div class="from-product-option">
                                    <span class="from-product-option__label">Use product from earlier step:</span>
                                    <select
                                      class="select-input select-input--narrow"
                                      value=""
                                      data-testid={'recipe-step-' +
                                        stepIndex +
                                        '-instrument-' +
                                        instIdx +
                                        '-from-product'}
                                      onchange={(e) => {
                                        const v = (e.currentTarget as HTMLSelectElement).value;
                                        if (v) {
                                          const [si, pi] = v.split(':').map(Number);
                                          if (!Number.isNaN(si) && !Number.isNaN(pi)) {
                                            state.setInstrumentFromProduct(stepIndex, instIdx, si, pi);
                                          }
                                          (e.currentTarget as HTMLSelectElement).value = '';
                                        }
                                      }}
                                    >
                                      <option value="">— Select from step…</option>
                                      {#each availableInstrumentProducts as opt (`${opt.stepIndex}:${opt.productIndex}`)}
                                        <option value="{opt.stepIndex}:{opt.productIndex}">
                                          Step {opt.stepIndex + 1}: {opt.name || 'Product'}
                                        </option>
                                      {/each}
                                    </select>
                                  </div>
                                {:else}
                                  <button
                                    type="button"
                                    class="reveal-link"
                                    onclick={() => toggleShowFromProduct(instKey)}
                                  >
                                    Use product from earlier step…
                                  </button>
                                {/if}
                              {/if}
                            {/if}
                          </FormField>
                          {#if showInstOptionGroup}
                            <FormField label="Option group">
                              <select
                                class="select-input select-input--narrow"
                                value={String(instrument.optionIndex ?? 0)}
                                data-testid={'recipe-step-' + stepIndex + '-instrument-' + instIdx + '-option-group'}
                                onchange={(e) => {
                                  const v = parseInt((e.currentTarget as HTMLSelectElement).value, 10);
                                  if (!Number.isNaN(v)) state.setInstrumentOptionIndex(stepIndex, instIdx, v);
                                }}
                              >
                                {#each optionGroupOptions as opt (opt.value)}
                                  <option value={opt.value}>{opt.label}</option>
                                {/each}
                              </select>
                            </FormField>
                          {:else}
                            <button type="button" class="reveal-link" onclick={() => toggleShowOptionGroup(instKey)}>
                              Set option group (OR)…
                            </button>
                          {/if}
                          <Button
                            type="button"
                            variant="default"
                            onclick={() => state.removeInstrumentFromStep(stepIndex, instIdx)}
                          >
                            Remove
                          </Button>
                        </div>
                      {/each}
                      <Button type="button" variant="default" onclick={() => state.addInstrumentToStep(stepIndex)}>
                        Add Instrument
                      </Button>
                    </div>

                    <div class="step-section">
                      <h3>Vessels</h3>
                      {#each step.vessels ?? [] as vessel, vesselIdx (`${stepIndex}-vessel-${vesselIdx}`)}
                        {@const vesselKey = optionalKey(stepIndex, 'vessel', vesselIdx)}
                        {@const availableForStepVessel = getAvailablePreviousProducts(state.recipe.steps, stepIndex)}
                        {@const availableVesselProducts = availableForStepVessel.filter(
                          (p) => p.type === RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_VESSEL,
                        )}
                        {@const vesselFromProduct = vessel.productOfRecipeStepIndex != null}
                        {@const vesselFromStepNum =
                          vessel.productOfRecipeStepIndex != null ? vessel.productOfRecipeStepIndex + 1 : 0}
                        {@const showVesselOptionGroup =
                          showOptionGroupByKey[vesselKey] || (vessel.optionIndex ?? 0) !== 0}
                        <div class="step-item">
                          <FormField label="Vessel {vesselIdx + 1}">
                            {#if vesselFromProduct}
                              <div
                                class="from-product-summary"
                                data-testid={'recipe-step-' +
                                  stepIndex +
                                  '-vessel-' +
                                  vesselIdx +
                                  '-from-product-summary'}
                              >
                                <span class="from-product-summary__label"
                                  >From Step {vesselFromStepNum}:
                                  <strong
                                    >{getReferencedProductName(
                                      state.recipe.steps,
                                      vessel.productOfRecipeStepIndex,
                                      vessel.productOfRecipeStepProductIndex,
                                    ) ||
                                      vessel.name ||
                                      'Unnamed product'}</strong
                                  ></span
                                >
                                <div class="from-product-summary__actions">
                                  {#if availableVesselProducts.length > 1}
                                    <select
                                      class="select-input select-input--inline"
                                      value={`${vessel.productOfRecipeStepIndex}:${vessel.productOfRecipeStepProductIndex}`}
                                      aria-label="Change which step’s product to use"
                                      onchange={(e) => {
                                        const v = (e.currentTarget as HTMLSelectElement).value;
                                        const [si, pi] = v.split(':').map(Number);
                                        if (!Number.isNaN(si) && !Number.isNaN(pi)) {
                                          state.setVesselFromProduct(stepIndex, vesselIdx, si, pi);
                                        }
                                      }}
                                    >
                                      {#each availableVesselProducts as opt (`${opt.stepIndex}:${opt.productIndex}`)}
                                        <option value="{opt.stepIndex}:{opt.productIndex}">
                                          Step {opt.stepIndex + 1}: {opt.name || 'Product'}
                                        </option>
                                      {/each}
                                    </select>
                                  {/if}
                                  <Button
                                    type="button"
                                    variant="default"
                                    onclick={() => state.clearVesselProductRef(stepIndex, vesselIdx)}
                                  >
                                    Clear
                                  </Button>
                                </div>
                              </div>
                            {:else}
                              <Autocomplete
                                bind:value={state.stepHelpers[stepIndex].vesselQueries[vesselIdx]}
                                placeholder="Search vessel"
                                dataTestId={'recipe-step-' + stepIndex + '-vessel-' + vesselIdx}
                                suggestions={(state.stepHelpers[stepIndex].vesselSuggestions[vesselIdx] ?? []).map(
                                  (vpv: ValidPreparationVessel) => ({
                                    id: vpv.id,
                                    label: (vpv as { vessel?: { name?: string } }).vessel?.name ?? vpv.id,
                                  }),
                                )}
                                onInput={(query: string) => debouncedVesselSearch(stepIndex, vesselIdx, query)}
                                onSelect={(item: { id: string; label: string }) => {
                                  const vpv = state.stepHelpers[stepIndex].vesselSuggestions[vesselIdx]?.find(
                                    (x: ValidPreparationVessel) => x.id === item.id,
                                  );
                                  if (vpv) state.setVessel(stepIndex, vesselIdx, vpv);
                                }}
                              />
                              {#if availableVesselProducts.length > 0}
                                {#if showFromProductByKey[vesselKey]}
                                  <div class="from-product-option">
                                    <span class="from-product-option__label">Use product from earlier step:</span>
                                    <select
                                      class="select-input select-input--narrow"
                                      value=""
                                      data-testid={'recipe-step-' +
                                        stepIndex +
                                        '-vessel-' +
                                        vesselIdx +
                                        '-from-product'}
                                      onchange={(e) => {
                                        const v = (e.currentTarget as HTMLSelectElement).value;
                                        if (v) {
                                          const [si, pi] = v.split(':').map(Number);
                                          if (!Number.isNaN(si) && !Number.isNaN(pi)) {
                                            state.setVesselFromProduct(stepIndex, vesselIdx, si, pi);
                                          }
                                          (e.currentTarget as HTMLSelectElement).value = '';
                                        }
                                      }}
                                    >
                                      <option value="">— Select from step…</option>
                                      {#each availableVesselProducts as opt (`${opt.stepIndex}:${opt.productIndex}`)}
                                        <option value="{opt.stepIndex}:{opt.productIndex}">
                                          Step {opt.stepIndex + 1}: {opt.name || 'Product'}
                                        </option>
                                      {/each}
                                    </select>
                                  </div>
                                {:else}
                                  <button
                                    type="button"
                                    class="reveal-link"
                                    onclick={() => toggleShowFromProduct(vesselKey)}
                                  >
                                    Use product from earlier step…
                                  </button>
                                {/if}
                              {/if}
                            {/if}
                          </FormField>
                          {#if showVesselOptionGroup}
                            <FormField label="Option group">
                              <select
                                class="select-input select-input--narrow"
                                value={String(vessel.optionIndex ?? 0)}
                                data-testid={'recipe-step-' + stepIndex + '-vessel-' + vesselIdx + '-option-group'}
                                onchange={(e) => {
                                  const v = parseInt((e.currentTarget as HTMLSelectElement).value, 10);
                                  if (!Number.isNaN(v)) state.setVesselOptionIndex(stepIndex, vesselIdx, v);
                                }}
                              >
                                {#each optionGroupOptions as opt (opt.value)}
                                  <option value={opt.value}>{opt.label}</option>
                                {/each}
                              </select>
                            </FormField>
                          {:else}
                            <button type="button" class="reveal-link" onclick={() => toggleShowOptionGroup(vesselKey)}>
                              Set option group (OR)…
                            </button>
                          {/if}
                          <Button
                            type="button"
                            variant="default"
                            onclick={() => state.removeVesselFromStep(stepIndex, vesselIdx)}
                          >
                            Remove
                          </Button>
                        </div>
                      {/each}
                      <Button type="button" variant="default" onclick={() => state.addVesselToStep(stepIndex)}>
                        Add Vessel
                      </Button>
                    </div>

                    <div class="step-section">
                      <h3>Ingredients</h3>
                      {#each step.ingredients as ingredient, ingIdx (`${stepIndex}-ing-${ingIdx}`)}
                        {@const ingKey = optionalKey(stepIndex, 'ingredient', ingIdx)}
                        {@const availableForStep = getAvailablePreviousProducts(state.recipe.steps, stepIndex)}
                        {@const availableIngredientProducts = availableForStep.filter(
                          (p) => p.type === RecipeStepProductType.RECIPE_STEP_PRODUCT_TYPE_INGREDIENT,
                        )}
                        {@const isFromProduct = ingredient.productOfRecipeStepIndex != null}
                        {@const isFromRecipe = ingredient.recipeStepProductRecipeId != null}
                        {@const fromStepNum =
                          ingredient.productOfRecipeStepIndex != null ? ingredient.productOfRecipeStepIndex + 1 : 0}
                        {@const showIngOptionGroup =
                          showOptionGroupByKey[ingKey] || (ingredient.optionIndex ?? 0) !== 0}
                        <div class="step-item step-item--ingredient">
                          <FormField label="Ingredient {ingIdx + 1}">
                            {#if isFromProduct}
                              <div
                                class="from-product-summary"
                                data-testid={'recipe-step-' +
                                  stepIndex +
                                  '-ingredient-' +
                                  ingIdx +
                                  '-from-product-summary'}
                              >
                                <span class="from-product-summary__label"
                                  >From Step {fromStepNum}:
                                  <strong
                                    >{getReferencedProductName(
                                      state.recipe.steps,
                                      ingredient.productOfRecipeStepIndex,
                                      ingredient.productOfRecipeStepProductIndex,
                                    ) ||
                                      ingredient.name ||
                                      'Unnamed product'}</strong
                                  ></span
                                >
                                <div class="from-product-summary__actions">
                                  {#if availableIngredientProducts.length > 1}
                                    <select
                                      class="select-input select-input--inline"
                                      value={`${ingredient.productOfRecipeStepIndex}:${ingredient.productOfRecipeStepProductIndex}`}
                                      aria-label="Change which step’s product to use"
                                      onchange={(e) => {
                                        const v = (e.currentTarget as HTMLSelectElement).value;
                                        const [si, pi] = v.split(':').map(Number);
                                        if (!Number.isNaN(si) && !Number.isNaN(pi)) {
                                          state.setIngredientFromProduct(stepIndex, ingIdx, si, pi);
                                        }
                                      }}
                                    >
                                      {#each availableIngredientProducts as opt (`${opt.stepIndex}:${opt.productIndex}`)}
                                        <option value="{opt.stepIndex}:{opt.productIndex}">
                                          Step {opt.stepIndex + 1}: {opt.name || 'Product'}
                                        </option>
                                      {/each}
                                    </select>
                                  {/if}
                                  <Button
                                    type="button"
                                    variant="default"
                                    onclick={() => state.clearIngredientProductRef(stepIndex, ingIdx)}
                                  >
                                    Clear
                                  </Button>
                                </div>
                              </div>
                            {:else if isFromRecipe}
                              <div
                                class="from-product-summary"
                                data-testid={'recipe-step-' +
                                  stepIndex +
                                  '-ingredient-' +
                                  ingIdx +
                                  '-from-recipe-summary'}
                              >
                                <span class="from-product-summary__label"
                                  >Recipe: <strong>{ingredient.name || 'Unnamed recipe'}</strong
                                  >{#if ingredient.recipeStepProductRecipeSlug}
                                    <span class="from-product-summary__slug"
                                      >({ingredient.recipeStepProductRecipeSlug})</span
                                    >{/if}</span
                                >
                                <div class="from-product-summary__actions">
                                  <Button
                                    type="button"
                                    variant="default"
                                    onclick={() => state.clearIngredientRecipeRef(stepIndex, ingIdx)}
                                  >
                                    Clear
                                  </Button>
                                </div>
                              </div>
                            {:else}
                              <Autocomplete
                                bind:value={state.stepHelpers[stepIndex].ingredientQueries[ingIdx]}
                                placeholder="Search ingredient"
                                dataTestId={'recipe-step-' + stepIndex + '-ingredient-' + ingIdx}
                                suggestions={(state.stepHelpers[stepIndex].ingredientSuggestions[ingIdx] ?? []).map(
                                  (ing: ValidIngredient) => ({ id: ing.id, label: ing.name ?? '' }),
                                )}
                                onInput={(query: string) => debouncedIngredientSearch(stepIndex, ingIdx, query)}
                                onSelect={async (item: { id: string; label: string }) => {
                                  const vips = await fetchIngredientPreparations(
                                    state.stepHelpers[stepIndex].selectedPreparation!.id,
                                  );
                                  const vip = vips.find(
                                    (v: ValidIngredientPreparation) =>
                                      (v.ingredient as { id?: string } | undefined)?.id === item.id,
                                  );
                                  const vimus = await fetch(
                                    `/api/mealplanning/search-measurement-units?ingredientId=${encodeURIComponent(item.id)}`,
                                  ).then((r: Response) => r.json());
                                  const vimuList: ValidIngredientMeasurementUnit[] = vimus.results ?? [];
                                  state.setIngredientMeasurementUnitSuggestions(stepIndex, ingIdx, vimuList);
                                  const ing = {
                                    id: item.id,
                                    name: item.label,
                                    validIngredientPreparationId: vip?.id,
                                  };
                                  const firstVimu = vimuList[0] ?? null;
                                  state.setIngredient(stepIndex, ingIdx, ing, firstVimu);
                                }}
                              />
                              {#if availableIngredientProducts.length > 0}
                                {#if showFromProductByKey[ingKey]}
                                  <div class="from-product-option">
                                    <span class="from-product-option__label">Use product from earlier step:</span>
                                    <select
                                      class="select-input select-input--narrow"
                                      value=""
                                      data-testid={'recipe-step-' +
                                        stepIndex +
                                        '-ingredient-' +
                                        ingIdx +
                                        '-from-product'}
                                      onchange={(e) => {
                                        const v = (e.currentTarget as HTMLSelectElement).value;
                                        if (v) {
                                          const [si, pi] = v.split(':').map(Number);
                                          if (!Number.isNaN(si) && !Number.isNaN(pi)) {
                                            state.setIngredientFromProduct(stepIndex, ingIdx, si, pi);
                                          }
                                          (e.currentTarget as HTMLSelectElement).value = '';
                                        }
                                      }}
                                    >
                                      <option value="">— Select from step…</option>
                                      {#each availableIngredientProducts as opt (`${opt.stepIndex}:${opt.productIndex}`)}
                                        <option value="{opt.stepIndex}:{opt.productIndex}">
                                          Step {opt.stepIndex + 1}: {opt.name || 'Product'}
                                        </option>
                                      {/each}
                                    </select>
                                  </div>
                                {:else}
                                  <button
                                    type="button"
                                    class="reveal-link"
                                    onclick={() => toggleShowFromProduct(ingKey)}
                                  >
                                    Use product from earlier step…
                                  </button>
                                {/if}
                              {/if}
                              {#if showFromRecipeByKey[ingKey]}
                                <div class="from-product-option">
                                  <span class="from-product-option__label">Use recipe as ingredient:</span>
                                  <Autocomplete
                                    value={recipeSearchBySlot[recipeSlotKey(stepIndex, ingIdx)] ?? ''}
                                    placeholder="Search recipe by name"
                                    dataTestId={'recipe-step-' + stepIndex + '-ingredient-' + ingIdx + '-from-recipe'}
                                    suggestions={(recipeSuggestionsBySlot[recipeSlotKey(stepIndex, ingIdx)] ?? []).map(
                                      (r: RecipeSearchHit) => ({ id: r.id, label: r.name }),
                                    )}
                                    onInput={(query: string) => {
                                      recipeSearchBySlot = {
                                        ...recipeSearchBySlot,
                                        [recipeSlotKey(stepIndex, ingIdx)]: query,
                                      };
                                      debouncedRecipeSearch(stepIndex, ingIdx, query);
                                    }}
                                    onSelect={(item: { id: string; label: string }) => {
                                      const hit = (
                                        recipeSuggestionsBySlot[recipeSlotKey(stepIndex, ingIdx)] ?? []
                                      ).find((r: RecipeSearchHit) => r.id === item.id);
                                      if (hit) {
                                        state.setIngredientRecipeRef(stepIndex, ingIdx, hit.id, hit.name, hit.slug);
                                        recipeSearchBySlot = {
                                          ...recipeSearchBySlot,
                                          [recipeSlotKey(stepIndex, ingIdx)]: '',
                                        };
                                        recipeSuggestionsBySlot = {
                                          ...recipeSuggestionsBySlot,
                                          [recipeSlotKey(stepIndex, ingIdx)]: [],
                                        };
                                      }
                                    }}
                                  />
                                </div>
                              {:else}
                                <button type="button" class="reveal-link" onclick={() => toggleShowFromRecipe(ingKey)}>
                                  Use recipe as ingredient…
                                </button>
                              {/if}
                            {/if}
                          </FormField>
                          {#if showIngOptionGroup}
                            <FormField label="Option group">
                              <select
                                class="select-input select-input--narrow"
                                value={String(ingredient.optionIndex ?? 0)}
                                data-testid={'recipe-step-' + stepIndex + '-ingredient-' + ingIdx + '-option-group'}
                                onchange={(e) => {
                                  const v = parseInt((e.currentTarget as HTMLSelectElement).value, 10);
                                  if (!Number.isNaN(v)) state.setIngredientOptionIndex(stepIndex, ingIdx, v);
                                }}
                              >
                                {#each optionGroupOptions as opt (opt.value)}
                                  <option value={opt.value}>{opt.label}</option>
                                {/each}
                              </select>
                            </FormField>
                          {:else}
                            <button type="button" class="reveal-link" onclick={() => toggleShowOptionGroup(ingKey)}>
                              Set option group (OR)…
                            </button>
                          {/if}
                          {#if !isFromProduct && (state.stepHelpers[stepIndex].ingredientMeasurementUnitSuggestions[ingIdx] ?? []).length > 0}
                            <FormField label="Measurement Unit" required>
                              <select
                                class="select-input"
                                value={ingredient.validIngredientMeasurementUnitId ?? ''}
                                data-testid={'recipe-step-' + stepIndex + '-ingredient-' + ingIdx + '-unit'}
                                onchange={(e) => {
                                  const id = (e.currentTarget as HTMLSelectElement).value;
                                  const vimu = (
                                    state.stepHelpers[stepIndex].ingredientMeasurementUnitSuggestions[ingIdx] ?? []
                                  ).find((v: ValidIngredientMeasurementUnit) => v.id === id);
                                  if (vimu) state.setIngredientMeasurementUnit(stepIndex, ingIdx, vimu);
                                }}
                              >
                                <option value="" disabled>Select unit</option>
                                {#each state.stepHelpers[stepIndex].ingredientMeasurementUnitSuggestions[ingIdx] ?? [] as vimu (vimu.id)}
                                  <option value={vimu.id}>
                                    {vimu.measurementUnit?.name ?? vimu.id}
                                  </option>
                                {/each}
                              </select>
                            </FormField>
                          {/if}
                          {#if !isFromProduct}
                            <FormField label="Quantity">
                              <NumberInput
                                bind:value={ingredient.minQuantity}
                                min={0}
                                step={0.25}
                                dataTestId={'recipe-step-' + stepIndex + '-ingredient-' + ingIdx + '-quantity'}
                              />
                            </FormField>
                          {/if}
                          <Button
                            type="button"
                            variant="default"
                            onclick={() => state.removeIngredientFromStep(stepIndex, ingIdx)}
                          >
                            Remove
                          </Button>
                        </div>
                      {/each}
                      <Button type="button" variant="default" onclick={() => state.addIngredientToStep(stepIndex)}>
                        Add Ingredient
                      </Button>
                    </div>

                    <div class="step-section">
                      <h3>Products</h3>
                      {#each step.products as product, prodIdx (`${stepIndex}-prod-${prodIdx}`)}
                        {@const isContinuous =
                          product.minItemQuantity === undefined && product.maxItemQuantity === undefined}
                        {@const productUnits =
                          state.stepHelpers[stepIndex]?.productMeasurementUnitSuggestions[prodIdx] ?? []}
                        <div class="step-item">
                          <FormField label="Product {prodIdx + 1}">
                            <Input
                              value={product.name}
                              placeholder="Product name"
                              dataTestId={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-name'}
                              oninput={(e) => state.updateProduct(stepIndex, prodIdx, { name: e.currentTarget.value })}
                            />
                          </FormField>
                          <FormField label="Type">
                            <select
                              class="select-input"
                              value={String(product.type)}
                              data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-type'}
                              onchange={(e) => {
                                const v = (e.currentTarget as HTMLSelectElement).value;
                                state.updateProduct(stepIndex, prodIdx, {
                                  type: Number(v) as RecipeStepProductType,
                                });
                              }}
                            >
                              {#each productTypeOptions as opt (opt.value)}
                                <option value={opt.value}>{opt.label}</option>
                              {/each}
                            </select>
                          </FormField>
                          <FormField label="Amount type">
                            <select
                              class="select-input select-input--narrow"
                              value={isContinuous ? 'continuous' : 'discrete'}
                              data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-amount-type'}
                              onchange={(e) => {
                                const v = (e.currentTarget as HTMLSelectElement).value;
                                if (v === 'continuous') {
                                  state.updateProduct(stepIndex, prodIdx, {
                                    minItemQuantity: undefined,
                                    maxItemQuantity: undefined,
                                    minMeasurementQuantity: product.minMeasurementQuantity ?? 1,
                                    maxMeasurementQuantity: product.maxMeasurementQuantity,
                                    measurementUnitId: product.measurementUnitId,
                                  });
                                } else {
                                  state.updateProduct(stepIndex, prodIdx, {
                                    minItemQuantity: product.minItemQuantity ?? 1,
                                    maxItemQuantity: product.maxItemQuantity,
                                    minMeasurementQuantity: product.minMeasurementQuantity ?? 1,
                                    maxMeasurementQuantity: product.maxMeasurementQuantity,
                                    measurementUnitId: product.measurementUnitId,
                                  });
                                }
                              }}
                            >
                              <option value="continuous">Continuous (e.g. 2 cups sauce)</option>
                              <option value="discrete">Discrete (e.g. 4 patties, 4 oz each)</option>
                            </select>
                          </FormField>
                          {#if isContinuous}
                            <FormField label="Quantity">
                              <div class="product-qty-row">
                                <input
                                  type="number"
                                  min={0}
                                  step={0.25}
                                  class="number-input"
                                  data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-qty'}
                                  value={product.minMeasurementQuantity ?? ''}
                                  oninput={(e) => {
                                    const v = (e.currentTarget as HTMLInputElement).value;
                                    const n = v === '' ? undefined : parseFloat(v);
                                    state.updateProduct(stepIndex, prodIdx, {
                                      minMeasurementQuantity: n ?? 0,
                                    });
                                  }}
                                />
                                <span class="product-qty-row__sep">–</span>
                                <input
                                  type="number"
                                  min={0}
                                  step={0.25}
                                  placeholder="max"
                                  class="number-input product-qty-max"
                                  data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-qty-max'}
                                  value={product.maxMeasurementQuantity ?? ''}
                                  oninput={(e) => {
                                    const v = (e.currentTarget as HTMLInputElement).value;
                                    const n = v === '' ? undefined : parseFloat(v);
                                    state.updateProduct(stepIndex, prodIdx, {
                                      maxMeasurementQuantity: n !== undefined && !Number.isNaN(n) ? n : undefined,
                                    });
                                  }}
                                />
                              </div>
                            </FormField>
                            <FormField label="Unit">
                              <select
                                class="select-input select-input--narrow"
                                value={product.measurementUnitId ?? ''}
                                data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-unit'}
                                onfocus={async () => {
                                  if (productUnits.length === 0) {
                                    const units = await fetchProductMeasurementUnits();
                                    state.setProductMeasurementUnitSuggestions(stepIndex, prodIdx, units);
                                  }
                                }}
                                onchange={(e) => {
                                  const id = (e.currentTarget as HTMLSelectElement).value;
                                  state.updateProduct(stepIndex, prodIdx, { measurementUnitId: id || undefined });
                                }}
                              >
                                <option value="">—</option>
                                {#each productUnits as mu (mu.id)}
                                  <option value={mu.id}>{(mu as { name?: string }).name ?? mu.id}</option>
                                {/each}
                              </select>
                            </FormField>
                          {:else}
                            <FormField label="Item count">
                              <div class="product-qty-row">
                                <input
                                  type="number"
                                  min={0}
                                  step={1}
                                  class="number-input"
                                  data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-item-count'}
                                  value={product.minItemQuantity ?? ''}
                                  oninput={(e) => {
                                    const v = (e.currentTarget as HTMLInputElement).value;
                                    const n = v === '' ? undefined : parseFloat(v);
                                    state.updateProduct(stepIndex, prodIdx, {
                                      minItemQuantity: n ?? 0,
                                    });
                                  }}
                                />
                                <span class="product-qty-row__sep">–</span>
                                <input
                                  type="number"
                                  min={0}
                                  step={1}
                                  placeholder="max"
                                  class="number-input product-qty-max"
                                  data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-item-count-max'}
                                  value={product.maxItemQuantity ?? ''}
                                  oninput={(e) => {
                                    const v = (e.currentTarget as HTMLInputElement).value;
                                    const n = v === '' ? undefined : parseFloat(v);
                                    state.updateProduct(stepIndex, prodIdx, {
                                      maxItemQuantity: n !== undefined && !Number.isNaN(n) ? n : undefined,
                                    });
                                  }}
                                />
                              </div>
                            </FormField>
                            <FormField label="Per item">
                              <div class="product-qty-row">
                                <input
                                  type="number"
                                  min={0}
                                  step={0.25}
                                  class="number-input"
                                  data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-per-item'}
                                  value={product.minMeasurementQuantity ?? ''}
                                  oninput={(e) => {
                                    const v = (e.currentTarget as HTMLInputElement).value;
                                    const n = v === '' ? undefined : parseFloat(v);
                                    state.updateProduct(stepIndex, prodIdx, {
                                      minMeasurementQuantity: n ?? 0,
                                    });
                                  }}
                                />
                                <select
                                  class="select-input select-input--narrow"
                                  value={product.measurementUnitId ?? ''}
                                  data-testid={'recipe-step-' + stepIndex + '-product-' + prodIdx + '-per-item-unit'}
                                  onfocus={async () => {
                                    if (productUnits.length === 0) {
                                      const units = await fetchProductMeasurementUnits();
                                      state.setProductMeasurementUnitSuggestions(stepIndex, prodIdx, units);
                                    }
                                  }}
                                  onchange={(e) => {
                                    const id = (e.currentTarget as HTMLSelectElement).value;
                                    state.updateProduct(stepIndex, prodIdx, { measurementUnitId: id || undefined });
                                  }}
                                >
                                  <option value="">—</option>
                                  {#each productUnits as mu (mu.id)}
                                    <option value={mu.id}>{(mu as { name?: string }).name ?? mu.id}</option>
                                  {/each}
                                </select>
                              </div>
                            </FormField>
                          {/if}
                          <Button
                            type="button"
                            variant="default"
                            onclick={() => state.removeProductFromStep(stepIndex, prodIdx)}
                          >
                            Remove
                          </Button>
                        </div>
                      {/each}
                      <Button type="button" variant="default" onclick={() => state.addProductToStep(stepIndex)}>
                        Add Product
                      </Button>
                    </div>

                    <FormField label="Instructions">
                      <textarea
                        class="textarea"
                        placeholder="Step instructions"
                        bind:value={step.explicitInstructions}
                        rows="3"
                        data-testid={'recipe-step-' + stepIndex + '-instructions'}
                      ></textarea>
                    </FormField>
                    <FormField label="Notes">
                      <textarea
                        class="textarea textarea--compact"
                        placeholder="Internal or structured notes (optional)"
                        bind:value={step.notes}
                        rows="2"
                        data-testid={'recipe-step-' + stepIndex + '-notes'}
                      ></textarea>
                    </FormField>
                  {/if}

                  <div class="step-actions">
                    <Button
                      type="button"
                      variant="default"
                      onclick={() => state.removeStep(stepIndex)}
                      disabled={state.recipe.steps.length <= 1}
                    >
                      Remove Step
                    </Button>
                  </div>
                </Card>
              {/each}

              <Button type="button" variant="default" onclick={() => state.addStep()}>Add Step</Button>
            </div>
          </section>
        </div>
      </div>

      <div class="form-actions">
        <Button type="submit" disabled={state.recipe.steps.length < 2}>Create Recipe</Button>
      </div>
    </div>
  </form>
</PageContainer>

<style>
  .recipe-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
    width: 100%;
  }

  .recipe-layout {
    display: grid;
    grid-template-columns: 30fr 40fr;
    grid-template-rows: 1fr;
    gap: var(--space-md);
    min-width: 0;
    min-height: calc(100vh - 10rem);
    align-items: stretch;
  }

  @media (max-width: 768px) {
    .recipe-layout {
      grid-template-columns: 1fr;
      min-height: 0;
    }
  }

  .recipe-layout__left {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    min-width: 0;
    min-height: 0;
  }

  .recipe-layout__right {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
  }

  .step-entry-scroll {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  /* Compact info section */
  .recipe-section--info {
    padding: var(--space-sm) var(--space-lg);
    padding-left: var(--space-lg);
    padding-right: var(--space-xl);
    flex-shrink: 0;
    box-sizing: border-box;
  }
  .recipe-section--info h2 {
    margin: 0 0 var(--space-sm);
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
  }
  .recipe-section--info :global(.form-field) {
    margin-bottom: var(--space-xs);
  }
  .recipe-section--info :global(.form-field:last-child) {
    margin-bottom: 0;
  }
  .recipe-section--info :global(label) {
    font-size: var(--font-size-sm);
  }
  .recipe-section--info :global(input),
  .recipe-section--info :global(select) {
    padding: var(--space-xs) var(--space-sm);
    font-size: var(--font-size-sm);
  }

  .info-row {
    display: flex;
    gap: var(--space-lg);
    align-items: flex-end;
  }
  .info-row__slug {
    flex: 1 1 auto;
    min-width: 0;
  }
  .info-row__portions {
    flex: 0 0 auto;
    min-width: 10rem;
  }
  .portions-range {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }
  .portions-range :global(.number-input),
  .portions-range input.number-input {
    width: 3.5rem;
    min-width: 0;
    box-sizing: border-box;
    padding: var(--space-xs) var(--space-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
  }
  .portions-range__sep {
    flex: 0 0 auto;
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
    user-select: none;
  }

  .portion-name-row {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }
  .portion-name-row :global(input) {
    flex: 1 1 auto;
    min-width: 0;
  }
  .portion-name-row__sep {
    flex: 0 0 auto;
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
    user-select: none;
  }
  .checkbox-label {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font-size: var(--font-size-sm);
    cursor: pointer;
  }
  .checkbox-label input[type='checkbox'] {
    width: auto;
  }

  .textarea--compact {
    min-height: 2.5rem;
    padding: var(--space-xs) var(--space-sm);
    font-size: var(--font-size-sm);
    resize: vertical;
  }

  .recipe-section__hint {
    margin: 0 0 var(--space-md);
    font-size: var(--font-size-sm);
    color: var(--color-text-muted);
  }
  .recipe-section--prep-tasks {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }
  .recipe-section--prep-tasks__header h2 {
    margin-bottom: var(--space-xs);
  }
  .recipe-section--prep-tasks__header .recipe-section__hint {
    margin-bottom: 0;
  }
  .prep-task-steps {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-md);
  }
  .prep-task-actions {
    margin-top: var(--space-sm);
  }
  .recipe-section--debug {
    flex-shrink: 0;
  }
  .debug-actions {
    display: flex;
    gap: var(--space-sm);
    margin-bottom: var(--space-sm);
  }
  .debug-actions + :global(.alert) {
    margin-bottom: var(--space-sm);
  }
  .debug-textarea {
    font-family: ui-monospace, monospace;
    font-size: var(--font-size-sm);
  }

  .dag-panel {
    flex-shrink: 0;
    padding: var(--space-sm) var(--space-md);
  }
  .dag-panel h2 {
    margin: 0 0 var(--space-sm);
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
  }
  .dag-panel__diagram {
    min-height: 100px;
  }
  .dag-panel__diagram :global(.mermaid) {
    display: flex;
    justify-content: center;
  }
  .dag-panel__diagram :global(svg) {
    max-width: 100%;
    height: auto;
  }

  .form-actions {
    grid-column: 1 / -1;
    margin-top: var(--space-md);
    flex-shrink: 0;
  }

  .recipe-section {
    padding: var(--space-md);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }

  /* Steps section: spacing between step cards and Add Step button */
  .recipe-section--steps {
    padding: var(--space-md) var(--space-lg);
  }
  .steps-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
    margin-top: var(--space-md);
  }

  .recipe-section h2 {
    margin: 0 0 var(--space-md);
    font-size: 1.25rem;
  }

  .step-section {
    margin-top: var(--space-md);
  }

  .step-section h3 {
    margin: 0 0 var(--space-sm);
    font-size: 1rem;
  }

  .step-item {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-md);
    align-items: center;
    margin-bottom: var(--space-md);
  }

  .step-item :global(.form-field) {
    flex: 1;
    min-width: 12rem;
  }
  .product-qty-row {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }
  .product-qty-row .number-input {
    width: 4rem;
    min-width: 0;
    padding: var(--space-xs) var(--space-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    font-size: var(--font-size-base);
  }
  .product-qty-row .product-qty-max {
    width: 3.5rem;
  }
  .product-qty-row__sep {
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
    user-select: none;
  }

  .step-actions {
    margin-top: var(--space-md);
  }

  .textarea {
    width: 100%;
    padding: var(--space-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    font-family: var(--font-sans);
    font-size: 1rem;
    color: var(--color-text);
    resize: vertical;
  }

  .textarea:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .select-input {
    width: 100%;
    padding: var(--space-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    font-family: var(--font-sans);
    font-size: var(--font-size-base);
    color: var(--color-text);
    background: var(--color-surface);
    cursor: pointer;
    box-shadow: var(--shadow-sm);
    transition:
      border-color var(--transition-fast),
      box-shadow var(--transition-fast);
  }

  .select-input:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px var(--color-primary-muted);
  }

  .select-input--narrow {
    min-width: 10rem;
    max-width: 14rem;
  }

  .select-input--inline {
    display: inline-block;
    min-width: 10rem;
    margin-right: var(--space-sm);
  }

  .from-product-summary {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-sm) var(--space-md);
    background: var(--color-surface-alt, #f5f5f5);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
  }

  .from-product-summary__label {
    flex: 1 1 auto;
    min-width: 10rem;
    font-size: var(--font-size-base);
    color: var(--color-text);
  }
  .from-product-summary__slug {
    font-size: var(--font-size-sm);
    color: var(--color-text-muted);
  }

  .from-product-summary__label strong {
    font-weight: var(--font-weight-medium);
  }

  .from-product-summary__actions {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    flex-shrink: 0;
  }

  .from-product-option {
    margin-bottom: var(--space-sm);
  }

  .from-product-option__label {
    display: block;
    font-size: var(--font-size-sm);
    color: var(--color-text-muted);
    margin-bottom: var(--space-xs);
  }

  .reveal-link {
    display: inline-block;
    margin-top: var(--space-xs);
    padding: 0;
    border: none;
    background: none;
    font-size: var(--font-size-sm);
    color: var(--color-primary, #2563eb);
    text-decoration: underline;
    cursor: pointer;
  }
  .reveal-link:hover {
    text-decoration: none;
  }
</style>
