# Meal Planning System Overview

## Core Concepts

The meal planning system is designed to help groups collaboratively decide what to eat through a democratic voting process. Here's how the key concepts relate:

- **[Recipe](recipes.md)**: A detailed cooking instruction with ingredients, steps, tools, and timing
- **[Meal](meals.md)**: A collection of one or more recipes that form a complete eating experience (e.g., "Dinner" = main course + side dish)
- **Meal Plan Event**: A specific time slot for eating (e.g., "Dinner on Tuesday")
- **Meal Plan Option**: A proposed meal for a specific event (e.g., "Pasta with salad for Tuesday dinner")
- **Meal Plan**: A collection of events with their options, covering a time period (e.g., "This week's meal plan")

## System Architecture

The meal planning system consists of several key components:

### 1. Core Domain Models

- **Meal Planning Manager**: Handles CRUD operations for meals, meal plans, and related entities
- **Recipe Manager**: Manages recipe creation, updates, and retrieval
- **Valid Enumerations Manager**: Manages system-defined valid values (ingredients, measurement units, etc.)

### 2. Voting and Decision Making

- **Schulze Voting Method**: Used for ranking meal plan options within each event
- **Vote Management**: Users can rank options and change votes until the deadline
- **Election Processing**: Background workers tally votes and determine winners

### 3. Background Workers

- **Meal Plan Finalizer**: Processes finalized meal plans and creates grocery lists
- **Grocery List Initializer**: Generates shopping lists from finalized meal plans
- **Task Creator**: Creates prep tasks and cooking assignments
- **Search Data Index Scheduler**: Updates search indices for recipe discovery

## Meal Planning Flow

### 1. Meal Plan Creation

1. User creates a meal plan with a name and time period (ad-hoc, at will)
2. Events are added to the meal plan (e.g., "Dinner on Tuesday")
3. For each event, multiple meal options are proposed
4. Each option references a meal (collection of recipes) - creating users can search for existing meals to add

### 2. Voting Phase

1. Users rank their preferences for each event's options
2. Users can change their votes until the deadline
3. Users can abstain from voting on some options while voting on others
4. The system tracks who has and hasn't voted
5. If no one votes on any options, all options are considered tied and one is chosen and marked as tiebroken

### 3. Finalization

Finalization is a durable saga — one named, linear sequence of steps with per-step state,
retries, and compensations, run on platform-go's `saga` package. It replaced three independently
scheduled jobs coordinated by two boolean columns.

1. A scheduled job (`meal_plan_finalization_starter`, every minute) finds meal plans the pipeline
   still owes something to and writes one saga instance for each, claiming the plan by recording
   the instance's ID in `meal_plans.finalization_saga_id`. Both writes are one transaction, so a
   plan can never be claimed by a saga that does not exist, nor claimed twice.
2. The saga worker advances each instance, running as many steps as one pass allows:
   - **`finalize_meal_plan`** — the Schulze method determines each event's winner and the plan is
     marked finalized.
   - **`create_meal_plan_tasks`** — prep tasks are generated from the winning recipes.
   - **`initialize_grocery_list`** — the grocery list is built from all winning recipes.

Finalizing on a user's request starts the same saga rather than doing the work in the request.

**Idempotency.** Each of the two later steps re-reads `meal_plans.tasks_created` /
`meal_plans.grocery_list_initialized` and does nothing if it is already set. Those flags are
written in the same transaction as the work they describe, which is a stronger guarantee than an
idempotency key committed separately could offer. They are no longer the coordinator — the saga
is — and a plan left half-processed by the jobs this replaced is picked up by the starter's query
and finished by a saga that skips whatever it already has.

**Compensation.** A step that exhausts its retry budget unwinds the saga in reverse, starting at
the step that failed. The two later steps delete the tasks and grocery items *that instance
recorded creating* and clear their flag; anything a user added themselves, and anything an
earlier build created, is untouched. Finalization itself is never undone: a finalized plan is a
decision the account's members can already see. An instance whose compensation also fails lands
in `stuck` and needs an operator — alert on `saga_instances_stuck`.

