//
//  MealPlanningHomeContent.swift
//  ios
//
//  Meal planning content for the home screen.
//

import SwiftProtobuf
import SwiftUI

struct MealPlanningHomeContent: View {
  @Environment(UserSettingsService.self) private var userSettingsService
  @Environment(EventReporterService.self) private var eventReporterService
  var viewModel: HomeViewModel?

  var body: some View {
    Group {
      if let viewModel = viewModel {
        ScrollView {
          VStack(spacing: 0) {
            // Meal Plans section (always first, with plus to create)
            mealPlansSection(viewModel: viewModel)
              .dsScreenPadding()
              .padding(.bottom, DSTheme.Spacing.lg)

            // Pending Votes
            if !viewModel.pendingVoteMealPlans.isEmpty {
              pendingVotesSection(viewModel: viewModel)
                .dsScreenPadding()
                .padding(.bottom, DSTheme.Spacing.lg)
            }

            // Upcoming Meal Plans (non-finalized)
            if !viewModel.upcomingMealPlans.isEmpty {
              upcomingMealPlansSection(viewModel: viewModel)
                .dsScreenPadding()
                .padding(.bottom, DSTheme.Spacing.lg)
            }

            // Future Meal Plans
            if !viewModel.futureFinalizedMealPlans.isEmpty {
              softSeparator

              futureMealPlansSection(viewModel: viewModel)
                .dsScreenPadding()
                .padding(.bottom, DSTheme.Spacing.xl)
            }
          }
        }
        .frame(maxHeight: .infinity)
      }
    }
  }

