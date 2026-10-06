//
//  RecipeListViewModel.swift
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
class RecipeListViewModel {
  var recipes: [Mealplanning_RecipeSummary] = []
  var searchResults: [Mealplanning_RecipeSummary] = []
  var isLoading = false
  var isSearching = false
  var errorMessage: String?
  var searchError: String?

  /// Recipe status filter for GetRecipes. Default "approved"; recipe reviewers can toggle to
  /// "submitted".
  var recipeStatusFilter: String = "approved"
  /// The caller's effective permissions, as GetPrincipal answers them. Closed until loaded.
  var permissions: CallerPermissions = .unloaded

  /// Whether the caller reviews submitted recipes, which is what the status toggle is for. It
  /// is the grant on UpdateRecipeStatus rather than a service role's name.
  var canReviewRecipes: Bool {
    permissions.allows(CallerPermissions.Name.updateRecipeStatus)
  }

  private let authManager: AuthenticationManager
  private var searchTask: Task<Void, Never>?
  private var hasLoadedPermissions = false

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  var displayedRecipes: [Mealplanning_RecipeSummary] {
    // If we have search results, show those; otherwise show all recipes
    return searchResults.isEmpty ? recipes : searchResults
  }

  var isInSearchMode: Bool {
    return !searchResults.isEmpty
  }

  func loadRecipes() async {
    isLoading = true
    errorMessage = nil

    do {
      // Use selected status (approved by default; service admins can toggle).
      var request = Mealplanning_GetRecipesRequest()
      request.status = recipeStatusFilter

      let response = try await authManager.authenticatedCall("getRecipes", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getRecipes(request, metadata: metadata, options: options)
      }

      self.recipes = response.results
    } catch {
      errorMessage = "Failed to load recipes: \(error.localizedDescription)"
    }

    isLoading = false
  }

  func searchRecipes(query: String) {
    // Cancel any existing search task
    searchTask?.cancel()

    let trimmedQuery = query.trimmingCharacters(in: .whitespacesAndNewlines)

    // If query is empty, clear search results
    if trimmedQuery.isEmpty {
      searchResults = []
      searchError = nil
      isSearching = false
      return
    }

    // Debounce: wait 500ms before executing search
    searchTask = Task {
      try? await Task.sleep(nanoseconds: 500_000_000)  // 500ms

      // Check if task was cancelled
      guard !Task.isCancelled else { return }

      await performSearch(query: trimmedQuery)
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

      searchResults = response.results
    } catch {
      searchError = "Failed to search recipes: \(error.localizedDescription)"
      searchResults = []
    }

    isSearching = false
  }

  /// Reads the caller's permissions, which decide whether the status toggle shows. Call once
  /// when the recipe list appears.
  func loadPermissions() async {
    guard !hasLoadedPermissions else { return }
    hasLoadedPermissions = true

    do {
      permissions = try await authManager.callerPermissions()
    } catch {
      // Non-fatal: the toggle stays hidden.
    }
  }

  /// Updates recipe status filter and reloads. For the recipe reviewers' toggle.
  func setRecipeStatusFilter(_ status: String) {
    recipeStatusFilter = status
  }
}
