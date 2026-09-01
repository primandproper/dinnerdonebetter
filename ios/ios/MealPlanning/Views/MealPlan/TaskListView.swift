//
//  TaskListView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftProtobuf
import SwiftUI

struct TaskListView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @State private var viewModel: TaskListViewModel

  init(
    mealPlan: any MealPlanDisplayable,
    tasks: [Mealplanning_MealPlanTask],
    authManager: AuthenticationManager,
    userSettingsService: UserSettingsService
  ) {
    _viewModel = State(
      initialValue: TaskListViewModel(
        mealPlan: mealPlan,
        tasks: tasks,
        authManager: authManager,
        userSettingsService: userSettingsService
      )
    )
  }

  var body: some View {
    ScrollView {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.xl) {
        // Header
        headerSection

        if viewModel.isLoading {
          DSLoadingView("Loading tasks...")
        } else if viewModel.tasks.isEmpty {
          emptyStateView
        } else {
          let readyNowGroups = viewModel.getReadyNowGroups()
          let doLaterGroups = viewModel.getDoLaterGroups()
          let finishedGroups = viewModel.getFinishedGroups()

          if !readyNowGroups.isEmpty {
            tasksSection(
              title: "Ready Now",
              groups: readyNowGroups,
              color: DSTheme.Colors.warning
            )
          }

          if !doLaterGroups.isEmpty {
            laterTasksSection(groups: doLaterGroups)
          }

          if !finishedGroups.isEmpty {
            tasksSection(
              title: "Completed",
              groups: finishedGroups,
              color: DSTheme.Colors.success
            )
          }
        }

        if let errorMessage = viewModel.errorMessage {
          Text(errorMessage)
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.error)
            .padding()
        }
      }
      .dsScreenPadding()
    }
    .navigationTitle(viewModel.mealPlan.notes.isEmpty ? "Tasks" : "Tasks")
    .navigationBarTitleDisplayMode(.large)
    .onAppear {
      eventReporterService.reporter.track(
        event: "task_list_viewed",
        properties: ["meal_plan_id": viewModel.mealPlan.id])
    }
    .task {
      await viewModel.loadTasks()
    }
  }

  private var isPast: Bool {
    MealPlanningHomeHelpers.isMealPlanInPast(viewModel.mealPlan)
  }

  private var headerSection: some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
      Text(MealPlanningHomeHelpers.formatMealPlanTimeRange(viewModel.mealPlan))
        .font(DSTheme.Typography.body)
        .foregroundColor(DSTheme.Colors.textSecondary)
        .strikethrough(isPast)

      Text("\(viewModel.tasks.count) task\(viewModel.tasks.count == 1 ? "" : "s")")
        .font(DSTheme.Typography.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
        .strikethrough(isPast)
    }
  }

  private var emptyStateView: some View {
    DSEmptyState(
      icon: "checklist",
      title: "No tasks",
      message: "Tasks will appear here once the meal plan is finalized."
    )
  }

  private func tasksSection(
    title: String,
    groups: [TaskGroup],
    color: Color
  ) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      Text(title)
        .font(DSTheme.Typography.label)
        .foregroundColor(color)
        .strikethrough(isPast)

      ForEach(groups, id: \.parent.id) { group in
        TaskGroupRow(
          group: group,
          viewModel: viewModel,
          loadedRecipes: viewModel.loadedRecipes,
          loadedPrepTasks: viewModel.loadedPrepTasks
        )
        .padding(.bottom, DSTheme.Spacing.sm)
      }
    }
  }

  private func laterTasksSection(groups: [TaskGroup]) -> some View {
    let buckets = bucketByStartTime(groups)

    return VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      Text("Later")
        .font(DSTheme.Typography.label)
        .foregroundColor(.secondary)
        .strikethrough(isPast)

      ForEach(buckets, id: \.label) { bucket in
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          Text(bucket.label)
            .font(DSTheme.Typography.caption)
            .fontWeight(.semibold)
            .foregroundColor(.orange)
            .strikethrough(isPast)

          ForEach(bucket.groups, id: \.parent.id) { group in
            TaskGroupRow(
              group: group,
              viewModel: viewModel,
              loadedRecipes: viewModel.loadedRecipes,
              loadedPrepTasks: viewModel.loadedPrepTasks
            )
            .padding(.bottom, DSTheme.Spacing.sm)
          }
        }
      }
    }
  }

  private struct TimeBucket {
    let label: String
    let groups: [TaskGroup]
  }

  private func bucketByStartTime(_ groups: [TaskGroup]) -> [TimeBucket] {
    var bucketMap: [(key: String, groups: [TaskGroup])] = []
    let now = Date()

    for group in groups {
      let label: String
      if let startDate = viewModel.canStartAt(task: group.parent) {
        label = formatCountdownLabel(from: now, to: startDate)
      } else {
        label = "Later"
      }

      if let index = bucketMap.firstIndex(where: { $0.key == label }) {
        bucketMap[index].groups.append(group)
      } else {
        bucketMap.append((key: label, groups: [group]))
      }
    }

    return bucketMap.map { TimeBucket(label: $0.key, groups: $0.groups) }
  }

  private func formatCountdownLabel(from now: Date, to date: Date) -> String {
    let seconds = Int(date.timeIntervalSince(now))
    if seconds <= 0 { return "Now" }

    let hours = seconds / 3600
    let days = hours / 24

    if hours < 1 {
      let minutes = seconds / 60
      return "In \(minutes) min"
    } else if hours < 24 {
      return "In \(hours) hr"
    } else if days == 1 {
      return "In 1 day"
    } else {
      return "In \(days) days"
    }
  }
}

