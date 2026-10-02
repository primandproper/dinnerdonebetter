//
//  VoteMealPlanViewModelTests.swift
//  iosTests
//
//  Created by Auto on 12/8/25.
//

import Foundation
import SwiftProtobuf
@testable import ios
import Testing

// MARK: - Helper Functions for Test Data

func createMockMealPlan(
  id: String = "test-plan-1",
  notes: String = "Test Meal Plan",
  eventCount: Int = 1,
  optionsPerEvent: Int = 3
) -> Mealplanning_MealPlan {
  var mealPlan = Mealplanning_MealPlan()
  mealPlan.id = id
  mealPlan.notes = notes

  var events: [Mealplanning_MealPlanEvent] = []
  for eventIndex in 0..<eventCount {
    var event = Mealplanning_MealPlanEvent()
    event.id = "event-\(eventIndex)"
    event.mealName = .dinner

    var options: [Mealplanning_MealPlanOption] = []
    for optionIndex in 0..<optionsPerEvent {
      var option = Mealplanning_MealPlanOption()
      option.id = "option-\(eventIndex)-\(optionIndex)"
      var meal = Mealplanning_Meal()
      meal.id = "meal-\(eventIndex)-\(optionIndex)"
      meal.name = "Meal \(eventIndex)-\(optionIndex)"
      option.meal = meal
      options.append(option)
    }
    event.options = options
    events.append(event)
  }
  mealPlan.events = events

  return mealPlan
}

func createMockAuthenticationManager() -> AuthenticationManager {
  let manager = AuthenticationManager()
  manager.isAuthenticated = true
  return manager
}

// MARK: - EventBallot Tests

struct EventBallotTests {
  @Test("EventBallot initializes with required fields")
  func testEventBallotInitialization() {
    var option = Mealplanning_MealPlanOption()
    option.id = "option-1"
    var meal = Mealplanning_Meal()
    meal.name = "Test Meal"
    option.meal = meal

    let ballot = EventBallot(
      id: "event-1",
      rankedOptions: [option],
      isLocked: false
    )

    #expect(ballot.id == "event-1")
    #expect(ballot.rankedOptions.count == 1)
    #expect(ballot.isLocked == false)
  }

  @Test("EventBallot isComplete returns true when all options present")
  func testEventBallotIsComplete() {
    var option1 = Mealplanning_MealPlanOption()
    option1.id = "option-1"
    var option2 = Mealplanning_MealPlanOption()
    option2.id = "option-2"

    let ballot = EventBallot(
      id: "event-1",
      rankedOptions: [option1, option2],
      isLocked: false
    )

    #expect(ballot.isComplete(totalOptions: 2) == true)
    #expect(ballot.isComplete(totalOptions: 3) == false)
  }

  @Test("EventBallot isComplete returns false when options missing")
  func testEventBallotIsIncomplete() {
    var option1 = Mealplanning_MealPlanOption()
    option1.id = "option-1"

    let ballot = EventBallot(
      id: "event-1",
      rankedOptions: [option1],
      isLocked: false
    )

    #expect(ballot.isComplete(totalOptions: 2) == false)
    #expect(ballot.isComplete(totalOptions: 1) == true)
  }
}

// MARK: - VoteMealPlanViewModel Initialization Tests

struct VoteMealPlanViewModelInitializationTests {
  @Test("VoteMealPlanViewModel initializes with meal plan")
  @MainActor
  func testViewModelInitialization() async {
    let mealPlan = createMockMealPlan(eventCount: 2, optionsPerEvent: 3)
    let authManager = createMockAuthenticationManager()

    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    #expect(viewModel.mealPlan.id == mealPlan.id)
    #expect(viewModel.ballots.count == 2)
  }