  // MARK: - Meal Plans Section
  private func mealPlansSection(viewModel: HomeViewModel) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      // Header with plus icon
      HStack {
        Label("Active Meal Plans", systemImage: "star.circle.fill")
          .font(DSTheme.Typography.title2)
          .foregroundColor(DSTheme.Colors.textPrimary)

        Spacer()

        NavigationLink(
          destination: CreateMealPlanWizardView(
            acceptedOccupiedDates: viewModel.acceptedOccupiedDates,
            proposedOccupiedDates: viewModel.proposedOccupiedDates
          )
        ) {
          Image(systemName: "plus.circle.fill")
            .font(.system(size: 24))
            .foregroundColor(DSTheme.Colors.primary)
        }
        .buttonStyle(.plain)
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "home_create_meal_plan_plus_tapped", properties: [:])
          })
      }

      if let activeMealPlan = viewModel.activeMealPlan {
        NavigationLink(
          destination: LoadedMealPlanView(mealPlanID: activeMealPlan.id) { plan in
            MealPlanDetailView(mealPlan: plan, groceryListItems: nil)
          }
        ) {
          UpcomingMealCardContent(mealPlan: activeMealPlan)
        }
        .buttonStyle(.plain)
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "home_meal_plan_card_tapped",
              properties: ["meal_plan_id": activeMealPlan.id])
          })

        activePlanTaskAndGrocery(viewModel: viewModel, mealPlan: activeMealPlan)
      } else {
        // Empty state: prominent Create Meal Plan CTA
        createMealPlanCTA(viewModel: viewModel)
      }
    }
  }

  private func createMealPlanCTA(viewModel: HomeViewModel) -> some View {
    NavigationLink(
      destination: CreateMealPlanWizardView(
        acceptedOccupiedDates: viewModel.acceptedOccupiedDates,
        proposedOccupiedDates: viewModel.proposedOccupiedDates
      )
    ) {
      VStack(spacing: DSTheme.Spacing.lg) {
        ZStack {
          Circle()
            .fill(DSTheme.Colors.primary.opacity(0.15))
            .frame(width: 64, height: 64)

          Image(systemName: "plus")
            .font(.system(size: 28, weight: .semibold))
            .foregroundColor(DSTheme.Colors.primary)
        }

        VStack(spacing: DSTheme.Spacing.xxs) {
          Text("Create Meal Plan")
            .font(DSTheme.Typography.label)
            .foregroundColor(DSTheme.Colors.textPrimary)

          Text("Plan your meals for the week")
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.textSecondary)
        }

        Image(systemName: "chevron.right")
          .font(.system(size: 14, weight: .semibold))
          .foregroundColor(DSTheme.Colors.textTertiary)
      }
      .frame(maxWidth: .infinity)
      .padding(DSTheme.Spacing.xl)
      .background(
        RoundedRectangle(cornerRadius: DSTheme.Radius.lg)
          .fill(DSTheme.Colors.cardBackground)
          .shadow(color: Color.black.opacity(0.06), radius: 8, x: 0, y: 2)
      )
      .overlay(
        RoundedRectangle(cornerRadius: DSTheme.Radius.lg)
          .stroke(DSTheme.Colors.primary.opacity(0.2), lineWidth: 1)
      )
    }
    .buttonStyle(.plain)
    .simultaneousGesture(
      TapGesture().onEnded {
        eventReporterService.reporter.track(
          event: "home_create_meal_plan_cta_tapped", properties: [:])
      })
  }

  // MARK: - Pending Votes Section
  private func pendingVotesSection(viewModel: HomeViewModel) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      HStack {
        Label("Pending Votes", systemImage: "hand.raised.fill")
          .font(DSTheme.Typography.title2)
          .foregroundColor(DSTheme.Colors.textPrimary)

        Spacer()

        Text("\(viewModel.pendingVoteMealPlans.count)")
          .font(DSTheme.Typography.caption)
          .fontWeight(.semibold)
          .foregroundColor(DSTheme.Colors.warning)
          .padding(.horizontal, DSTheme.Spacing.sm)
          .padding(.vertical, DSTheme.Spacing.xxs)
          .background(DSTheme.Colors.warning.opacity(0.15))
          .cornerRadius(DSTheme.Radius.full)
      }

      ForEach(viewModel.pendingVoteMealPlans, id: \.id) { mealPlan in
        NavigationLink(
          destination: LoadedMealPlanView(mealPlanID: mealPlan.id) { plan in
            VoteMealPlanView(mealPlan: plan)
          }
        ) {
          PendingVoteCardContent(
            mealPlan: mealPlan,
            hasVoted: viewModel.hasUserVoted(on: mealPlan),
            timeUntilDeadline: viewModel.timeUntilDeadline(mealPlan.votingDeadline)
          )
        }
        .buttonStyle(.plain)
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "home_pending_vote_tapped",
              properties: ["meal_plan_id": mealPlan.id])
          })
      }
    }
  }

  private func activePlanTaskAndGrocery(
    viewModel: HomeViewModel, mealPlan: any MealPlanDisplayable
  ) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
      if let taskEntry = viewModel.mealPlansWithTasks.first(where: { $0.mealPlanID == mealPlan.id })
      {
        let summary = taskSummary(
          tasks: taskEntry.tasks, mealPlan: mealPlan, includeDateForFuture: false)
        NavigationLink(
          destination: TaskListView(
            mealPlan: mealPlan,
            tasks: [],
            authManager: viewModel.authManager,
            userSettingsService: userSettingsService
          )
        ) {
          DSInfoRow(
            icon: "checklist",
            text: summary.text,
            color: summary.color,
            strikethrough: MealPlanningHomeHelpers.isMealPlanInPast(mealPlan)
          )
        }
        .buttonStyle(.plain)
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "home_task_summary_tapped",
              properties: ["meal_plan_id": mealPlan.id])
          })
      }

      if let groceryEntry = viewModel.mealPlansWithGroceryLists.first(where: {
        $0.mealPlanID == mealPlan.id
      }) {
        let neededCount = groceryEntry.items.filter { $0.status == .needs || $0.status == .unknown }
          .count
        NavigationLink(
          destination: GroceryListView(
            mealPlan: mealPlan,
            items: [],
            authManager: viewModel.authManager
          )
        ) {
          DSInfoRow(
            icon: "cart.fill",
            text: neededCount > 0
              ? "Grocery List (\(neededCount) ingredient\(neededCount == 1 ? "" : "s") needed)"
              : "All ingredients acquired",
            color: neededCount > 0 ? DSTheme.Colors.primary : DSTheme.Colors.success,
            strikethrough: MealPlanningHomeHelpers.isMealPlanInPast(mealPlan)
          )
        }
        .buttonStyle(.plain)
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "home_grocery_list_tapped",
              properties: ["meal_plan_id": mealPlan.id])
          })
      }
    }
  }

  private var softSeparator: some View {
    DSSoftDivider(verticalSpacing: DSTheme.Spacing.sm)
  }

  private func futureMealPlansSection(viewModel: HomeViewModel) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      DSSectionHeader(title: "Future Meal Plans", systemImage: "calendar.badge.clock")

      ForEach(viewModel.futureFinalizedMealPlans, id: \.id) { mealPlan in
        futureMealPlanBlock(viewModel: viewModel, mealPlan: mealPlan)
      }
    }
  }

  private func futureMealPlanBlock(
    viewModel: HomeViewModel, mealPlan: any MealPlanDisplayable
  ) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
      NavigationLink(
        destination: LoadedMealPlanView(mealPlanID: mealPlan.id) { plan in
          MealPlanDetailView(mealPlan: plan, groceryListItems: nil)
        }
      ) {
        UpcomingMealCardContent(mealPlan: mealPlan)
      }
      .buttonStyle(.plain)
      .simultaneousGesture(
        TapGesture().onEnded {
          eventReporterService.reporter.track(
            event: "home_meal_plan_card_tapped",
            properties: ["meal_plan_id": mealPlan.id])
        })

      VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
        if let taskEntry = viewModel.mealPlansWithTasks.first(where: {
          $0.mealPlanID == mealPlan.id
        }) {
          let summary = taskSummary(
            tasks: taskEntry.tasks, mealPlan: mealPlan, includeDateForFuture: true)
          NavigationLink(
            destination: TaskListView(
              mealPlan: mealPlan,
              tasks: [],
              authManager: viewModel.authManager,
              userSettingsService: userSettingsService
            )
          ) {
            DSInfoRow(
              icon: "checklist",
              text: summary.text,
              color: summary.color,
              strikethrough: MealPlanningHomeHelpers.isMealPlanInPast(mealPlan)
            )
          }
          .buttonStyle(.plain)
          .simultaneousGesture(
            TapGesture().onEnded {
              eventReporterService.reporter.track(
                event: "home_task_summary_tapped",
                properties: ["meal_plan_id": mealPlan.id])
            })
        }

        if let groceryEntry = viewModel.mealPlansWithGroceryLists.first(where: {
          $0.mealPlanID == mealPlan.id
        }) {
          let neededCount = groceryEntry.items.filter {
            $0.status == .needs || $0.status == .unknown
          }.count
          NavigationLink(
            destination: GroceryListView(
              mealPlan: mealPlan,
              items: [],
              authManager: viewModel.authManager
            )
          ) {
            DSInfoRow(
              icon: "cart.fill",
              text: neededCount > 0
                ? "Grocery List (\(neededCount) ingredient\(neededCount == 1 ? "" : "s") needed)"
                : "All ingredients acquired",
              color: neededCount > 0 ? DSTheme.Colors.primary : DSTheme.Colors.success,
              strikethrough: MealPlanningHomeHelpers.isMealPlanInPast(mealPlan)
            )
          }
          .buttonStyle(.plain)
          .simultaneousGesture(
            TapGesture().onEnded {
              eventReporterService.reporter.track(
                event: "home_grocery_list_tapped",
                properties: ["meal_plan_id": mealPlan.id])
            })
        }
      }
    }
  }

  // MARK: - Upcoming Meal Plans Section
  private func upcomingMealPlansSection(viewModel: HomeViewModel) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      DSSectionHeader(title: "Upcoming Meal Plans", systemImage: "calendar") {
        Text(
          "\(viewModel.upcomingMealPlans.count) plan\(viewModel.upcomingMealPlans.count == 1 ? "" : "s")"
        )
        .font(DSTheme.Typography.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
      }

      ForEach(viewModel.upcomingMealPlans, id: \.id) { mealPlan in
        NavigationLink(
          destination: LoadedMealPlanView(mealPlanID: mealPlan.id) { plan in
            MealPlanDetailView(mealPlan: plan, groceryListItems: nil)
          }
        ) {
          UpcomingMealCardContent(mealPlan: mealPlan)
        }
        .buttonStyle(.plain)
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(
              event: "home_meal_plan_card_tapped",
              properties: ["meal_plan_id": mealPlan.id])
          })
      }
    }
  }

  private func taskSummary(
    tasks: [Mealplanning_MealPlanTask], mealPlan: any MealPlanDisplayable,
    includeDateForFuture: Bool = false
  ) -> (text: String, color: Color) {
    let now = Date()
    let unfinished = tasks.filter { $0.status == .unfinished }

    guard !unfinished.isEmpty else {
      return ("All tasks done", DSTheme.Colors.success)
    }

    var readyCount = 0
    var earliestStart: Date?

    for task in unfinished {
      if let startDate = taskStartDate(task: task, mealPlan: mealPlan) {
        if now >= startDate {
          readyCount += 1
        } else if earliestStart.map({ startDate < $0 }) ?? true {
          earliestStart = startDate
        }
      } else {
        readyCount += 1
      }
    }

    if readyCount > 0 {
      return (
        "Prep Tasks (\(readyCount) ready)",
        DSTheme.Colors.warning
      )
    }

    if let earliest = earliestStart {
      let timeFormatter = DateFormatter()
      timeFormatter.dateStyle = .none
      timeFormatter.timeStyle = .short
      if includeDateForFuture {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "M/d/yy"
        return (
          "Next task on \(dateFormatter.string(from: earliest)) at \(timeFormatter.string(from: earliest))",
          DSTheme.Colors.textSecondary
        )
      } else {
        return (
          "Next task at \(timeFormatter.string(from: earliest))",
          DSTheme.Colors.textSecondary
        )
      }
    }

    return ("No tasks ready yet", DSTheme.Colors.textSecondary)
  }

  private func taskStartDate(
    task: Mealplanning_MealPlanTask, mealPlan: any MealPlanDisplayable
  ) -> Date? {
    guard task.hasRecipePrepTask,
      task.recipePrepTask.hasMaxTimeBufferBeforeRecipeInSeconds,
      task.hasMealPlanOption
    else { return nil }
    let eventID = task.mealPlanOption.belongsToMealPlanEvent
    guard !eventID.isEmpty,
      let event = mealPlan.displayEvents.first(where: { $0.id == eventID })
    else { return nil }
    let eventTime = HomeViewModel.timestampToDate(event.startsAt)
    return eventTime.addingTimeInterval(
      -Double(task.recipePrepTask.maxTimeBufferBeforeRecipeInSeconds))
  }

}

