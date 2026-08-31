//
//  CreateMealPlanViewModel.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import Foundation
import GRPCCore
import SwiftProtobuf
import SwiftUI

// MARK: - Wizard Step

enum CreateMealPlanWizardStep: Int, CaseIterable {
  case weekSelection = 1
  case mealAssignment = 2
  case optionSelection = 3
}

// MARK: - Date Occupancy

enum DateOccupancy {
  case accepted  // finalized meal plan
  case proposed  // awaiting votes
}

// MARK: - Create Meal Plan ViewModel (Wizard Mode)

@Observable
@MainActor
// swiftlint:disable:next type_body_length
class CreateMealPlanViewModel {
  // Wizard state
  var wizardStep: CreateMealPlanWizardStep = .weekSelection
  var selectedWeekOffset: Int = 0  // 0 = current week, 1 = next week
  var selectedDates: [Date] = []
  var dayMeals: [Date: Mealplanning_Meal] = [:]

  // Search state (for meal assignment step)
  var searchQuery: String = ""
  var searchResults: [Mealplanning_MealSummary] = []
  var isSearching: Bool = false
  var searchError: String?
  var isAssigningMeal: Bool = false
  var assignmentError: String?

  // Current day index when iterating through meal assignment (0-based)
  var currentDayIndex: Int = 0

  // Meal plan creation state
  var mealPlanName: String = "Meal Plan"
  var isCreating: Bool = false
  var creationError: String?
  var createdMealPlanID: String?

  // Option selections: recipeID -> (stepID -> (ingredientIndex -> selectedOptionIndex))
  var recipeOptionSelections: [String: [String: [UInt32: UInt32]]] = [:]

  // Meal scale per date (default 1.0)
  var dayMealScales: [Date: Float] = [:]

  private let authManager: AuthenticationManager
  private let acceptedOccupiedDates: Set<Date>
  private let proposedOccupiedDates: Set<Date>

  private var calendar: Calendar {
    var cal = Calendar.current
    cal.firstWeekday = 2  // Monday
    return cal
  }

  init(
    authManager: AuthenticationManager,
    acceptedOccupiedDates: Set<Date> = [],
    proposedOccupiedDates: Set<Date> = []
  ) {
    self.authManager = authManager
    self.acceptedOccupiedDates = acceptedOccupiedDates
    self.proposedOccupiedDates = proposedOccupiedDates
  }

  // MARK: - Week and Date Helpers

  /// Start of the displayed week (Monday)
  var displayedWeekStart: Date {
    let now = Date()
    let weekStart = calendar.dateInterval(of: .weekOfYear, for: now)?.start ?? now
    return calendar.date(byAdding: .day, value: selectedWeekOffset * 7, to: weekStart) ?? weekStart
  }

  /// All 7 days of the displayed week
  var displayedWeekDays: [Date] {
    (0..<7).compactMap { offset in
      calendar.date(byAdding: .day, value: offset, to: displayedWeekStart)
    }
  }

  func isDateSelected(_ date: Date) -> Bool {
    selectedDates.contains { calendar.isDate($0, inSameDayAs: date) }
  }

  /// Whether the user can plan dinner for this date. Past dates are not planable.
  /// If it's after 6PM, today is also not planable.
  /// Dates occupied by accepted or proposed meal plans are not planable.
  func isDatePlanable(_ date: Date) -> Bool {
    let dateNorm = calendar.startOfDay(for: date)
    if acceptedOccupiedDates.contains(dateNorm) || proposedOccupiedDates.contains(dateNorm) {
      return false
    }
    let now = Date()
    let todayStart = calendar.startOfDay(for: now)

    if dateNorm < todayStart {
      return false  // yesterday and earlier
    }
    if dateNorm > todayStart {
      return true  // future dates
    }
    // today
    let hour = calendar.component(.hour, from: now)
    return hour < 18
  }

