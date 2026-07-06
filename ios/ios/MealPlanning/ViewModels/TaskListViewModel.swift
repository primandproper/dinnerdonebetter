//
//  TaskListViewModel.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import SwiftProtobuf
import SwiftUI

@Observable
@MainActor
class TaskListViewModel {
  // Data
  var tasks: [Mealplanning_MealPlanTask] = []
  var mealPlan: Mealplanning_MealPlan
  var loadedRecipes: [String: Mealplanning_Recipe] = [:]
  var loadedPrepTasks: [String: Mealplanning_RecipePrepTask] = [:]

  // Track expanded state for parent tasks: [taskID: isExpanded]
  var expandedTasks: [String: Bool] = [:]

  // Track subtask (step) completion state: [parentTaskID: [stepID: isCompleted]]
  var subtaskCompletionState: [String: [String: Bool]] = [:]

  // Loading states
  var isLoading = false
  var isUpdating = false
  var errorMessage: String?

  /// Temperature display unit from user_temperature_unit setting. Reads from UserSettingsService.
  var temperatureUnit: String {
    userSettingsService.value(for: "user_temperature_unit", default: "fahrenheit")
  }

  private let authManager: AuthenticationManager
  private let userSettingsService: UserSettingsService

  init(
    mealPlan: Mealplanning_MealPlan,
    tasks: [Mealplanning_MealPlanTask],
    authManager: AuthenticationManager,
    userSettingsService: UserSettingsService
  ) {
    self.mealPlan = mealPlan
    self.tasks = tasks
    self.authManager = authManager
    self.userSettingsService = userSettingsService
  }

  func loadTasks() async {
    isLoading = true
    errorMessage = nil

    do {
      let fetchedTasks = try await fetchTasks()
      self.tasks = fetchedTasks

      // Log task statuses after reload
      for task in fetchedTasks {
        print("📋 Reloaded task \(task.id): status = \(task.status)")
      }

      // Initialize subtask completion state
      initializeSubtaskCompletionState()

      // Load recipes for all tasks (will try to find recipe IDs from meal options if needed)
      await loadRecipesForTasks()
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      errorMessage = "Failed to load tasks: \(error.localizedDescription)"
      print("❌ Error loading tasks: \(error)")
    }

    isLoading = false
  }

  private func initializeSubtaskCompletionState() {
    // Initialize subtask completion state
    // Preserve existing completion state if task status hasn't changed
    for task in tasks where task.hasRecipePrepTask {
      let isTaskFinished = task.status == .finished
      let existingStates = subtaskCompletionState[task.id] ?? [:]

      var stepStates: [String: Bool] = [:]
      for taskStep in task.recipePrepTask.taskSteps {
        // Preserve existing state if task status hasn't changed, otherwise use task status
        if isTaskFinished {
          stepStates[taskStep.id] = true
        } else {
          // Preserve existing completion state for unfinished tasks
          stepStates[taskStep.id] = existingStates[taskStep.id] ?? false
        }
      }
      subtaskCompletionState[task.id] = stepStates
    }
  }

  func toggleExpanded(taskID: String) {
    expandedTasks[taskID] = !(expandedTasks[taskID] ?? false)  // Default to false (minimized)
  }

  func isExpanded(taskID: String) -> Bool {
    return expandedTasks[taskID] ?? false  // Default to false (minimized)
  }

  func toggleSubtaskCompletion(parentTaskID: String, stepID: String) {
    if subtaskCompletionState[parentTaskID] == nil {
      subtaskCompletionState[parentTaskID] = [:]
    }
    let currentState = subtaskCompletionState[parentTaskID]?[stepID] ?? false
    let newState = !currentState
    subtaskCompletionState[parentTaskID]?[stepID] = newState

    // Don't auto-complete parent task - user must click parent checkbox when all subtasks are done
    // But if a subtask is unchecked and parent is finished, mark parent as unfinished
    guard let parentTask = tasks.first(where: { $0.id == parentTaskID }) else {
      return
    }

    let allCompleted = areAllSubtasksCompleted(parentTaskID: parentTaskID)

    // If any subtask is unchecked and parent is finished, mark parent as unfinished
    if !allCompleted && parentTask.status == .finished {
      Task {
        await toggleTaskStatus(parentTask)
      }
    }
  }

  func isSubtaskCompleted(parentTaskID: String, stepID: String) -> Bool {
    return subtaskCompletionState[parentTaskID]?[stepID] ?? false
  }

