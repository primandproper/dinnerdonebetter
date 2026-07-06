//
//  MealDetailView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftProtobuf
import SwiftUI

struct MealDetailView: View {
  private enum MealFlowMode: String, CaseIterable, Identifiable {
    case unified = "Unified"
    case byComponent = "By Component"

    var id: String { rawValue }
  }

  private static let componentTagColors: [Color] = [
    .indigo, .green, .yellow,
    .purple, .orange, .blue,
    .mint, .red, .teal,
    .pink, .cyan, .brown,
    Color(red: 0.85, green: 0.20, blue: 0.20),
    Color(red: 0.90, green: 0.45, blue: 0.10),
    Color(red: 0.75, green: 0.65, blue: 0.15),
    Color(red: 0.25, green: 0.65, blue: 0.25),
    Color(red: 0.15, green: 0.70, blue: 0.55),
    Color(red: 0.15, green: 0.65, blue: 0.70),
    Color(red: 0.20, green: 0.45, blue: 0.85),
    Color(red: 0.30, green: 0.35, blue: 0.85),
    Color(red: 0.45, green: 0.25, blue: 0.80),
    Color(red: 0.70, green: 0.20, blue: 0.70),
    Color(red: 0.85, green: 0.20, blue: 0.50),
    Color(red: 0.60, green: 0.35, blue: 0.20),
    Color(red: 0.55, green: 0.20, blue: 0.20),
    Color(red: 0.60, green: 0.35, blue: 0.10),
    Color(red: 0.55, green: 0.50, blue: 0.10),
    Color(red: 0.20, green: 0.55, blue: 0.20),
    Color(red: 0.10, green: 0.55, blue: 0.45),
    Color(red: 0.15, green: 0.45, blue: 0.70),
    Color(red: 0.25, green: 0.25, blue: 0.65),
    Color(red: 0.55, green: 0.20, blue: 0.60),
  ]

  @Environment(AuthenticationManager.self) private var authManager
  @State private var viewModel: MealDetailViewModel?
  @State private var loadedRecipes: [String: (recipe: Mealplanning_Recipe, scale: Float)] = [:]
  // Maps component recipe ID to base scale from meal
  @State private var baseComponentScales: [String: Float] = [:]
  // Meal-level scale multiplier
  @State private var mealScale: Float = 1.0
  @State private var mealScaleText: String = "1.0"
  @FocusState private var isMealScaleFocused: Bool
  @State private var isInstrumentsVesselsExpanded = false
  @State private var isIngredientsExpanded = false
  @State private var checkedIngredients: Set<String> = []
  @State private var checkedInstrumentsVessels: Set<String> = []
  @State private var componentViewModels: [String: PerformRecipeViewModel] = [:]
  /// Dismisses wash-hands reminder when user completes any step in this session
  @State private var hasDismissedWashHandsReminder = false
  @State private var mealFlowMode: MealFlowMode = .unified
  @State private var showMealCompletedSteps = false
  @State private var customUpNextOrder: [String] = []
  @State private var customForLaterOrder: [String] = []
  @State private var isDAGSectionExpanded = false
  @State private var showDiagramFullScreen = false
  @State private var mealTimerTick: Int = 0
  @State private var mealTimerRefresh: Timer?
  @State private var completedStepsFromTasks: [String: Set<String>] = [:]
  let mealID: String
  /// When true, scale editing is hidden because the scale is set by the meal plan.
  let isFromMealPlan: Bool
  /// When viewing from meal plan, the plan's scale for this meal (e.g. 4.0 for 4x). Used to display scaled portions.
  let mealPlanScale: Float?
  /// When viewing from meal plan, the meal plan ID for fetching completed tasks.
  let mealPlanID: String?
  /// When viewing from meal plan, the meal plan option ID for filtering tasks.
  let mealPlanOptionID: String?

  init(
    mealID: String,
    isFromMealPlan: Bool = false,
    mealPlanScale: Float? = nil,
    mealPlanID: String? = nil,
    mealPlanOptionID: String? = nil
  ) {
    self.mealID = mealID
    self.isFromMealPlan = isFromMealPlan
    self.mealPlanScale = mealPlanScale
    self.mealPlanID = mealPlanID
    self.mealPlanOptionID = mealPlanOptionID
  }