// MARK: - Task Group Row

struct TaskGroupRow: View {
  let group: TaskGroup
  let viewModel: TaskListViewModel
  let loadedRecipes: [String: Mealplanning_Recipe]
  let loadedPrepTasks: [String: Mealplanning_RecipePrepTask]

  var hasSubtasks: Bool {
    !group.subtasks.isEmpty
  }

  var isExpanded: Bool {
    viewModel.isExpanded(taskID: group.parent.id)
  }

  var body: some View {
    VStack(alignment: .leading, spacing: 0) {
      // Parent task
      TaskRow(
        task: group.parent,
        viewModel: viewModel,
        loadedRecipes: loadedRecipes,
        loadedPrepTasks: loadedPrepTasks,
        isParent: hasSubtasks,
        isExpanded: isExpanded,
        onToggleExpand: {
          viewModel.toggleExpanded(taskID: group.parent.id)
        }
      )

      // Steps preview - "What's involved" when expanded (read-only, tappable to open recipe)
      if hasSubtasks && isExpanded {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
          Text("What's involved")
            .font(DSTheme.Typography.labelSmall)
            .foregroundColor(DSTheme.Colors.textSecondary)
            .padding(.bottom, DSTheme.Spacing.xs)

          ForEach(group.subtasks) { subtask in
            StepPreviewRow(
              subtask: subtask,
              parentTask: group.parent,
              viewModel: viewModel
            )
          }
        }
        .padding(DSTheme.Spacing.md)
        .background(DSTheme.Colors.cardBackgroundElevated)
        .cornerRadius(DSTheme.Radius.sm)
        .padding(.leading, DSTheme.Spacing.lg)  // Align with task content
        .padding(.top, DSTheme.Spacing.sm)
      }
    }
  }
}

// MARK: - Step Preview Row

/// Read-only step display—tappable to open recipe. No individual checkboxes; the whole task is the unit.
struct StepPreviewRow: View {
  let subtask: SubtaskDisplayInfo
  let parentTask: Mealplanning_MealPlanTask
  let viewModel: TaskListViewModel

  private var recipeScaleFromMealPlan: Float? {
    guard let recipeID = subtask.recipeID else { return nil }
    return parentTask.recipeScaleForMealPlan(recipeID: recipeID)
  }

  private var label: String {
    subtask.ingredientNames.isEmpty
      ? subtask.description
      : "\(subtask.description) — \(subtask.ingredientNames.joined(separator: ", "))"
  }

  var body: some View {
    Group {
      if let recipeID = subtask.recipeID, let recipeStepID = subtask.recipeStepID {
        NavigationLink {
          PerformRecipeView(
            recipeID: recipeID,
            highlightedStepIDs: Set([recipeStepID]),
            initialScale: recipeScaleFromMealPlan
          )
        } label: {
          stepContent
        }
        .buttonStyle(.plain)
      } else {
        stepContent
      }
    }
  }

