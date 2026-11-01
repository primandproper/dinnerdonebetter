//
//  RecipePerformanceContentView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import DesignSystem
import SwiftProtobuf
import SwiftUI

/// A reusable view for displaying recipe performance content (ingredients, instruments, vessels, and steps)
/// This can be embedded in PerformRecipeView, Meal views, or any other context where recipe performance is needed
struct RecipePerformanceContentView: View {  // swiftlint:disable:this type_body_length
  @Environment(EventReporterService.self) private var eventReporterService
  @Binding var checkedIngredients: Set<String>
  @Binding var checkedInstrumentsVessels: Set<String>
  @Binding var isInstrumentsVesselsExpanded: Bool
  @Binding var isIngredientsExpanded: Bool
  @State private var isPrepTasksExpanded: Bool = false
  @State private var isDAGSectionExpanded: Bool = false
  @State private var showDiagramFullScreen: Bool = false

  let recipe: Mealplanning_Recipe
  let viewModel: PerformRecipeViewModel
  var hideIngredientsAndInstruments: Bool = false
  var mealPlanSelections: [Mealplanning_MealPlanRecipeOptionSelection]?
  var highlightedStepIDs: Set<String>?
  var prepTaskContext: PerformRecipeView.PrepTaskContext?
  @Binding var externalScale: Float?  // Optional external scale binding (for meal components)
  var showWashHandsStepCard: Bool = true
  var sharedWashHandsCompleted: Binding<Bool>?
  var sharedCompletedStepsVisibility: Binding<Bool>?
  var allowCompletedStepsToggle: Bool = false
  var onStepCompleted: (() -> Void)?

  // State for option selections (for interactive selection outside meal plan context)
  // optionGroupID -> selectedOptionIndex
  @State private var selectedIngredientOptions: [String: UInt32] = [:]
  // optionGroupID -> selectedOptionIndex
  @State private var selectedInstrumentOptions: [String: UInt32] = [:]
  // optionGroupID -> selectedOptionIndex
  @State private var selectedVesselOptions: [String: UInt32] = [:]

  // State for recipe scaling
  @State private var internalRecipeScale: Float = 1.0
  @State private var scaleText: String = "1.0"
  @FocusState private var isScaleFocused: Bool
  @State private var showCompletedSteps: Bool = true
  @State private var customUpNextOrder: [String] = []
  @State private var customForLaterOrder: [String] = []
  @State private var timerTick: Int = 0
  @State private var timerRefresh: Timer?
  /// When true, show all steps (including non-task) for context. Only used when opened from meal plan prep task.
  @State private var showAllStepsFromPrepTask: Bool = false

  private var isShowingCompletedSteps: Bool {
    sharedCompletedStepsVisibility?.wrappedValue ?? showCompletedSteps
  }

  private func setShowingCompletedSteps(_ newValue: Bool) {
    if let sharedCompletedStepsVisibility {
      sharedCompletedStepsVisibility.wrappedValue = newValue
    } else {
      showCompletedSteps = newValue
    }
  }

  // Helper to get current scale value
  private var recipeScale: Float {
    externalScale ?? internalRecipeScale
  }

  // Helper to set scale value
  private func setRecipeScale(_ newValue: Float) {
    if externalScale != nil {
      externalScale = newValue
    } else {
      internalRecipeScale = newValue
    }
  }

  private func startTimerRefresh() {
    guard timerRefresh == nil else { return }
    timerRefresh = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { _ in
      timerTick += 1
    }
    if let timer = timerRefresh {
      RunLoop.main.add(timer, forMode: .common)
    }
  }

  private func stopTimerRefresh() {
    timerRefresh?.invalidate()
    timerRefresh = nil
  }

  init(
    checkedIngredients: Binding<Set<String>>,
    checkedInstrumentsVessels: Binding<Set<String>>,
    isInstrumentsVesselsExpanded: Binding<Bool>,
    isIngredientsExpanded: Binding<Bool>,
    recipe: Mealplanning_Recipe,
    viewModel: PerformRecipeViewModel,
    hideIngredientsAndInstruments: Bool = false,
    mealPlanSelections: [Mealplanning_MealPlanRecipeOptionSelection]? = nil,
    highlightedStepIDs: Set<String>? = nil,
    prepTaskContext: PerformRecipeView.PrepTaskContext? = nil,
    externalScale: Binding<Float?> = .constant(nil),
    showWashHandsStepCard: Bool = true,
    sharedWashHandsCompleted: Binding<Bool>? = nil,
    sharedCompletedStepsVisibility: Binding<Bool>? = nil,
    allowCompletedStepsToggle: Bool = false,
    onStepCompleted: (() -> Void)? = nil
  ) {
    self._checkedIngredients = checkedIngredients
    self._checkedInstrumentsVessels = checkedInstrumentsVessels
    self._isInstrumentsVesselsExpanded = isInstrumentsVesselsExpanded
    self._isIngredientsExpanded = isIngredientsExpanded
    self.recipe = recipe
    self.viewModel = viewModel
    self.hideIngredientsAndInstruments = hideIngredientsAndInstruments
    self.mealPlanSelections = mealPlanSelections
    self.highlightedStepIDs = highlightedStepIDs
    self.prepTaskContext = prepTaskContext
    self._externalScale = externalScale
    self.showWashHandsStepCard = showWashHandsStepCard
    self.sharedWashHandsCompleted = sharedWashHandsCompleted
    self.sharedCompletedStepsVisibility = sharedCompletedStepsVisibility
    self.allowCompletedStepsToggle = allowCompletedStepsToggle
    self.onStepCompleted = onStepCompleted
  }

  @Environment(AuthenticationManager.self) private var authManager