  var body: some View {
    Group {
      if let viewModel = viewModel {
        DSContentState(
          isLoading: viewModel.isLoading,
          loadingMessage: "Loading meal...",
          error: viewModel.errorMessage,
          onRetry: { await viewModel.loadMeal() },
          content: {
            if let meal = viewModel.meal {
              ScrollView {
                VStack(alignment: .leading, spacing: DSTheme.Spacing.xl) {
                  // Overall Info Section
                  overallInfoSection(meal: meal)

                  // Equipment, Ingredients & Graph (collapsibles with uniform spacing)
                  if !meal.components.isEmpty {
                    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
                      aggregatedInstrumentsVesselsSection
                      aggregatedIngredientsSection
                      mealDAGSection(viewModel: viewModel)
                    }
                  }

                  // Keep screen awake
                  if !meal.components.isEmpty {
                    DSKeepScreenAwakeButton(inline: true)
                  }

                  if !meal.components.isEmpty {
                    madeAheadSection(meal: meal)
                  }

                  if !meal.components.isEmpty {
                    mealFlowModeSection
                    if mealFlowMode == .unified {
                      unifiedMealStepsSection(meal: meal)
                        .id(mealTimerTick)
                    } else {
                      componentsSection(meal: meal)
                    }
                  }
                }
                .dsScreenPadding()
              }
            } else {
              DSEmptyState(
                icon: "fork.knife",
                title: "Meal not found",
                message: "This meal could not be loaded."
              )
            }
          })
      } else {
        DSInitializingView()
      }
    }
    .navigationTitle(viewModel?.meal?.name ?? "Meal")
    .navigationBarTitleDisplayMode(.large)
    .onAppear {
      if viewModel == nil {
        viewModel = MealDetailViewModel(mealID: mealID, authManager: authManager)
      }
      // When viewing from meal plan, apply the plan's scale so portions and ingredients reflect it
      if isFromMealPlan, let scale = mealPlanScale, scale > 0 {
        mealScale = scale
        mealScaleText = String(format: "%.2f", mealScale)
      }
      if let viewModel = viewModel {
        Task {
          await viewModel.loadMeal()
          // Initialize base scales from meal components after meal loads
          if let meal = viewModel.meal, baseComponentScales.isEmpty {
            for component in meal.components {
              baseComponentScales[component.recipe.id] = component.recipeScale
            }
          }
          if let meal = viewModel.meal {
            ensureComponentDataLoaded(for: meal)
            loadStepOrderFromUserDefaults(mealID: meal.id)
            if isFromMealPlan, let planID = mealPlanID {
              await loadCompletedStepsFromTasks(mealPlanID: planID, meal: meal)
            }
          }
        }
      }
      if hasActiveMealStepTimers {
        startMealTimerRefresh()
      }
    }
    .onChange(of: hasActiveMealStepTimers) { _, hasActive in
      if hasActive {
        startMealTimerRefresh()
      } else {
        stopMealTimerRefresh()
      }
    }
    .onDisappear {
      stopMealTimerRefresh()
    }
    .onChange(of: viewModel?.meal?.id) { _, _ in
      if let meal = viewModel?.meal {
        ensureComponentDataLoaded(for: meal)
        loadStepOrderFromUserDefaults(mealID: meal.id)
      }
    }
  }

  // MARK: - Overall Info Section

  private func overallInfoSection(meal: Mealplanning_Meal) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      // Description
      if !meal.description_p.isEmpty {
        Text(meal.description_p)
          .font(DSTheme.Typography.body)
          .foregroundColor(DSTheme.Colors.textSecondary)
      }

      // Estimated portions (scaled by meal scale when applicable, e.g. from meal plan)
      HStack(spacing: DSTheme.Spacing.sm) {
        Image(systemName: "person.2")
          .foregroundColor(DSTheme.Colors.textSecondary)
        Text(
          "Estimated Portions: \(PortionsFormatter.formatScaled(min: meal.minEstimatedPortions, max: meal.hasMaxEstimatedPortions ? meal.maxEstimatedPortions : nil, scale: mealScale))"
        )
        .font(DSTheme.Typography.body)
        .foregroundColor(DSTheme.Colors.textSecondary)
      }

      // Meal Scale Control (hidden when viewing from meal plan – scale is set by the plan)
      if !isFromMealPlan {
        DSDivider()
          .padding(.vertical, DSTheme.Spacing.xs)

        HStack(spacing: DSTheme.Spacing.md) {
          Text("Meal Scale:")
            .font(DSTheme.Typography.label)

          HStack(spacing: DSTheme.Spacing.sm) {
            TextField("1.0", text: $mealScaleText)
              .keyboardType(.decimalPad)
              .textFieldStyle(.roundedBorder)
              .frame(width: 80)
              .focused($isMealScaleFocused)
              .onSubmit {
                updateMealScaleFromText()
              }
              .onChange(of: isMealScaleFocused) { _, isFocused in
                if !isFocused {
                  updateMealScaleFromText()
                }
              }
              .onChange(of: mealScaleText) { _, newValue in
                // Filter to only allow numbers and a single decimal point
                var filtered = newValue.filter { $0.isNumber || $0 == "." }
                // Ensure only one decimal point
                let parts = filtered.split(separator: ".", omittingEmptySubsequences: false)
                if parts.count > 2 {
                  filtered = parts[0] + "." + parts.dropFirst().joined()
                }
                if filtered != newValue {
                  mealScaleText = filtered
                }
              }

            Text("x")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)

            DSStepperButtons(
              onDecrement: { adjustMealScale(by: -0.25) },
              onIncrement: { adjustMealScale(by: 0.25) }
            )
          }
        }

        Slider(
          value: Binding(
            get: { Double(mealScale) },
            set: { setMealScale(Float($0)) }
          ),
          in: 0.25...4.0,
          step: 0.25
        )
      }
    }
    .padding(DSTheme.Spacing.lg)
    .frame(maxWidth: .infinity, alignment: .leading)
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.lg)
  }

  // MARK: - Meal DAG Section

  private func mealDAGSection(viewModel: MealDetailViewModel) -> some View {
    VStack(alignment: .leading, spacing: 0) {
      Button(
        action: {
          withAnimation {
            isDAGSectionExpanded.toggle()
          }
        },
        label: {
          HStack {
            Text("Meal Graph")
              .font(DSTheme.Typography.title3)
              .foregroundColor(DSTheme.Colors.textPrimary)
            Spacer()
            if let mermaidSource = viewModel.mermaidDiagram, !mermaidSource.isEmpty {
              Button {
                showDiagramFullScreen = true
              } label: {
                Image(systemName: "arrow.up.left.and.arrow.down.right")
                  .font(DSTheme.Typography.caption)
                  .foregroundColor(DSTheme.Colors.textSecondary)
              }
              .buttonStyle(.plain)
            }
            Image(systemName: isDAGSectionExpanded ? "chevron.down" : "chevron.right")
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
          .padding(DSTheme.Spacing.lg)
          .background(DSTheme.Colors.cardBackground)
        }
      )
      .buttonStyle(.plain)
      .fullScreenCover(isPresented: $showDiagramFullScreen) {
        DiagramFullScreenSheet(
          mermaidSource: viewModel.mermaidDiagram ?? "",
          title: "Meal Graph",
          onDismiss: { showDiagramFullScreen = false }
        )
      }

      if isDAGSectionExpanded {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          if viewModel.isLoadingMermaid {
            HStack {
              ProgressView()
                .scaleEffect(0.8)
              Text("Loading diagram...")
                .font(DSTheme.Typography.body)
                .foregroundColor(DSTheme.Colors.textSecondary)
            }
            .frame(maxWidth: .infinity)
            .padding(DSTheme.Spacing.lg)
          } else if let error = viewModel.mermaidError {
            VStack(spacing: DSTheme.Spacing.md) {
              Text(error)
                .font(DSTheme.Typography.body)
                .foregroundColor(DSTheme.Colors.textSecondary)
                .multilineTextAlignment(.center)
              DSButton("Retry", icon: "arrow.clockwise", style: .primary, size: .small) {
                Task {
                  await viewModel.loadMermaidDiagram()
                }
              }
            }
            .frame(maxWidth: .infinity)
            .padding(DSTheme.Spacing.lg)
          } else if let mermaidSource = viewModel.mermaidDiagram, !mermaidSource.isEmpty {
            MermaidDiagramView(source: mermaidSource)
              .padding(DSTheme.Spacing.lg)
          } else {
            Text("No diagram available for this meal")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)
              .frame(maxWidth: .infinity)
              .padding(DSTheme.Spacing.lg)
          }
        }
        .padding(.vertical, DSTheme.Spacing.md)
        .background(Color(.systemBackground))
        .onAppear {
          Task {
            await viewModel.loadMermaidDiagram()
          }
        }
      }
    }
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.lg)
  }

  private func updateMealScaleFromText() {
    if let scale = Float(mealScaleText), scale > 0 {
      setMealScale(scale)
    } else {
      mealScaleText = String(format: "%.2f", mealScale)
    }
  }

  private func setMealScale(_ newScale: Float) {
    mealScale = min(max(newScale, 0.25), 4.0)
    mealScaleText = String(format: "%.2f", mealScale)
    // Update loadedRecipes with new scales
    for (recipeID, baseScale) in baseComponentScales {
      let newScale = baseScale * mealScale
      if let recipeData = loadedRecipes[recipeID] {
        loadedRecipes[recipeID] = (recipe: recipeData.recipe, scale: newScale)
      }
    }
  }

  private func adjustMealScale(by delta: Float) {
    setMealScale(mealScale + delta)
  }

  private func formatComponentType(_ type: Mealplanning_MealComponentType) -> String {
    switch type {
    case .amuseBouche:
      return "Amuse Bouche"
    case .appetizer:
      return "Appetizer"
    case .soup:
      return "Soup"
    case .main:
      return "Main"
    case .salad:
      return "Salad"
    case .beverage:
      return "Beverage"
    case .side:
      return "Side"
    case .dessert:
      return "Dessert"
    default:
      return ""
    }
  }

  private var washHandsReminderText: some View {
    HStack(spacing: DSTheme.Spacing.sm) {
      Image(systemName: "hands.sparkles")
        .font(DSTheme.Typography.body)
        .foregroundColor(DSTheme.Colors.textSecondary)
      Text("Wash your hands before cooking.")
        .font(DSTheme.Typography.body)
        .foregroundColor(DSTheme.Colors.textPrimary)
    }
    .padding(.vertical, DSTheme.Spacing.md)
    .padding(.horizontal, DSTheme.Spacing.sm)
    .frame(maxWidth: .infinity, alignment: .leading)
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.sm)
  }

  private var mealFlowModeSection: some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
      Text("Step View")
        .font(DSTheme.Typography.label)
        .foregroundColor(DSTheme.Colors.textSecondary)

      Picker("Step View", selection: $mealFlowMode) {
        ForEach(MealFlowMode.allCases) { mode in
          Text(mode.rawValue).tag(mode)
        }
      }
      .pickerStyle(.segmented)
    }
  }

  private func madeAheadSection(meal: Mealplanning_Meal) -> some View {
    let uniqueAssociated = uniqueAssociatedRecipesFromLoaded
    guard !uniqueAssociated.isEmpty else {
      return AnyView(EmptyView())
    }

    return AnyView(
      VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
        Text("Made Ahead")
          .font(DSTheme.Typography.title2)

        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          ForEach(uniqueAssociated, id: \.id) { associatedRecipe in
            madeAheadRow(associatedRecipe: associatedRecipe)
          }
        }
      }
    )
  }

  private var uniqueAssociatedRecipesFromLoaded: [Mealplanning_Recipe] {
    var byID: [String: Mealplanning_Recipe] = [:]
    for (_, recipeData) in loadedRecipes {
      for assoc in recipeData.recipe.associatedRecipes {
        byID[assoc.id] = assoc
      }
    }
    return Array(byID.values).sorted {
      ($0.name.isEmpty ? "Unnamed" : $0.name) < ($1.name.isEmpty ? "Unnamed" : $1.name)
    }
  }

  private func madeAheadRow(associatedRecipe: Mealplanning_Recipe) -> some View {
    let allComplete = isAssociatedRecipeFullyComplete(associatedRecipe)
    return HStack {
      Text(associatedRecipe.name.isEmpty ? "Unnamed Recipe" : associatedRecipe.name)
        .font(DSTheme.Typography.body)
        .foregroundColor(DSTheme.Colors.textPrimary)

      Spacer()

      Button(
        action: {
          UIImpactFeedbackGenerator(style: .light).impactOccurred()
          markAssociatedRecipeAsMadeInAllComponents(associatedRecipe: associatedRecipe)
        },
        label: {
          HStack(spacing: DSTheme.Spacing.xs) {
            Image(systemName: allComplete ? "checkmark.circle.fill" : "checkmark.circle")
              .foregroundColor(allComplete ? .green : DSTheme.Colors.primary)
            Text("I made this")
              .font(DSTheme.Typography.label)
              .foregroundColor(allComplete ? .green : DSTheme.Colors.primary)
          }
        }
      )
      .buttonStyle(.plain)
      .disabled(allComplete)
    }
    .padding(DSTheme.Spacing.md)
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.md)
  }

  private func isAssociatedRecipeFullyComplete(_ associatedRecipe: Mealplanning_Recipe) -> Bool {
    for (componentID, recipeData) in loadedRecipes {
      guard recipeData.recipe.associatedRecipes.contains(where: { $0.id == associatedRecipe.id })
      else {
        continue
      }
      guard let vm = componentViewModels[componentID] else { continue }
      if associatedRecipe.steps.allSatisfy({
        vm.isStepCompleted(recipeID: associatedRecipe.id, stepID: $0.id)
      }) {
        return true
      }
    }
    return false
  }

  private func markAssociatedRecipeAsMadeInAllComponents(associatedRecipe: Mealplanning_Recipe) {
    for (componentID, recipeData) in loadedRecipes {
      guard recipeData.recipe.associatedRecipes.contains(where: { $0.id == associatedRecipe.id })
      else {
        continue
      }
      componentViewModels[componentID]?.markAssociatedRecipeAsCompleted(
        associatedRecipe: associatedRecipe)
    }
    mealTimerTick += 1
  }

  private var aggregatedInstrumentsVesselsSection: some View {
    let aggregatedItems = getAggregatedInstrumentsAndVessels()

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
              .font(DSTheme.Typography.title3)
              .foregroundColor(DSTheme.Colors.textPrimary)
            Spacer()
            Image(systemName: isInstrumentsVesselsExpanded ? "chevron.down" : "chevron.right")
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
          .padding(DSTheme.Spacing.lg)
          .background(DSTheme.Colors.cardBackground)
        }
      )
      .buttonStyle(.plain)

      if isInstrumentsVesselsExpanded && !aggregatedItems.isEmpty {
        aggregatedInstrumentsVesselsList(aggregatedItems)
      }
    }
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.lg)
  }

  @ViewBuilder
  private func aggregatedInstrumentsVesselsList(_ items: [AggregatedInstrumentVessel]) -> some View
  {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
      ForEach(items, id: \.itemID) { item in
        aggregatedInstrumentVesselRow(item: item)
      }
    }
    .padding(.vertical, DSTheme.Spacing.sm)
    .background(Color(.systemBackground))
  }

  private func aggregatedInstrumentVesselRow(item: AggregatedInstrumentVessel) -> some View {
    HStack(spacing: DSTheme.Spacing.md) {
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
          .foregroundColor(
            checkedInstrumentsVessels.contains(item.itemID)
              ? DSTheme.Colors.success
              : DSTheme.Colors.textTertiary
          )
        }
      )
      .buttonStyle(.plain)

      HStack(spacing: DSTheme.Spacing.sm) {
        Image(
          systemName: item.type == .instrument
            ? "wrench.and.screwdriver" : "square.stack.3d.up"
        )
        .font(DSTheme.Typography.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
        .frame(width: 20)

        HStack {
          Text(item.name)
            .font(DSTheme.Typography.body)
            .foregroundColor(
              checkedInstrumentsVessels.contains(item.itemID)
                ? DSTheme.Colors.textSecondary
                : DSTheme.Colors.textPrimary
            )
            .strikethrough(checkedInstrumentsVessels.contains(item.itemID))

          if let sourceName = item.sourceRecipeName, !sourceName.isEmpty {
            Text("(from \(sourceName))")
              .font(DSTheme.Typography.captionSmall)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
        }
      }

      Spacer()
    }
    .padding(.horizontal, DSTheme.Spacing.lg)
    .padding(.vertical, DSTheme.Spacing.xs)
  }

  private static func stepOrderKey(mealID: String, group: String) -> String {
    "stepOrder_meal_\(mealID)_\(group)"
  }

  private func loadStepOrderFromUserDefaults(mealID: String) {
    let upNext =
      UserDefaults.standard.stringArray(forKey: Self.stepOrderKey(mealID: mealID, group: "upNext"))
      ?? []
    let forLater =
      UserDefaults.standard.stringArray(
        forKey: Self.stepOrderKey(mealID: mealID, group: "forLater")) ?? []
    customUpNextOrder = upNext
    customForLaterOrder = forLater
  }

  private func saveStepOrderToUserDefaults(mealID: String, upNext: [String], forLater: [String]) {
    UserDefaults.standard.set(upNext, forKey: Self.stepOrderKey(mealID: mealID, group: "upNext"))
    UserDefaults.standard.set(
      forLater, forKey: Self.stepOrderKey(mealID: mealID, group: "forLater"))
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

  private func unifiedMealStepsSection(meal: Mealplanning_Meal) -> some View {
    let allSteps = collectUnifiedMealStepsWithMerging(
      meal: meal,
      componentViewModels: componentViewModels,
      loadedRecipes: loadedRecipes,
      baseComponentScales: baseComponentScales,
      mealScale: mealScale,
      formatComponentType: formatComponentType
    )
    let upNext = allSteps.filter { $0.category == .upNext }
    let forLater = allSteps.filter { $0.category == .forLater }

    let orderedUpNext = applyOrder(upNext, order: customUpNextOrder)
    let orderedForLater = applyOrder(forLater, order: customForLaterOrder)

    let focusedGroups = [
      StepFlowGroup(title: "Up Next", color: DSTheme.Colors.warning, items: orderedUpNext),
      StepFlowGroup(title: "Not Yet", color: DSTheme.Colors.primary, items: orderedForLater),
    ]

    return StepFlowSection(
      showCompleted: $showMealCompletedSteps,
      allSteps: allSteps,
      focusedGroups: focusedGroups,
      allStepsTitle: "All Steps",
      emptyMessage: "Loading component steps...",
      showStepsOverlay: false,
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
          mealID: meal.id,
          upNext: customUpNextOrder,
          forLater: customForLaterOrder
        )
      },
      headerContent: {
        Text("Cook Flow")
          .font(DSTheme.Typography.title2)
      },
      allModeLeadingContent: {
        if !hasDismissedWashHandsReminder {
          washHandsReminderText
        }
      },
      focusModeLeadingContent: {
        if !hasDismissedWashHandsReminder {
          washHandsReminderText
        }
      },
      rowContent: { item in
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          let tagStyle = componentTagStyle(for: item.componentIndex)
          Text(item.componentNamesForTag)
            .font(DSTheme.Typography.captionSmall)
            .padding(.horizontal, DSTheme.Spacing.sm)
            .padding(.vertical, DSTheme.Spacing.xxs)
            .background(tagStyle.background)
            .foregroundColor(tagStyle.foreground)
            .cornerRadius(DSTheme.Radius.sm)

          StepCardView(
            step: item.step,
            index: item.stepIndex,
            viewModel: item.viewModel,
            formatStepTitle: { step, _ in
              formatUnifiedStepTitle(step: step, index: item.stepIndex)
            },
            recipeID: item.recipeID,
            isAssociatedRecipeStep: item.isAssociatedRecipeStep,
            associatedRecipeName: item.associatedRecipeName,
            scale: item.isMerged ? 1.0 : item.scale,
            isCompletedOverride: item.isMerged
              ? item.sources.allSatisfy {
                $0.viewModel.isStepCompleted(recipeID: $0.recipeID, stepID: $0.step.id)
              }
              : nil,
            canCheckOverride: item.isMerged
              ? (item.sources.allSatisfy { source in
                source.viewModel.canCheckStep(recipeID: source.recipeID, stepID: source.step.id)
              }
                || item.sources.contains { source in
                  source.viewModel.isStepTimerActive(
                    recipeID: source.recipeID, stepID: source.step.id)
                }
                || item.sources.allSatisfy { source in
                  source.viewModel.isStepCompleted(
                    recipeID: source.recipeID, stepID: source.step.id)
                })
              : nil,
            onToggleOverride: item.isMerged
              ? {
                for source in item.sources {
                  source.viewModel.toggleStep(recipeID: source.recipeID, stepID: source.step.id)
                }
              }
              : nil,
            onStepCompleted: { hasDismissedWashHandsReminder = true },
            ingredientBreakdownBySource: item.ingredientBreakdownBySource,
            timerElapsedSeconds: item.isMerged
              ? mergedStepTimerElapsed(sources: item.sources)
              : nil,
            timerDurationSeconds: item.isMerged
              ? mergedStepTimerDuration(sources: item.sources)
              : nil,
            onSkipTimerOverride: item.isMerged
              ? {
                for source in item.sources {
                  source.viewModel.skipStepTimer(
                    recipeID: source.recipeID, stepID: source.step.id)
                }
              }
              : nil,
            canSkipTimerOverride: item.isMerged
              ? item.sources.contains { source in
                source.viewModel.canSkipStepTimer(
                  recipeID: source.recipeID, stepID: source.step.id)
              }
              : nil,
            timerMinSeconds: item.isMerged
              ? mergedStepTimerMin(sources: item.sources)
              : nil,
            timerMaxSeconds: item.isMerged
              ? mergedStepTimerMax(sources: item.sources)
              : nil
          )
        }
      }
    )
  }

  private var hasActiveMealStepTimers: Bool {
    componentViewModels.values.contains { $0.hasActiveStepTimers }
  }

  private func startMealTimerRefresh() {
    guard mealTimerRefresh == nil else { return }
    mealTimerRefresh = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { _ in
      mealTimerTick += 1
    }
    if let timer = mealTimerRefresh {
      RunLoop.main.add(timer, forMode: .common)
    }
  }

  private func stopMealTimerRefresh() {
    mealTimerRefresh?.invalidate()
    mealTimerRefresh = nil
  }

  private func formatUnifiedStepTitle(step: Mealplanning_RecipeStep, index: Int) -> String {
    if step.hasPreparation, !step.preparation.name.isEmpty {
      return step.preparation.name
    }
    return "Step \(index + 1)"
  }

  private func mergedStepTimerElapsed(sources: [UnifiedMealStepSource]) -> TimeInterval? {
    var best: (elapsed: TimeInterval, duration: UInt32)?
    for source in sources {
      guard source.viewModel.isStepTimerActive(recipeID: source.recipeID, stepID: source.step.id),
        let elapsed = source.viewModel.stepTimerElapsedSeconds(
          recipeID: source.recipeID, stepID: source.step.id),
        let duration = source.viewModel.stepTimerDurationSeconds(
          recipeID: source.recipeID, stepID: source.step.id)
      else { continue }
      let remaining = Double(duration) - elapsed
      if best.map({ remaining > Double($0.duration) - $0.elapsed }) ?? true {
        best = (elapsed, duration)
      }
    }
    return best?.elapsed
  }

  private func mergedStepTimerDuration(sources: [UnifiedMealStepSource]) -> UInt32? {
    var best: (elapsed: TimeInterval, duration: UInt32)?
    for source in sources {
      guard source.viewModel.isStepTimerActive(recipeID: source.recipeID, stepID: source.step.id),
        let elapsed = source.viewModel.stepTimerElapsedSeconds(
          recipeID: source.recipeID, stepID: source.step.id),
        let duration = source.viewModel.stepTimerDurationSeconds(
          recipeID: source.recipeID, stepID: source.step.id)
      else { continue }
      let remaining = Double(duration) - elapsed
      if best.map({ remaining > Double($0.duration) - $0.elapsed }) ?? true {
        best = (elapsed, duration)
      }
    }
    return best?.duration
  }

  private func mergedStepTimerMin(sources: [UnifiedMealStepSource]) -> UInt32? {
    var best: (elapsed: TimeInterval, duration: UInt32)?
    var bestSource: UnifiedMealStepSource?
    for source in sources {
      guard source.viewModel.isStepTimerActive(recipeID: source.recipeID, stepID: source.step.id),
        let elapsed = source.viewModel.stepTimerElapsedSeconds(
          recipeID: source.recipeID, stepID: source.step.id),
        let duration = source.viewModel.stepTimerDurationSeconds(
          recipeID: source.recipeID, stepID: source.step.id)
      else { continue }
      let remaining = Double(duration) - elapsed
      if best.map({ remaining > Double($0.duration) - $0.elapsed }) ?? true {
        best = (elapsed, duration)
        bestSource = source
      }
    }
    return bestSource.flatMap {
      $0.viewModel.stepTimerMinSeconds(recipeID: $0.recipeID, stepID: $0.step.id)
    }
  }

  private func mergedStepTimerMax(sources: [UnifiedMealStepSource]) -> UInt32? {
    var best: (elapsed: TimeInterval, duration: UInt32)?
    var bestSource: UnifiedMealStepSource?
    for source in sources {
      guard source.viewModel.isStepTimerActive(recipeID: source.recipeID, stepID: source.step.id),
        let elapsed = source.viewModel.stepTimerElapsedSeconds(
          recipeID: source.recipeID, stepID: source.step.id),
        let duration = source.viewModel.stepTimerDurationSeconds(
          recipeID: source.recipeID, stepID: source.step.id)
      else { continue }
      let remaining = Double(duration) - elapsed
      if best.map({ remaining > Double($0.duration) - $0.elapsed }) ?? true {
        best = (elapsed, duration)
        bestSource = source
      }
    }
    return bestSource.flatMap {
      $0.viewModel.stepTimerMaxSeconds(recipeID: $0.recipeID, stepID: $0.step.id)
    }
  }

  private func componentTagStyle(for componentIndex: Int) -> (background: Color, foreground: Color)
  {
    let color = Self.componentTagColors[componentIndex % Self.componentTagColors.count]
    return (color.opacity(0.16), color)
  }

  private func ensureComponentDataLoaded(for meal: Mealplanning_Meal) {
    for component in meal.components {
      let recipeID = component.recipe.id
      let baseScale = baseComponentScales[recipeID] ?? component.recipeScale

      if let existingViewModel = componentViewModels[recipeID] {
        if let recipe = existingViewModel.recipe, loadedRecipes[recipeID] == nil {
          loadedRecipes[recipeID] = (recipe: recipe, scale: baseScale * mealScale)
        }
        continue
      }

      let recipeViewModel = PerformRecipeViewModel(recipeID: recipeID, authManager: authManager)
      recipeViewModel.washHandsCompleted = true  // No gate in meal plan view
      recipeViewModel.completedSteps = completedStepsFromTasks[recipeID] ?? []
      componentViewModels[recipeID] = recipeViewModel

      Task { @MainActor in
        await recipeViewModel.loadRecipe()
        if let recipe = recipeViewModel.recipe {
          loadedRecipes[recipeID] = (recipe: recipe, scale: baseScale * mealScale)
        }
      }
    }
  }

  @MainActor
  private func loadCompletedStepsFromTasks(mealPlanID: String, meal: Mealplanning_Meal) async {
    do {
      var request = Mealplanning_GetMealPlanTasksRequest()
      request.mealPlanID = mealPlanID

      let response = try await authManager.authenticatedCall("getMealPlanTasks", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMealPlanTasks(
          request, metadata: metadata, options: options)
      }

      var completedStepsByRecipe: [String: Set<String>] = [:]
      let mealRecipeIDs = Set(meal.components.map { $0.recipe.id })

      for task in response.results {
        guard task.status == .finished, task.hasRecipePrepTask else { continue }

        let prepTask = task.recipePrepTask
        let prepTaskRecipeID = prepTask.belongsToRecipe.isEmpty ? nil : prepTask.belongsToRecipe

        for taskStep in prepTask.taskSteps where !taskStep.belongsToRecipeStep.isEmpty {
          let stepID = taskStep.belongsToRecipeStep
          let recipeID: String? = {
            if let rid = prepTaskRecipeID, mealRecipeIDs.contains(rid) {
              return rid
            }
            return resolveStepToRecipe(stepID: stepID, meal: meal)?.recipeID
          }()
          if let recipeID = recipeID, mealRecipeIDs.contains(recipeID) {
            completedStepsByRecipe[recipeID, default: []].insert("\(recipeID):\(stepID)")
          }
        }
      }

      completedStepsFromTasks = completedStepsByRecipe

      for (recipeID, viewModel) in componentViewModels {
        if let steps = completedStepsByRecipe[recipeID] {
          viewModel.completedSteps.formUnion(steps)
        }
      }

      mealTimerTick += 1
    } catch {
      // Non-fatal: leave completed-steps state as-is
    }
  }

  private func resolveStepToRecipe(
    stepID: String,
    meal: Mealplanning_Meal
  ) -> (recipeID: String, stepID: String)? {
    for component in meal.components {
      let recipe = component.recipe
      if recipe.steps.contains(where: { $0.id == stepID }) {
        return (recipe.id, stepID)
      }
      for associatedRecipe in recipe.associatedRecipes
      where associatedRecipe.steps.contains(where: { $0.id == stepID }) {
        return (associatedRecipe.id, stepID)
      }
    }
    return nil
  }

  private var aggregatedIngredientsSection: some View {
    let aggregatedIngredients = getAggregatedIngredients()

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
              .font(DSTheme.Typography.title3)
              .foregroundColor(DSTheme.Colors.textPrimary)
            Spacer()
            Image(systemName: isIngredientsExpanded ? "chevron.down" : "chevron.right")
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
          .padding(DSTheme.Spacing.lg)
          .background(DSTheme.Colors.cardBackground)
        }
      )
      .buttonStyle(.plain)

      if isIngredientsExpanded && !aggregatedIngredients.isEmpty {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          ForEach(aggregatedIngredients, id: \.ingredientID) { aggregated in
            HStack(spacing: DSTheme.Spacing.md) {
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
                    checkedIngredients.contains(aggregated.ingredientID)
                      ? DSTheme.Colors.success
                      : DSTheme.Colors.textTertiary
                  )
                }
              )
              .buttonStyle(.plain)

              VStack(alignment: .leading, spacing: DSTheme.Spacing.xxs) {
                HStack {
                  if let quantityText = aggregated.quantityText {
                    Text(quantityText)
                      .font(DSTheme.Typography.label)
                      .foregroundColor(DSTheme.Colors.textSecondary)
                  }

                  Text(aggregated.name)
                    .font(DSTheme.Typography.body)
                    .foregroundColor(
                      checkedIngredients.contains(aggregated.ingredientID)
                        ? DSTheme.Colors.textSecondary
                        : DSTheme.Colors.textPrimary
                    )
                    .strikethrough(checkedIngredients.contains(aggregated.ingredientID))

                  if let sourceName = aggregated.sourceRecipeName, !sourceName.isEmpty {
                    Text("(from \(sourceName))")
                      .font(DSTheme.Typography.captionSmall)
                      .foregroundColor(DSTheme.Colors.textSecondary)
                  }
                }

                if !aggregated.quantityNotes.isEmpty {
                  Text(aggregated.quantityNotes)
                    .font(DSTheme.Typography.caption)
                    .foregroundColor(DSTheme.Colors.textSecondary)
                }
              }

              Spacer()
            }
            .padding(.horizontal, DSTheme.Spacing.lg)
            .padding(.vertical, DSTheme.Spacing.xs)
          }
        }
        .padding(.vertical, DSTheme.Spacing.sm)
        .background(Color(.systemBackground))
      }
    }
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.lg)
  }

}

