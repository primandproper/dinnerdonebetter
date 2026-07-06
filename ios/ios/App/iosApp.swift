//
//  iosApp.swift
//  ios
//
//  Created by Jeffrey Dorrycott on 12/8/25.
//

import Observability
import RevenueCat
import SwiftUI

@main
struct IOSApp: App {
  @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
  @State private var authManager: AuthenticationManager
  @State private var eventReporterService: EventReporterService
  @State private var userSettingsService = UserSettingsService()
  @State private var deepLinkHandler = DeepLinkHandler()

  /// Retained for the app's lifetime so its MetricKit subscription stays alive.
  private let diagnostics: MetricKitDiagnostics

  init() {
    // Must run before any view accesses Purchases.shared. AuthManager init (during
    // @State setup) fires a Task that calls Purchases.shared; we configure here
    // and AuthManager defers its Task with yield so this runs first.
    if RevenueCatConfiguration.isConfigured {
      Purchases.configure(withAPIKey: RevenueCatConfiguration.revenueCatAPIKey)
    }

    // Bootstrap observability at the composition root and subscribe to MetricKit
    // (delivers on device only; a harmless no-op in the simulator).
    let platform = PlatformServices.shared
    diagnostics = MetricKitDiagnostics(logger: platform.logger("MetricKit"))
    diagnostics.start()

    let am = AuthenticationManager()
    _authManager = State(initialValue: am)
    _eventReporterService = State(initialValue: EventReporterService(authManager: am))
  }

  var body: some Scene {
    WindowGroup {
      ContentView()
        .environment(eventReporterService)
        .environment(authManager)
        .environment(userSettingsService)
        .environment(deepLinkHandler)
        .onAppear {
          userSettingsService.configure(authManager: authManager)
          DeviceTokenRegistrationService.shared.configure(authManager: authManager)
          Task { await authManager.logInToRevenueCatIfNeeded() }
        }
        .task(id: authManager.isAuthenticated) {
          if authManager.isAuthenticated {
            await userSettingsService.load()
          } else {
            userSettingsService.clear()
          }
        }
        .onOpenURL { url in
          deepLinkHandler.handleURL(url)
        }
    }
  }
}
