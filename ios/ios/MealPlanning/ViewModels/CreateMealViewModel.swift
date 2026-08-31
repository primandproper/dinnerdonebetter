//
//  CreateMealViewModel.swift
//  ios
//
//  View model for the meal creation flow: basic info, add recipes, review with validation.
//

import Foundation
import GRPCCore
import SwiftProtobuf
import SwiftUI

// MARK: - Draft Component

struct CreateMealDraftComponent: Identifiable {
  let id: String
  let recipe: Mealplanning_Recipe
  var componentType: Mealplanning_MealComponentType
  var recipeScale: Float
}

// MARK: - Create Meal ViewModel

@Observable
@MainActor
class CreateMealViewModel {
  // Basic info
  var mealName: String = ""
  var mealDescription: String = ""

  // Recipe search
  var searchQuery: String = ""
  var searchResults: [Mealplanning_RecipeSummary] = []
  var isSearching: Bool = false
  var searchError: String?

  // Draft components (recipes added to the meal)
  var draftComponents: [CreateMealDraftComponent] = []
  var isAddingRecipe: Bool = false
  var addRecipeError: String?

  // Creation
  var isCreating: Bool = false
  var creationError: String?
  var createdMealID: String?

  // Wizard step (1=add recipes, 2=review)
  var wizardStep: Int = 1

  private let authManager: AuthenticationManager

  /// Sets mealName to component names joined by " & " when empty. Call when navigating to review.
  func ensureDefaultMealNameFromComponents() {
    guard mealName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
      !draftComponents.isEmpty
    else { return }
    let names = draftComponents.map { component in
      component.recipe.name.isEmpty ? "Unnamed Recipe" : component.recipe.name
    }
    mealName = names.joined(separator: " & ")
  }
  private var searchTask: Task<Void, Never>?

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  var hasAtLeastOneMain: Bool {
    draftComponents.contains { $0.componentType == .main }
  }

  var canCreateMeal: Bool {
    !mealName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
      && !draftComponents.isEmpty
      && hasAtLeastOneMain
  }

  // MARK: - Search

  /// Search for recipes. Uses SearchForRecipes (same as RecipeListView) for reliable search.
  func searchRecipes(query: String) {
    searchTask?.cancel()
    searchQuery = query

    let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
    if trimmed.isEmpty {
      searchResults = []
      searchError = nil
      isSearching = false
      return
    }

    searchTask = Task {
      try? await Task.sleep(nanoseconds: 500_000_000)  // 500ms debounce (match RecipeListView)
      guard !Task.isCancelled else { return }
      await performSearch(query: trimmed)
    }
  }

  private func performSearch(query: String) async {
    isSearching = true
    searchError = nil

    do {
      var request = Mealplanning_SearchForRecipesRequest()
      request.query = query
      request.useSearchService = Features.useSearchService

      let response = try await authManager.authenticatedCall("searchForRecipes", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.searchForRecipes(
          request, metadata: metadata, options: options)
      }

      if searchQuery.trimmingCharacters(in: .whitespacesAndNewlines) == query {
        searchResults = response.results
      }
    } catch {
      if searchQuery.trimmingCharacters(in: .whitespacesAndNewlines) == query {
        searchError = "Failed to search: \(error.localizedDescription)"
        searchResults = []
      }
    }

    isSearching = false
  }

  /// Recipes from search that are not yet in the draft
  func searchResultsNotInDraft() -> [Mealplanning_RecipeSummary] {
    let draftIDs = Set(draftComponents.map { $0.recipe.id })
    return searchResults.filter { !draftIDs.contains($0.id) }
  }

  // MARK: - Add Recipe

  /// Adds a recipe to the draft. Fetches full recipe via GetRecipe for validation (steps, vessels, temps).
  func addRecipe(_ recipe: Mealplanning_RecipeSummary) async {
    let recipeID = recipe.id
    guard !draftComponents.contains(where: { $0.recipe.id == recipeID }) else { return }

    isAddingRecipe = true
    addRecipeError = nil

    do {
      var request = Mealplanning_GetRecipeRequest()
      request.recipeID = recipeID

      let response = try await authManager.authenticatedCall("getRecipe", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getRecipe(request, metadata: metadata, options: options)
      }

      let fullRecipe = response.result
      let componentType: Mealplanning_MealComponentType = hasAtLeastOneMain ? .side : .main

      let component = CreateMealDraftComponent(
        id: recipeID,
        recipe: fullRecipe,
        componentType: componentType,
        recipeScale: 1.0
      )
      draftComponents.append(component)
    } catch {
      addRecipeError = "Failed to add recipe: \(error.localizedDescription)"
    }

    isAddingRecipe = false
  }

  func removeRecipe(atOffsets offsets: IndexSet) {
    draftComponents.remove(atOffsets: offsets)
  }

  func setComponentType(_ type: Mealplanning_MealComponentType, for componentID: String) {
    guard let index = draftComponents.firstIndex(where: { $0.id == componentID }) else { return }
    draftComponents[index].componentType = type
  }

  func setRecipeScale(_ scale: Float, for componentID: String) {
    guard let index = draftComponents.firstIndex(where: { $0.id == componentID }) else { return }
    let clamped = min(max(scale, 0.25), 4.0)
    draftComponents[index].recipeScale = clamped
  }

  // MARK: - Create Meal

  func createMeal() async -> Bool {
    guard canCreateMeal else {
      creationError = "Please add a name and at least one main recipe."
      return false
    }

    isCreating = true
    creationError = nil

    do {
      var input = Mealplanning_MealCreationRequestInput()
      input.name = mealName.trimmingCharacters(in: .whitespacesAndNewlines)
      input.description_p = mealDescription.trimmingCharacters(in: .whitespacesAndNewlines)
      input.eligibleForMealPlans = true

      if let first = draftComponents.first {
        input.minEstimatedPortions = first.recipe.minEstimatedPortions
        if first.recipe.hasMaxEstimatedPortions {
          input.maxEstimatedPortions = first.recipe.maxEstimatedPortions
        }
      } else {
        input.minEstimatedPortions = 4
      }

      for component in draftComponents {
        var compInput = Mealplanning_MealComponentCreationRequestInput()
        compInput.recipeID = component.recipe.id
        compInput.componentType = component.componentType
        compInput.recipeScale = component.recipeScale
        input.components.append(compInput)
      }

      var request = Mealplanning_CreateMealRequest()
      request.input = input

      let response = try await authManager.authenticatedCall("createMeal") {
        client, metadata, options in
        try await client.mealPlanning.createMeal(request, metadata: metadata, options: options)
      }

      createdMealID = response.created.id
      isCreating = false
      return true
    } catch let error as GRPCCore.RPCError {
      creationError =
        error.code == .alreadyExists
        ? "A meal with this name and components already exists."
        : "Failed to create meal: \(error.localizedDescription)"
      isCreating = false
      return false
    } catch {
      creationError = "Failed to create meal: \(error.localizedDescription)"
      isCreating = false
      return false
    }
  }
}