  func areAllSubtasksCompleted(parentTaskID: String) -> Bool {
    guard let stepStates = subtaskCompletionState[parentTaskID] else {
      print("🔍 areAllSubtasksCompleted: No step states for task \(parentTaskID)")
      return false
    }
    // If no subtasks, return false (or could return true if you want empty tasks to be considered complete)
    if stepStates.isEmpty {
      print("🔍 areAllSubtasksCompleted: Empty step states for task \(parentTaskID)")
      return false
    }
    let allCompleted = stepStates.values.allSatisfy { $0 }
    print(
      "🔍 areAllSubtasksCompleted for task \(parentTaskID): \(allCompleted) (total steps: \(stepStates.count), completed: \(stepStates.values.filter { $0 }.count))"
    )
    return allCompleted
  }

  // Group tasks with their recipe prep task steps as subtasks
  // Each task becomes a parent, with its prep task steps as subtasks
  func getGroupedTasks() -> [TaskGroup] {
    var taskGroups: [TaskGroup] = []

    // Sort tasks by creation time for consistent ordering
    let sortedTasks = tasks.sorted { task1, task2 in
      let time1 = task1.createdAt.seconds
      let time2 = task2.createdAt.seconds
      if time1 != time2 {
        return time1 < time2
      }
      return task1.id < task2.id
    }

    // Each task becomes a parent with its steps as subtasks
    for task in sortedTasks {
      // Get the step data and recipeID for this task
      let (stepData, recipeID) = getStepDataAndRecipeIDForTask(task)

      print("📋 Task \(task.id): Found \(stepData.count) steps")

      // Create subtask items from the step data
      let subtasks = stepData.map { stepInfo in
        createSubtaskFromStep(task: task, stepInfo: stepInfo, recipeID: recipeID)
      }

      print("📋 Task \(task.id): Created \(subtasks.count) subtasks")

      taskGroups.append(TaskGroup(parent: task, subtasks: subtasks))
    }

    print("📋 Total task groups: \(taskGroups.count)")
    return taskGroups
  }

  // Helper to get step data and the resolved recipeID for a task
  private func getStepDataAndRecipeIDForTask(_ task: Mealplanning_MealPlanTask) -> (
    [StepInfo], String?
  ) {
    guard task.hasRecipePrepTask else {
      print("⚠️ Task \(task.id) has no recipePrepTask")
      return ([], nil)
    }

    let prepTask = task.recipePrepTask
    var recipeID = prepTask.belongsToRecipe

    print("🔍 Task \(task.id): prepTask.belongsToRecipe = '\(recipeID)'")
    print("🔍 Task \(task.id): prepTask.taskSteps.count = \(prepTask.taskSteps.count)")

    // If belongsToRecipe is empty, try to find it from the meal option
    if recipeID.isEmpty && task.hasMealPlanOption {
      let meal = task.mealPlanOption.meal
      for component in meal.components {
        let candidateRecipeID = component.recipe.id
        for taskStep in prepTask.taskSteps where !taskStep.belongsToRecipeStep.isEmpty {
          if component.recipe.steps.contains(where: { $0.id == taskStep.belongsToRecipeStep }) {
            recipeID = candidateRecipeID
            print("✅ Found recipe ID \(recipeID) from meal option")
            break
          }
        }
        if !recipeID.isEmpty {
          break
        }
      }
    }

    if recipeID.isEmpty {
      print("⚠️ Task \(task.id): No recipe ID found")
      return ([], nil)
    }

    guard let recipe = loadedRecipes[recipeID] else {
      print(
        "⚠️ Task \(task.id): Recipe \(recipeID) not loaded yet. Loaded recipes: \(loadedRecipes.keys.joined(separator: ", "))"
      )
      return ([], recipeID)
    }

    print("✅ Task \(task.id): Recipe \(recipeID) loaded, has \(recipe.steps.count) steps")

    var stepData: [StepInfo] = []

    for taskStep in prepTask.taskSteps where !taskStep.belongsToRecipeStep.isEmpty {
      if let stepIndex = recipe.steps.firstIndex(where: { $0.id == taskStep.belongsToRecipeStep }) {
        let step = recipe.steps[stepIndex]
        let formatted = formatStepTitle(step: step)
        if !formatted.isEmpty {
          let ingredients = step.ingredients.compactMap { ingredient -> String? in
            ingredient.name.isEmpty ? nil : ingredient.name
          }
          stepData.append(
            StepInfo(
              stepID: taskStep.id,
              description: formatted,
              ingredientNames: ingredients,
              recipeStepID: taskStep.belongsToRecipeStep
            ))
          print("✅ Added step: \(formatted)")
        }
      } else {
        print("⚠️ Step \(taskStep.belongsToRecipeStep) not found in recipe")
      }
    }

    print("📋 Task \(task.id): Returning \(stepData.count) step data items")
    return (stepData, recipeID)
  }