  var body: some View {
    ScrollView {
      VStack(alignment: .leading, spacing: 16) {
        // Recipe header
        recipeHeader(recipe: recipe, viewModel: viewModel)

        // Associated recipes section
        if !recipe.associatedRecipes.isEmpty {
          associatedRecipesSection(recipe: recipe)
        }

        // DAG section (hidden when embedded in meal view)
        if !hideIngredientsAndInstruments {
          recipeDAGSection(recipe: recipe, viewModel: viewModel)
        }

        // Instruments & Vessels section (hidden when embedded in meal view)
        if !hideIngredientsAndInstruments {
          instrumentsVesselsSection(recipe: recipe)
        }

        // Ingredients section (hidden when embedded in meal view)
        if !hideIngredientsAndInstruments {
          ingredientsSection(recipe: recipe)
        }

        // Prep Tasks section (hidden when embedded in meal view)
        if !hideIngredientsAndInstruments {
          prepTasksSection(recipe: recipe, viewModel: viewModel)
        }

        // Steps list (timerTick forces refresh when step timers are active)
        stepsList(recipe: recipe, viewModel: viewModel, scale: recipeScale)
          .id(timerTick)
      }
      .padding()
    }
    .onAppear {
      if let sharedWashHandsCompleted {
        viewModel.washHandsCompleted = sharedWashHandsCompleted.wrappedValue
      }
      if allowCompletedStepsToggle && sharedCompletedStepsVisibility == nil {
        showCompletedSteps = false
      }
      loadStepOrderFromUserDefaults(recipeID: recipe.id)
      if viewModel.hasActiveStepTimers {
        startTimerRefresh()
      }
    }
    .onChange(of: viewModel.hasActiveStepTimers) { _, hasActive in
      if hasActive {
        startTimerRefresh()
      } else {
        stopTimerRefresh()
      }
    }
    .onDisappear {
      stopTimerRefresh()
    }
    .onChange(of: sharedWashHandsValue) { _, newValue in
      if sharedWashHandsCompleted != nil {
        viewModel.washHandsCompleted = newValue
      }
    }
    .onChange(of: recipe.id) { _, newID in
      loadStepOrderFromUserDefaults(recipeID: newID)
    }
  }

  // MARK: - Associated Recipes Section

  private func associatedRecipesSection(recipe: Mealplanning_Recipe) -> some View {
    VStack(alignment: .leading, spacing: 12) {
      Text("Prerequisite Recipes")
        .font(.headline)
        .padding(.horizontal, 4)

      ForEach(recipe.associatedRecipes, id: \.id) { associatedRecipe in
        AssociatedRecipeCard(
          recipe: associatedRecipe,
          viewModel: viewModel,
          parentRecipe: recipe
        )
      }
    }
  }

  // MARK: - Recipe DAG Section

  private func recipeDAGSection(recipe: Mealplanning_Recipe, viewModel: PerformRecipeViewModel)
    -> some View
  {
    VStack(alignment: .leading, spacing: 0) {
      Button(
        action: {
          withAnimation {
            isDAGSectionExpanded.toggle()
          }
        },
        label: {
          HStack {
            Text("Recipe Graph")
              .font(.headline)
              .foregroundColor(.primary)
            Spacer()
            if let mermaidSource = viewModel.mermaidDiagram, !mermaidSource.isEmpty {
              Button {
                showDiagramFullScreen = true
              } label: {
                Image(systemName: "arrow.up.left.and.arrow.down.right")
                  .font(.caption)
                  .foregroundColor(.secondary)
              }
              .buttonStyle(.plain)
            }
            Image(systemName: isDAGSectionExpanded ? "chevron.down" : "chevron.right")
              .font(.caption)
              .foregroundColor(.secondary)
          }
          .padding()
          .background(Color(.systemGray6))
        }
      )
      .buttonStyle(.plain)
      .fullScreenCover(isPresented: $showDiagramFullScreen) {
        DiagramFullScreenSheet(
          mermaidSource: viewModel.mermaidDiagram ?? "",
          title: "Recipe Graph",
          onDismiss: { showDiagramFullScreen = false }
        )
      }

      if isDAGSectionExpanded {
        VStack(alignment: .leading, spacing: 8) {
          if viewModel.isLoadingMermaid {
            HStack {
              ProgressView()
                .scaleEffect(0.8)
              Text("Loading diagram...")
                .font(.subheadline)
                .foregroundColor(.secondary)
            }
            .frame(maxWidth: .infinity)
            .padding()
          } else if let error = viewModel.mermaidError {
            VStack(spacing: 12) {
              Text(error)
                .font(.subheadline)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
              DSButton("Retry", icon: "arrow.clockwise", style: .primary, size: .small) {
                Task {
                  await viewModel.loadMermaidDiagram()
                }
              }
            }
            .frame(maxWidth: .infinity)
            .padding()
          } else if let mermaidSource = viewModel.mermaidDiagram, !mermaidSource.isEmpty {
            MermaidDiagramView(source: mermaidSource)
              .padding()
          } else {
            Text("No diagram available for this recipe")
              .font(.subheadline)
              .foregroundColor(.secondary)
              .frame(maxWidth: .infinity)
              .padding()
          }
        }
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
        .onAppear {
          Task {
            await viewModel.loadMermaidDiagram()
          }
        }
      }
    }
    .background(Color(.systemGray6))
    .cornerRadius(12)
  }

  // MARK: - Recipe Header

  private func formatEventTime(_ date: Date) -> String {
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short
    return formatter.string(from: date)
  }