// MARK: - Meal Detail View Extensions

extension MealDetailView {
  func getAggregatedInstrumentsAndVessels() -> [AggregatedInstrumentVessel] {
    var aggregated: [String: AggregatedInstrumentVessel] = [:]

    for (_, recipeData) in loadedRecipes {
      let recipe = recipeData.recipe
      let scale = recipeData.scale

      func processSteps(
        _ steps: [Mealplanning_RecipeStep], sourceRecipeID: String?, sourceRecipeName: String?
      ) {
        for step in steps {
          for instrument in step.instruments where instrument.hasInstrument {
            let validInstrument = instrument.instrument
            if validInstrument.displayInSummaryLists {
              let itemID = validInstrument.id
              if !itemID.isEmpty {
                if aggregated[itemID] == nil {
                  aggregated[itemID] = AggregatedInstrumentVessel(
                    itemID: itemID,
                    name: instrument.name,
                    type: .instrument,
                    sourceRecipeID: sourceRecipeID,
                    sourceRecipeName: sourceRecipeName
                  )
                }

                if var current = aggregated[itemID] {
                  let effectiveScale =
                    scale * (instrument.scaleFactor > 0 ? instrument.scaleFactor : 1.0)
                  let scaledMax: UInt32? =
                    instrument.hasMaxQuantity
                    ? (effectiveScale > 1
                      ? DiscreteQuantityScaling.scaledMax(
                        instrument.maxQuantity, scale: effectiveScale)
                      : instrument.maxQuantity)
                    : nil
                  current.addQuantity(min: instrument.minQuantity, max: scaledMax)
                  aggregated[itemID] = current
                }
              }
            }
          }

          for vessel in step.vessels where vessel.hasVessel {
            let validVessel = vessel.vessel
            if validVessel.displayInSummaryLists {
              let itemID = validVessel.id
              if !itemID.isEmpty {
                if aggregated[itemID] == nil {
                  aggregated[itemID] = AggregatedInstrumentVessel(
                    itemID: itemID,
                    name: vessel.name,
                    type: .vessel,
                    sourceRecipeID: sourceRecipeID,
                    sourceRecipeName: sourceRecipeName
                  )
                }

                if var current = aggregated[itemID] {
                  let effectiveScale = scale * (vessel.scaleFactor > 0 ? vessel.scaleFactor : 1.0)
                  let scaledMax: UInt32? =
                    vessel.hasMaxQuantity
                    ? (effectiveScale > 1
                      ? UInt32(
                        DiscreteQuantityScaling.scaledMax(
                          UInt16(vessel.maxQuantity), scale: effectiveScale))
                      : vessel.maxQuantity)
                    : nil
                  current.addQuantity(min: vessel.minQuantity, max: scaledMax)
                  aggregated[itemID] = current
                }
              }
            }
          }
        }
      }

      // Process associated (prerequisite) recipe steps first
      for associatedRecipe in recipe.associatedRecipes {
        processSteps(
          associatedRecipe.steps,
          sourceRecipeID: associatedRecipe.id,
          sourceRecipeName: associatedRecipe.name.isEmpty ? nil : associatedRecipe.name
        )
      }
      // Process main recipe steps
      processSteps(recipe.steps, sourceRecipeID: nil, sourceRecipeName: nil)
    }

    return Array(aggregated.values).sorted { $0.name < $1.name }
  }