// MARK: - Pending Vote Card
struct PendingVoteCard: View {
  let mealPlan: any MealPlanDisplayable
  let hasVoted: Bool
  let timeUntilDeadline: String
  let onTap: () -> Void

  var body: some View {
    Button(action: onTap) {
      PendingVoteCardContent(
        mealPlan: mealPlan,
        hasVoted: hasVoted,
        timeUntilDeadline: timeUntilDeadline
      )
    }
    .buttonStyle(.plain)
  }
}

// MARK: - Pending Vote Card Content (reusable for NavigationLink)
struct PendingVoteCardContent: View {
  let mealPlan: any MealPlanDisplayable
  let hasVoted: Bool
  let timeUntilDeadline: String

  var body: some View {
    HStack(spacing: DSTheme.Spacing.md) {
      // Status indicator
      ZStack {
        Circle()
          .fill(
            hasVoted ? DSTheme.Colors.success.opacity(0.15) : DSTheme.Colors.warning.opacity(0.15)
          )
          .frame(width: 44, height: 44)

        Image(systemName: hasVoted ? "checkmark" : "hand.raised.fill")
          .font(.system(size: 18, weight: .medium))
          .foregroundColor(hasVoted ? DSTheme.Colors.success : DSTheme.Colors.warning)
      }

      VStack(alignment: .leading, spacing: DSTheme.Spacing.xxs) {
        Text(MealPlanningHomeHelpers.mealPlanDisplayTitle(mealPlan, fallback: "Meal Plan"))
          .font(DSTheme.Typography.label)
          .foregroundColor(DSTheme.Colors.textPrimary)

        Text(MealPlanningHomeHelpers.formatMealPlanTimeRange(mealPlan))
          .font(DSTheme.Typography.caption)
          .foregroundColor(DSTheme.Colors.textSecondary)

        HStack(spacing: DSTheme.Spacing.xs) {
          Image(systemName: "clock")
            .font(.system(size: 11))
          Text(timeUntilDeadline)
            .font(DSTheme.Typography.caption)
        }
        .foregroundColor(hasVoted ? DSTheme.Colors.textTertiary : DSTheme.Colors.warning)
      }

      Spacer()

      if !hasVoted {
        Text("Vote")
          .font(DSTheme.Typography.caption)
          .fontWeight(.semibold)
          .foregroundColor(.white)
          .padding(.horizontal, DSTheme.Spacing.md)
          .padding(.vertical, DSTheme.Spacing.sm)
          .background(DSTheme.Colors.primary)
          .cornerRadius(DSTheme.Radius.full)
      } else {
        Image(systemName: "chevron.right")
          .font(.system(size: 14, weight: .semibold))
          .foregroundColor(DSTheme.Colors.textTertiary)
      }
    }
    .padding(DSTheme.Spacing.md)
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.lg)
    .overlay(
      RoundedRectangle(cornerRadius: DSTheme.Radius.lg)
        .stroke(
          hasVoted ? DSTheme.Colors.border : DSTheme.Colors.warning.opacity(0.3),
          lineWidth: 1
        )
    )
  }
}