  /// Occupancy status for UI coloring. Accepted checked first, then proposed.
  func dateOccupancy(for date: Date) -> DateOccupancy? {
    let dateNorm = calendar.startOfDay(for: date)
    if acceptedOccupiedDates.contains(dateNorm) { return .accepted }
    if proposedOccupiedDates.contains(dateNorm) { return .proposed }
    return nil
  }

  func toggleDateSelection(_ date: Date) {
    let normalized = calendar.startOfDay(for: date)
    if let index = selectedDates.firstIndex(where: { calendar.isDate($0, inSameDayAs: date) }) {
      selectedDates.remove(at: index)
    } else {
      guard isDatePlanable(date) else { return }
      selectedDates.append(normalized)
      selectedDates.sort()
    }
    updateMealPlanName()
  }

  /// Select range from start to end (inclusive), replacing current selection.
  /// Only planable dates are included.
  func setDateRangeSelection(from start: Date, to end: Date) {
    var current = calendar.startOfDay(for: start)
    let endNorm = calendar.startOfDay(for: end)
    var newSelection: [Date] = []

    while current <= endNorm {
      if isDatePlanable(current) {
        newSelection.append(current)
      }
      guard let next = calendar.date(byAdding: .day, value: 1, to: current) else { break }
      current = next
    }

    selectedDates = newSelection.sorted()
    updateMealPlanName()
  }

  func goToNextWeek() {
    selectedWeekOffset = min(selectedWeekOffset + 1, 4)  // Cap at 4 weeks ahead
    selectedDates = []
    updateMealPlanName()
  }

  func goToPreviousWeek() {
    selectedWeekOffset = max(selectedWeekOffset - 1, 0)
    selectedDates = []
    updateMealPlanName()
  }

  private func updateMealPlanName() {
    guard !selectedDates.isEmpty else {
      mealPlanName = "Meal Plan"
      return
    }
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    let sorted = selectedDates.sorted()
    guard let firstDate = sorted.first, let lastDate = sorted.last else {
      mealPlanName = "Meal Plan"
      return
    }
    mealPlanName =
      "Dinners \(formatter.string(from: firstDate)) - \(formatter.string(from: lastDate))"
  }

  /// Date at 7PM for meal plan event
  private func dateAt7PM(for date: Date) -> Date {
    calendar.date(bySettingHour: 19, minute: 0, second: 0, of: date) ?? date
  }

  private func autoEndDate(from startDate: Date) -> Date {
    calendar.date(byAdding: .hour, value: 1, to: startDate) ?? startDate
  }

  // MARK: - Meal Assignment

  /// Assigns an already-fetched meal to a day.
  func assignMeal(_ meal: Mealplanning_Meal, to date: Date) {
    let normalized = selectedDates.first { calendar.isDate($0, inSameDayAs: date) } ?? date
    dayMeals[normalized] = meal
    searchQuery = ""
    searchResults = []
    searchError = nil
  }

  /// Assigns a meal picked from search. Search returns MealSummary, whose
  /// components carry recipes without steps, so the whole meal is fetched
  /// first -- the option-selection step downstream reads step ingredients to
  /// find option groups, and a summary has none.
  func assignMeal(fromSearchResult summary: Mealplanning_MealSummary, to date: Date) async {
    isAssigningMeal = true
    assignmentError = nil

    do {
      var request = Mealplanning_GetMealRequest()
      request.mealID = summary.id

      let response = try await authManager.authenticatedCall("getMeal", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMeal(request, metadata: metadata, options: options)
      }

      assignMeal(response.result, to: date)
    } catch {
      assignmentError = "Failed to add meal: \(error.localizedDescription)"
    }

    isAssigningMeal = false
  }

  func removeMeal(from date: Date) {
    let normalized = selectedDates.first { calendar.isDate($0, inSameDayAs: date) } ?? date
    dayMeals.removeValue(forKey: normalized)
  }