  func getAggregatedIngredients() -> [AggregatedIngredient] {
    var aggregated: [String: AggregatedIngredient] = [:]

    for (_, recipeData) in loadedRecipes {
      let recipe = recipeData.recipe
      let scale = recipeData.scale

      func processSteps(
        _ steps: [Mealplanning_RecipeStep], sourceRecipeID: String?, sourceRecipeName: String?
      ) {
        for step in steps {
          for ingredient in step.ingredients where ingredient.hasIngredient {
            let validIngredient = ingredient.ingredient
            let key = validIngredient.id
            if !key.isEmpty {
              if aggregated[key] == nil {
                aggregated[key] = AggregatedIngredient(
                  ingredientID: key,
                  name: ingredient.name,
                  quantityNotes: ingredient.quantityNotes,
                  measurementUnit: ingredient.hasMeasurementUnit ? ingredient.measurementUnit : nil,
                  sourceRecipeID: sourceRecipeID,
                  sourceRecipeName: sourceRecipeName
                )
              }

              if var current = aggregated[key] {
                let effectiveScale =
                  scale * (ingredient.scaleFactor > 0 ? ingredient.scaleFactor : 1.0)
                let scaledMin = ingredient.minQuantity * effectiveScale
                let scaledMax: Float? =
                  ingredient.hasMaxQuantity
                  ? ingredient.maxQuantity * effectiveScale : nil
                current.addQuantity(min: scaledMin, max: scaledMax)
                aggregated[key] = current
              }
            }
          }
        }
      }

      // Process associated (prerequisite) recipe steps first
      for associatedRecipe in recipe.associatedRecipes {
        processSteps(
          associatedRecipe.steps,
          sourceRecipeID: associatedRecipe.id,
          sourceRecipeName: associatedRecipe.name.isEmpty ? nil : associatedRecipe.name
        )
      }
      // Process main recipe steps
      processSteps(recipe.steps, sourceRecipeID: nil, sourceRecipeName: nil)
    }

    return Array(aggregated.values).sorted { $0.name < $1.name }
  }

