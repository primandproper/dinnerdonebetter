import { randomUUID } from 'node:crypto';
import { describe, it, expect } from 'vitest';
import { createRecipeCreatorState } from './RecipeCreatorState';
import { renderMermaidForRecipeCreationInput } from './recipeMermaid';

describe('renderMermaidForRecipeCreationInput', () => {
  it('says there is nothing to draw for a recipe with no steps', () => {
    const state = createRecipeCreatorState();
    state.removeStep(1);
    state.removeStep(0);

    expect(renderMermaidForRecipeCreationInput(state.recipe, [])).toContain('No steps yet');
  });

  it('draws one node per step, named for its preparation', () => {
    const state = createRecipeCreatorState();
    const preparations = [randomUUID(), randomUUID()];

    const chart = renderMermaidForRecipeCreationInput(state.recipe, preparations);

    expect(chart).toContain(`Step1["Step #1 (${preparations[0]})"];`);
    expect(chart).toContain(`Step2["Step #2 (${preparations[1]})"];`);
  });

  it('labels the edge with everything one step hands the next', () => {
    const state = createRecipeCreatorState();
    state.addVesselToStep(1);
    state.setIngredientFromProduct(1, 0, 0, 0);
    state.setVesselFromProduct(1, 0, 0, 0);

    const chart = renderMermaidForRecipeCreationInput(state.recipe, [randomUUID(), randomUUID()]);

    expect(chart).toContain('Step1 -->|"1 ingredient and 1 vessel"| Step2;');
    expect(chart).not.toContain('Step2 -->');
  });

  it('keeps a quote in a preparation name from closing the label', () => {
    const state = createRecipeCreatorState();
    const name = `${randomUUID()}"]; injected`;

    const chart = renderMermaidForRecipeCreationInput(state.recipe, [name]);

    expect(chart).not.toContain(name);
    expect(chart).toContain(name.replace(/"/g, '&#34;'));
  });

  it('gathers a prep task’s steps into a subgraph', () => {
    const state = createRecipeCreatorState();
    const taskName = randomUUID();
    state.addPrepTask();
    state.updatePrepTaskField(0, 'name', taskName);
    state.setPrepTaskRecipeSteps(0, [{ belongsToRecipeStepIndex: 1, satisfiesRecipeStep: false }]);

    const chart = renderMermaidForRecipeCreationInput(state.recipe, []);

    expect(chart).toContain(`subgraph prep0 ["${taskName} (prep task #1)"]\n\t\tStep2\n\tend`);
  });
});
