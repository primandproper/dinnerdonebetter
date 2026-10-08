/**
 * Renders a Mermaid flowchart for RecipeCreationRequestInput, mirroring the backend
 * RenderMermaidDiagramForRecipe logic (recipe_analyzer.go).
 */

import type { RecipeCreationRequestInput, RecipeStepCreationRequestInput } from './client-types';

/** Escape a label for Mermaid: quotes and backslashes. */
function escapeMermaidLabel(s: string): string {
  return s.replace(/\\/g, '\\\\').replace(/"/g, '&#34;');
}

/** Pluralize for edge labels (e.g. "1 ingredient", "2 instruments"). */
function pluralize(count: number, singular: string, plural: string): string {
  return count === 1 ? `${count} ${singular}` : `${count} ${plural}`;
}

/**
 * Count what step `fromIndex` provides to step `toIndex` via productOfRecipeStepIndex
 * (ingredients, instruments, vessels). Returns a single label string for the edge.
 */
function stepProvidesWhatFromTo(steps: RecipeStepCreationRequestInput[], fromIndex: number, toIndex: number): string {
  const from = steps[fromIndex];
  const to = steps[toIndex];
  if (!from || !to) return '';

  let ingredients = 0;
  let instruments = 0;
  let vessels = 0;

  // to.ingredients that reference from's products (by productOfRecipeStepIndex)
  for (const ing of to.ingredients) {
    if (ing.productOfRecipeStepIndex === fromIndex) ingredients++;
  }
  for (const inst of to.instruments) {
    if (inst.productOfRecipeStepIndex === fromIndex) instruments++;
  }
  for (const v of to.vessels) {
    if (v.productOfRecipeStepIndex === fromIndex) vessels++;
  }

  const parts: string[] = [];
  if (ingredients > 0) parts.push(pluralize(ingredients, 'ingredient', 'ingredients'));
  if (instruments > 0) parts.push(pluralize(instruments, 'instrument', 'instruments'));
  if (vessels > 0) parts.push(pluralize(vessels, 'vessel', 'vessels'));

  if (parts.length === 0) return '';
  if (parts.length === 1) return parts[0];
  return parts.slice(0, -1).join(', ') + ' and ' + parts[parts.length - 1];
}

/**
 * Render a Mermaid flowchart TD string for the given recipe creation input.
 * @param recipe - The in-progress recipe (steps, optional prepTasks).
 * @param stepLabels - Preparation name per step index (e.g. from stepHelpers[i].selectedPreparation?.name).
 */
export function renderMermaidForRecipeCreationInput(
  recipe: RecipeCreationRequestInput,
  stepLabels: (string | undefined | null)[],
): string {
  const steps = recipe.steps ?? [];
  if (steps.length === 0) {
    return 'flowchart TD;\n\tempty["No steps yet"];\n';
  }

  const lines: string[] = ['flowchart TD;'];

  // Nodes: one per main-recipe step (1-based graph ID)
  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    const gid = (step?.index ?? i) + 1;
    const prepName = stepLabels[i]?.trim() || '…';
    const label = `Step #${gid} (${prepName})`;
    lines.push(`\tStep${gid}["${escapeMermaidLabel(label)}"];`);
  }

  // Edges: from step i to step j when j consumes i's products
  for (let i = 0; i < steps.length; i++) {
    for (let j = 0; j < steps.length; j++) {
      if (i === j) continue;
      const provides = stepProvidesWhatFromTo(steps, i, j);
      if (provides) {
        const fromGid = (steps[i]?.index ?? i) + 1;
        const toGid = (steps[j]?.index ?? j) + 1;
        lines.push(`\tStep${fromGid} -->|"${escapeMermaidLabel(provides)}"| Step${toGid};`);
      }
    }
  }

  // Prep task subgraphs (if present)
  const prepTasks = recipe.prepTasks ?? [];
  for (let ptIdx = 0; ptIdx < prepTasks.length; ptIdx++) {
    const prep = prepTasks[ptIdx];
    if (!prep?.recipeSteps?.length) continue;
    const name = prep.name?.trim() || `Prep task #${ptIdx + 1}`;
    lines.push(`\tsubgraph prep${ptIdx} ["${escapeMermaidLabel(name)} (prep task #${ptIdx + 1})"]`);
    for (const rs of prep.recipeSteps) {
      const stepIdx = rs.belongsToRecipeStepIndex;
      if (stepIdx >= 0 && stepIdx < steps.length) {
        const gid = (steps[stepIdx]?.index ?? stepIdx) + 1;
        lines.push(`\t\tStep${gid}`);
      }
    }
    lines.push('\tend');
  }

  return lines.join('\n');
}
