//
//  HomeViewModel.swift
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
class HomeViewModel {
  // Type alias for convenience
  typealias Timestamp = SwiftProtobuf.Google_Protobuf_Timestamp

  // Static utility function to convert protobuf timestamp to Date
  // Can be used from both HomeViewModel and HomeView
  static func timestampToDate(_ timestamp: SwiftProtobuf.Google_Protobuf_Timestamp) -> Date {
    // Google_Protobuf_Timestamp.seconds is Int64, nanos is Int32
    // TimeInterval is a typealias for Double
    let seconds = TimeInterval(timestamp.seconds)
    let nanos = TimeInterval(timestamp.nanos) / 1_000_000_000.0
    return Date(timeIntervalSince1970: seconds + nanos)
  }
  // Data
  var allMealPlans: [Mealplanning_MealPlan] = []
  var userTasks: [Mealplanning_MealPlanTask] = []
  var tasksByMealPlan: [String: [Mealplanning_MealPlanTask]] = [:]
  var groceryLists: [String: [Mealplanning_MealPlanGroceryListItem]] = [:]  // Keyed by meal plan ID
  var currentUser: Identity_User?

  /// Display name for welcome message: first name if set, otherwise username.
  var currentUserDisplayName: String {
    guard let user = currentUser else {
      return authManager.username
    }
    if !user.firstName.isEmpty {
      return user.firstName
    }
    return user.username.isEmpty ? authManager.username : user.username
  }

  // Loading states
  var isLoading = false
  var errorMessage: String?
  var errorTitle: String = "Error"
  var errorIcon: String = "exclamationmark.triangle"
  var errorIconColor = DSTheme.Colors.warning
  var isServerDownError = false

  // Computed properties
  var pendingVoteMealPlans: [Mealplanning_MealPlan] {
    let now = Date()
    return allMealPlans.filter { mealPlan in
      // Meal plan is pending if voting deadline hasn't passed and status indicates voting
      guard mealPlan.status == .awaitingVotes else {
        return false
      }

      let deadline = timestampToDate(mealPlan.votingDeadline)
      return deadline > now
    }
  }

  /// The active meal plan: finalized plan whose start/end range is closest to the current date.
  var activeMealPlan: Mealplanning_MealPlan? {
    let candidates = allMealPlans.filter { mealPlan in
      mealPlan.status == .finalized && !mealPlan.events.isEmpty
    }
    guard !candidates.isEmpty else { return nil }
    return candidates.min { lhs, rhs in
      distanceFromNow(to: lhs) < distanceFromNow(to: rhs)
    }
  }

  /// Distance from now to a plan's date range: 0 if within, otherwise seconds to nearest boundary.
  private func distanceFromNow(to mealPlan: Mealplanning_MealPlan) -> TimeInterval {
    let now = Date()
    let start = mealPlanStartDate(mealPlan)
    let end = mealPlanEndDate(mealPlan)
    if now < start { return start.timeIntervalSince(now) }
    if now > end { return now.timeIntervalSince(end) }
    return 0
  }

  /// Finalized meal plans that are not the active plan (e.g. next week's plan).
  var futureFinalizedMealPlans: [Mealplanning_MealPlan] {
    guard let active = activeMealPlan else {
      return allMealPlans.filter { $0.status == .finalized && !$0.events.isEmpty }
        .sorted { mealPlanStartDate($0) < mealPlanStartDate($1) }
    }
    return allMealPlans.filter { mealPlan in
      mealPlan.status == .finalized
        && !mealPlan.events.isEmpty
        && mealPlan.id != active.id
        && mealPlanStartDate(mealPlan) >= mealPlanEndDate(active)
    }
    .sorted { mealPlanStartDate($0) < mealPlanStartDate($1) }
  }

  /// Meal plans that are not finalized and start after the active meal plan's end date.
  /// Only meaningful when there are such plans; used to conditionally show "Upcoming Meal Plans".
  var upcomingMealPlans: [Mealplanning_MealPlan] {
    let now = Date()
    let activeEnd = activeMealPlan.map { mealPlanEndDate($0) } ?? now
    return allMealPlans.filter { mealPlan in
      guard mealPlan.status != .finalized else { return false }
      guard !mealPlan.events.isEmpty else { return false }
      let planStart = mealPlanStartDate(mealPlan)
      return planStart > activeEnd
    }
  }

  private func mealPlanStartDate(_ mealPlan: Mealplanning_MealPlan) -> Date {
    mealPlan.events.map { timestampToDate($0.startsAt) }.min() ?? Date.distantPast
  }

  private func mealPlanEndDate(_ mealPlan: Mealplanning_MealPlan) -> Date {
    mealPlan.events.map { timestampToDate($0.endsAt) }.max() ?? Date.distantFuture
  }