  private func recipeHeader(recipe: Mealplanning_Recipe, viewModel: PerformRecipeViewModel)
    -> some View
  {
    VStack(alignment: .leading, spacing: 8) {
      // Show prep task context if available
      if let context = prepTaskContext {
        VStack(alignment: .leading, spacing: 4) {
          if let prepTaskName = context.prepTaskName, !prepTaskName.isEmpty {
            Text(prepTaskName)
              .font(.headline)
              .foregroundColor(.blue)
          }

          HStack(spacing: 4) {
            if let recipeName = context.recipeName, !recipeName.isEmpty {
              Text("for \(recipeName)")
                .font(.subheadline)
                .foregroundColor(.secondary)
            }

            if let eventName = context.eventName, let eventTime = context.eventTime {
              Text("• \(eventName) at \(formatEventTime(eventTime))")
                .font(.subheadline)
                .foregroundColor(.secondary)
            }
          }
        }
        .padding(.bottom, 8)
      }

      Text(recipe.name.isEmpty ? "Unnamed Recipe" : recipe.name)
        .font(.title)
        .fontWeight(.bold)

      if !recipe.description_p.isEmpty {
        Text(recipe.description_p)
          .font(.subheadline)
          .foregroundColor(.secondary)
      }

      if !recipe.source.isEmpty {
        RecipeSourceView(source: recipe.source)
      }

      // Progress indicator and time estimate
      let completedCount =
        viewModel.completedSteps.count + ((showWashHandsStepCard && sharedWashHandsValue) ? 1 : 0)
      let totalSteps = recipe.steps.count + (showWashHandsStepCard ? 1 : 0)
      HStack(spacing: 12) {
        Text("\(completedCount) of \(totalSteps) steps completed")
          .font(.caption)
          .foregroundColor(.secondary)
        if let estimate = RecipeTimeEstimation.estimate(steps: recipe.steps) {
          Label(
            RecipeTimeEstimation.format(
              minSeconds: estimate.minSeconds, maxSeconds: estimate.maxSeconds),
            systemImage: "clock"
          )
          .font(.caption)
          .foregroundColor(.secondary)
        }
      }
      .padding(.top, 4)

      // Scale control
      if !hideIngredientsAndInstruments {
        DSDivider()
          .padding(.vertical, 8)

        HStack(spacing: 12) {
          Text("Scale:")
            .font(.subheadline)
            .fontWeight(.medium)

          HStack(spacing: 8) {
            TextField("1.0", text: $scaleText)
              .keyboardType(.decimalPad)
              .textFieldStyle(.roundedBorder)
              .frame(width: 80)
              .focused($isScaleFocused)
              .onSubmit {
                updateScaleFromText()
              }
              .onChange(of: isScaleFocused) { _, isFocused in
                if !isFocused {
                  updateScaleFromText()
                }
              }
              .onChange(of: scaleText) { _, newValue in
                // Filter to only allow numbers and a single decimal point
                var filtered = newValue.filter { $0.isNumber || $0 == "." }
                // Ensure only one decimal point
                let parts = filtered.split(separator: ".", omittingEmptySubsequences: false)
                if parts.count > 2 {
                  filtered = parts[0] + "." + parts.dropFirst().joined()
                }
                if filtered != newValue {
                  scaleText = filtered
                }
              }

            Text("x")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)

            DSStepperButtons(
              onDecrement: { adjustRecipeScale(by: -0.25) },
              onIncrement: { adjustRecipeScale(by: 0.25) }
            )
          }
        }

        Slider(
          value: Binding(
            get: { Double(recipeScale) },
            set: { setRecipeScale(Float($0)) }
          ),
          in: 0.25...4.0,
          step: 0.25
        )
      }
    }
    .padding()
    .frame(maxWidth: .infinity, alignment: .leading)
    .background(Color(.systemGray6))
    .cornerRadius(12)
    .onAppear {
      scaleText = String(format: "%.2f", recipeScale)
    }
    .onChange(of: externalScale) { _, newValue in
      if !isScaleFocused, let newValue = newValue {
        scaleText = String(format: "%.2f", newValue)
      }
    }
    .onChange(of: internalRecipeScale) { _, newValue in
      if !isScaleFocused, externalScale == nil {
        scaleText = String(format: "%.2f", newValue)
      }
    }
  }

  private func updateScaleFromText() {
    if let scale = Float(scaleText), scale > 0 {
      setRecipeScale(min(max(scale, 0.25), 4.0))
      scaleText = String(format: "%.2f", recipeScale)
    } else {
      // Reset to current scale if invalid input
      scaleText = String(format: "%.2f", recipeScale)
    }
  }

  private func adjustRecipeScale(by delta: Float) {
    let next = min(max(recipeScale + delta, 0.25), 4.0)
    setRecipeScale(next)
    scaleText = String(format: "%.2f", next)
    eventReporterService.reporter.track(
      event: "perform_recipe_scale_changed",
      properties: ["scale": next])
  }

  // MARK: - Instruments & Vessels Section

  private func instrumentsVesselsSection(recipe: Mealplanning_Recipe) -> some View {
    let (regularItems, instrumentOptionGroups, vesselOptionGroups) =
      getAggregatedInstrumentsAndVessels(
        from: recipe,
        selectedInstrumentOptions: selectedInstrumentOptions,
        selectedVesselOptions: selectedVesselOptions,
        mealPlanSelections: mealPlanSelections,
        scale: recipeScale
      )

    // Filter option groups based on meal plan selections or user selections
    let filteredInstrumentGroups = filterInstrumentOptionGroupsForDisplay(instrumentOptionGroups)
    let filteredVesselGroups = filterVesselOptionGroupsForDisplay(vesselOptionGroups)

    return VStack(alignment: .leading, spacing: 0) {
      Button(
        action: {
          withAnimation {
            isInstrumentsVesselsExpanded.toggle()
          }
        },
        label: {
          HStack {
            Text("Equipment")
              .font(.headline)
              .foregroundColor(.primary)
            Spacer()
            Image(systemName: isInstrumentsVesselsExpanded ? "chevron.down" : "chevron.right")
              .font(.caption)
              .foregroundColor(.secondary)
          }
          .padding()
          .background(Color(.systemGray6))
        }
      )
      .buttonStyle(.plain)

      if isInstrumentsVesselsExpanded {
        VStack(alignment: .leading, spacing: 8) {
          // Regular items
          if !regularItems.isEmpty {
            ForEach(regularItems, id: \.itemID) { item in
              instrumentVesselRow(item: item)
            }
          }

          // Instrument option groups
          if !filteredInstrumentGroups.isEmpty {
            Text("Options")
              .font(.subheadline)
              .fontWeight(.semibold)
              .foregroundColor(.secondary)
              .padding(.top, 8)
              .padding(.horizontal)

            ForEach(filteredInstrumentGroups) { group in
              InteractiveInstrumentOptionGroupView(
                group: group,
                selectedOptionIndex: Binding(
                  get: {
                    // Check meal plan selections first
                    if let selections = mealPlanSelections,
                      let selection = selections.first(where: { sel in
                        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
                          && sel.selectionType == .instrument
                      })
                    {
                      return selection.selectedOptionIndex
                    }
                    // Then check user selections, or use sentinel value if none
                    return selectedInstrumentOptions[group.id] ?? UInt32.max
                  },
                  set: { newValue in
                    if newValue != UInt32.max {
                      selectedInstrumentOptions[group.id] = newValue
                    } else {
                      selectedInstrumentOptions.removeValue(forKey: group.id)
                    }
                  }
                ),
                scale: recipeScale
              )
            }
          }

          // Vessel option groups
          if !filteredVesselGroups.isEmpty {
            if !filteredInstrumentGroups.isEmpty {
              Text("Equipment Options")
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundColor(.secondary)
                .padding(.top, 8)
                .padding(.horizontal)
            } else {
              Text("Options")
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundColor(.secondary)
                .padding(.top, 8)
                .padding(.horizontal)
            }

            ForEach(filteredVesselGroups) { group in
              InteractiveVesselOptionGroupView(
                group: group,
                selectedOptionIndex: Binding(
                  get: {
                    // Check meal plan selections first
                    if let selections = mealPlanSelections,
                      let selection = selections.first(where: { sel in
                        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
                          && sel.selectionType == .vessel
                      })
                    {
                      return selection.selectedOptionIndex
                    }
                    // Then check user selections, or use sentinel value if none
                    return selectedVesselOptions[group.id] ?? UInt32.max
                  },
                  set: { newValue in
                    if newValue != UInt32.max {
                      selectedVesselOptions[group.id] = newValue
                    } else {
                      selectedVesselOptions.removeValue(forKey: group.id)
                    }
                  }
                ),
                scale: recipeScale
              )
            }
          }
        }
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
      }
    }
    .background(Color(.systemGray6))
    .cornerRadius(12)
  }

  private func instrumentVesselRow(item: AggregatedInstrumentVessel) -> some View {
    HStack(spacing: 12) {
      // Checkbox
      Button(
        action: {
          if checkedInstrumentsVessels.contains(item.itemID) {
            checkedInstrumentsVessels.remove(item.itemID)
          } else {
            checkedInstrumentsVessels.insert(item.itemID)
          }
        },
        label: {
          Image(
            systemName: checkedInstrumentsVessels.contains(item.itemID)
              ? "checkmark.circle.fill" : "circle"
          )
          .font(.title3)
          .foregroundColor(checkedInstrumentsVessels.contains(item.itemID) ? .green : .gray)
        }
      )
      .buttonStyle(.plain)

      HStack(spacing: 8) {
        Image(
          systemName: item.type == .instrument
            ? "wrench.and.screwdriver" : "square.stack.3d.up"
        )
        .font(.caption)
        .foregroundColor(.secondary)
        .frame(width: 20)

        HStack {
          Text(item.name)
            .font(.subheadline)
            .foregroundColor(
              checkedInstrumentsVessels.contains(item.itemID) ? .secondary : .primary
            )
            .strikethrough(checkedInstrumentsVessels.contains(item.itemID))

          if let sourceRecipeName = item.sourceRecipeName {
            Text("(from \(sourceRecipeName))")
              .font(.caption2)
              .foregroundColor(.secondary)
          }
        }
      }

      Spacer()
    }
    .padding(.horizontal)
    .padding(.vertical, 4)
  }

  private func filterInstrumentOptionGroups(
    _ groups: [InstrumentOptionGroupAggregate]
  ) -> [InstrumentOptionGroupAggregate] {
    guard let selections = mealPlanSelections else {
      return groups
    }

    return groups.compactMap { group in
      let selection = selections.first { sel in
        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
          && sel.selectionType == .instrument
      }

      if let selection = selection {
        let selectedOptions = group.options.filter {
          $0.optionIndex == selection.selectedOptionIndex
        }
        if !selectedOptions.isEmpty {
          return InstrumentOptionGroupAggregate(
            id: group.id,
            recipeID: group.recipeID,
            stepID: group.stepID,
            stepIndex: group.stepIndex,
            index: group.index,
            options: selectedOptions,
            selectedOptionIndex: selection.selectedOptionIndex,
            sourceRecipeID: group.sourceRecipeID,
            sourceRecipeName: group.sourceRecipeName
          )
        }
        return nil
      }

      return group
    }
  }

  private func filterVesselOptionGroups(
    _ groups: [VesselOptionGroupAggregate]
  ) -> [VesselOptionGroupAggregate] {
    guard let selections = mealPlanSelections else {
      return groups
    }

    return groups.compactMap { group in
      let selection = selections.first { sel in
        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
          && sel.selectionType == .vessel
      }

      if let selection = selection {
        let selectedOptions = group.options.filter {
          $0.optionIndex == selection.selectedOptionIndex
        }
        if !selectedOptions.isEmpty {
          return VesselOptionGroupAggregate(
            id: group.id,
            recipeID: group.recipeID,
            stepID: group.stepID,
            stepIndex: group.stepIndex,
            index: group.index,
            options: selectedOptions,
            selectedOptionIndex: selection.selectedOptionIndex,
            sourceRecipeID: group.sourceRecipeID,
            sourceRecipeName: group.sourceRecipeName
          )
        }
        return nil
      }

      return group
    }
  }

  // MARK: - Ingredients Section

  private func ingredientsSection(recipe: Mealplanning_Recipe) -> some View {
    let (regularIngredients, optionGroups) = getAggregatedIngredients(
      from: recipe,
      selectedIngredientOptions: selectedIngredientOptions,
      mealPlanSelections: mealPlanSelections,
      scale: recipeScale
    )

    // Filter option groups based on meal plan selections or user selections
    let filteredOptionGroups = filterIngredientOptionGroupsForDisplay(optionGroups)

    return VStack(alignment: .leading, spacing: 0) {
      Button(
        action: {
          withAnimation {
            isIngredientsExpanded.toggle()
          }
        },
        label: {
          HStack {
            Text("Ingredients")
              .font(.headline)
              .foregroundColor(.primary)
            Spacer()
            Image(systemName: isIngredientsExpanded ? "chevron.down" : "chevron.right")
              .font(.caption)
              .foregroundColor(.secondary)
          }
          .padding()
          .background(Color(.systemGray6))
        }
      )
      .buttonStyle(.plain)

      if isIngredientsExpanded {
        VStack(alignment: .leading, spacing: 8) {
          // Regular ingredients
          if !regularIngredients.isEmpty {
            ForEach(regularIngredients, id: \.ingredientID) { aggregated in
              ingredientRow(aggregated: aggregated)
            }
          }

          // Options section (only ingredients have selectable options)
          if !filteredOptionGroups.isEmpty {
            Text("Options")
              .font(.subheadline)
              .fontWeight(.semibold)
              .foregroundColor(.secondary)
              .padding(.top, 8)
              .padding(.horizontal)

            ForEach(filteredOptionGroups) { group in
              InteractiveIngredientOptionGroupView(
                group: group,
                selectedOptionIndex: Binding(
                  get: {
                    // Check meal plan selections first
                    if let selections = mealPlanSelections,
                      let selection = selections.first(where: { sel in
                        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
                          && sel.ingredientIndex == group.index && sel.selectionType == .ingredient
                      })
                    {
                      return selection.selectedOptionIndex
                    }
                    // Then check user selections, or use sentinel value if none
                    return selectedIngredientOptions[group.id] ?? UInt32.max
                  },
                  set: { newValue in
                    if newValue != UInt32.max {
                      selectedIngredientOptions[group.id] = newValue
                    } else {
                      selectedIngredientOptions.removeValue(forKey: group.id)
                    }
                  }
                ),
                scale: recipeScale
              )
            }
          }
        }
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
      }
    }
    .background(Color(.systemGray6))
    .cornerRadius(12)
  }

  private func ingredientRow(aggregated: AggregatedIngredient) -> some View {
    HStack(spacing: 12) {
      // Checkbox
      Button(
        action: {
          if checkedIngredients.contains(aggregated.ingredientID) {
            checkedIngredients.remove(aggregated.ingredientID)
          } else {
            checkedIngredients.insert(aggregated.ingredientID)
          }
        },
        label: {
          Image(
            systemName: checkedIngredients.contains(aggregated.ingredientID)
              ? "checkmark.circle.fill" : "circle"
          )
          .font(.title3)
          .foregroundColor(
            checkedIngredients.contains(aggregated.ingredientID) ? .green : .gray
          )
        }
      )
      .buttonStyle(.plain)

      VStack(alignment: .leading, spacing: 2) {
        HStack {
          if let quantityText = aggregated.quantityText(scale: recipeScale) {
            Text(quantityText)
              .font(.subheadline)
              .fontWeight(.medium)
              .foregroundColor(.secondary)
          }

          Text(aggregated.name)
            .font(.subheadline)
            .foregroundColor(
              checkedIngredients.contains(aggregated.ingredientID) ? .secondary : .primary
            )
            .strikethrough(checkedIngredients.contains(aggregated.ingredientID))

          if let sourceRecipeName = aggregated.sourceRecipeName {
            Text("(from \(sourceRecipeName))")
              .font(.caption2)
              .foregroundColor(.secondary)
          }
        }

        if !aggregated.quantityNotes.isEmpty {
          Text(aggregated.quantityNotes)
            .font(.caption)
            .foregroundColor(.secondary)
        }
      }

      Spacer()
    }
    .padding(.horizontal)
    .padding(.vertical, 4)
  }

  // Helper to get selected option index for an ingredient option group
  private func getSelectedIngredientOptionIndex(for group: OptionGroupAggregate) -> UInt32 {
    if let selections = mealPlanSelections {
      if let selection = selections.first(where: { sel in
        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
          && sel.ingredientIndex == group.index && sel.selectionType == .ingredient
      }) {
        return selection.selectedOptionIndex
      }
    }
    // Default to user selection or optionIndex 0
    return selectedIngredientOptions[group.id]
      ?? (group.options.first(where: { $0.optionIndex == 0 })?.optionIndex ?? group.options.first?
        .optionIndex ?? 0)
  }

  // Helper to get selected option index for an instrument option group
  private func getSelectedInstrumentOptionIndex(for group: InstrumentOptionGroupAggregate) -> UInt32
  {
    if let selections = mealPlanSelections {
      if let selection = selections.first(where: { sel in
        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
          && sel.selectionType == .instrument
      }) {
        return selection.selectedOptionIndex
      }
    }
    // Default to user selection or optionIndex 0
    return selectedInstrumentOptions[group.id]
      ?? (group.options.first(where: { $0.optionIndex == 0 })?.optionIndex ?? group.options.first?
        .optionIndex ?? 0)
  }

  // Helper to get selected option index for a vessel option group
  private func getSelectedVesselOptionIndex(for group: VesselOptionGroupAggregate) -> UInt32 {
    if let selections = mealPlanSelections {
      if let selection = selections.first(where: { sel in
        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
          && sel.selectionType == .vessel
      }) {
        return selection.selectedOptionIndex
      }
    }
    // Default to user selection or optionIndex 0
    return selectedVesselOptions[group.id]
      ?? (group.options.first(where: { $0.optionIndex == 0 })?.optionIndex ?? group.options.first?
        .optionIndex ?? 0)
  }

  private func filterIngredientOptionGroupsForDisplay(
    _ groups: [OptionGroupAggregate]
  ) -> [OptionGroupAggregate] {
    guard let selections = mealPlanSelections else {
      // No meal plan selections - return all groups for interactive selection
      return groups
    }

    // Filter based on meal plan selections
    return groups.compactMap { group in
      let selection = selections.first { sel in
        sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
          && sel.ingredientIndex == group.index && sel.selectionType == .ingredient
      }

      if let selection = selection {
        // Only show the selected option
        let selectedOptions = group.options.filter {
          $0.optionIndex == selection.selectedOptionIndex
        }
        if !selectedOptions.isEmpty {
          return OptionGroupAggregate(
            id: group.id,
            recipeID: group.recipeID,
            stepID: group.stepID,
            stepIndex: group.stepIndex,
            index: group.index,
            options: selectedOptions,
            selectedOptionIndex: selection.selectedOptionIndex,
            sourceRecipeID: group.sourceRecipeID,
            sourceRecipeName: group.sourceRecipeName
          )
        }
        return nil
      }

      // No selection - show all options
      return group
    }
  }

  private func filterInstrumentOptionGroupsForDisplay(
    _ groups: [InstrumentOptionGroupAggregate]
  ) -> [InstrumentOptionGroupAggregate] {
    return groups.map { group in
      let selectedIndex: UInt32?
      // Check meal plan selections first
      if let selections = mealPlanSelections,
        let selection = selections.first(where: { sel in
          sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
            && sel.selectionType == .instrument
        })
      {
        selectedIndex = selection.selectedOptionIndex
      } else {
        // Check user selections - use sentinel value if none
        let userSelection = selectedInstrumentOptions[group.id]
        selectedIndex = userSelection == UInt32.max ? nil : userSelection
      }

      // If a selection has been made, filter to show only that option
      if let selectedIndex = selectedIndex {
        let filteredOptions = group.options.filter { $0.optionIndex == selectedIndex }
        return InstrumentOptionGroupAggregate(
          id: group.id,
          recipeID: group.recipeID,
          stepID: group.stepID,
          stepIndex: group.stepIndex,
          index: group.index,
          options: filteredOptions.isEmpty ? group.options : filteredOptions,
          selectedOptionIndex: selectedIndex,
          sourceRecipeID: group.sourceRecipeID,
          sourceRecipeName: group.sourceRecipeName
        )
      }

      // No selection - show all options
      return InstrumentOptionGroupAggregate(
        id: group.id,
        recipeID: group.recipeID,
        stepID: group.stepID,
        stepIndex: group.stepIndex,
        index: group.index,
        options: group.options,
        selectedOptionIndex: nil,
        sourceRecipeID: group.sourceRecipeID,
        sourceRecipeName: group.sourceRecipeName
      )
    }
  }

  private func filterVesselOptionGroupsForDisplay(
    _ groups: [VesselOptionGroupAggregate]
  ) -> [VesselOptionGroupAggregate] {
    return groups.map { group in
      let selectedIndex: UInt32?
      // Check meal plan selections first
      if let selections = mealPlanSelections,
        let selection = selections.first(where: { sel in
          sel.recipeID == group.recipeID && sel.recipeStepID == group.stepID
            && sel.selectionType == .vessel
        })
      {
        selectedIndex = selection.selectedOptionIndex
      } else {
        // Check user selections - use sentinel value if none
        let userSelection = selectedVesselOptions[group.id]
        selectedIndex = userSelection == UInt32.max ? nil : userSelection
      }

      // If a selection has been made, filter to show only that option
      if let selectedIndex = selectedIndex {
        let filteredOptions = group.options.filter { $0.optionIndex == selectedIndex }
        return VesselOptionGroupAggregate(
          id: group.id,
          recipeID: group.recipeID,
          stepID: group.stepID,
          stepIndex: group.stepIndex,
          index: group.index,
          options: filteredOptions.isEmpty ? group.options : filteredOptions,
          selectedOptionIndex: selectedIndex,
          sourceRecipeID: group.sourceRecipeID,
          sourceRecipeName: group.sourceRecipeName
        )
      }

      // No selection - show all options
      return VesselOptionGroupAggregate(
        id: group.id,
        recipeID: group.recipeID,
        stepID: group.stepID,
        stepIndex: group.stepIndex,
        index: group.index,
        options: group.options,
        selectedOptionIndex: nil,
        sourceRecipeID: group.sourceRecipeID,
        sourceRecipeName: group.sourceRecipeName
      )
    }
  }

  // MARK: - Prep Tasks Section

  private func prepTasksSection(recipe: Mealplanning_Recipe, viewModel: PerformRecipeViewModel)
    -> some View
  {
    return VStack(alignment: .leading, spacing: 0) {
      Button(
        action: {
          withAnimation {
            isPrepTasksExpanded.toggle()
          }
        },
        label: {
          HStack {
            Text("Prep Tasks")
              .font(.headline)
              .foregroundColor(.primary)
            Spacer()
            Image(systemName: isPrepTasksExpanded ? "chevron.down" : "chevron.right")
              .font(.caption)
              .foregroundColor(.secondary)
          }
          .padding()
          .background(Color(.systemGray6))
        }
      )
      .buttonStyle(.plain)

      if isPrepTasksExpanded {
        VStack(alignment: .leading, spacing: 8) {
          if viewModel.isLoadingPrepTasks {
            HStack {
              ProgressView()
                .scaleEffect(0.8)
              Text("Loading prep tasks...")
                .font(.subheadline)
                .foregroundColor(.secondary)
            }
            .padding()
          } else if viewModel.prepTasks.isEmpty {
            Text("No prep tasks available for this recipe")
              .font(.subheadline)
              .foregroundColor(.secondary)
              .padding()
          } else {
            ForEach(viewModel.prepTasks, id: \.id) { prepTask in
              prepTaskRow(prepTask: prepTask, recipe: recipe, viewModel: viewModel)
            }
          }
        }
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
      }
    }
    .background(Color(.systemGray6))
    .cornerRadius(12)
  }

  private func prepTaskRow(
    prepTask: Mealplanning_RecipePrepTask,
    recipe: Mealplanning_Recipe,
    viewModel: PerformRecipeViewModel
  ) -> some View {
    // Get step IDs for this prep task
    let stepIDs = Set(
      prepTask.taskSteps.compactMap { taskStep in
        taskStep.belongsToRecipeStep.isEmpty ? nil : taskStep.belongsToRecipeStep
      })

    let hasSteps = !stepIDs.isEmpty
    let isCompleted = viewModel.isPrepTaskCompleted(prepTask.id)

    return VStack(alignment: .leading, spacing: 8) {
      HStack(spacing: 12) {
        // Checkbox for prep task completion
        Button {
          viewModel.togglePrepTask(prepTask)
        } label: {
          Image(systemName: isCompleted ? "checkmark.circle.fill" : "circle")
            .font(.title3)
            .foregroundColor(isCompleted ? .green : .secondary)
        }
        .buttonStyle(.plain)

        // Prep task content (clickable if it has steps)
        if hasSteps {
          NavigationLink(
            destination: PerformRecipeView(
              recipeID: recipe.id,
              highlightedStepIDs: stepIDs
            )
            .environment(authManager)
          ) {
            prepTaskContent(prepTask: prepTask, stepIDs: stepIDs, isCompleted: isCompleted)
          }
          .buttonStyle(.plain)
        } else {
          prepTaskContent(prepTask: prepTask, stepIDs: stepIDs, isCompleted: isCompleted)
        }
      }
    }
    .padding(.horizontal)
    .padding(.vertical, 8)
  }

  private func prepTaskContent(
    prepTask: Mealplanning_RecipePrepTask,
    stepIDs: Set<String>,
    isCompleted: Bool
  ) -> some View {
    VStack(alignment: .leading, spacing: 4) {
      HStack {
        Text(prepTask.name.isEmpty ? "Unnamed Prep Task" : prepTask.name)
          .font(.subheadline)
          .fontWeight(.semibold)
          .foregroundColor(isCompleted ? .secondary : .primary)
          .strikethrough(isCompleted)

        Spacer()

        if !stepIDs.isEmpty {
          Image(systemName: "chevron.right")
            .font(.caption)
            .foregroundColor(.secondary)
        }
      }

      if !prepTask.description_p.isEmpty {
        Text(prepTask.description_p)
          .font(.caption)
          .foregroundColor(isCompleted ? .secondary : .secondary)
          .strikethrough(isCompleted)
      }

      if !prepTask.notes.isEmpty {
        Text(prepTask.notes)
          .font(.caption2)
          .foregroundColor(.secondary)
          .italic()
          .strikethrough(isCompleted)
      }

      if !stepIDs.isEmpty {
        Text("\(stepIDs.count) step\(stepIDs.count == 1 ? "" : "s")")
          .font(.caption2)
          .foregroundColor(isCompleted ? .green : .blue)
          .padding(.top, 2)
      }

      if prepTask.optional {
        Label("Optional", systemImage: "info.circle")
          .font(.caption2)
          .foregroundColor(.orange)
      }
    }
  }

  // MARK: - Steps List

  private static func stepOrderKey(recipeID: String, group: String) -> String {
    "stepOrder_recipe_\(recipeID)_\(group)"
  }

  private func loadStepOrderFromUserDefaults(recipeID: String) {
    let upNext =
      UserDefaults.standard.stringArray(
        forKey: Self.stepOrderKey(recipeID: recipeID, group: "upNext")) ?? []
    let forLater =
      UserDefaults.standard.stringArray(
        forKey: Self.stepOrderKey(recipeID: recipeID, group: "forLater")) ?? []
    customUpNextOrder = upNext
    customForLaterOrder = forLater
  }

  private func saveStepOrderToUserDefaults(recipeID: String, upNext: [String], forLater: [String]) {
    UserDefaults.standard.set(
      upNext, forKey: Self.stepOrderKey(recipeID: recipeID, group: "upNext"))
    UserDefaults.standard.set(
      forLater, forKey: Self.stepOrderKey(recipeID: recipeID, group: "forLater"))
  }

  private func applyOrder<T: Identifiable>(_ items: [T], order: [String]) -> [T]
  where T.ID == String {
    guard !order.isEmpty else { return items }
    return items.sorted { firstItem, secondItem in
      let firstIndex = order.firstIndex(of: firstItem.id) ?? Int.max
      let secondIndex = order.firstIndex(of: secondItem.id) ?? Int.max
      return firstIndex < secondIndex
    }
  }

  private func shouldShowStep(stepID: String) -> Bool {
    guard let highlightedStepIDs = highlightedStepIDs else { return true }
    if showAllStepsFromPrepTask { return true }
    return highlightedStepIDs.contains(stepID)
  }

  // Step info for categorization
  private struct StepInfo: Identifiable {
    var id: String {
      "\(recipeID):\(step.id)"
    }
    let step: Mealplanning_RecipeStep
    let index: Int
    let recipeID: String
    let isAssociatedRecipeStep: Bool
    let associatedRecipeName: String?
  }

  // Collect all steps from recipe and associated recipes
  private func collectAllSteps(recipe: Mealplanning_Recipe) -> [StepInfo] {
    var allSteps: [StepInfo] = []

    // Collect steps from associated recipes
    for associatedRecipe in recipe.associatedRecipes {
      for (index, step) in associatedRecipe.steps.enumerated() where shouldShowStep(stepID: step.id)
      {
        allSteps.append(
          StepInfo(
            step: step,
            index: index,
            recipeID: associatedRecipe.id,
            isAssociatedRecipeStep: true,
            associatedRecipeName: associatedRecipe.name
          ))
      }
    }

    // Collect steps from main recipe
    for (index, step) in recipe.steps.enumerated() where shouldShowStep(stepID: step.id) {
      allSteps.append(
        StepInfo(
          step: step,
          index: index,
          recipeID: recipe.id,
          isAssociatedRecipeStep: false,
          associatedRecipeName: nil
        ))
    }

    return allSteps
  }

  private func stepsList(
    recipe: Mealplanning_Recipe, viewModel: PerformRecipeViewModel, scale: Float
  )
    -> some View
  {
    let allSteps = collectAllSteps(recipe: recipe)

    // Categorize steps
    let upNextSteps = allSteps.filter { stepInfo in
      viewModel.categorizeStep(recipeID: stepInfo.recipeID, stepID: stepInfo.step.id) == .upNext
    }

    let forLaterSteps = allSteps.filter { stepInfo in
      viewModel.categorizeStep(recipeID: stepInfo.recipeID, stepID: stepInfo.step.id) == .forLater
    }

    let orderedUpNext = applyOrder(upNextSteps, order: customUpNextOrder)
    let orderedForLater = applyOrder(forLaterSteps, order: customForLaterOrder)

    let focusedGroups = [
      StepFlowGroup(title: "Up Next", color: .orange, items: orderedUpNext),
      StepFlowGroup(title: "Not Yet", color: .blue, items: orderedForLater),
    ]

    return StepFlowSection(
      showCompleted: Binding(
        get: { isShowingCompletedSteps },
        set: { setShowingCompletedSteps($0) }
      ),
      allSteps: allSteps,
      focusedGroups: focusedGroups,
      allowToggle: allowCompletedStepsToggle,
      allStepsTitle: "All Steps",
      showStepsOverlay: !sharedWashHandsValue
        && (showWashHandsStepCard || sharedWashHandsCompleted != nil),
      onReorder: { groupId, source, destination in
        let items = groupId == "Up Next" ? orderedUpNext : orderedForLater
        var mutable = items
        mutable.move(fromOffsets: source, toOffset: destination)
        let newOrder = mutable.map(\.id)
        if groupId == "Up Next" {
          customUpNextOrder = newOrder
        } else {
          customForLaterOrder = newOrder
        }
        saveStepOrderToUserDefaults(
          recipeID: recipe.id,
          upNext: customUpNextOrder,
          forLater: customForLaterOrder
        )
      },
      headerContent: {
        HStack(spacing: 8) {
          Text("Steps")
            .font(.headline)
            .padding(.horizontal, 4)
          if highlightedStepIDs != nil {
            Button(showAllStepsFromPrepTask ? "Task only" : "Show all steps") {
              eventReporterService.reporter.track(
                event: "perform_recipe_show_all_steps_toggled",
                properties: ["showing_all": !showAllStepsFromPrepTask])
              withAnimation { showAllStepsFromPrepTask.toggle() }
            }
            .font(.caption)
            .buttonStyle(.bordered)
            .controlSize(.small)
          }
        }
      },
      allModeLeadingContent: {
        VStack(alignment: .leading, spacing: 12) {
          DSKeepScreenAwakeButton(inline: true) {
            eventReporterService.reporter.track(
              event: "perform_recipe_keep_awake_toggled",
              properties: ["enabled": $0])
          }
          if showWashHandsStepCard, !sharedWashHandsValue, allSteps.first != nil {
            washHandsStepCard(viewModel: viewModel)
          }
        }
      },
      focusModeLeadingContent: {
        VStack(alignment: .leading, spacing: 12) {
          DSKeepScreenAwakeButton(inline: true) {
            eventReporterService.reporter.track(
              event: "perform_recipe_keep_awake_toggled",
              properties: ["enabled": $0])
          }
          if showWashHandsStepCard, !sharedWashHandsValue, upNextSteps.first != nil {
            washHandsStepCard(viewModel: viewModel)
          }
        }
      },
      rowContent: { stepInfo in
        StepCardView(
          step: stepInfo.step,
          index: stepInfo.index,
          viewModel: viewModel,
          formatStepTitle: formatStepTitle,
          recipeID: stepInfo.recipeID,
          mealPlanSelections: mealPlanSelections,
          isAssociatedRecipeStep: stepInfo.isAssociatedRecipeStep,
          associatedRecipeName: stepInfo.associatedRecipeName,
          highlightedStepIDs: highlightedStepIDs,
          selectedIngredientOptions: selectedIngredientOptions,
          selectedInstrumentOptions: selectedInstrumentOptions,
          selectedVesselOptions: selectedVesselOptions,
          scale: scale,
          onStepCompleted: onStepCompleted
        )
      }
    )
  }

  // MARK: - Associated Recipe Steps Header

  private func associatedRecipeStepsHeader(recipe: Mealplanning_Recipe) -> some View {
    VStack(alignment: .leading, spacing: 4) {
      HStack(spacing: 8) {
        Text("PREREQUISITE:")
          .font(.caption2)
          .fontWeight(.semibold)
          .foregroundColor(.purple)
          .textCase(.uppercase)
          .tracking(0.5)

        NavigationLink(
          destination: {
            PerformRecipeView(recipeID: recipe.id)
              .environment(authManager)
          },
          label: {
            Text(recipe.name)
              .font(.subheadline)
              .fontWeight(.semibold)
              .foregroundColor(.purple)
          }
        )
      }

      if !recipe.description_p.isEmpty {
        Text(recipe.description_p)
          .font(.caption)
          .foregroundColor(.secondary)
      }
    }
    .padding()
    .frame(maxWidth: .infinity, alignment: .leading)
    .background(Color.purple.opacity(0.1))
    .cornerRadius(8)
    .overlay(
      RoundedRectangle(cornerRadius: 8)
        .stroke(Color.purple.opacity(0.3), lineWidth: 2)
    )
    .padding(.vertical, 4)
  }

  // MARK: - Main Recipe Steps Header

  private func mainRecipeStepsHeader() -> some View {
    HStack {
      Text("MAIN RECIPE STEPS")
        .font(.caption2)
        .fontWeight(.semibold)
        .foregroundColor(.blue)
        .textCase(.uppercase)
        .tracking(0.5)
    }
    .padding()
    .frame(maxWidth: .infinity, alignment: .leading)
    .background(Color.blue.opacity(0.1))
    .cornerRadius(8)
    .overlay(
      RoundedRectangle(cornerRadius: 8)
        .stroke(Color.blue.opacity(0.3), lineWidth: 2)
    )
    .padding(.vertical, 4)
  }

  // MARK: - Wash Hands Step Card

  private func washHandsStepCard(viewModel: PerformRecipeViewModel) -> some View {
    let isCompleted = sharedWashHandsValue
    let canCheck = viewModel.canCheckStep(PerformRecipeViewModel.washHandsStepIndex)

    return VStack(alignment: .leading, spacing: 12) {
      // Step header with checkbox
      HStack(alignment: .top, spacing: 12) {
        // Checkbox
        Button(
          action: {
            toggleWashHandsStep()
          },
          label: {
            Image(systemName: isCompleted ? "checkmark.circle.fill" : "circle")
              .font(.title2)
              .foregroundColor(
                canCheck ? (isCompleted ? .green : .blue) : .gray
              )
          }
        )
        .disabled(!canCheck)

        // Step title
        VStack(alignment: .leading, spacing: 4) {
          HStack(spacing: 6) {
            Image(systemName: "hands.sparkles")
              .font(.headline)
            Text("Wash your hands")
              .font(.headline)
          }
          .foregroundColor(isCompleted ? .secondary : .primary)
          .italic(isCompleted)
        }

        Spacer()
      }
    }
    .padding()
    .background(
      isCompleted ? Color(.systemGray6) : Color(.systemBackground)
    )
    .cornerRadius(12)
    .overlay(
      RoundedRectangle(cornerRadius: 12)
        .stroke(
          isCompleted ? Color.green.opacity(0.3) : Color.clear,
          lineWidth: 2
        )
    )
  }

  private var sharedWashHandsValue: Bool {
    sharedWashHandsCompleted?.wrappedValue ?? viewModel.washHandsCompleted
  }

  private func toggleWashHandsStep() {
    if let sharedWashHandsCompleted {
      let nextValue = !sharedWashHandsCompleted.wrappedValue
      sharedWashHandsCompleted.wrappedValue = nextValue
      viewModel.washHandsCompleted = nextValue
      if !nextValue {
        viewModel.completedSteps.removeAll()
      }
      return
    }

    viewModel.toggleStep(PerformRecipeViewModel.washHandsStepIndex)
  }

}