Lifecycle events (`saga.started`, `saga.step_completed`, `saga.compensating`, `saga.stuck`, …)
go through the outbox on the `saga.lifecycle` topic, in the transaction that records the
transition they describe.

### 4. Execution

1. Users can view their assigned tasks and grocery lists
2. Grocery lists can be modified by any account member (mark items as acquired, edit quantities, etc.)
3. Prep tasks can be completed ahead of time
4. Cooking tasks are completed on the scheduled day

## Key Design Decisions

### Why Schulze Voting?

The Schulze method was chosen after research comparing various election tallying schemes. While no voting system can be perfectly fair in all circumstances, Schulze was found to be the closest to ideal. It handles complex preference rankings well and produces consistent, defensible results.

**TODO**: Implement instant runoff voting as a fallback for very small groups (2-3 people) where Schulze doesn't work as effectively.

### Why Background Workers?

The system uses background workers to handle computationally intensive tasks like:

- Vote tallying and election processing
- Grocery list generation and ingredient consolidation
- Task creation and assignment
- Search index updates

This keeps the API responsive while ensuring complex operations complete reliably.

### Why a Saga for Finalization?

The three stages of finalization were three independently scheduled jobs, each rediscovering its
own work with a "finalized but not yet X" query over two boolean columns. That cost three things:

- **Latency was the sum of the intervals.** A plan whose deadline passed waited for the
  finalizer's tick, then the task creator's, then the grocery list initializer's. A saga pass runs
  as many steps as its budget allows, so the whole chain runs at once.
- **There was no compensation.** A stage that failed for good left the plan half-processed
  forever, with nothing to unwind it and nothing to alert on.
- **The state was not inspectable.** There was no record of which stage a plan was in, when it
  got there, how many attempts it had had, or why it was stuck. `saga_instances` is that record.

The pipeline is linear, which is the constraint the `saga` package declares up front: no
branching, no versioning of in-flight definitions, no parallel fan-out. If finalization ever needs
any of those, the answer is Temporal rather than a bigger saga package.

## Data Models

### Meal Plan Events

Events represent specific eating times within a meal plan. They include:

- **Meal Name**: Enum values like "breakfast", "lunch", "dinner", "snack"
- **Start/End Times**: Flexible timing to accommodate busy family schedules
- **Notes**: Additional context (e.g., "Sarah has volleyball, so dinner is at 7:30")

### Meal Plan Options

Options are proposed meals for specific events:

- **Meal Reference**: Points to a [meal](meals.md) (collection of recipes)
- **Assigned Cook/Dishwasher**: Optional role assignments
- **Notes**: Additional context (e.g. "this was a hit last time")

### Voting System

- **Ranking**: Users rank options in order of preference
- **Abstention**: Users can abstain from voting on some options
- **Vote Changes**: Votes can be modified until the deadline
- **Deadline Enforcement**: Votes can be changed after finalization, but changes have no effect on the already-determined winners

**TODO**: Add validation to prevent vote changes after meal plan finalization to avoid user confusion.

## Recipe Option Selections