  func mealForDate(_ date: Date) -> Mealplanning_Meal? {
    selectedDates.first { calendar.isDate($0, inSameDayAs: date) }.flatMap { dayMeals[$0] }
  }

  /// Look up an assigned meal by ID. Search results are summaries and are not
  /// consulted here -- a caller wanting a whole meal wants an assigned one.
  func meal(forId id: String) -> Mealplanning_Meal? {
    dayMeals.values.first { $0.id == id }
  }

  var allDaysHaveMeals: Bool {
    guard !selectedDates.isEmpty else { return false }
    return selectedDates.allSatisfy { dayMeals[$0] != nil }
  }

  var allSelectedMeals: [Mealplanning_Meal] {
    selectedDates.compactMap { dayMeals[$0] }
  }

  /// The date we're currently planning for (in meal assignment step)
  var currentPlanningDate: Date? {
    let sorted = selectedDates.sorted()
    guard currentDayIndex >= 0, currentDayIndex < sorted.count else { return nil }
    return sorted[currentDayIndex]
  }

  var canGoToPreviousDay: Bool {
    currentDayIndex > 0
  }

  var canGoToNextDay: Bool {
    currentDayIndex < selectedDates.sorted().count - 1
  }

  func goToPreviousDay() {
    guard canGoToPreviousDay else { return }
    currentDayIndex -= 1
    searchQuery = ""
    searchResults = []
    searchError = nil
  }

  func goToNextDay() {
    guard canGoToNextDay else { return }
    currentDayIndex += 1
    searchQuery = ""
    searchResults = []
    searchError = nil
  }

  /// Reset day index when entering meal assignment (call from view onAppear)
  func resetMealAssignmentState() {
    currentDayIndex = 0
    searchQuery = ""
    searchResults = []
    searchError = nil
  }

  // MARK: - Search

  func searchForMeals() async {
    let queryWhenStarted = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !queryWhenStarted.isEmpty else {
      searchResults = []
      return
    }

    isSearching = true
    searchError = nil

    do {
      var request = Mealplanning_SearchForMealsRequest()
      request.query = queryWhenStarted
      request.useSearchService = Features.useSearchService

      let response = try await authManager.authenticatedCall("searchForMeals", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.searchForMeals(request, metadata: metadata, options: options)
      }

      if searchQuery.trimmingCharacters(in: .whitespacesAndNewlines) == queryWhenStarted {
        searchResults = response.results
        AnalyticsConfiguration.provideEventReporter().track(
          event: "meal_plan_meal_assignment_search",
          properties: ["query_length": queryWhenStarted.count])
      }
    } catch {
      if searchQuery.trimmingCharacters(in: .whitespacesAndNewlines) == queryWhenStarted {
        searchError = "Failed to search meals: \(error.localizedDescription)"
        searchResults = []
      }
    }

    isSearching = false
  }

  /// Filter search results. Excludes only the meal for the given date so when going
  /// back to change a day, other meals (including those on other days) still appear.
  func filteredSearchResults(for date: Date?) -> [Mealplanning_MealSummary] {
    guard let date = date else { return searchResults }
    let currentMealID = mealForDate(date)?.id
    return searchResults.filter { currentMealID == nil || $0.id != currentMealID }
  }

  /// Returns the date this meal is assigned to, if assigned to a day other than the given one.
  func assignedDayForMeal(mealID: String, excludingDate: Date) -> Date? {
    let exclNorm = selectedDates.first { calendar.isDate($0, inSameDayAs: excludingDate) }
    for (date, meal) in dayMeals where meal.id == mealID {
      switch exclNorm {
      case nil:
        return date
      case let excl?:
        if !calendar.isDate(date, inSameDayAs: excl) {
          return date
        }
      }
    }
    return nil
  }

  func mealScale(for date: Date) -> Float {
    let normalized = selectedDates.first { calendar.isDate($0, inSameDayAs: date) } ?? date
    return dayMealScales[normalized] ?? 1.0
  }