  private var stepContent: some View {
    HStack(alignment: .center, spacing: DSTheme.Spacing.sm) {
      Image(systemName: "circle.fill")
        .font(.system(size: 4))
        .foregroundColor(DSTheme.Colors.textTertiary)

      Text(label)
        .font(DSTheme.Typography.bodySmall)
        .foregroundColor(DSTheme.Colors.textPrimary)

      Spacer(minLength: 0)

      if subtask.recipeID != nil && subtask.recipeStepID != nil {
        Image(systemName: "arrow.up.right")
          .font(.caption2)
          .foregroundColor(DSTheme.Colors.textTertiary)
      }
    }
    .padding(.vertical, DSTheme.Spacing.xs)
    .padding(.horizontal, DSTheme.Spacing.sm)
    .contentShape(Rectangle())
  }
}

// MARK: - Task Row

struct TaskRow: View {
  @Environment(EventReporterService.self) private var eventReporterService
  let task: Mealplanning_MealPlanTask
  let viewModel: TaskListViewModel
  let loadedRecipes: [String: Mealplanning_Recipe]
  let loadedPrepTasks: [String: Mealplanning_RecipePrepTask]
  let isParent: Bool
  let isExpanded: Bool
  let onToggleExpand: (() -> Void)?

  // Extract recipe ID and step IDs from prep task
  private var recipeID: String? {
    guard task.hasRecipePrepTask else { return nil }
    let prepTask = task.recipePrepTask
    var recipeID = prepTask.belongsToRecipe

    if recipeID.isEmpty && !prepTask.id.isEmpty {
      if let loadedPrepTask = loadedPrepTasks[prepTask.id] {
        recipeID = loadedPrepTask.belongsToRecipe
      }
    }

    if recipeID.isEmpty && task.hasMealPlanOption {
      recipeID = findRecipeIDFromMealOption(task: task, prepTask: prepTask)
    }

    return recipeID.isEmpty ? nil : recipeID
  }

  private var highlightedStepIDs: Set<String>? {
    guard task.hasRecipePrepTask else { return nil }
    let prepTask = task.recipePrepTask
    let stepIDs = prepTask.taskSteps
      .filter { !$0.belongsToRecipeStep.isEmpty }
      .map { $0.belongsToRecipeStep }
    return stepIDs.isEmpty ? nil : Set(stepIDs)
  }

  // Get prep task context information
  private var recipeScaleFromMealPlan: Float? {
    guard let recipeID = recipeID else { return nil }
    return task.recipeScaleForMealPlan(recipeID: recipeID)
  }

  private var prepTaskContext: PerformRecipeView.PrepTaskContext? {
    guard task.hasRecipePrepTask else { return nil }
    let prepTask = task.recipePrepTask

    // Get event information from meal plan option
    var eventName: String?
    var eventTime: Date?

    if task.hasMealPlanOption {
      let eventID = task.mealPlanOption.belongsToMealPlanEvent
      if !eventID.isEmpty,
        let event = viewModel.mealPlan.displayEvents.first(where: { $0.id == eventID })
      {
        eventName = MealPlanningUtils.formatMealName(event.mealName)
        eventTime = HomeViewModel.timestampToDate(event.startsAt)
      }
    }

    // Get recipe name
    var recipeName: String?
    if let recipeID = recipeID, let recipe = loadedRecipes[recipeID] {
      recipeName = recipe.name
    }

    // Get prep task name
    let prepTaskName = prepTask.name.isEmpty ? nil : prepTask.name

    return PerformRecipeView.PrepTaskContext(
      prepTaskName: prepTaskName,
      recipeName: recipeName,
      eventName: eventName,
      eventTime: eventTime
    )
  }

