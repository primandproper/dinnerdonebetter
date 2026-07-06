//
//  MealPlanDetailView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftProtobuf
import SwiftUI

struct MealPlanDetailView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Environment(UserSettingsService.self) private var userSettingsService
  @Environment(\.dismiss) private var dismiss
  @State var mealPlan: Mealplanning_MealPlan
  let groceryListItems: [Mealplanning_MealPlanGroceryListItem]?
  @State private var taskCount: Int?
  @State private var groceryListItemCount: Int?
  @State private var showArchiveConfirmation = false
  @State private var isArchiving = false
  @State private var archiveError: String?
  @State private var moveSwapSheetEvent: IdentifiableMealPlanEvent?
  @State private var cancelEventConfirmationEvent: Mealplanning_MealPlanEvent?
  @State private var isCancellingEvent = false
  @State private var cancelEventError: String?

  init(mealPlan: Mealplanning_MealPlan, groceryListItems: [Mealplanning_MealPlanGroceryListItem]?) {
    _mealPlan = State(initialValue: mealPlan)
    self.groceryListItems = groceryListItems
  }

  var body: some View {
    ScrollView {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.xl) {
        // Header with title, status, and range inline
        headerSection

        // Voting deadline (if not finalized) - shown in a card if needed
        if mealPlan.status == .awaitingVotes {
          votingDeadlineCard
        }

        // Events
        eventsSection

        // Tasks Link (if finalized and tasks created) - matches home page order (prep tasks above grocery)
        if mealPlan.status == .finalized && mealPlan.tasksCreated {
          tasksSection
        }

        // Grocery List Link (if finalized and has items)
        if mealPlan.status == .finalized {
          groceryListSection
        }
      }
      .dsScreenPadding()
    }
    .navigationBarTitleDisplayMode(.inline)
    .toolbar {
      ToolbarItem(placement: .primaryAction) {
        Button {
          showArchiveConfirmation = true
        } label: {
          Label("Archive", systemImage: "archivebox")
        }
      }
    }
    .alert("Archive Meal Plan", isPresented: $showArchiveConfirmation) {
      Button("Archive", role: .destructive) {
        eventReporterService.reporter.track(event: "meal_plan_archived", properties: [:])
        Task { await archiveMealPlan() }
      }
      Button("Cancel", role: .cancel) {}
    } message: {
      Text(
        "This will remove the meal plan from your active plans. You can still access archived plans later."
      )
    }
    .alert("Archive Failed", isPresented: .constant(archiveError != nil)) {
      Button("OK") { archiveError = nil }
    } message: {
      if let error = archiveError {
        Text(error)
      }
    }
    .disabled(isArchiving || isCancellingEvent)
    .onAppear {
      eventReporterService.reporter.track(
        event: "meal_plan_detail_viewed",
        properties: ["meal_plan_id": mealPlan.id, "status": mealPlan.status.rawValue])
    }
    .task {
      await loadTaskCount()
      await loadGroceryListItemCount()
    }
    .sheet(item: $moveSwapSheetEvent) { identifiable in
      MoveSwapEventSheet(
        mealPlan: mealPlan,
        event: identifiable.event,
        onSuccess: { await refetchMealPlan() },
        startInMoveMode: identifiable.startInMoveMode,
        startInSwapMode: identifiable.startInSwapMode
      )
      .environment(authManager)
      .environment(eventReporterService)
    }
    .alert(
      "Cancel Meal",
      isPresented: Binding(
        get: { cancelEventConfirmationEvent != nil },
        set: { if !$0 { cancelEventConfirmationEvent = nil } }
      )
    ) {
      Button("Cancel", role: .cancel) {}
      Button("Remove from Plan", role: .destructive) {
        if let evt = cancelEventConfirmationEvent {
          eventReporterService.reporter.track(
            event: "meal_plan_event_removed",
            properties: ["event_id": evt.id])
          cancelEventConfirmationEvent = nil
          Task { await cancelMealPlanEvent(evt) }
        }
      }
    } message: {
      Text(
        "This will remove the meal from your plan. Any prep tasks for this meal will be cancelled.")
    }
    .alert("Cancel Failed", isPresented: .constant(cancelEventError != nil)) {
      Button("OK") { cancelEventError = nil }
    } message: {
      if let err = cancelEventError {
        Text(err)
      }
    }
  }

  private func refetchMealPlan() async {
    do {
      var request = Mealplanning_GetMealPlanRequest()
      request.mealPlanID = mealPlan.id

      let response = try await authManager.authenticatedCall("getMealPlan", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMealPlan(request, metadata: metadata, options: options)
      }

      mealPlan = response.result
    } catch {
      // Non-fatal: keep the current meal plan on refresh failure
    }
  }

  private func archiveMealPlan() async {
    isArchiving = true
    archiveError = nil

    do {
      var request = Mealplanning_ArchiveMealPlanRequest()
      request.mealPlanID = mealPlan.id

      _ = try await authManager.authenticatedCall("archiveMealPlan") {
        client, metadata, options in
        try await client.mealPlanning.archiveMealPlan(request, metadata: metadata, options: options)
      }

      NotificationCenter.default.post(name: .mealPlanArchived, object: nil)
      dismiss()
    } catch {
      archiveError = error.localizedDescription
    }

    isArchiving = false
  }

  private func cancelMealPlanEvent(_ event: Mealplanning_MealPlanEvent) async {
    isCancellingEvent = true
    cancelEventError = nil

    do {
      var request = Mealplanning_ArchiveMealPlanEventRequest()
      request.mealPlanID = mealPlan.id
      request.mealPlanEventID = event.id

      _ = try await authManager.authenticatedCall("archiveMealPlanEvent") {
        client, metadata, options in
        try await client.mealPlanning.archiveMealPlanEvent(
          request, metadata: metadata, options: options)
      }

      NotificationCenter.default.post(name: .mealPlanEventsUpdated, object: nil)
      await refetchMealPlan()
    } catch {
      cancelEventError = error.localizedDescription
    }

    isCancellingEvent = false
  }

  private func loadTaskCount() async {
    guard mealPlan.status == .finalized && mealPlan.tasksCreated else {
      return
    }

    do {
      var request = Mealplanning_GetMealPlanTasksRequest()
      request.mealPlanID = mealPlan.id

      let response = try await authManager.authenticatedCall("getMealPlanTasks", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMealPlanTasks(
          request, metadata: metadata, options: options)
      }

      // Deduplicate by ID and get count
      var seenIDs: Set<String> = []
      let uniqueCount = response.results.filter { task in
        if seenIDs.contains(task.id) {
          return false
        }
        seenIDs.insert(task.id)
        return true
      }.count

      taskCount = uniqueCount
    } catch {
      taskCount = 0
    }
  }

  private func loadGroceryListItemCount() async {
    guard mealPlan.status == .finalized else {
      return
    }

    do {
      var request = Mealplanning_GetMealPlanGroceryListItemsForMealPlanRequest()
      request.mealPlanID = mealPlan.id

      let response = try await authManager.authenticatedCall(
        "getMealPlanGroceryListItemsForMealPlan", idempotent: true
      ) { client, metadata, options in
        try await client.mealPlanning.getMealPlanGroceryListItemsForMealPlan(
          request, metadata: metadata, options: options)
      }

      groceryListItemCount = response.results.count
    } catch {
      groceryListItemCount = 0
    }
  }

  private var isPast: Bool {
    MealPlanningHomeHelpers.isMealPlanInPast(mealPlan)
  }

  private var headerSection: some View {
    VStack(alignment: .leading, spacing: 8) {
      // Title with status badge and range inline
      HStack(alignment: .firstTextBaseline, spacing: 8) {
        Text(mealPlan.notes.isEmpty ? "Meal Plan" : mealPlan.notes)
          .font(.title2)
          .fontWeight(.bold)
          .strikethrough(isPast)

        statusBadge

        Spacer()
      }

      // Time range on second line
      Text(MealPlanningHomeHelpers.formatMealPlanTimeRange(mealPlan))
        .font(.subheadline)
        .foregroundColor(.secondary)
        .strikethrough(isPast)
    }
  }

  private var votingDeadlineCard: some View {
    votingDeadlineView
      .padding()
      .background(Color(.systemGray6))
      .cornerRadius(10)
  }

  private var statusBadge: some View {
    let (text, color) = statusInfo(mealPlan.status)
    return Text(text)
      .font(.caption)
      .fontWeight(.semibold)
      .strikethrough(isPast)
      .padding(.horizontal, 8)
      .padding(.vertical, 4)
      .background(color.opacity(0.2))
      .foregroundColor(color)
      .cornerRadius(6)
  }

  private var votingDeadlineView: some View {
    let deadline = HomeViewModel.timestampToDate(mealPlan.votingDeadline)
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short

    return HStack {
      Image(systemName: "clock")
        .foregroundColor(.secondary)
      Text("Voting deadline: \(formatter.string(from: deadline))")
        .font(.subheadline)
        .foregroundColor(.secondary)
        .strikethrough(isPast)
    }
  }

  private func statusInfo(_ status: Mealplanning_MealPlanStatus) -> (String, Color) {
    switch status {
    case .awaitingVotes:
      return ("Awaiting Votes", .orange)
    case .finalized:
      return ("Finalized", .green)
    default:
      return ("Unknown", .gray)
    }
  }

  private var upcomingEvents: [Mealplanning_MealPlanEvent] {
    let now = Date()
    return mealPlan.events.filter { HomeViewModel.timestampToDate($0.startsAt) >= now }
  }

  private var pastEvents: [Mealplanning_MealPlanEvent] {
    let now = Date()
    return mealPlan.events.filter { HomeViewModel.timestampToDate($0.startsAt) < now }
  }

  private var eventsSection: some View {
    VStack(alignment: .leading, spacing: 12) {
      if !upcomingEvents.isEmpty {
        Text("Upcoming")
          .font(.title2)
          .fontWeight(.bold)
          .strikethrough(isPast)

        ForEach(upcomingEvents, id: \.id) { event in
          EventCard(
            event: event,
            mealPlan: mealPlan,
            canEdit: mealPlan.status == .finalized,
            onMove: {
              eventReporterService.reporter.track(
                event: "meal_plan_move_swap_tapped", properties: ["mode": "move"])
              moveSwapSheetEvent = IdentifiableMealPlanEvent(event: event, startInMoveMode: true)
            },
            onSwap: {
              eventReporterService.reporter.track(
                event: "meal_plan_move_swap_tapped", properties: ["mode": "swap"])
              moveSwapSheetEvent = IdentifiableMealPlanEvent(event: event, startInSwapMode: true)
            },
            onCancel: { cancelEventConfirmationEvent = event }
          )
        }
      }

      if !pastEvents.isEmpty {
        Text("Past")
          .font(.title2)
          .fontWeight(.bold)
          .foregroundColor(.secondary)
          .strikethrough(isPast)

        ForEach(pastEvents, id: \.id) { event in
          EventCard(
            event: event,
            mealPlan: mealPlan,
            canEdit: mealPlan.status == .finalized,
            onMove: {
              eventReporterService.reporter.track(
                event: "meal_plan_move_swap_tapped", properties: ["mode": "move"])
              moveSwapSheetEvent = IdentifiableMealPlanEvent(event: event, startInMoveMode: true)
            },
            onSwap: {
              eventReporterService.reporter.track(
                event: "meal_plan_move_swap_tapped", properties: ["mode": "swap"])
              moveSwapSheetEvent = IdentifiableMealPlanEvent(event: event, startInSwapMode: true)
            },
            onCancel: { cancelEventConfirmationEvent = event }
          )
          .opacity(0.7)
        }
      }
    }
  }

  private var groceryListSection: some View {
    let count = groceryListItemCount ?? 0
    let hasItems = count > 0

    return Group {
      if hasItems {
        NavigationLink(
          destination: GroceryListView(
            mealPlan: mealPlan,
            items: [],  // Always start with empty array, GroceryListView will fetch fresh data
            authManager: authManager
          )
        ) {
          groceryListCardContent(count: count)
        }
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "meal_plan_grocery_tapped",
              properties: ["meal_plan_id": mealPlan.id])
          }
        )
        .buttonStyle(.plain)
      } else {
        groceryListCardContent(count: count)
          .opacity(0.6)
      }
    }
  }

  private func groceryListCardContent(count: Int) -> some View {
    HStack {
      Image(systemName: "cart")
        .foregroundColor(.blue)
      Text("View Grocery List\(count > 0 ? " (\(count))" : "")")
        .font(.headline)
        .strikethrough(isPast)
      Spacer()
      if count > 0 {
        Image(systemName: "chevron.right")
          .foregroundColor(.secondary)
          .font(.caption)
      }
    }
    .padding()
    .background(Color(.systemGray6))
    .cornerRadius(10)
  }

  private var tasksSection: some View {
    let count = taskCount ?? 0
    let hasTasks = count > 0

    return Group {
      if hasTasks {
        NavigationLink(
          destination: TaskListView(
            mealPlan: mealPlan,
            tasks: [],  // Always start with empty array, TaskListView will fetch fresh data
            authManager: authManager,
            userSettingsService: userSettingsService
          )
        ) {
          tasksCardContent(count: count)
        }
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "meal_plan_tasks_tapped",
              properties: ["meal_plan_id": mealPlan.id])
          }
        )
        .buttonStyle(.plain)
      } else {
        tasksCardContent(count: count)
          .opacity(0.6)
      }
    }
  }

  private func tasksCardContent(count: Int) -> some View {
    HStack {
      Image(systemName: "checklist")
        .foregroundColor(.blue)
      Text("Prep Tasks (\(count))")
        .font(.headline)
        .strikethrough(isPast)
      Spacer()
      if count > 0 {
        Image(systemName: "chevron.right")
          .foregroundColor(.secondary)
          .font(.caption)
      }
    }
    .padding()
    .background(Color(.systemGray6))
    .cornerRadius(10)
  }
}

