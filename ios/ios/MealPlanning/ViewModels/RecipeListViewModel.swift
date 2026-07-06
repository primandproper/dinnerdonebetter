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
  var recipes: [Mealplanning_Recipe] = []
  var searchResults: [Mealplanning_Recipe] = []
  var isLoading = false
  var isSearching = false
  var errorMessage: String?
  var searchError: String?

  /// Recipe status filter for GetRecipes. Default "approved"; service admins can toggle to "submitted".
  var recipeStatusFilter: String = "approved"
  /// True when current user has service_role = "service_admin". Enables status toggle in UI.
  var isServiceAdmin = false

  private let authManager: AuthenticationManager
  private var searchTask: Task<Void, Never>?
  private var hasCheckedServiceAdmin = false

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  var displayedRecipes: [Mealplanning_Recipe] {
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

  /// Fetches current user to determine if they are a service_admin. Call once when recipe list appears.
  func loadCurrentUserForAdminCheck() async {
    guard !hasCheckedServiceAdmin else { return }
    hasCheckedServiceAdmin = true

    do {
      let user = try await CurrentUserService.shared.currentUser(using: authManager)
      isServiceAdmin = user.serviceRole == "service_admin"
    } catch {
      // Non-fatal: just leave isServiceAdmin false
    }
  }

  /// Updates recipe status filter and reloads. For service admin toggle.
  func setRecipeStatusFilter(_ status: String) {
    recipeStatusFilter = status
  }
}