  var body: some View {
    let hasNavigation = recipeID != nil && highlightedStepIDs != nil
    let isCompleted = task.status == .finished
    let context = prepTaskContext

    VStack(alignment: .leading, spacing: 0) {
      HStack(alignment: .top, spacing: DSTheme.Spacing.md) {
        // Checkbox - single action for the whole task (parent or standalone)
        Button {
          Task {
            await viewModel.toggleTaskStatus(task)
          }
        } label: {
          Image(systemName: task.status == .finished ? "checkmark.circle.fill" : "circle")
            .font(.title2)
            .foregroundColor(
              task.status == .finished ? DSTheme.Colors.success : DSTheme.Colors.info
            )
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .disabled(viewModel.isUpdating)
        .frame(minWidth: 44, minHeight: 24)

        // Task content + meta - tappable to navigate (when has recipe)
        Group {
          if hasNavigation, let recipeID = recipeID, let highlightedStepIDs = highlightedStepIDs {
            NavigationLink {
              PerformRecipeView(
                recipeID: recipeID,
                highlightedStepIDs: highlightedStepIDs,
                prepTaskContext: context,
                initialScale: recipeScaleFromMealPlan
              )
            } label: {
              VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
                taskDescriptionContent(isCompleted: isCompleted)
                if task.status != .finished {
                  taskMetaRow
                }
              }
              .frame(maxWidth: .infinity, alignment: .leading)
              .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .simultaneousGesture(
              TapGesture().onEnded {
                eventReporterService.reporter.track(
                  event: "task_recipe_tapped",
                  properties: ["recipe_id": recipeID, "meal_plan_id": viewModel.mealPlan.id])
              })
          } else {
            VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
              taskDescriptionContent(isCompleted: isCompleted)
              if task.status != .finished {
                taskMetaRow
              }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
          }
        }

        // Disclosure indicator for parent tasks (only this excludes navigation)
        if isParent {
          Button {
            onToggleExpand?()
          } label: {
            Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
              .font(.subheadline.weight(.semibold))
              .foregroundColor(DSTheme.Colors.textSecondary)
              .frame(width: 24)
          }
          .buttonStyle(.plain)
        }
      }
    }
    .padding(DSTheme.Spacing.lg)
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.md)
    .opacity(task.status == .finished ? 0.7 : 1.0)
  }