// MARK: - Upcoming Meal Card Content
struct UpcomingMealCardContent: View {
  let mealPlan: any MealPlanDisplayable

  private var isPast: Bool {
    MealPlanningHomeHelpers.isMealPlanInPast(mealPlan)
  }

  private var usesDerivedTitle: Bool {
    let notes = mealPlan.notes.trimmingCharacters(in: .whitespacesAndNewlines)
    return notes.isEmpty
      || MealPlanningHomeHelpers.isDefaultMealPlanTitle(notes, mealPlan: mealPlan)
  }

  private var subtitleText: String {
    usesDerivedTitle
      ? MealPlanningHomeHelpers.formatMealPlanTimeRangeCompact(mealPlan)
      : MealPlanningHomeHelpers.formatMealPlanTimeRange(mealPlan)
  }

  var body: some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
      HStack {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.xxs) {
          Text(MealPlanningHomeHelpers.mealPlanDisplayTitle(mealPlan, fallback: "Meal Plan"))
            .font(DSTheme.Typography.label)
            .foregroundColor(DSTheme.Colors.textPrimary)
            .strikethrough(isPast)

          if !subtitleText.isEmpty {
            Text(subtitleText)
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
              .strikethrough(isPast)
          }
        }

        Spacer()

        Image(systemName: "chevron.right")
          .font(.system(size: 14, weight: .semibold))
          .foregroundColor(DSTheme.Colors.textTertiary)
      }

      if !mealPlan.displayEvents.isEmpty {
        DSDivider()

        VStack(spacing: DSTheme.Spacing.xs) {
          ForEach(mealPlan.displayEvents.prefix(3), id: \.id) { event in
            HStack {
              Circle()
                .fill(mealColor(for: event.mealName))
                .frame(width: 8, height: 8)

              Text(eventDisplayLabel(event))
                .font(DSTheme.Typography.body)
                .foregroundColor(DSTheme.Colors.textPrimary)
                .strikethrough(isPast)

              Spacer()

              Text(formatEventDate(event, compact: mealPlan.status == .finalized))
                .font(DSTheme.Typography.caption)
                .foregroundColor(DSTheme.Colors.textSecondary)
                .strikethrough(isPast)
            }
          }
        }

        if mealPlan.displayEvents.count > 3 {
          Text("+ \(mealPlan.displayEvents.count - 3) more")
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.textTertiary)
            .strikethrough(isPast)
        }
      }
    }
    .padding(DSTheme.Spacing.md)
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.lg)
    .overlay(
      RoundedRectangle(cornerRadius: DSTheme.Radius.lg)
        .stroke(DSTheme.Colors.border, lineWidth: 1)
    )
  }

  private func eventDisplayLabel(_ event: any MealPlanEventDisplayable) -> String {
    guard mealPlan.status == .finalized,
      let chosenMealName = event.chosenMealDisplayName
    else {
      return MealPlanningUtils.formatMealName(event.mealName)
    }
    return chosenMealName
  }

  private func formatEventDate(_ event: any MealPlanEventDisplayable, compact: Bool) -> String {
    let date = HomeViewModel.timestampToDate(event.startsAt)
    if compact {
      let weekdayFormatter = DateFormatter()
      weekdayFormatter.dateFormat = "EEE"
      let timeFormatter = DateFormatter()
      timeFormatter.dateStyle = .none
      timeFormatter.timeStyle = .short
      return "\(weekdayFormatter.string(from: date)) \(timeFormatter.string(from: date))"
    }
    let formatter = DateFormatter()
    formatter.dateStyle = .short
    formatter.timeStyle = .short
    return formatter.string(from: date)
  }

  private func mealColor(for mealName: Mealplanning_MealPlanEventName) -> Color {
    switch mealName {
    case .breakfast, .secondBreakfast:
      return .orange
    case .brunch:
      return .yellow
    case .lunch:
      return DSTheme.Colors.primary
    case .supper, .dinner:
      return .purple
    default:
      return DSTheme.Colors.secondary
    }
  }
}

// MARK: - Upcoming Meal Card (legacy wrapper)
struct UpcomingMealCard: View {
  let mealPlan: any MealPlanDisplayable
  let onTap: () -> Void

  var body: some View {
    UpcomingMealCardContent(mealPlan: mealPlan)
  }
}
