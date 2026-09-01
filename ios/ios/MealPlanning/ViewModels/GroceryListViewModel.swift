//
//  GroceryListViewModel.swift
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
class GroceryListViewModel {
  // Data
  var items: [Mealplanning_MealPlanGroceryListItem] = []
  // A summary is enough here: this reads the plan's ID and its date range.
  var mealPlan: any MealPlanDisplayable

  // Loading states
  var isLoading = false
  var isUpdating = false
  var errorMessage: String?

  private let authManager: AuthenticationManager

  init(
    mealPlan: any MealPlanDisplayable, items: [Mealplanning_MealPlanGroceryListItem],
    authManager: AuthenticationManager
  ) {
    self.mealPlan = mealPlan
    self.items = items
    self.authManager = authManager
  }

  func loadItems() async {
    isLoading = true
    errorMessage = nil

    do {
      let fetchedItems = try await fetchGroceryListItems()
      self.items = fetchedItems
    } catch {
      errorMessage = "Failed to load grocery list: \(error.localizedDescription)"
    }

    isLoading = false
  }

  private func fetchGroceryListItems() async throws -> [Mealplanning_MealPlanGroceryListItem] {
    var request = Mealplanning_GetMealPlanGroceryListItemsForMealPlanRequest()
    request.mealPlanID = mealPlan.id

    let response = try await authManager.authenticatedCall(
      "getMealPlanGroceryListItemsForMealPlan", idempotent: true
    ) { client, metadata, options in
      try await client.mealPlanning.getMealPlanGroceryListItemsForMealPlan(
        request, metadata: metadata, options: options)
    }

    return response.results
  }

  func updateItem(
    _ item: Mealplanning_MealPlanGroceryListItem,
    status: Mealplanning_MealPlanGroceryListItemStatus? = nil,
    quantityNeededMin: Float? = nil,
    quantityNeededMax: Float? = nil,
    quantityPurchased: Float? = nil
  ) async {
    isUpdating = true
    errorMessage = nil

    do {
      var updateInput = Mealplanning_MealPlanGroceryListItemUpdateRequestInput()

      // Update status if provided
      if let status = status {
        updateInput.status = status
      }

      // Update quantity needed if provided
      if let min = quantityNeededMin {
        updateInput.minQuantityNeeded = min
        if let max = quantityNeededMax {
          updateInput.maxQuantityNeeded = max
        }
        // When max is nil, we omit it so backend clears max (single-value edit)
      }

      // Update quantity purchased if provided
      if let purchased = quantityPurchased {
        updateInput.quantityPurchased = purchased
      }

      let updated = try await updateGroceryListItem(itemID: item.id, input: updateInput)

      // Optimistically update local state instead of full reload
      if let index = items.firstIndex(where: { $0.id == item.id }) {
        items[index] = updated
      }
    } catch {
      errorMessage = "Failed to update item: \(error.localizedDescription)"
    }

    isUpdating = false
  }

  func markAsAcquired(_ item: Mealplanning_MealPlanGroceryListItem) async {
    await updateItem(item, status: .acquired)
  }

  func markAsAlreadyOwned(_ item: Mealplanning_MealPlanGroceryListItem) async {
    await updateItem(item, status: .alreadyOwned)
  }

  func markAsNeeds(_ item: Mealplanning_MealPlanGroceryListItem) async {
    await updateItem(item, status: .needs)
  }

  func markAsUnavailable(_ item: Mealplanning_MealPlanGroceryListItem) async {
    await updateItem(item, status: .unavailable)
  }

  func updateQuantityNeeded(
    _ item: Mealplanning_MealPlanGroceryListItem,
    min: Float,
    max: Float?
  ) async {
    await updateItem(
      item,
      quantityNeededMin: min,
      quantityNeededMax: max
    )
  }

  func updateQuantityPurchased(
    _ item: Mealplanning_MealPlanGroceryListItem,
    quantity: Float
  ) async {
    await updateItem(
      item,
      quantityPurchased: quantity
    )
  }

  private func updateGroceryListItem(
    itemID: String,
    input: Mealplanning_MealPlanGroceryListItemUpdateRequestInput
  ) async throws -> Mealplanning_MealPlanGroceryListItem {
    var request = Mealplanning_UpdateMealPlanGroceryListItemRequest()
    request.mealPlanID = mealPlan.id
    request.mealPlanGroceryListItemID = itemID
    request.input = input

    let response = try await authManager.authenticatedCall("updateMealPlanGroceryListItem") {
      client, metadata, options in
      try await client.mealPlanning.updateMealPlanGroceryListItem(
        request, metadata: metadata, options: options)
    }
    return response.updated
  }

  // Computed properties for filtering
  var itemsByStatus:
    [Mealplanning_MealPlanGroceryListItemStatus: [Mealplanning_MealPlanGroceryListItem]]
  {
    Dictionary(grouping: items) { $0.status }
  }

  var needsItems: [Mealplanning_MealPlanGroceryListItem] {
    items.filter { $0.status == .needs }
  }

  var alreadyOwnedItems: [Mealplanning_MealPlanGroceryListItem] {
    items.filter { $0.status == .alreadyOwned }
  }

  var acquiredItems: [Mealplanning_MealPlanGroceryListItem] {
    items.filter { $0.status == .acquired }
  }

  var unavailableItems: [Mealplanning_MealPlanGroceryListItem] {
    items.filter { $0.status == .unavailable }
  }

  /// Combined acquired + already owned for the "Have" section.
  var haveItems: [Mealplanning_MealPlanGroceryListItem] {
    items.filter { $0.status == .acquired || $0.status == .alreadyOwned }
  }
}
