//
//  MealDetailViewModel.swift
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
class MealDetailViewModel {
  var meal: Mealplanning_Meal?
  var isLoading = false
  var errorMessage: String?

  var mermaidDiagram: String?
  var isLoadingMermaid = false
  var mermaidError: String?

  private let mealID: String
  private let authManager: AuthenticationManager

  init(mealID: String, authManager: AuthenticationManager) {
    self.mealID = mealID
    self.authManager = authManager
  }

  func loadMeal() async {
    isLoading = true
    errorMessage = nil

    do {
      var request = Mealplanning_GetMealRequest()
      request.mealID = mealID

      let response = try await authManager.authenticatedCall("getMeal", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMeal(request, metadata: metadata, options: options)
      }

      self.meal = response.result
    } catch {
      errorMessage = "Failed to load meal: \(error.localizedDescription)"
    }

    isLoading = false
  }

  func loadMermaidDiagram() async {
    guard mermaidDiagram == nil else { return }
    isLoadingMermaid = true
    mermaidError = nil

    do {
      var request = Mealplanning_GetMermaidDiagramForMealRequest()
      request.mealID = mealID

      let response = try await authManager.authenticatedCall(
        "getMermaidDiagramForMeal", idempotent: true
      ) { client, metadata, options in
        try await client.mealPlanning.getMermaidDiagramForMeal(
          request, metadata: metadata, options: options)
      }

      self.mermaidDiagram = response.response
    } catch {
      mermaidError = "Failed to load diagram: \(error.localizedDescription)"
    }

    isLoadingMermaid = false
  }
}