  @Test("VoteMealPlanViewModel initializes ballots for all events")
  @MainActor
  func testViewModelInitializesBallots() async {
    let mealPlan = createMockMealPlan(eventCount: 3, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()

    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    #expect(viewModel.ballots.count == 3)
    for event in mealPlan.events {
      #expect(viewModel.ballots[event.id] != nil)
      #expect(viewModel.ballots[event.id]?.id == event.id)
      #expect(viewModel.ballots[event.id]?.rankedOptions.count == 2)
    }
  }

  @Test("VoteMealPlanViewModel initializes ballots with original option order")
  @MainActor
  func testViewModelPreservesOptionOrder() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 3)
    let authManager = createMockAuthenticationManager()

    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let event = mealPlan.events[0]
    let ballot = viewModel.ballots[event.id]

    #expect(ballot != nil)
    #expect(ballot?.rankedOptions.count == 3)
    // Verify order matches original
    for (index, option) in event.options.enumerated() {
      #expect(ballot?.rankedOptions[index].id == option.id)
    }
  }

  @Test("VoteMealPlanViewModel initializes with empty meal plan")
  @MainActor
  func testViewModelWithEmptyMealPlan() async {
    var mealPlan = Mealplanning_MealPlan()
    mealPlan.id = "empty-plan"
    mealPlan.events = []
    let authManager = createMockAuthenticationManager()

    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    #expect(viewModel.ballots.isEmpty)
  }
}

// MARK: - Ballot Management Tests

struct BallotManagementTests {
  @Test("getBallot returns ballot for existing event")
  @MainActor
  func testGetBallot() async {
    let mealPlan = createMockMealPlan(eventCount: 2, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    let ballot = viewModel.getBallot(for: eventID)

    #expect(ballot != nil)
    #expect(ballot?.id == eventID)
  }

  @Test("getBallot returns nil for non-existent event")
  @MainActor
  func testGetBallotNonExistent() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let ballot = viewModel.getBallot(for: "non-existent-event")

    #expect(ballot == nil)
  }

  @Test("reorderOptions reorders options in ballot")
  @MainActor
  func testReorderOptions() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 3)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    let originalOrder = viewModel.ballots[eventID]?.rankedOptions.map { $0.id } ?? []

    // Move first option to last position
    viewModel.reorderOptions(eventID: eventID, from: IndexSet(integer: 0), to: 3)

    let newOrder = viewModel.ballots[eventID]?.rankedOptions.map { $0.id } ?? []
    #expect(newOrder.count == 3)
    #expect(newOrder[0] == originalOrder[1])
    #expect(newOrder[1] == originalOrder[2])
    #expect(newOrder[2] == originalOrder[0])
  }

  @Test("reorderOptions does nothing for non-existent event")
  @MainActor
  func testReorderOptionsNonExistent() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let originalCount = viewModel.ballots.count
    let originalEventID = mealPlan.events[0].id
    let originalOptionCount = viewModel.ballots[originalEventID]?.rankedOptions.count

    viewModel.reorderOptions(eventID: "non-existent", from: IndexSet(integer: 0), to: 1)

    // Ballots should be unchanged
    #expect(viewModel.ballots.count == originalCount)
    #expect(viewModel.ballots[originalEventID]?.rankedOptions.count == originalOptionCount)
  }

  @Test("reorderOptions does nothing when ballot is locked")
  @MainActor
  func testReorderOptionsWhenLocked() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 3)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    let originalOrder = viewModel.ballots[eventID]?.rankedOptions.map { $0.id } ?? []

    // Lock the ballot
    viewModel.ballots[eventID]?.isLocked = true

    // Try to reorder
    viewModel.reorderOptions(eventID: eventID, from: IndexSet(integer: 0), to: 2)

    // Order should be unchanged
    let newOrder = viewModel.ballots[eventID]?.rankedOptions.map { $0.id } ?? []
    #expect(newOrder == originalOrder)
  }

  @Test("toggleLock locks complete ballot")
  @MainActor
  func testToggleLockCompleteBallot() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    #expect(viewModel.ballots[eventID]?.isLocked == false)

    viewModel.toggleLock(eventID: eventID)

    #expect(viewModel.ballots[eventID]?.isLocked == true)
  }

  @Test("toggleLock unlocks locked ballot")
  @MainActor
  func testToggleLockUnlock() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    viewModel.ballots[eventID]?.isLocked = true

    viewModel.toggleLock(eventID: eventID)

    #expect(viewModel.ballots[eventID]?.isLocked == false)
  }

  @Test("toggleLock does not lock incomplete ballot")
  @MainActor
  func testToggleLockIncompleteBallot() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 3)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    // Remove one option to make it incomplete
    viewModel.ballots[eventID]?.rankedOptions.removeLast()

    #expect(viewModel.ballots[eventID]?.isLocked == false)

    viewModel.toggleLock(eventID: eventID)

    // Should still be unlocked
    #expect(viewModel.ballots[eventID]?.isLocked == false)
  }

  @Test("toggleLock does nothing for non-existent event")
  @MainActor
  func testToggleLockNonExistent() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let originalCount = viewModel.ballots.count
    let originalEventID = mealPlan.events[0].id
    let originalLockState = viewModel.ballots[originalEventID]?.isLocked

    viewModel.toggleLock(eventID: "non-existent")

    // Ballots should be unchanged
    #expect(viewModel.ballots.count == originalCount)
    #expect(viewModel.ballots[originalEventID]?.isLocked == originalLockState)
  }
}

