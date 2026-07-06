//
//  PerformRecipeViewModel.swift
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
class PerformRecipeViewModel {
  var recipe: Mealplanning_Recipe?
  var prepTasks: [Mealplanning_RecipePrepTask] = []
  var isLoading = false
  var isLoadingPrepTasks = false
  var errorMessage: String?

  var mermaidDiagram: String?
  var isLoadingMermaid = false
  var mermaidError: String?

  // Track which steps are completed (by step ID: "recipeID:stepID")
  var completedSteps: Set<String> = []

  // Track when user checked steps with timers (stepKey -> start Date)
  // Step stays in Up Next until elapsed >= estimatedTimeInSeconds.min
  var stepTimerStartTimes: [String: Date] = [:]

  // Track which prep tasks are completed (by prep task ID)
  var completedPrepTasks: Set<String> = []

  // Track which completion conditions are checked (by key: "recipeID:stepID:conditionID")
  var completedStepCompletionConditions: Set<String> = []

  // Track which step timers are completed (by key: "recipeID:stepID")
  // Timer is a synthetic completion condition; when done, user can mark step complete
  var completedStepTimers: Set<String> = []

  // Special step: wash hands (index -1)
  var washHandsCompleted: Bool = false

  // Map from product ID to the step that produces it (recipeID:stepID)
  var productIDToStepKey: [String: String] = [:]

  // Special index for wash hands step
  static let washHandsStepIndex = -1

  private let recipeID: String
  private let authManager: AuthenticationManager

  init(recipeID: String, authManager: AuthenticationManager) {
    self.recipeID = recipeID
    self.authManager = authManager
  }

  func loadRecipe() async {
    isLoading = true
    errorMessage = nil

    do {
      var request = Mealplanning_GetRecipeRequest()
      request.recipeID = recipeID

      let response = try await authManager.authenticatedCall("getRecipe", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getRecipe(request, metadata: metadata, options: options)
      }

      self.recipe = response.result
      buildProductIDToStepIndexMapping()

      // Load prep tasks after recipe is loaded
      await loadPrepTasks()
    } catch {
      errorMessage = "Failed to load recipe: \(error.localizedDescription)"
    }

    isLoading = false
  }

  func loadPrepTasks() async {
    isLoadingPrepTasks = true

    do {
      var request = Mealplanning_GetRecipePrepTasksRequest()
      request.recipeID = recipeID

      let response = try await authManager.authenticatedCall("getRecipePrepTasks", idempotent: true)
      {
        client, metadata, options in
        try await client.mealPlanning.getRecipePrepTasks(
          request, metadata: metadata, options: options)
      }

      self.prepTasks = response.results
    } catch {
      // Non-critical: don't surface a prep-tasks error to the user
    }

    isLoadingPrepTasks = false
  }

  func loadMermaidDiagram() async {
    guard mermaidDiagram == nil else { return }
    isLoadingMermaid = true
    mermaidError = nil

    do {
      var request = Mealplanning_GetMermaidDiagramForRecipeRequest()
      request.recipeID = recipeID

      let response = try await authManager.authenticatedCall(
        "getMermaidDiagramForRecipe", idempotent: true
      ) { client, metadata, options in
        try await client.mealPlanning.getMermaidDiagramForRecipe(
          request, metadata: metadata, options: options)
      }

      self.mermaidDiagram = response.response
    } catch {
      mermaidError = "Failed to load diagram: \(error.localizedDescription)"
    }

    isLoadingMermaid = false
  }

  // Build a mapping from recipe step product IDs to the step that produces them
  private func buildProductIDToStepIndexMapping() {
    guard let recipe = recipe else { return }

    // Map products from main recipe
    for step in recipe.steps {
      let stepKey = "\(recipe.id):\(step.id)"
      for product in step.products {
        productIDToStepKey[product.id] = stepKey
      }
    }

    // Map products from associated recipes
    for associatedRecipe in recipe.associatedRecipes {
      for step in associatedRecipe.steps {
        let stepKey = "\(associatedRecipe.id):\(step.id)"
        for product in step.products {
          productIDToStepKey[product.id] = stepKey
        }
      }
    }
  }

  // Helper to create a step key from recipeID and stepID
  private func stepKey(recipeID: String, stepID: String) -> String {
    return "\(recipeID):\(stepID)"
  }