  func setMealScale(_ scale: Float, for date: Date) {
    let normalized = selectedDates.first { calendar.isDate($0, inSameDayAs: date) } ?? date
    let clamped = min(max(scale, 0.25), 4.0)
    dayMealScales[normalized] = clamped
  }

  func adjustMealScale(for date: Date, by delta: Float) {
    setMealScale(mealScale(for: date) + delta, for: date)
  }

  func mealScaleText(for date: Date) -> String {
    String(format: "%.2fx", mealScale(for: date))
  }

  // MARK: - Meal Plan Creation

  func canCreateMealPlan() -> Bool {
    !selectedDates.isEmpty && allDaysHaveMeals
  }

  func createMealPlan() async -> Bool {
    guard canCreateMealPlan() else {
      creationError = "Each day must have a meal assigned"
      return false
    }

    isCreating = true
    creationError = nil

    do {
      var input = Mealplanning_MealPlanCreationRequestInput()
      input.notes = mealPlanName
      // Solo flow: set voting deadline to now so cron finalizes immediately
      input.votingDeadline = dateToTimestamp(Date())
      input.electionMethod = .schulze

      let sortedDates = selectedDates.sorted()
      for date in sortedDates {
        guard let meal = dayMeals[date] else { continue }

        var eventInput = Mealplanning_MealPlanEventCreationRequestInput()
        let startAt = dateAt7PM(for: date)
        eventInput.startsAt = dateToTimestamp(startAt)
        eventInput.endsAt = dateToTimestamp(autoEndDate(from: startAt))
        eventInput.mealName = .dinner
        eventInput.notes = ""

        var optionInput = Mealplanning_MealPlanOptionCreationRequestInput()
        optionInput.mealID = meal.id
        optionInput.mealScale = mealScale(for: date)
        optionInput.notes = ""

        for component in meal.components {
          let recipeID = component.recipe.id
          if let stepSelections = recipeOptionSelections[recipeID] {
            for (stepID, indexSelections) in stepSelections {
              for (ingredientIndex, selectedOptionIndex) in indexSelections {
                var selectionInput =
                  Mealplanning_MealPlanRecipeOptionSelectionCreationRequestInput()
                selectionInput.recipeID = recipeID
                selectionInput.recipeStepID = stepID
                selectionInput.ingredientIndex = ingredientIndex
                selectionInput.selectedOptionIndex = selectedOptionIndex
                selectionInput.selectionType = .ingredient
                optionInput.selections.append(selectionInput)
              }
            }
          }
        }

        eventInput.options.append(optionInput)
        input.events.append(eventInput)
      }

      var request = Mealplanning_CreateMealPlanRequest()
      request.input = input

      let response = try await authManager.authenticatedCall("createMealPlan") {
        client, metadata, options in
        try await client.mealPlanning.createMealPlan(request, metadata: metadata, options: options)
      }

      let createdPlan = response.created

      // Solo flow: create votes for each option so finalizer can pick winner
      let userID = authManager.userID
      if !userID.isEmpty {
        for event in createdPlan.events where event.options.count == 1 {
          let option = event.options[0]
          var voteInput = Mealplanning_MealPlanOptionVoteCreationRequestInput()
          var vote = Mealplanning_MealPlanOptionVoteCreationInput()
          vote.byUser = userID
          vote.belongsToMealPlanOption = option.id
          vote.rank = 0
          vote.abstain = false
          voteInput.votes.append(vote)

          var voteRequest = Mealplanning_CreateMealPlanOptionVoteRequest()
          voteRequest.mealPlanID = createdPlan.id
          voteRequest.mealPlanEventID = event.id
          voteRequest.input = voteInput

          _ = try? await authManager.authenticatedCall("createMealPlanOptionVote") {
            client, metadata, options in
            try await client.mealPlanning.createMealPlanOptionVote(
              voteRequest, metadata: metadata, options: options)
          }
        }
      }

      createdMealPlanID = createdPlan.id
      isCreating = false
      return true
    } catch let error as GRPCCore.RPCError {
      creationError =
        error.code == .alreadyExists
        ? "One or more meals are already in this plan."
        : "Failed to create meal plan: \(error.localizedDescription)"
      AnalyticsConfiguration.provideEventReporter().track(
        event: "meal_plan_creation_failed",
        properties: ["error": creationError ?? "Unknown error"])
      isCreating = false
      return false
    } catch {
      creationError = "Failed to create meal plan: \(error.localizedDescription)"
      AnalyticsConfiguration.provideEventReporter().track(
        event: "meal_plan_creation_failed",
        properties: ["error": creationError ?? "Unknown error"])
      isCreating = false
      return false
    }
  }