  private func createSubtaskFromStep(
    task: Mealplanning_MealPlanTask,
    stepInfo: StepInfo,
    recipeID: String?
  ) -> SubtaskDisplayInfo {
    SubtaskDisplayInfo(
      id: "\(task.id)_step_\(stepInfo.stepID)",
      parentTaskID: task.id,
      stepID: stepInfo.stepID,
      description: stepInfo.description,
      ingredientNames: stepInfo.ingredientNames,
      recipeID: recipeID,
      recipeStepID: stepInfo.recipeStepID
    )
  }

  // Determines the earliest start date for a task based on its time buffer and event time
  func canStartAt(task: Mealplanning_MealPlanTask) -> Date? {
    guard task.hasRecipePrepTask else { return nil }
    let prepTask = task.recipePrepTask
    guard prepTask.hasMaxTimeBufferBeforeRecipeInSeconds else { return nil }

    guard task.hasMealPlanOption else { return nil }
    let eventID = task.mealPlanOption.belongsToMealPlanEvent
    guard !eventID.isEmpty,
      let event = mealPlan.events.first(where: { $0.id == eventID })
    else { return nil }

    let eventTime = HomeViewModel.timestampToDate(event.startsAt)
    let maxBuffer = prepTask.maxTimeBufferBeforeRecipeInSeconds
    return eventTime.addingTimeInterval(-Double(maxBuffer))
  }

  private func isStartableNow(_ task: Mealplanning_MealPlanTask) -> Bool {
    guard let startDate = canStartAt(task: task) else { return true }
    return Date() >= startDate
  }

  // Get all task groups, separated by status and startability
  func getReadyNowGroups() -> [TaskGroup] {
    let allGroups = getGroupedTasks()
    return allGroups.filter { $0.parent.status == .unfinished && isStartableNow($0.parent) }
  }

  func getDoLaterGroups() -> [TaskGroup] {
    let allGroups = getGroupedTasks()
    return
      allGroups
      .filter { $0.parent.status == .unfinished && !isStartableNow($0.parent) }
      .sorted { group1, group2 in
        let date1 = canStartAt(task: group1.parent) ?? .distantFuture
        let date2 = canStartAt(task: group2.parent) ?? .distantFuture
        return date1 < date2
      }
  }

  func getFinishedGroups() -> [TaskGroup] {
    let allGroups = getGroupedTasks()
    return allGroups.filter { $0.parent.status == .finished }
  }

  private func loadRecipesForTasks() async {
    // Collect unique recipe IDs from tasks and loaded prep tasks
    var recipeIDs: Set<String> = []

    for task in tasks where task.hasRecipePrepTask {
      let prepTask = task.recipePrepTask
      let recipeID = findRecipeIDForTask(task: task, prepTask: prepTask)
      if !recipeID.isEmpty {
        recipeIDs.insert(recipeID)
      }
    }

    // Load each recipe
    for recipeID in recipeIDs {
      do {
        let recipe = try await fetchRecipe(recipeID: recipeID)
        loadedRecipes[recipeID] = recipe
      } catch {
        await authManager.invalidateCredentialsIfSessionError(error)
        print("⚠️ Failed to load recipe \(recipeID): \(error)")
      }
    }
  }

  private func findRecipeIDForTask(
    task: Mealplanning_MealPlanTask,
    prepTask: Mealplanning_RecipePrepTask
  ) -> String {
    var recipeID = prepTask.belongsToRecipe

    // If belongsToRecipe is empty, check loaded prep task
    if recipeID.isEmpty && !prepTask.id.isEmpty {
      if let loadedPrepTask = loadedPrepTasks[prepTask.id] {
        recipeID = loadedPrepTask.belongsToRecipe
      }
    }

    // If belongsToRecipe is still empty, try to find it from the meal option
    if recipeID.isEmpty && task.hasMealPlanOption {
      recipeID = findRecipeIDFromMealOption(task: task, prepTask: prepTask)
    }

    return recipeID
  }

  private func findRecipeIDFromMealOption(
    task: Mealplanning_MealPlanTask,
    prepTask: Mealplanning_RecipePrepTask
  ) -> String {
    let meal = task.mealPlanOption.meal
    // Try to find the recipe by matching step IDs
    for component in meal.components {
      let candidateRecipeID = component.recipe.id
      // Check if any task step belongs to this recipe's steps
      for taskStep in prepTask.taskSteps where !taskStep.belongsToRecipeStep.isEmpty {
        // Check if this step ID exists in the component's recipe
        if component.recipe.steps.contains(where: { $0.id == taskStep.belongsToRecipeStep }) {
          return candidateRecipeID
        }
      }
    }
    return ""
  }