  @ViewBuilder
  private var taskMetaRow: some View {
    let hasStorage =
      task.hasRecipePrepTask
      && (!task.recipePrepTask.explicitStorageInstructions.isEmpty
        || !task.recipePrepTask.storageType.isEmpty
        || task.recipePrepTask.hasMinStorageTemperatureInCelsius
        || task.recipePrepTask.hasMaxStorageTemperatureInCelsius)
    let hasCountdown = eventStartTime != nil && task.status != .finished
    let (startDate, endDate) = taskTimeRange

    if hasStorage || hasCountdown {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
        if hasCountdown, let end = endDate {
          if let start = startDate {
            TaskTimeRangeView(startDate: start, endDate: end)
          } else {
            TaskCountdownTimer(dueDate: end)
          }
        }
        if hasStorage {
          storagePill(prepTask: task.recipePrepTask)
        }
      }
    }
  }

  /// Returns (startDate, endDate) when we have a time buffer. Otherwise (nil, eventTime).
  private var taskTimeRange: (Date?, Date?) {
    guard let eventTime = eventStartTime else { return (nil, nil) }
    guard task.hasRecipePrepTask,
      task.recipePrepTask.minTimeBufferBeforeRecipeInSeconds > 0
    else {
      return (nil, eventTime)
    }
    let startTime = eventTime.addingTimeInterval(
      -Double(task.recipePrepTask.minTimeBufferBeforeRecipeInSeconds))
    return (startTime, eventTime)
  }

  @ViewBuilder
  private func storagePill(prepTask: Mealplanning_RecipePrepTask) -> some View {
    let hasExplicit = !prepTask.explicitStorageInstructions.isEmpty
    let hasType = !prepTask.storageType.isEmpty
    let temp = formatStorageTemperature(
      min: prepTask.hasMinStorageTemperatureInCelsius
        ? prepTask.minStorageTemperatureInCelsius : nil,
      max: prepTask.hasMaxStorageTemperatureInCelsius
        ? prepTask.maxStorageTemperatureInCelsius : nil,
      unit: viewModel.temperatureUnit)

    if hasExplicit || hasType || !temp.isEmpty {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
        if hasExplicit {
          Label {
            Text(prepTask.explicitStorageInstructions)
              .lineLimit(2)
              .truncationMode(.tail)
          } icon: {
            Image(systemName: "info.circle")
          }
        }
        if hasType || !temp.isEmpty {
          HStack(spacing: DSTheme.Spacing.xs) {
            if hasType {
              Label(prepTask.storageType, systemImage: "archivebox")
            }
            if !temp.isEmpty {
              Label(temp, systemImage: "thermometer.medium")
            }
          }
        }
      }
      .font(DSTheme.Typography.caption)
      .foregroundColor(DSTheme.Colors.textSecondary)
    }
  }

  // Get event start time for countdown
  private var eventStartTime: Date? {
    guard task.hasMealPlanOption else { return nil }
    let eventID = task.mealPlanOption.belongsToMealPlanEvent
    guard !eventID.isEmpty,
      let event = viewModel.mealPlan.displayEvents.first(where: { $0.id == eventID })
    else {
      return nil
    }
    return HomeViewModel.timestampToDate(event.startsAt)
  }

  private func taskDescriptionContent(isCompleted: Bool) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
      // Task description - show prep task name or creation explanation
      if task.hasRecipePrepTask && !task.recipePrepTask.name.isEmpty {
        Text(task.recipePrepTask.name)
          .font(DSTheme.Typography.body)
          .fontWeight(.semibold)
          .strikethrough(isCompleted)
          .foregroundColor(
            isCompleted ? DSTheme.Colors.textSecondary : DSTheme.Colors.textPrimary
          )
      } else {
        Text(task.creationExplanation)
          .font(DSTheme.Typography.body)
          .fontWeight(.semibold)
          .strikethrough(isCompleted)
          .foregroundColor(
            isCompleted ? DSTheme.Colors.textSecondary : DSTheme.Colors.textPrimary
          )
      }

      if !task.statusExplanation.isEmpty {
        Text(task.statusExplanation)
          .font(DSTheme.Typography.caption)
          .foregroundColor(DSTheme.Colors.textSecondary)
      }
    }
  }

  private func formatStorageTemperature(min minVal: Float?, max maxVal: Float?, unit: String)
    -> String
  {
    let useFahrenheit = unit == "fahrenheit"

    func formatCelsius(_ c: Float) -> String {
      if useFahrenheit {
        let f = (c * 9 / 5) + 32
        return "\(Int(f.rounded()))°F"
      }
      return "\(Int(c.rounded()))°C"
    }

    if let minVal = minVal, let maxVal = maxVal {
      if minVal == maxVal {
        return formatCelsius(minVal)
      }
      if useFahrenheit {
        let minF = (minVal * 9 / 5) + 32
        let maxF = (maxVal * 9 / 5) + 32
        return "\(Int(minF.rounded()))–\(Int(maxF.rounded()))°F"
      }
      return "\(Int(minVal.rounded()))–\(Int(maxVal.rounded()))°C"
    }
    if let maxVal = maxVal {
      return "below \(formatCelsius(maxVal))"
    }
    if let minVal = minVal {
      return "above \(formatCelsius(minVal))"
    }
    return ""
  }

  // Helper struct for step data
  struct StepInfo {
    let stepID: String
    let description: String
  }

  private func getStepData(task: Mealplanning_MealPlanTask) -> [StepInfo] {
    guard task.hasRecipePrepTask else {
      print("⚠️ Task \(task.id) has no recipePrepTask")
      return []
    }

    let prepTask = task.recipePrepTask
    let recipeID = findRecipeID(for: task, prepTask: prepTask)

    if recipeID.isEmpty {
      print("⚠️ Task \(task.id) prep task has no recipe ID and couldn't find it from meal option")
      return []
    }

    guard let recipe = loadedRecipes[recipeID] else {
      print(
        "⚠️ Recipe \(recipeID) not loaded yet for task \(task.id). Loaded recipes: \(loadedRecipes.keys.joined(separator: ", "))"
      )
      return []
    }

    return buildStepData(prepTask: prepTask, recipe: recipe, recipeID: recipeID)
  }

  private func findRecipeID(
    for task: Mealplanning_MealPlanTask,
    prepTask: Mealplanning_RecipePrepTask
  ) -> String {
    var recipeID = prepTask.belongsToRecipe

    print("🔍 Task \(task.id): prepTask.id = '\(prepTask.id)'")
    print(
      "🔍 Task \(task.id): prepTask.belongsToRecipe = '\(recipeID)' (isEmpty: \(recipeID.isEmpty))")
    print("🔍 Task \(task.id): prepTask.name = '\(prepTask.name)'")
    print("🔍 Task \(task.id): prepTask.taskSteps.count = \(prepTask.taskSteps.count)")

    // If belongsToRecipe is empty, check loaded prep task
    if recipeID.isEmpty && !prepTask.id.isEmpty {
      if let loadedPrepTask = loadedPrepTasks[prepTask.id] {
        recipeID = loadedPrepTask.belongsToRecipe
        print(
          "✅ Found recipe ID \(recipeID) from loaded prep task \(prepTask.id) for task \(task.id)")
      } else {
        print("⚠️ Prep task \(prepTask.id) not found in loadedPrepTasks")
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
          print("✅ Found recipe ID \(candidateRecipeID) from meal option for task \(task.id)")
          return candidateRecipeID
        }
      }
    }
    return ""
  }

  private func buildStepData(
    prepTask: Mealplanning_RecipePrepTask,
    recipe: Mealplanning_Recipe,
    recipeID: String
  ) -> [StepInfo] {
    var stepData: [StepInfo] = []

    for taskStep in prepTask.taskSteps where !taskStep.belongsToRecipeStep.isEmpty {
      // Try to find step by ID in the loaded recipe
      if let stepIndex = recipe.steps.firstIndex(where: { $0.id == taskStep.belongsToRecipeStep }) {
        let step = recipe.steps[stepIndex]
        let formatted = formatStepTitle(step: step)
        if !formatted.isEmpty {
          stepData.append(StepInfo(stepID: taskStep.id, description: formatted))
        } else {
          print("⚠️ Step \(taskStep.belongsToRecipeStep) formatted to empty string")
        }
      } else {
        print("⚠️ Step \(taskStep.belongsToRecipeStep) not found in recipe \(recipeID)")
      }
    }

    if stepData.isEmpty {
      print(
        "⚠️ No step descriptions found, prep task has \(prepTask.taskSteps.count) task steps"
      )
    }

    return stepData
  }

  private func formatStepTitle(step: Mealplanning_RecipeStep) -> String {
    if step.hasPreparation && !step.preparation.name.isEmpty {
      return step.preparation.name
    }
    return "Step \(Int(step.index) + 1)"
  }
}