  private func completionConditionKey(recipeID: String, stepID: String, conditionIdentifier: String)
    -> String
  {
    return "\(recipeID):\(stepID):\(conditionIdentifier)"
  }

  private func stepFor(recipeID: String, stepID: String) -> Mealplanning_RecipeStep? {
    guard let recipe = recipe else {
      return nil
    }

    if recipeID == recipe.id {
      return recipe.steps.first { $0.id == stepID }
    }

    return recipe.associatedRecipes.first { $0.id == recipeID }?.steps.first { $0.id == stepID }
  }

  private func stepHasTimer(_ step: Mealplanning_RecipeStep) -> Bool {
    step.hasMinEstimatedTimeInSeconds && step.minEstimatedTimeInSeconds > 0
  }

  /// Display total (max when both min/max exist, else min) - for "elapsed / total" display
  private func stepTimerDurationSeconds(_ step: Mealplanning_RecipeStep) -> UInt32? {
    guard step.hasMinEstimatedTimeInSeconds, step.minEstimatedTimeInSeconds > 0 else {
      return nil
    }
    if step.hasMaxEstimatedTimeInSeconds,
      step.maxEstimatedTimeInSeconds >= step.minEstimatedTimeInSeconds
    {
      return step.maxEstimatedTimeInSeconds
    }
    return step.minEstimatedTimeInSeconds
  }

  /// Done threshold - step auto-completes when elapsed >= this (max when both exist, else min)
  private func stepTimerDoneThresholdSeconds(_ step: Mealplanning_RecipeStep) -> UInt32? {
    guard step.hasMinEstimatedTimeInSeconds, step.minEstimatedTimeInSeconds > 0 else {
      return nil
    }
    if step.hasMaxEstimatedTimeInSeconds,
      step.maxEstimatedTimeInSeconds >= step.minEstimatedTimeInSeconds
    {
      return step.maxEstimatedTimeInSeconds
    }
    return step.minEstimatedTimeInSeconds
  }

  /// Min threshold - user can skip only when elapsed >= this
  private func stepTimerMinSeconds(_ step: Mealplanning_RecipeStep) -> UInt32? {
    guard step.hasMinEstimatedTimeInSeconds, step.minEstimatedTimeInSeconds > 0 else {
      return nil
    }
    return step.minEstimatedTimeInSeconds
  }

  func stepCompletionConditionIdentifier(
    condition: Mealplanning_RecipeStepCompletionCondition,
    index: Int
  ) -> String {
    if !condition.id.isEmpty {
      return condition.id
    }

    return "condition-\(index)"
  }

  func isStepCompletionConditionCompleted(
    recipeID: String,
    stepID: String,
    conditionIdentifier: String
  ) -> Bool {
    let key = completionConditionKey(
      recipeID: recipeID,
      stepID: stepID,
      conditionIdentifier: conditionIdentifier
    )
    return completedStepCompletionConditions.contains(key)
  }

  func toggleStepCompletionCondition(
    recipeID: String,
    stepID: String,
    conditionIdentifier: String
  ) {
    let key = completionConditionKey(
      recipeID: recipeID,
      stepID: stepID,
      conditionIdentifier: conditionIdentifier
    )

    if completedStepCompletionConditions.contains(key) {
      completedStepCompletionConditions.remove(key)

      let stepKey = stepKey(recipeID: recipeID, stepID: stepID)
      if completedSteps.contains(stepKey) {
        uncheckStepAndDependents(stepKey: stepKey)
      }
      return
    }

    completedStepCompletionConditions.insert(key)
  }

  // MARK: - Timer as completion condition

  func isTimerConditionCompleted(recipeID: String, stepID: String) -> Bool {
    completedStepTimers.contains(stepKey(recipeID: recipeID, stepID: stepID))
  }

  func completeTimerCondition(recipeID: String, stepID: String) {
    let key = stepKey(recipeID: recipeID, stepID: stepID)
    stepTimerStartTimes.removeValue(forKey: key)
    StepTimerNotificationService.cancelNotification(stepKey: key)
    completedStepTimers.insert(key)
  }

  func toggleTimerCondition(recipeID: String, stepID: String) {
    let key = stepKey(recipeID: recipeID, stepID: stepID)
    guard completedStepTimers.contains(key) else { return }
    completedStepTimers.remove(key)

    if completedSteps.contains(key) {
      uncheckStepAndDependents(stepKey: key)
    }
  }