  private func componentsSection(meal: Mealplanning_Meal) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      Text("Components")
        .font(DSTheme.Typography.title2)

      if !hasDismissedWashHandsReminder {
        washHandsReminderText
      }

      ForEach(meal.components, id: \.recipe.id) { component in
        EmbeddedRecipeView(
          recipeID: component.recipe.id,
          recipeScale: Binding(
            get: {
              let baseScale = baseComponentScales[component.recipe.id] ?? component.recipeScale
              return baseScale * mealScale
            },
            set: { _ in
              // Scale is controlled at meal level, so we don't allow individual changes
            }
          ),
          componentType: component.componentType,
          sharedViewModel: componentViewModels[component.recipe.id],
          globalWashHandsCompleted: true,  // No gate in meal plan view
          sharedShowCompletedSteps: $showMealCompletedSteps,
          onGlobalWashHandsToggle: nil,
          onStepCompleted: { hasDismissedWashHandsReminder = true },
          onViewModelReady: { recipeViewModel in
            componentViewModels[component.recipe.id] = recipeViewModel
          },
          onRecipeLoaded: { recipe in
            Task { @MainActor in
              let baseScale = baseComponentScales[component.recipe.id] ?? component.recipeScale
              let currentScale = baseScale * mealScale
              loadedRecipes[component.recipe.id] = (recipe: recipe, scale: currentScale)
            }
          }
        )
      }
    }
    .onAppear {
      // Initialize base scales from meal components if not already set
      if baseComponentScales.isEmpty {
        for component in meal.components {
          baseComponentScales[component.recipe.id] = component.recipeScale
        }
      }
    }
    .onChange(of: mealScale) { _, _ in
      // Update loadedRecipes when meal scale changes
      for (recipeID, baseScale) in baseComponentScales {
        let newScale = baseScale * mealScale
        if let recipeData = loadedRecipes[recipeID] {
          loadedRecipes[recipeID] = (recipe: recipeData.recipe, scale: newScale)
        }
      }
    }
  }
}