// MARK: - Identifiable Meal Plan Event

private struct IdentifiableMealPlanEvent: Identifiable {
  let event: Mealplanning_MealPlanEvent
  var startInMoveMode = false
  var startInSwapMode = false
  var id: String { event.id }
}

// MARK: - Event Card

struct EventCard: View {
  @Environment(EventReporterService.self) private var eventReporterService
  let event: Mealplanning_MealPlanEvent
  var mealPlan: Mealplanning_MealPlan?
  var canEdit: Bool = false
  var onMove: (() -> Void)?
  var onSwap: (() -> Void)?
  var onCancel: (() -> Void)?

  /// True when this specific event is in the past (startsAt < now).
  private var isPast: Bool {
    HomeViewModel.timestampToDate(event.startsAt) < Date()
  }

  private var canSwap: Bool {
    guard let plan = mealPlan else { return false }
    return plan.events.filter { $0.id != event.id }.count >= 1
  }

  private func eventTimeRange(event: Mealplanning_MealPlanEvent) -> some View {
    let startDate = HomeViewModel.timestampToDate(event.startsAt)
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short

    return Text(formatter.string(from: startDate))
      .font(.caption)
      .foregroundColor(.secondary)
      .strikethrough(isPast)
  }

  var body: some View {
    VStack(alignment: .leading, spacing: 12) {
      // Event header
      HStack {
        VStack(alignment: .leading, spacing: 4) {
          Text(MealPlanningUtils.formatMealName(event.mealName))
            .font(.headline)
            .strikethrough(isPast)

          eventTimeRange(event: event)
        }

        Spacer()

        if canEdit {
          Menu {
            if let onMove = onMove {
              Button {
                onMove()
              } label: {
                Label("Move to another day", systemImage: "calendar")
              }
            }
            if canSwap, let onSwap = onSwap {
              Button {
                onSwap()
              } label: {
                Label("Swap with another event", systemImage: "arrow.triangle.2.circlepath")
              }
            }
            if let onCancel = onCancel {
              Button(role: .destructive) {
                onCancel()
              } label: {
                Label("Cancel event", systemImage: "xmark.circle")
              }
            }
          } label: {
            Image(systemName: "ellipsis.circle")
              .font(.body)
              .foregroundColor(.secondary)
          }
        }
      }

      // Selected meals
      if !event.options.isEmpty {
        VStack(alignment: .leading, spacing: 8) {
          ForEach(event.options.filter { $0.chosen }, id: \.id) { option in
            NavigationLink(
              destination: MealDetailView(
                mealID: option.meal.id,
                isFromMealPlan: true,
                mealPlanScale: option.mealScale > 0 ? option.mealScale : 1.0,
                mealPlanID: mealPlan?.id,
                mealPlanOptionID: option.id
              )
            ) {
              MealOptionCard(option: option, isChosen: true, strikethrough: isPast)
            }
            .buttonStyle(.plain)
            .simultaneousGesture(
              TapGesture().onEnded {
                eventReporterService.reporter.track(
                  event: "meal_plan_event_card_tapped",
                  properties: ["event_id": event.id, "meal_id": option.meal.id])
              })
          }
        }
      } else {
        Text("No meals selected")
          .font(.subheadline)
          .foregroundColor(.secondary)
          .strikethrough(isPast)
      }
    }
    .padding()
    .background(Color(.systemGray6))
    .cornerRadius(10)
  }
}