  func startStepTimer(recipeID: String, stepID: String) {
    let key = stepKey(recipeID: recipeID, stepID: stepID)
    guard let step = stepFor(recipeID: recipeID, stepID: stepID), stepHasTimer(step) else { return }
    stepTimerStartTimes[key] = Date()
    let recipeName = recipe?.name ?? "???"
    let stepName = step.hasPreparation ? step.preparation.name : stepID
    if let minSeconds = stepTimerMinSeconds(step) {
      StepTimerNotificationService.scheduleNotification(
        stepKey: key,
        recipeName: recipeName,
        stepName: stepName,
        minSeconds: minSeconds
      )
    }
    print(
      "⏱ [\(recipeName)] startStepTimer '\(stepName)' → timer started"
    )
  }

  func areStepCompletionConditionsCompleted(recipeID: String, stepID: String) -> Bool {
    guard let step = stepFor(recipeID: recipeID, stepID: stepID) else {
      return false
    }

    let nonOptionalConditions = step.completionConditions.enumerated().filter {
      !$0.element.optional
    }

    if nonOptionalConditions.isEmpty {
      return true
    }

    for (index, condition) in step.completionConditions.enumerated() where !condition.optional {
      let conditionIdentifier = stepCompletionConditionIdentifier(
        condition: condition, index: index)
      let key = completionConditionKey(
        recipeID: recipeID, stepID: stepID, conditionIdentifier: conditionIdentifier)
      let isCompleted = completedStepCompletionConditions.contains(key)
      if !isCompleted {
        return false
      }
    }

    return true
  }

  func clearStepCompletionConditionProgress() {
    completedStepCompletionConditions.removeAll()
  }

  /// Marks all steps for an associated (prerequisite) recipe as complete.
  /// Use when the user has already made this recipe ahead of time (e.g., garlic confit on Sunday).
  func markAssociatedRecipeAsCompleted(associatedRecipe: Mealplanning_Recipe) {
    let recipeName = associatedRecipe.name.isEmpty ? "Unnamed" : associatedRecipe.name
    for step in associatedRecipe.steps {
      let currentStepKey = stepKey(recipeID: associatedRecipe.id, stepID: step.id)
      completedSteps.insert(currentStepKey)

      // Mark completion conditions so canCheckStep logic remains satisfied
      for (condIndex, condition) in step.completionConditions.enumerated() where !condition.optional
      {
        let conditionIdentifier = stepCompletionConditionIdentifier(
          condition: condition, index: condIndex)
        let key = completionConditionKey(
          recipeID: associatedRecipe.id,
          stepID: step.id,
          conditionIdentifier: conditionIdentifier
        )
        completedStepCompletionConditions.insert(key)
      }

      if stepHasTimer(step) {
        completedStepTimers.insert(currentStepKey)
      }
    }
    print(
      "✅ [\(recipe?.name ?? "???")] markAssociatedRecipeAsCompleted '\(recipeName)' → \(associatedRecipe.steps.count) steps marked done"
    )
  }

  // Check if a step can be checked off (all prerequisites are completed)
  // Supports both old API (stepIndex for main recipe) and new API (recipeID + stepID)
  func canCheckStep(_ stepIndex: Int) -> Bool {
    // Wash hands step can always be checked
    if stepIndex == Self.washHandsStepIndex {
      return true
    }

    guard let recipe = recipe, stepIndex >= 0, stepIndex < recipe.steps.count else {
      return false
    }
    let step = recipe.steps[stepIndex]
    return canCheckStep(recipeID: recipe.id, stepID: step.id)
  }

  /// Can the user press "Complete Step" to mark this step done?
  /// Requires: wash hands + prerequisites + all non-optional conditions + timer condition (if step has timer)
  func canCompleteStep(recipeID: String, stepID: String) -> Bool {
    if !washHandsCompleted {
      return false
    }

    guard let step = stepFor(recipeID: recipeID, stepID: stepID) else {
      return false
    }

    for ingredient in step.ingredients where ingredient.hasRecipeStepProductID {
      let productID = ingredient.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID] {
        if !completedSteps.contains(prerequisiteStepKey) {
          return false
        }
      }
    }