// MARK: - Embedded Recipe View

struct EmbeddedRecipeView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @State private var viewModel: PerformRecipeViewModel?
  @State private var isInstrumentsVesselsExpanded = false
  @State private var isIngredientsExpanded = false
  @State private var checkedIngredients: Set<String> = []
  @State private var checkedInstrumentsVessels: Set<String> = []
  @State private var isExpanded = false

  let recipeID: String
  @Binding var recipeScale: Float
  let componentType: Mealplanning_MealComponentType
  let sharedViewModel: PerformRecipeViewModel?
  let globalWashHandsCompleted: Bool
  let sharedShowCompletedSteps: Binding<Bool>?
  let onGlobalWashHandsToggle: ((Bool) -> Void)?
  let onStepCompleted: (() -> Void)?
  let onViewModelReady: ((PerformRecipeViewModel) -> Void)?
  let onRecipeLoaded: ((Mealplanning_Recipe) -> Void)?

  var body: some View {
    VStack(alignment: .leading, spacing: 0) {
      // Collapsible header
      VStack(alignment: .leading, spacing: 0) {
        Button(
          action: {
            withAnimation {
              isExpanded.toggle()
            }
          },
          label: {
            HStack {
              // Component type badge
              if componentType != .unspecified {
                Text(formatComponentType(componentType))
                  .font(DSTheme.Typography.labelSmall)
                  .padding(.horizontal, DSTheme.Spacing.sm)
                  .padding(.vertical, DSTheme.Spacing.xs)
                  .background(DSTheme.Colors.primary.opacity(0.2))
                  .foregroundColor(DSTheme.Colors.primary)
                  .cornerRadius(DSTheme.Radius.sm)
              }

              // Recipe name
              if let recipe = viewModel?.recipe {
                Text(recipe.name.isEmpty ? "Unnamed Recipe" : recipe.name)
                  .font(DSTheme.Typography.label)
                  .foregroundColor(DSTheme.Colors.textPrimary)
              } else {
                Text("Loading...")
                  .font(DSTheme.Typography.body)
                  .foregroundColor(DSTheme.Colors.textSecondary)
              }

              Spacer()

              // Chevron
              Image(systemName: isExpanded ? "chevron.down" : "chevron.right")
                .font(DSTheme.Typography.caption)
                .foregroundColor(DSTheme.Colors.textSecondary)
            }
            .padding(DSTheme.Spacing.lg)
            .background(DSTheme.Colors.cardBackground)
          }
        )
        .buttonStyle(.plain)
      }

      // Recipe content (collapsible)
      if isExpanded {
        if let viewModel = viewModel {
          if viewModel.isLoading {
            ProgressView("Loading recipe...")
              .frame(maxWidth: .infinity)
              .padding(DSTheme.Spacing.lg)
          } else if let errorMessage = viewModel.errorMessage {
            VStack(spacing: DSTheme.Spacing.sm) {
              Image(systemName: "exclamationmark.triangle")
                .foregroundColor(DSTheme.Colors.warning)
              Text("Error loading recipe: \(errorMessage)")
                .font(DSTheme.Typography.caption)
                .foregroundColor(DSTheme.Colors.textSecondary)
            }
            .frame(maxWidth: .infinity)
            .padding(DSTheme.Spacing.lg)
          } else if let recipe = viewModel.recipe {
            RecipePerformanceContentView(
              checkedIngredients: $checkedIngredients,
              checkedInstrumentsVessels: $checkedInstrumentsVessels,
              isInstrumentsVesselsExpanded: $isInstrumentsVesselsExpanded,
              isIngredientsExpanded: $isIngredientsExpanded,
              recipe: recipe,
              viewModel: viewModel,
              hideIngredientsAndInstruments: true,
              externalScale: Binding(
                get: { recipeScale },
                set: { newValue in
                  if let scale = newValue {
                    recipeScale = scale
                  }
                }
              ),
              showWashHandsStepCard: false,
              sharedWashHandsCompleted: Binding(
                get: { globalWashHandsCompleted },
                set: { newValue in
                  onGlobalWashHandsToggle?(newValue)
                }
              ),
              sharedCompletedStepsVisibility: sharedShowCompletedSteps,
              allowCompletedStepsToggle: true,
              onStepCompleted: onStepCompleted
            )
            .onAppear {
              onRecipeLoaded?(recipe)
            }
          }
        } else {
          ProgressView("Initializing...")
            .frame(maxWidth: .infinity)
            .padding(DSTheme.Spacing.lg)
        }
      }
    }
    .background(DSTheme.Colors.cardBackground)
    .cornerRadius(DSTheme.Radius.md)
    .onAppear {
      if viewModel == nil {
        if let sharedViewModel {
          viewModel = sharedViewModel
          onViewModelReady?(sharedViewModel)
        } else {
          let createdViewModel = PerformRecipeViewModel(
            recipeID: recipeID, authManager: authManager)
          viewModel = createdViewModel
          onViewModelReady?(createdViewModel)
        }
      }

      if let viewModel, viewModel.recipe == nil, !viewModel.isLoading {
        Task {
          await viewModel.loadRecipe()
        }
      }
      viewModel?.washHandsCompleted = globalWashHandsCompleted
      if !globalWashHandsCompleted {
        viewModel?.completedSteps.removeAll()
      }
    }
    .onChange(of: globalWashHandsCompleted) { _, newValue in
      viewModel?.washHandsCompleted = newValue
      if !newValue {
        viewModel?.completedSteps.removeAll()
      }
    }
    .onChange(of: viewModel?.recipe) { _, newRecipe in
      if let recipe = newRecipe {
        onRecipeLoaded?(recipe)
      }
    }
  }

  private func formatComponentType(_ type: Mealplanning_MealComponentType) -> String {
    switch type {
    case .amuseBouche:
      return "Amuse Bouche"
    case .appetizer:
      return "Appetizer"
    case .soup:
      return "Soup"
    case .main:
      return "Main"
    case .salad:
      return "Salad"
    case .beverage:
      return "Beverage"
    case .side:
      return "Side"
    case .dessert:
      return "Dessert"
    default:
      return ""
    }
  }
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "John Doe"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  return NavigationView {
    MealDetailView(mealID: "meal123")
  }
  .environment(authManager)
}
