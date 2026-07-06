//
//  MealListViewModel.swift
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
class MealListViewModel {
  var meals: [Mealplanning_Meal] = []
  var searchResults: [Mealplanning_Meal] = []
  var isLoading = false
  var isSearching = false
  var errorMessage: String?
  var searchError: String?

  private let authManager: AuthenticationManager
  private var searchTask: Task<Void, Never>?

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  var displayedMeals: [Mealplanning_Meal] {
    // If we have search results, show those; otherwise show all meals
    return searchResults.isEmpty ? meals : searchResults
  }

  var isInSearchMode: Bool {
    return !searchResults.isEmpty
  }

  func loadMeals() async {
    isLoading = true
    errorMessage = nil

    do {
      let request = Mealplanning_GetMealsRequest()

      let response = try await authManager.authenticatedCall("getMeals", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMeals(request, metadata: metadata, options: options)
      }

      self.meals = response.results
    } catch {
      errorMessage = "Failed to load meals: \(error.localizedDescription)"
    }

    isLoading = false
  }

  func searchMeals(query: String) {
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
      var request = Mealplanning_SearchForMealsRequest()
      request.query = query
      request.useSearchService = Features.useSearchService

      let response = try await authManager.authenticatedCall("searchForMeals", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.searchForMeals(request, metadata: metadata, options: options)
      }

      searchResults = response.results
    } catch {
      searchError = "Failed to search meals: \(error.localizedDescription)"
      searchResults = []
    }

    isSearching = false
  }
}