  // MARK: - Option Selection Helpers

  func setOptionSelections(ingredientSelections: [String: [String: [UInt32: UInt32]]]) {
    recipeOptionSelections = ingredientSelections
  }

  /// Returns true if any recipe in this meal has ingredient option groups.
  func mealHasIngredientOptions(_ meal: Mealplanning_Meal) -> Bool {
    !collectRecipesWithOptions(from: [meal]).isEmpty
  }

  func collectRecipesWithOptions(from meals: [Mealplanning_Meal]) -> Set<String> {
    var recipeIDsWithOptions: Set<String> = []
    var allRecipes: [Mealplanning_Recipe] = []
    for meal in meals {
      for component in meal.components {
        allRecipes.append(component.recipe)
        allRecipes.append(contentsOf: component.recipe.associatedRecipes)
      }
    }

    for recipe in allRecipes {
      guard !recipe.steps.isEmpty else { continue }
      let hasOptions = recipe.steps.contains { step in
        var indexGroups: [UInt32: [Mealplanning_RecipeStepIngredient]] = [:]
        for ingredient in step.ingredients {
          let index = ingredient.index
          if indexGroups[index] == nil { indexGroups[index] = [] }
          indexGroups[index]?.append(ingredient)
        }
        return indexGroups.values.contains { $0.count > 1 }
      }
      if hasOptions { recipeIDsWithOptions.insert(recipe.id) }
    }
    return recipeIDsWithOptions
  }

  func getAllRecipes(from meals: [Mealplanning_Meal]) -> [Mealplanning_Recipe] {
    var allRecipes: [Mealplanning_Recipe] = []
    for meal in meals {
      for component in meal.components {
        allRecipes.append(component.recipe)
        allRecipes.append(contentsOf: component.recipe.associatedRecipes)
      }
    }
    var seen: Set<String> = []
    return allRecipes.filter { seen.insert($0.id).inserted }
  }

  func getDefaultOptionSelections(for recipe: Mealplanning_Recipe) -> [String: [UInt32: UInt32]] {
    var defaults: [String: [UInt32: UInt32]] = [:]
    for step in recipe.steps {
      var optionGroupsByIndex: [UInt32: [Mealplanning_RecipeStepIngredient]] = [:]
      for ingredient in step.ingredients {
        if optionGroupsByIndex[ingredient.index] == nil {
          optionGroupsByIndex[ingredient.index] = []
        }
        optionGroupsByIndex[ingredient.index]?.append(ingredient)
      }
      for (index, group) in optionGroupsByIndex where group.count > 1 {
        if defaults[step.id] == nil { defaults[step.id] = [:] }
        defaults[step.id]?[index] = 0
      }
    }
    return defaults
  }

  private func dateToTimestamp(_ date: Date) -> SwiftProtobuf.Google_Protobuf_Timestamp {
    var timestamp = SwiftProtobuf.Google_Protobuf_Timestamp()
    timestamp.seconds = Int64(date.timeIntervalSince1970)
    timestamp.nanos = Int32(
      (date.timeIntervalSince1970 - Double(timestamp.seconds)) * 1_000_000_000)
    return timestamp
  }
}