  func toggleTaskStatus(_ task: Mealplanning_MealPlanTask) async {
    print("🔄 toggleTaskStatus called for task \(task.id), current status: \(task.status)")
    isUpdating = true
    errorMessage = nil

    do {
      let newStatus: Mealplanning_MealPlanTaskStatus =
        task.status == .finished
        ? .unfinished : .finished

      print("🔄 Updating task \(task.id) to status: \(newStatus)")

      // Keep subtask state in sync: unchecking clears all; checking marks all done
      if newStatus == .unfinished {
        if let stepStates = subtaskCompletionState[task.id] {
          for stepID in stepStates.keys {
            subtaskCompletionState[task.id]?[stepID] = false
          }
        }
      } else if task.hasRecipePrepTask {
        for taskStep in task.recipePrepTask.taskSteps {
          if subtaskCompletionState[task.id] == nil {
            subtaskCompletionState[task.id] = [:]
          }
          subtaskCompletionState[task.id]?[taskStep.id] = true
        }
      }

      // Optimistically update the task status in the local array
      if let taskIndex = tasks.firstIndex(where: { $0.id == task.id }) {
        var updatedTask = tasks[taskIndex]
        updatedTask.status = newStatus
        tasks[taskIndex] = updatedTask
        print("✅ Optimistically updated task status in local array")
      }

      try await updateTaskStatus(taskID: task.id, status: newStatus)
      print("✅ Task status updated successfully on server")

      // Reload tasks to get updated data from server
      await loadTasks()
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      errorMessage = "Failed to update task: \(error.localizedDescription)"
      print("❌ Error updating task: \(error)")

      // Revert optimistic update on error
      if let taskIndex = tasks.firstIndex(where: { $0.id == task.id }) {
        var revertedTask = tasks[taskIndex]
        revertedTask.status = task.status  // Revert to original status
        tasks[taskIndex] = revertedTask
      }
    }

    isUpdating = false
  }

  // Computed properties for filtering
  var unfinishedTasks: [Mealplanning_MealPlanTask] {
    tasks.filter { $0.status == .unfinished }
  }

  var finishedTasks: [Mealplanning_MealPlanTask] {
    tasks.filter { $0.status == .finished }
  }
}

// TaskGroup represents a parent task with its subtasks
struct TaskGroup {
  let parent: Mealplanning_MealPlanTask
  let subtasks: [SubtaskDisplayInfo]
}

// SubtaskDisplayInfo carries everything SubtaskRow needs for display and navigation
struct SubtaskDisplayInfo: Identifiable {
  let id: String
  let parentTaskID: String
  let stepID: String
  let description: String
  let ingredientNames: [String]
  let recipeID: String?
  let recipeStepID: String?
}

// StepInfo represents a recipe step for display (internal to view model)
struct StepInfo {
  let stepID: String
  let description: String
  let ingredientNames: [String]
  let recipeStepID: String?
}

// MARK: - TaskListViewModel Helpers

extension TaskListViewModel {
  // Helper to format step title
  private func formatStepTitle(step: Mealplanning_RecipeStep) -> String {
    if step.hasPreparation && !step.preparation.name.isEmpty {
      return step.preparation.name
    }
    return "Step \(Int(step.index) + 1)"
  }

  private func fetchRecipe(recipeID: String) async throws -> Mealplanning_Recipe {
    var request = Mealplanning_GetRecipeRequest()
    request.recipeID = recipeID

    let response = try await authManager.authenticatedCall("getRecipe", idempotent: true) {
      client, metadata, options in
      try await client.mealPlanning.getRecipe(request, metadata: metadata, options: options)
    }

    return response.result
  }

  private func fetchTasks() async throws -> [Mealplanning_MealPlanTask] {
    var request = Mealplanning_GetMealPlanTasksRequest()
    request.mealPlanID = mealPlan.id

    let response = try await authManager.authenticatedCall("getMealPlanTasks", idempotent: true) {
      client, metadata, options in
      try await client.mealPlanning.getMealPlanTasks(request, metadata: metadata, options: options)
    }

    // Deduplicate tasks by ID (keep first occurrence)
    var seenIDs: Set<String> = []
    let uniqueTasks = response.results.filter { task in
      if seenIDs.contains(task.id) {
        return false
      }
      seenIDs.insert(task.id)
      return true
    }

    return uniqueTasks
  }

  private func updateTaskStatus(
    taskID: String,
    status: Mealplanning_MealPlanTaskStatus
  ) async throws {
    var request = Mealplanning_UpdateMealPlanTaskStatusRequest()
    request.mealPlanID = mealPlan.id
    request.mealPlanTaskID = taskID
    var input = Mealplanning_MealPlanTaskStatusChangeRequestInput()
    input.status = status
    request.input = input

    _ = try await authManager.authenticatedCall("updateMealPlanTaskStatus") {
      client, metadata, options in
      try await client.mealPlanning.updateMealPlanTaskStatus(
        request, metadata: metadata, options: options)
    }
  }
}