// Make MealPlanTask Identifiable
extension Mealplanning_MealPlanTask: Identifiable {
  // Already has id property, so this extension just makes it conform to Identifiable
}

// MARK: - Recipe Scale from Meal Plan

extension Mealplanning_MealPlanTask {
  /// Returns the recipe scale from the meal plan option for the given recipe ID.
  /// Scale = component.recipeScale * option.mealScale. Treats 0 as 1.0.
  /// Returns nil if task has no meal plan option or recipe not found in meal.
  func recipeScaleForMealPlan(recipeID: String) -> Float? {
    guard hasMealPlanOption else { return nil }
    let meal = mealPlanOption.meal
    let optionMealScale = mealPlanOption.mealScale
    let mealScale = (optionMealScale == 0 ? 1.0 : optionMealScale)
    for component in meal.components where component.recipe.id == recipeID {
      let compScale = component.recipeScale
      let baseScale = (compScale == 0 ? 1.0 : compScale)
      return baseScale * mealScale
    }
    return nil
  }
}

// MARK: - Task Time Range View

struct TaskTimeRangeView: View {
  let startDate: Date
  let endDate: Date
  @State private var timeRemaining: TimeInterval = 0
  @State private var timer: Timer?

  var body: some View {
    HStack(spacing: 6) {
      Image(systemName: "clock.fill")
        .font(.caption2)
      Text(formattedDisplayText)
        .font(.caption)
        .fontWeight(.semibold)
        .monospacedDigit()
    }
    .foregroundColor(timeRemainingColor)
    .onAppear {
      updateTimeRemaining()
      startTimer()
    }
    .onDisappear {
      stopTimer()
    }
  }

  private var formattedDisplayText: String {
    let startStr = formatShortDate(startDate)
    let endStr = formatShortDate(endDate)
    let countdownStr = formattedTimeRemaining
    return "Between \(startStr) and \(endStr) (\(countdownStr))"
  }

  private var formattedTimeRemaining: String {
    if timeRemaining <= 0 {
      return "Overdue"
    }
    let days = Int(timeRemaining) / 86400
    let hours = (Int(timeRemaining) % 86400) / 3600
    let minutes = (Int(timeRemaining) % 3600) / 60
    let seconds = Int(timeRemaining) % 60
    if days > 0 {
      return String(format: "%dd %02d:%02d:%02d", days, hours, minutes, seconds)
    } else if hours > 0 {
      return String(format: "%02d:%02d:%02d", hours, minutes, seconds)
    } else {
      return String(format: "%02d:%02d", minutes, seconds)
    }
  }