    for instrument in step.instruments where instrument.hasRecipeStepProductID {
      let productID = instrument.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID] {
        if !completedSteps.contains(prerequisiteStepKey) {
          return false
        }
      }
    }

    for vessel in step.vessels where vessel.hasRecipeStepProductID {
      let productID = vessel.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID] {
        if !completedSteps.contains(prerequisiteStepKey) {
          return false
        }
      }
    }

    let conditionsOk = areStepCompletionConditionsCompleted(recipeID: recipeID, stepID: stepID)
    if !conditionsOk {
      return false
    }

    if stepHasTimer(step) && !isTimerConditionCompleted(recipeID: recipeID, stepID: stepID) {
      return false
    }

    return true
  }

  func markStepComplete(recipeID: String, stepID: String) {
    let currentStepKey = stepKey(recipeID: recipeID, stepID: stepID)
    guard canCompleteStep(recipeID: recipeID, stepID: stepID) else { return }
    completedSteps.insert(currentStepKey)
  }

  func canCheckStep(recipeID: String, stepID: String) -> Bool {
    canCompleteStep(recipeID: recipeID, stepID: stepID)
  }

  /// Prerequisites only (wash hands + step dependencies). Does NOT require completion conditions.
  /// Use for starting/canceling timers. Use canCheckStep when marking a step complete without a timer.
  func canStartStep(recipeID: String, stepID: String) -> Bool {
    if !washHandsCompleted { return false }
    guard let step = stepFor(recipeID: recipeID, stepID: stepID) else { return false }

    for ingredient in step.ingredients where ingredient.hasRecipeStepProductID {
      let productID = ingredient.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID],
        !completedSteps.contains(prerequisiteStepKey)
      {
        return false
      }
    }
    for instrument in step.instruments where instrument.hasRecipeStepProductID {
      let productID = instrument.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID],
        !completedSteps.contains(prerequisiteStepKey)
      {
        return false
      }
    }
    for vessel in step.vessels where vessel.hasRecipeStepProductID {
      let productID = vessel.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID],
        !completedSteps.contains(prerequisiteStepKey)
      {
        return false
      }
    }
    return true
  }

  // Toggle step completion
  // Supports both old API (stepIndex for main recipe) and new API (recipeID + stepID)
  func toggleStep(_ stepIndex: Int) {
    // Handle wash hands step
    if stepIndex == Self.washHandsStepIndex {
      washHandsCompleted.toggle()
      // If unchecking wash hands, uncheck all other steps
      if !washHandsCompleted {
        completedSteps.removeAll()
        stepTimerStartTimes.removeAll()
        completedStepTimers.removeAll()
        clearStepCompletionConditionProgress()
      }
      return
    }

    guard let recipe = recipe, stepIndex >= 0, stepIndex < recipe.steps.count else {
      return
    }
    let step = recipe.steps[stepIndex]
    toggleStep(recipeID: recipe.id, stepID: step.id)
  }

  func toggleStep(recipeID: String, stepID: String) {
    let currentStepKey = stepKey(recipeID: recipeID, stepID: stepID)

    // Basic checks: wash hands + prerequisites (NOT completion conditions)
    guard canStartStep(recipeID: recipeID, stepID: stepID) else {
      return
    }

    if stepTimerStartTimes[currentStepKey] != nil {
      stepTimerStartTimes.removeValue(forKey: currentStepKey)
      StepTimerNotificationService.cancelNotification(stepKey: currentStepKey)
    } else if completedSteps.contains(currentStepKey) {
      uncheckStepAndDependents(stepKey: currentStepKey)
    } else {
      if let step = stepFor(recipeID: recipeID, stepID: stepID), stepHasTimer(step) {
        startStepTimer(recipeID: recipeID, stepID: stepID)
      } else {
        markStepComplete(recipeID: recipeID, stepID: stepID)
      }
    }
  }

  // Recursively uncheck a step and all steps that depend on it
  private func uncheckStepAndDependents(stepKey: String) {
    guard let recipe = recipe else { return }
    completedSteps.remove(stepKey)
    stepTimerStartTimes.removeValue(forKey: stepKey)
    completedStepTimers.remove(stepKey)
    StepTimerNotificationService.cancelNotification(stepKey: stepKey)

    // Parse recipeID and stepID from stepKey
    let components = stepKey.split(separator: ":", maxSplits: 1)
    guard components.count == 2,
      let stepID = String(components[1]) as String?
    else {
      return
    }
    let recipeID = String(components[0])

    // Find the step to get its products
    let step: Mealplanning_RecipeStep?
    if recipeID == recipe.id {
      step = recipe.steps.first { $0.id == stepID }
    } else {
      step = recipe.associatedRecipes.first { $0.id == recipeID }?.steps.first { $0.id == stepID }
    }

    guard let step = step else { return }

    // Find all steps (in main recipe and associated recipes) that depend on this step's products
    for product in step.products {
      // Check main recipe steps
      processDependentSteps(
        steps: recipe.steps,
        recipeID: recipe.id,
        productID: product.id,
        excludeStepKey: stepKey
      )

      // Check associated recipe steps
      for associatedRecipe in recipe.associatedRecipes {
        processDependentSteps(
          steps: associatedRecipe.steps,
          recipeID: associatedRecipe.id,
          productID: product.id,
          excludeStepKey: stepKey
        )
      }
    }
  }

  // Helper function to check if a step depends on a product
  private func stepDependsOnProduct(_ step: Mealplanning_RecipeStep, productID: String) -> Bool {
    // Check ingredients
    for ingredient in step.ingredients {
      if ingredient.hasRecipeStepProductID && ingredient.recipeStepProductID == productID {
        return true
      }
    }

    // Check instruments
    for instrument in step.instruments {
      if instrument.hasRecipeStepProductID && instrument.recipeStepProductID == productID {
        return true
      }
    }

    // Check vessels
    for vessel in step.vessels {
      if vessel.hasRecipeStepProductID && vessel.recipeStepProductID == productID {
        return true
      }
    }

    return false
  }

  // Helper function to process dependent steps for a given recipe
  private func processDependentSteps(
    steps: [Mealplanning_RecipeStep],
    recipeID: String,
    productID: String,
    excludeStepKey: String
  ) {
    for dependentStep in steps {
      let dependentStepKey = self.stepKey(recipeID: recipeID, stepID: dependentStep.id)
      if dependentStepKey == excludeStepKey {
        continue  // Skip self
      }

      if stepDependsOnProduct(dependentStep, productID: productID)
        && completedSteps.contains(dependentStepKey)
      {
        uncheckStepAndDependents(stepKey: dependentStepKey)
      }
    }
  }

  // Get the step at the given index
  func getStep(_ stepIndex: Int) -> Mealplanning_RecipeStep? {
    guard let recipe = recipe, stepIndex < recipe.steps.count else {
      return nil
    }
    return recipe.steps[stepIndex]
  }

  // Check if a step is completed
  // Supports both old API (stepIndex for main recipe) and new API (recipeID + stepID)
  func isStepCompleted(_ stepIndex: Int) -> Bool {
    if stepIndex == Self.washHandsStepIndex {
      return washHandsCompleted
    }
    guard let recipe = recipe, stepIndex >= 0, stepIndex < recipe.steps.count else {
      return false
    }
    let step = recipe.steps[stepIndex]
    return isStepCompleted(recipeID: recipe.id, stepID: step.id)
  }

  func isStepCompleted(recipeID: String, stepID: String) -> Bool {
    let currentStepKey = stepKey(recipeID: recipeID, stepID: stepID)
    return completedSteps.contains(currentStepKey)
  }

  // MARK: - Step timer state (for UI)

  func isStepTimerActive(recipeID: String, stepID: String) -> Bool {
    let key = stepKey(recipeID: recipeID, stepID: stepID)
    guard let step = stepFor(recipeID: recipeID, stepID: stepID),
      stepHasTimer(step),
      stepTimerStartTimes[key] != nil
    else {
      return false
    }
    // If step has max time, auto-expire the timer when max is reached
    if step.hasMaxEstimatedTimeInSeconds,
      let doneThreshold = stepTimerDoneThresholdSeconds(step),
      let startDate = stepTimerStartTimes[key]
    {
      return Date().timeIntervalSince(startDate) < Double(doneThreshold)
    }
    return true  // No max: timer active until explicitly skipped
  }

  func stepTimerElapsedSeconds(recipeID: String, stepID: String) -> TimeInterval? {
    guard isStepTimerActive(recipeID: recipeID, stepID: stepID),
      let startDate = stepTimerStartTimes[stepKey(recipeID: recipeID, stepID: stepID)]
    else {
      return nil
    }
    return Date().timeIntervalSince(startDate)
  }

  func stepTimerDurationSeconds(recipeID: String, stepID: String) -> UInt32? {
    guard let step = stepFor(recipeID: recipeID, stepID: stepID) else {
      return nil
    }
    return stepTimerDurationSeconds(step)
  }

  func stepTimerMinSeconds(recipeID: String, stepID: String) -> UInt32? {
    guard let step = stepFor(recipeID: recipeID, stepID: stepID) else {
      return nil
    }
    return stepTimerMinSeconds(step)
  }

  func stepTimerMaxSeconds(recipeID: String, stepID: String) -> UInt32? {
    guard let step = stepFor(recipeID: recipeID, stepID: stepID) else {
      return nil
    }
    guard step.hasMaxEstimatedTimeInSeconds,
      step.maxEstimatedTimeInSeconds >= step.minEstimatedTimeInSeconds
    else {
      return nil
    }
    return step.maxEstimatedTimeInSeconds
  }

  func canSkipStepTimer(recipeID: String, stepID: String) -> Bool {
    guard let step = stepFor(recipeID: recipeID, stepID: stepID),
      let minSeconds = stepTimerMinSeconds(step),
      let startDate = stepTimerStartTimes[stepKey(recipeID: recipeID, stepID: stepID)]
    else {
      return false
    }
    return Date().timeIntervalSince(startDate) >= Double(minSeconds)
  }

  func skipStepTimer(recipeID: String, stepID: String) {
    let currentStepKey = stepKey(recipeID: recipeID, stepID: stepID)
    guard stepTimerStartTimes[currentStepKey] != nil else { return }
    completeTimerCondition(recipeID: recipeID, stepID: stepID)
  }

  func cancelStepTimer(recipeID: String, stepID: String) {
    let key = stepKey(recipeID: recipeID, stepID: stepID)
    stepTimerStartTimes.removeValue(forKey: key)
    StepTimerNotificationService.cancelNotification(stepKey: key)
  }

  var hasActiveStepTimers: Bool {
    for (stepKey, startDate) in stepTimerStartTimes {
      let components = stepKey.split(separator: ":", maxSplits: 1)
      guard components.count == 2,
        let recipeID = String(components[0]) as String?,
        let stepID = String(components[1]) as String?
      else {
        continue
      }
      guard let step = stepFor(recipeID: recipeID, stepID: stepID) else {
        continue
      }
      // If step has max time, check elapsed; otherwise it's always active until explicitly skipped
      if step.hasMaxEstimatedTimeInSeconds,
        let doneThreshold = stepTimerDoneThresholdSeconds(step)
      {
        if Date().timeIntervalSince(startDate) < Double(doneThreshold) {
          return true
        }
      } else {
        return true  // No max: always active until explicitly skipped
      }
    }
    return false
  }

  // Get all prerequisite step indices for a given step (old API for main recipe)
  func getPrerequisiteStepIndices(_ stepIndex: Int) -> [Int] {
    guard let recipe = recipe, stepIndex >= 0, stepIndex < recipe.steps.count else {
      return []
    }
    let step = recipe.steps[stepIndex]
    let stepKeys = getPrerequisiteStepKeys(recipeID: recipe.id, stepID: step.id)

    // Convert step keys back to indices for main recipe steps only
    var prerequisites: Set<Int> = []
    for stepKey in stepKeys {
      let components = stepKey.split(separator: ":", maxSplits: 1)
      if components.count == 2, String(components[0]) == recipe.id {
        if let stepID = String(components[1]) as String?,
          let index = recipe.steps.firstIndex(where: { $0.id == stepID })
        {
          prerequisites.insert(index)
        }
      }
    }
    return Array(prerequisites).sorted()
  }

  // Step categorization
  enum StepCategory {
    case upNext
    case forLater
    case done
  }

  // Categorize a step based on its completion status and dependencies
  func categorizeStep(recipeID: String, stepID: String) -> StepCategory {
    let currentStepKey = self.stepKey(recipeID: recipeID, stepID: stepID)

    if let startDate = stepTimerStartTimes[currentStepKey],
      let step = stepFor(recipeID: recipeID, stepID: stepID),
      let doneThreshold = stepTimerDoneThresholdSeconds(step)
    {
      let elapsed = Date().timeIntervalSince(startDate)
      // Only auto-complete timer if step has a defined max time; steps with only min require explicit skip
      if elapsed >= Double(doneThreshold) && step.hasMaxEstimatedTimeInSeconds {
        completeTimerCondition(recipeID: recipeID, stepID: stepID)
        if canCompleteStep(recipeID: recipeID, stepID: stepID) {
          markStepComplete(recipeID: recipeID, stepID: stepID)
          return .done
        }
        return .upNext  // Timer done but other conditions not met
      }
      return .upNext  // Timer still running - stay in Up Next
    }

    if completedSteps.contains(currentStepKey) {
      return .done
    }

    let prerequisiteKeys = getPrerequisiteStepKeys(recipeID: recipeID, stepID: stepID)
    let allDependenciesDone = prerequisiteKeys.allSatisfy { completedSteps.contains($0) }

    if allDependenciesDone {
      return .upNext
    } else {
      return .forLater
    }
  }

  // Get all prerequisite step keys for a given step
  func getPrerequisiteStepKeys(recipeID: String, stepID: String) -> [String] {
    guard let recipe = recipe else { return [] }

    // Find the step
    let step: Mealplanning_RecipeStep?
    if recipeID == recipe.id {
      step = recipe.steps.first { $0.id == stepID }
    } else {
      step = recipe.associatedRecipes.first { $0.id == recipeID }?.steps.first { $0.id == stepID }
    }

    guard let step = step else { return [] }

    var prerequisites: Set<String> = []

    // Check all ingredients
    for ingredient in step.ingredients where ingredient.hasRecipeStepProductID {
      let productID = ingredient.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID] {
        prerequisites.insert(prerequisiteStepKey)
      }
    }

    // Check all instruments
    for instrument in step.instruments where instrument.hasRecipeStepProductID {
      let productID = instrument.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID] {
        prerequisites.insert(prerequisiteStepKey)
      }
    }

    // Check all vessels
    for vessel in step.vessels where vessel.hasRecipeStepProductID {
      let productID = vessel.recipeStepProductID
      if let prerequisiteStepKey = productIDToStepKey[productID] {
        prerequisites.insert(prerequisiteStepKey)
      }
    }

    return Array(prerequisites).sorted()
  }

  // Get the step key that produces a given product ID
  func getStepKeyForProductID(_ productID: String) -> String? {
    return productIDToStepKey[productID]
  }

  // Get the step index for a product ID (for main recipe only, for display purposes)
  // Returns nil if the product is from an associated recipe
  func getStepIndexForProductID(_ productID: String) -> Int? {
    guard let recipe = recipe,
      let stepKey = productIDToStepKey[productID]
    else {
      return nil
    }

    // Parse step key
    let components = stepKey.split(separator: ":", maxSplits: 1)
    guard components.count == 2,
      String(components[0]) == recipe.id,
      let stepID = String(components[1]) as String?
    else {
      return nil  // Product is from associated recipe
    }

    // Find the step index in main recipe
    return recipe.steps.firstIndex(where: { $0.id == stepID })
  }

  /// Returns a display label for a step (e.g. "Drain", "Add (2nd)") for use in "(from X)" references.
  /// Works for both main recipe and associated recipe steps.
  func getStepDisplayLabelForStepKey(_ stepKey: String) -> String? {
    let components = stepKey.split(separator: ":", maxSplits: 1)
    guard components.count == 2,
      let recipeID = String(components[0]) as String?,
      let stepID = String(components[1]) as String?
    else {
      return nil
    }

    guard let step = stepFor(recipeID: recipeID, stepID: stepID) else {
      return nil
    }

    let steps: [Mealplanning_RecipeStep]
    let recipeName: String?
    if recipeID == recipe?.id {
      steps = recipe?.steps ?? []
      recipeName = nil
    } else if let associated = recipe?.associatedRecipes.first(where: { $0.id == recipeID }) {
      steps = associated.steps
      recipeName = associated.name
    } else {
      return nil
    }

    let prepName: String
    if step.hasPreparation && !step.preparation.name.isEmpty {
      prepName = step.preparation.name
    } else {
      prepName = "Step \(Int(step.index) + 1)"
    }

    // Disambiguate when multiple steps share the same preparation name
    var label = prepName
    if step.hasPreparation && !step.preparation.name.isEmpty {
      let samePrepSteps = steps.filter { s in
        s.hasPreparation && !s.preparation.name.isEmpty && s.preparation.name == prepName
      }
      if samePrepSteps.count > 1,
        let occurrenceIndex = samePrepSteps.firstIndex(where: { $0.id == step.id })
      {
        let ordinal = Self.ordinalSuffix(occurrenceIndex + 1)
        label = "\(prepName) (\(ordinal))"
      }
    }

    if let recipeName = recipeName, !recipeName.isEmpty {
      return "\(recipeName): \(label)"
    }
    return label
  }

  /// Returns (label, stepID) for a product from the main recipe, or nil if from an associated recipe.
  /// Used for "(from X)" display and tappable navigation.
  func getStepDisplayInfoForProductID(_ productID: String) -> (label: String, stepID: String)? {
    guard let recipe = recipe,
      let stepKey = productIDToStepKey[productID]
    else {
      return nil
    }

    let components = stepKey.split(separator: ":", maxSplits: 1)
    guard components.count == 2,
      String(components[0]) == recipe.id,
      let stepID = String(components[1]) as String?
    else {
      return nil  // Product is from associated recipe
    }

    guard let label = getStepDisplayLabelForStepKey(stepKey) else {
      return nil
    }
    return (label, stepID)
  }

  private static func ordinalSuffix(_ n: Int) -> String {
    switch n {
    case 1: return "1st"
    case 2: return "2nd"
    case 3: return "3rd"
    case 21: return "21st"
    case 22: return "22nd"
    case 23: return "23rd"
    case 31: return "31st"
    default:
      let mod10 = n % 10
      let mod100 = n % 100
      if mod100 >= 11 && mod100 <= 13 {
        return "\(n)th"
      }
      switch mod10 {
      case 1: return "\(n)st"
      case 2: return "\(n)nd"
      case 3: return "\(n)rd"
      default: return "\(n)th"
      }
    }
  }

  // MARK: - Prep Task Completion

  // Check if a prep task is completed
  func isPrepTaskCompleted(_ prepTaskID: String) -> Bool {
    return completedPrepTasks.contains(prepTaskID)
  }

  // Toggle prep task completion and mark/unmark associated steps
  func togglePrepTask(_ prepTask: Mealplanning_RecipePrepTask) {
    let prepTaskID = prepTask.id

    if completedPrepTasks.contains(prepTaskID) {
      // Uncheck prep task
      completedPrepTasks.remove(prepTaskID)

      // Uncheck all associated steps
      for taskStep in prepTask.taskSteps where !taskStep.belongsToRecipeStep.isEmpty {
        let stepID = taskStep.belongsToRecipeStep
        // Find which recipe this step belongs to
        if let recipe = recipe {
          // Check main recipe steps
          if recipe.steps.contains(where: { $0.id == stepID }) {
            let stepKey = stepKey(recipeID: recipe.id, stepID: stepID)
            uncheckStepAndDependents(stepKey: stepKey)
          } else {
            // Check associated recipe steps
            for associatedRecipe in recipe.associatedRecipes
            where associatedRecipe.steps.contains(where: { $0.id == stepID }) {
              let stepKey = stepKey(recipeID: associatedRecipe.id, stepID: stepID)
              uncheckStepAndDependents(stepKey: stepKey)
              break
            }
          }
        }
      }
    } else {
      // Check prep task
      completedPrepTasks.insert(prepTaskID)

      // Check all associated steps (allow checking even if prerequisites aren't satisfied,
      // since the user has already done the prep task ahead of time)
      for taskStep in prepTask.taskSteps where !taskStep.belongsToRecipeStep.isEmpty {
        let stepID = taskStep.belongsToRecipeStep
        // Find which recipe this step belongs to
        if let recipe = recipe {
          // Check main recipe steps
          if recipe.steps.contains(where: { $0.id == stepID }) {
            let stepKey = stepKey(recipeID: recipe.id, stepID: stepID)
            // Check the step (don't require prerequisites since prep task is done)
            completedSteps.insert(stepKey)
          } else {
            // Check associated recipe steps
            for associatedRecipe in recipe.associatedRecipes
            where associatedRecipe.steps.contains(where: { $0.id == stepID }) {
              let stepKey = stepKey(recipeID: associatedRecipe.id, stepID: stepID)
              // Check the step (don't require prerequisites since prep task is done)
              completedSteps.insert(stepKey)
              break
            }
          }
        }
      }
    }
  }
}