// MARK: - Meal Option Card

struct MealOptionCard: View {
  let option: Mealplanning_MealPlanOption
  var isChosen: Bool = true
  var strikethrough: Bool = false

  var body: some View {
    HStack {
      if isChosen {
        Image(systemName: "checkmark.circle.fill")
          .foregroundColor(.green)
          .font(.caption)
      }

      VStack(alignment: .leading, spacing: 4) {
        HStack(spacing: 6) {
          Text(option.meal.name.isEmpty ? "Unnamed Meal" : option.meal.name)
            .font(.subheadline)
            .fontWeight(isChosen ? .semibold : .regular)
            .strikethrough(strikethrough)

          // Tiebroken indicator
          if isChosen && option.tieBroken {
            HStack(spacing: 2) {
              Image(systemName: "dice")
                .font(.caption2)
              Text("Tiebroken")
                .font(.caption2)
            }
            .foregroundColor(.orange)
            .padding(.horizontal, 4)
            .padding(.vertical, 2)
            .background(Color.orange.opacity(0.2))
            .cornerRadius(4)
          }
        }

        if !option.meal.components.isEmpty {
          let recipeNames = option.meal.components.compactMap { component -> String? in
            component.recipe.name.isEmpty ? nil : component.recipe.name
          }
          if !recipeNames.isEmpty {
            Text(recipeNames.joined(separator: ", "))
              .font(.caption)
              .foregroundColor(.secondary)
              .lineLimit(2)
              .strikethrough(strikethrough)
          }
        }

        if option.mealScale != 1.0 {
          Text("Scale: \(String(format: "%.1f", option.mealScale))x")
            .font(.caption2)
            .foregroundColor(.secondary)
            .strikethrough(strikethrough)
        }
      }

      Spacer()

      // Show chevron only for chosen meals (indicating they're clickable)
      if isChosen {
        Image(systemName: "chevron.right")
          .foregroundColor(.secondary)
          .font(.caption)
      }
    }
    .padding(.vertical, 4)
    .opacity(isChosen ? 1.0 : 0.6)
    .contentShape(Rectangle())
  }
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "John Doe"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  // Create a sample meal plan
  var mealPlan = Mealplanning_MealPlan()
  mealPlan.id = "mealplan123"
  mealPlan.notes = "Sample Meal Plan"
  mealPlan.status = .finalized
  mealPlan.groceryListInitialized = true
  mealPlan.tasksCreated = true

  var event = Mealplanning_MealPlanEvent()
  event.id = "event123"
  event.mealName = .dinner
  var startTime = Google_Protobuf_Timestamp()
  startTime.seconds = Int64(Date().timeIntervalSince1970)
  event.startsAt = startTime
  var endTime = Google_Protobuf_Timestamp()
  endTime.seconds = Int64(Date().timeIntervalSince1970 + 7200)
  event.endsAt = endTime

  var option = Mealplanning_MealPlanOption()
  option.id = "option123"
  option.chosen = true
  var meal = Mealplanning_Meal()
  meal.name = "Chicken Dinner"
  option.meal = meal
  option.mealScale = 1.0
  event.options = [option]

  mealPlan.events = [event]

  return NavigationView {
    MealPlanDetailView(mealPlan: mealPlan, groceryListItems: nil)
  }
  .environment(authManager)
}