  private var timeRemainingColor: Color {
    if timeRemaining <= 0 { return .red }
    if timeRemaining < 3600 { return .red }
    if timeRemaining < 86400 { return .orange }
    return .secondary
  }

  private func formatShortDate(_ date: Date) -> String {
    let formatter = DateFormatter()
    let calendar = Calendar.current
    if calendar.isDateInToday(date) {
      formatter.dateFormat = "'today at' h:mm a"
    } else if calendar.isDateInTomorrow(date) {
      formatter.dateFormat = "'tomorrow at' h:mm a"
    } else {
      formatter.dateFormat = "EEE h:mm a"
    }
    return formatter.string(from: date)
  }

  private func updateTimeRemaining() {
    timeRemaining = max(0, endDate.timeIntervalSince(Date()))
  }

  private func startTimer() {
    timer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { _ in
      updateTimeRemaining()
    }
    if let timer = timer {
      RunLoop.main.add(timer, forMode: .common)
    }
  }

  private func stopTimer() {
    timer?.invalidate()
    timer = nil
  }
}

// MARK: - Task Countdown Timer

struct TaskCountdownTimer: View {
  let dueDate: Date
  @State private var timeRemaining: TimeInterval = 0
  @State private var timer: Timer?

  var body: some View {
    HStack(spacing: 6) {
      Image(systemName: "clock.fill")
        .font(.caption2)
      Text(formattedDisplayText)
        .font(.caption)
        .fontWeight(.semibold)
        .monospacedDigit()
    }
    .foregroundColor(timeRemainingColor)
    .onAppear {
      updateTimeRemaining()
      startTimer()
    }
    .onDisappear {
      stopTimer()
    }
  }

  private var formattedDate: String {
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short
    return formatter.string(from: dueDate)
  }

  private var formattedTimeRemaining: String {
    if timeRemaining <= 0 {
      return "Overdue"
    }

    let days = Int(timeRemaining) / 86400
    let hours = (Int(timeRemaining) % 86400) / 3600
    let minutes = (Int(timeRemaining) % 3600) / 60
    let seconds = Int(timeRemaining) % 60

    if days > 0 {
      return String(format: "%dd %02d:%02d:%02d", days, hours, minutes, seconds)
    } else if hours > 0 {
      return String(format: "%02d:%02d:%02d", hours, minutes, seconds)
    } else {
      return String(format: "%02d:%02d", minutes, seconds)
    }
  }

  private var formattedDisplayText: String {
    let dateString = formattedDate
    let countdownString = formattedTimeRemaining
    return "by \(dateString) (\(countdownString))"
  }

  private var timeRemainingColor: Color {
    if timeRemaining <= 0 {
      return .red
    } else if timeRemaining < 3600 {  // Less than 1 hour
      return .red
    } else if timeRemaining < 86400 {  // Less than 1 day
      return .orange
    } else {
      return .secondary
    }
  }

  private func updateTimeRemaining() {
    let now = Date()
    timeRemaining = max(0, dueDate.timeIntervalSince(now))
  }

  private func startTimer() {
    timer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { _ in
      updateTimeRemaining()
    }
    if let timer = timer {
      RunLoop.main.add(timer, forMode: .common)
    }
  }

  private func stopTimer() {
    timer?.invalidate()
    timer = nil
  }
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "John Doe"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  // Create a sample meal plan and tasks
  var mealPlan = Mealplanning_MealPlan()
  mealPlan.id = "mealplan123"
  mealPlan.notes = "Sample Meal Plan"

  var task1 = Mealplanning_MealPlanTask()
  task1.id = "task1"
  task1.creationExplanation = "Preheat oven to 350°F"
  task1.status = .unfinished

  var task2 = Mealplanning_MealPlanTask()
  task2.id = "task2"
  task2.creationExplanation = "Chop vegetables"
  task2.status = .finished

  let userSettingsService = UserSettingsService()
  userSettingsService.configure(authManager: authManager)

  return NavigationView {
    TaskListView(
      mealPlan: mealPlan,
      tasks: [task1, task2],
      authManager: authManager,
      userSettingsService: userSettingsService
    )
  }
  .environment(authManager)
  .environment(userSettingsService)
}
