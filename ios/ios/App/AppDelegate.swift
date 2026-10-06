//
//  AppDelegate.swift
//  ios
//
//  Handles push notification lifecycle (device token registration).
//  Configures RevenueCat here (before any SwiftUI views) so Purchases.shared
//  is never accessed before configure().
//

import Observability
import RevenueCat
import UIKit
import UserNotifications

@available(macOS 15.0, iOS 18.0, watchOS 11.0, tvOS 18.0, visionOS 2.0, *)
class AppDelegate: NSObject, UIApplicationDelegate {
  private let logger = PlatformServices.shared.logger("AppDelegate")

  /// Where an APNs token goes: the AuthenticationManager, whose ClientManager's `Devices`
  /// registers it and persists the registration. The app sets it on first appearance; a token
  /// that arrives before then waits in `pendingDeviceToken`.
  weak var authManager: AuthenticationManager? {
    didSet {
      if let authManager, let token = pendingDeviceToken {
        pendingDeviceToken = nil
        authManager.registerDeviceToken(token)
      }
    }
  }
  private var pendingDeviceToken: Data?

  override init() {
    super.init()
    if RevenueCatConfiguration.isConfigured {
      Purchases.configure(withAPIKey: RevenueCatConfiguration.revenueCatAPIKey)
    }
  }

  func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
  ) -> Bool {
    requestNotificationPermissionAndRegister()
    return true
  }

  func application(
    _ application: UIApplication,
    didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
  ) {
    if let authManager {
      authManager.registerDeviceToken(deviceToken)
    } else {
      pendingDeviceToken = deviceToken
    }
  }

  func application(
    _ application: UIApplication,
    didFailToRegisterForRemoteNotificationsWithError error: Error
  ) {
    logger.error("registering for remote notifications", error)
  }

  private func requestNotificationPermissionAndRegister() {
    UNUserNotificationCenter.current().requestAuthorization(
      options: [.alert, .badge, .sound]
    ) { granted, error in
      if let error {
        self.logger.error("requesting notification permission", error)
        return
      }
      if granted {
        Task { @MainActor in
          UIApplication.shared.registerForRemoteNotifications()
        }
      }
    }
  }
}