Recipes can include [option groups](recipes.md#option-groups-alternative-ingredients-instruments-and-vessels) - alternative ingredients, instruments, or vessels where any one can be used. The meal planning system allows users to specify their preferences for these alternatives through **selections**.

### Selection Fields

When creating a meal plan, users can include selections that specify which alternative to use:

- **Selection Type**: Can be `ingredient`, `instrument`, or `vessel`
- **Recipe ID**: The recipe containing the option group
- **Recipe Step ID**: The step within the recipe containing the option group
- **Ingredient Index**: Which option group in the step (maps to the `Index` field in the recipe)
- **Selected Option Index**: The user's chosen alternative (maps to the `OptionIndex` field in the recipe)

### Grocery List Generation Behavior

When generating grocery lists from finalized meal plans, the system handles selections as follows:

1. **With User Selection**: If the user specified a selection for an option group, only the selected alternative appears in the grocery list
2. **Without Selection (Default)**: If no selection was made, the system defaults to `optionIndex: 0` (the first alternative in the group)

This ensures grocery lists are practical and actionable - users won't see both "butter" and "margarine" for the same recipe step.

### Creating Selections at Meal Plan Creation

Selections can be provided when creating a meal plan:

```json
{
  "votingDeadline": "2024-01-15T18:00:00Z",
  "events": [...],
  "selections": [
    {
      "recipeId": "abc123",
      "recipeStepId": "step456",
      "ingredientIndex": 0,
      "selectedOptionIndex": 1,
      "selectionType": "ingredient"
    }
  ]
}
```

This example selects the second alternative (`optionIndex: 1`, e.g., margarine) for the first option group (`ingredientIndex: 0`) in the specified recipe step.

### Managing Selections

Selections can be:

- **Created**: When initially specifying preferences
- **Updated**: To change the selected option index
- **Archived**: To remove a selection (reverts to default behavior)

### Design Rationale

The selection system was designed with these principles:

1. **Sensible Defaults**: Without explicit selections, the first option (`optionIndex: 0`) is used - typically the "primary" or most common ingredient
2. **User Control**: Users who care about specific alternatives can override defaults
3. **Clean Grocery Lists**: Only selected alternatives appear, avoiding confusion
4. **Flexibility**: Supports ingredients, instruments, and vessels to cover all recipe variation scenarios

## Integration Points

### User Ingredient Preferences

Users can set preferences for ingredients they prefer or prefer not to eat. This data is currently not used by the system but is intended for future features like:

- Warning users about recipes containing disliked ingredients
- Highlighting recipes featuring preferred ingredients

**TODO**: Implement recipe filtering based on user ingredient preferences.

### Recipe Analysis

The system analyzes [recipes](recipes.md) to:

- Identify prep tasks that can be done ahead of time
- Determine cooking dependencies and timing
- Generate appropriate task assignments

## Current Limitations

1. **Polling to start**: The saga runs its whole chain in one pass, but a plan still waits up to
   the starter's interval to enter the pipeline; the starter polls even when there is no work
2. **No parallel steps**: The saga package is linear by design. Building the grocery list and the
   prep tasks concurrently is outside what it does, and the answer for that would be Temporal
3. **No Real-time Updates**: Users must refresh to see finalization results
4. **Limited Election Methods**: Only Schulze is implemented (Instant Runoff is defined but not used)
5. **Manual Task Assignment**: Prep tasks aren't automatically assigned to users
6. **No UI**: This is a backend API service only - access via gRPC/HTTP endpoints (e.g., Postman)
7. **Recipe Management**: Only service admins can create recipes; users can clone existing recipes

## Future Improvements

### Known Edge Cases

- **Recipe Deletion**: If a [recipe](recipes.md) is deleted after a meal plan is created but before finalization, it could permanently break the meal plan
- **Account Membership Changes**: If a user leaves an account after voting but before finalization, their vote is still counted
- **Overlapping Events**: The system doesn't prevent overlapping meal times

**TODO**: Add validation to prevent recipe deletion when referenced by active meal plans.
**TODO**: Add validation to prevent meal plan modifications after finalization.
**TODO**: Add integration tests for account membership changes during voting.

### Planned Improvements

1. **Queue-Based Architecture**: Move from cron jobs to queue-based processing for better scalability
2. **Enhanced Validation**: Add comprehensive checks for recipe/meal/account interactions
3. **Grocery List Consolidation**: Build interfaces for combining similar ingredients
4. **Notification System**: Ensure user notifications are sent when meal plans are created and finalized

**TODO**: Add integration tests to verify notification behavior.
**TODO**: Add integration tests to verify email sending behavior.

## Testing

The system includes comprehensive integration tests covering:

- Complete meal plan lifecycle (creation, voting, finalization)
- [Recipe](recipes.md) management and validation
- User ingredient preferences
- Valid enumeration management
- Voting logic and election processing

**TODO**: Add performance tests for background workers.
**TODO**: Add integration tests for edge cases around account membership and recipe deletion.