  /// Calendar days with finalized (accepted) meal plan events. Shown in red.
  var acceptedOccupiedDates: Set<Date> {
    let cal = Calendar.current
    var occupied = Set<Date>()
    for plan in allMealPlans where plan.status == .finalized {
      for event in plan.events {
        let date = Self.timestampToDate(event.startsAt)
        occupied.insert(cal.startOfDay(for: date))
      }
    }
    return occupied
  }

  /// Calendar days with awaiting-votes (proposed) meal plan events. Shown in yellow.
  /// Excludes dates already in acceptedOccupiedDates.
  var proposedOccupiedDates: Set<Date> {
    let cal = Calendar.current
    var occupied = Set<Date>()
    let accepted = acceptedOccupiedDates
    for plan in allMealPlans where plan.status == .awaitingVotes {
      for event in plan.events {
        let date = cal.startOfDay(for: Self.timestampToDate(event.startsAt))
        if !accepted.contains(date) { occupied.insert(date) }
      }
    }
    return occupied
  }

  var activeTaskLists: [(mealPlanID: String, tasks: [Mealplanning_MealPlanTask])] {
    return tasksByMealPlan.compactMap { mealPlanID, tasks in
      let mealPlan = allMealPlans.first { $0.id == mealPlanID }
      guard let mealPlan = mealPlan,
        mealPlan.status == .finalized,
        mealPlan.tasksCreated
      else {
        return nil
      }

      let unfinishedTasks = tasks.filter { $0.status != .finished }
      return unfinishedTasks.isEmpty ? nil : (mealPlanID, tasks)
    }
  }

  var mealPlansWithTasks: [(mealPlanID: String, tasks: [Mealplanning_MealPlanTask])] {
    return tasksByMealPlan.compactMap { mealPlanID, tasks in
      guard let mealPlan = allMealPlans.first(where: { $0.id == mealPlanID }),
        mealPlan.status == .finalized,
        mealPlan.tasksCreated,
        !tasks.isEmpty
      else { return nil }
      return (mealPlanID, tasks)
    }
  }

  var mealPlansWithGroceryLists:
    [(mealPlanID: String, items: [Mealplanning_MealPlanGroceryListItem])]
  {
    return groceryLists.compactMap { mealPlanID, items in
      guard let mealPlan = allMealPlans.first(where: { $0.id == mealPlanID }),
        mealPlan.status == .finalized,
        mealPlan.groceryListInitialized,
        !items.isEmpty
      else { return nil }
      return (mealPlanID, items)
    }
  }

  var readyNowTaskCount: Int {
    let now = Date()
    return tasksByMealPlan.reduce(0) { total, entry in
      let (mealPlanID, tasks) = entry
      guard let mealPlan = allMealPlans.first(where: { $0.id == mealPlanID }) else { return total }
      return total
        + tasks.filter { task in
          guard task.status == .unfinished else { return false }
          guard task.hasRecipePrepTask,
            task.recipePrepTask.hasMaxTimeBufferBeforeRecipeInSeconds,
            task.hasMealPlanOption
          else { return true }
          let eventID = task.mealPlanOption.belongsToMealPlanEvent
          guard !eventID.isEmpty,
            let event = mealPlan.events.first(where: { $0.id == eventID })
          else { return true }
          let eventTime = Self.timestampToDate(event.startsAt)
          let startDate = eventTime.addingTimeInterval(
            -Double(task.recipePrepTask.maxTimeBufferBeforeRecipeInSeconds))
          return now >= startDate
        }.count
    }
  }

  var neededIngredientCount: Int {
    groceryLists.values.reduce(0) { total, items in
      total + items.filter { $0.status == .needs || $0.status == .unknown }.count
    }
  }

  var activeGroceryLists: [(mealPlanID: String, items: [Mealplanning_MealPlanGroceryListItem])] {
    return groceryLists.compactMap { mealPlanID, items in
      // Only show grocery lists for finalized meal plans with unacquired items
      let mealPlan = allMealPlans.first { $0.id == mealPlanID }
      guard let mealPlan = mealPlan,
        mealPlan.status == .finalized,
        mealPlan.groceryListInitialized
      else {
        return nil
      }

      // Filter to show only items that aren't fully acquired
      let activeItems = items.filter { item in
        item.status != .acquired
      }

      return activeItems.isEmpty ? nil : (mealPlanID, activeItems)
    }
  }

  let authManager: AuthenticationManager