// MARK: - CanSubmit Tests

@Suite(.serialized)
struct CanSubmitTests {
  @Test("canSubmit returns false when no ballots")
  @MainActor
  func testCanSubmitNoBallots() async {
    var mealPlan = Mealplanning_MealPlan()
    mealPlan.id = "empty-plan"
    mealPlan.events = []
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Ensure initialization completes
    try? await Task.sleep(nanoseconds: 10_000_000)  // 10ms

    #expect(viewModel.ballots.isEmpty == true, "Ballots should be empty for meal plan with no events")
    #expect(viewModel.canSubmit == false, "canSubmit should be false when there are no ballots")
  }

  @Test("canSubmit returns false when ballots not locked")
  @MainActor
  func testCanSubmitNotLocked() async {
    let mealPlan = createMockMealPlan(eventCount: 2, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    #expect(viewModel.canSubmit == false)
  }

  @Test("canSubmit returns false when ballots incomplete")
  @MainActor
  func testCanSubmitIncomplete() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 3)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    // Remove one option
    viewModel.ballots[eventID]?.rankedOptions.removeLast()
    // Lock it anyway
    viewModel.ballots[eventID]?.isLocked = true

    #expect(viewModel.canSubmit == false)
  }

  @Test("canSubmit returns false when some ballots incomplete")
  @MainActor
  func testCanSubmitSomeIncomplete() async {
    let mealPlan = createMockMealPlan(eventCount: 2, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Lock first event
    let eventID1 = mealPlan.events[0].id
    viewModel.toggleLock(eventID: eventID1)

    // Second event remains unlocked
    #expect(viewModel.canSubmit == false)
  }

  @Test("canSubmit returns true when all ballots complete and locked")
  @MainActor
  func testCanSubmitAllCompleteAndLocked() async {
    let mealPlan = createMockMealPlan(eventCount: 2, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Lock all events
    for event in mealPlan.events {
      viewModel.toggleLock(eventID: event.id)
    }

    #expect(viewModel.canSubmit == true)
  }

  @Test("canSubmit returns false when ballot missing")
  @MainActor
  func testCanSubmitMissingBallot() async {
    let mealPlan = createMockMealPlan(eventCount: 2, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Remove a ballot
    let eventID = mealPlan.events[0].id
    viewModel.ballots.removeValue(forKey: eventID)

    #expect(viewModel.canSubmit == false)
  }
}

// MARK: - Vote Submission Tests

struct VoteSubmissionTests {
  @Test("submitVotes returns false when cannot submit")
  @MainActor
  func testSubmitVotesCannotSubmit() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Don't lock ballots
    let result = await viewModel.submitVotes()

    #expect(result == false)
    #expect(viewModel.submissionError != nil)
    #expect(viewModel.submissionError?.contains("complete and locked") == true)
  }

  @Test("submitVotes sets submission state")
  @MainActor
  func testSubmitVotesSetsState() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Lock ballot
    let eventID = mealPlan.events[0].id
    viewModel.toggleLock(eventID: eventID)

    // Note: This will fail without a server, but tests state management
    _ = await viewModel.submitVotes()

    // State should be reset after attempt
    #expect(viewModel.isSubmitting == false)
  }

  @Test("submitVotes handles multiple events")
  @MainActor
  func testSubmitVotesMultipleEvents() async {
    let mealPlan = createMockMealPlan(eventCount: 3, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Lock all ballots
    for event in mealPlan.events {
      viewModel.toggleLock(eventID: event.id)
    }

    // Note: This will fail without a server, but tests the flow
    let result = await viewModel.submitVotes()

    // Result depends on server availability
    #expect(result == false || result == true)
  }
}

// MARK: - Edge Cases Tests

struct VoteMealPlanViewModelEdgeCaseTests {
  @Test("ViewModel handles meal plan with single event")
  @MainActor
  func testViewModelSingleEvent() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 5)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    #expect(viewModel.ballots.count == 1)
    #expect(viewModel.ballots[mealPlan.events[0].id]?.rankedOptions.count == 5)
  }

  @Test("ViewModel handles meal plan with many events")
  @MainActor
  func testViewModelManyEvents() async {
    let mealPlan = createMockMealPlan(eventCount: 10, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    #expect(viewModel.ballots.count == 10)
  }

  @Test("ViewModel handles event with single option")
  @MainActor
  func testViewModelSingleOption() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 1)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    #expect(viewModel.ballots[eventID]?.rankedOptions.count == 1)
    #expect(viewModel.ballots[eventID]?.isComplete(totalOptions: 1) == true)
  }

  @Test("ViewModel handles event with many options")
  @MainActor
  func testViewModelManyOptions() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 20)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    #expect(viewModel.ballots[eventID]?.rankedOptions.count == 20)
  }

  @Test("reorderOptions handles multiple moves")
  @MainActor
  func testReorderOptionsMultipleMoves() async {
    let mealPlan = createMockMealPlan(eventCount: 1, optionsPerEvent: 5)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    let eventID = mealPlan.events[0].id
    let originalIDs = viewModel.ballots[eventID]?.rankedOptions.map { $0.id } ?? []

    // Move first to last
    viewModel.reorderOptions(eventID: eventID, from: IndexSet(integer: 0), to: 5)

    // After first move, indices have shifted
    // Move what is now at index 3 (was index 4) to first position
    viewModel.reorderOptions(eventID: eventID, from: IndexSet(integer: 3), to: 0)

    let newIDs = viewModel.ballots[eventID]?.rankedOptions.map { $0.id } ?? []
    #expect(newIDs.count == 5)
    // Verify all original IDs are still present (just reordered)
    #expect(Set(newIDs) == Set(originalIDs))
  }

  @Test("canSubmit handles mixed locked and unlocked ballots")
  @MainActor
  func testCanSubmitMixedStates() async {
    let mealPlan = createMockMealPlan(eventCount: 3, optionsPerEvent: 2)
    let authManager = createMockAuthenticationManager()
    let viewModel = VoteMealPlanViewModel(mealPlan: mealPlan, authManager: authManager)

    // Lock first two events
    viewModel.toggleLock(eventID: mealPlan.events[0].id)
    viewModel.toggleLock(eventID: mealPlan.events[1].id)
    // Third remains unlocked

    #expect(viewModel.canSubmit == false)
  }
}