  init(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  func loadData() async {
    isLoading = true
    errorMessage = nil
    errorTitle = "Error"
    errorIcon = "exclamationmark.triangle"
    errorIconColor = DSTheme.Colors.warning
    isServerDownError = false

    do {
      // Fetch current user for welcome message
      if let user = try? await fetchCurrentUser() {
        self.currentUser = user
      }

      // First fetch meal plans
      let mealPlans = try await fetchMealPlans()
      self.allMealPlans = mealPlans

      // Then fetch tasks (which needs meal plans to be loaded)
      let tasks = try await fetchUserTasks()
      self.userTasks = tasks

      // Fetch grocery lists for finalized meal plans
      await fetchGroceryLists(
        for: mealPlans.filter { $0.status == .finalized && $0.groceryListInitialized }
      )
    } catch {
      let display = ErrorDisplayFormatter.format(error, context: "load data")
      errorMessage = display.message
      errorTitle = display.title
      errorIcon = display.icon
      errorIconColor = display.iconColor
      isServerDownError = ErrorDisplayFormatter.isServerDown(error)
    }

    isLoading = false
  }

  private func fetchCurrentUser() async throws -> Identity_User {
    try await CurrentUserService.shared.currentUser(using: authManager)
  }

  private func fetchMealPlans() async throws -> [Mealplanning_MealPlan] {
    let response = try await authManager.authenticatedCall(
      "getMealPlansForAccount", idempotent: true
    ) { client, metadata, options in
      try await client.mealPlanning.getMealPlansForAccount(
        Mealplanning_GetMealPlansForAccountRequest(), metadata: metadata, options: options)
    }

    return response.results
  }

  private func fetchUserTasks() async throws -> [Mealplanning_MealPlanTask] {
    // We need to get tasks for all meal plans and filter by user
    // For now, let's get tasks from all finalized meal plans
    let finalizedMealPlans = allMealPlans.filter {
      $0.status == .finalized && $0.tasksCreated
    }

    var allUserTasks: [Mealplanning_MealPlanTask] = []
    var grouped: [String: [Mealplanning_MealPlanTask]] = [:]

    for mealPlan in finalizedMealPlans {
      var request = Mealplanning_GetMealPlanTasksRequest()
      request.mealPlanID = mealPlan.id

      do {
        let response = try await authManager.authenticatedCall(
          "getMealPlanTasks", idempotent: true
        ) { client, metadata, options in
          try await client.mealPlanning.getMealPlanTasks(
            request, metadata: metadata, options: options)
        }

        grouped[mealPlan.id] = response.results

        let userTasks = response.results.filter { task in
          task.assignedToUser == authManager.userID
        }
        allUserTasks.append(contentsOf: userTasks)
      } catch {
        // Non-fatal: skip this meal plan's tasks (session errors handled by authenticatedCall)
      }
    }

    self.tasksByMealPlan = grouped
    return allUserTasks
  }

  private func fetchGroceryLists(for mealPlans: [Mealplanning_MealPlan]) async {
    for mealPlan in mealPlans {
      var request = Mealplanning_GetMealPlanGroceryListItemsForMealPlanRequest()
      request.mealPlanID = mealPlan.id

      do {
        let response = try await authManager.authenticatedCall(
          "getMealPlanGroceryListItemsForMealPlan", idempotent: true
        ) { client, metadata, options in
          try await client.mealPlanning.getMealPlanGroceryListItemsForMealPlan(
            request, metadata: metadata, options: options)
        }

        await MainActor.run {
          groceryLists[mealPlan.id] = response.results
        }
      } catch {
        // Non-fatal: skip this grocery list (session errors handled by authenticatedCall)
      }
    }
  }

  // Instance method that delegates to static method for consistency
  func timestampToDate(_ timestamp: Timestamp) -> Date {
    return Self.timestampToDate(timestamp)
  }

  // Check if user has voted on a meal plan
  func hasUserVoted(on mealPlan: Mealplanning_MealPlan) -> Bool {
    for event in mealPlan.events {
      for option in event.options
      where option.votes.contains(where: { $0.byUser == authManager.userID && !$0.abstain }) {
        return true
      }
    }
    return false
  }

  // Format time until deadline
  func timeUntilDeadline(_ deadline: SwiftProtobuf.Google_Protobuf_Timestamp) -> String {
    let deadlineDate = timestampToDate(deadline)
    let now = Date()

    if deadlineDate <= now {
      return "Deadline passed"
    }

    let components = Calendar.current.dateComponents(
      [.day, .hour, .minute], from: now, to: deadlineDate)

    if let days = components.day, days > 0 {
      return "\(days) day\(days == 1 ? "" : "s") left"
    } else if let hours = components.hour, hours > 0 {
      return "\(hours) hour\(hours == 1 ? "" : "s") left"
    } else if let minutes = components.minute, minutes > 0 {
      return "\(minutes) minute\(minutes == 1 ? "" : "s") left"
    } else {
      return "Less than a minute left"
    }
  }
}
