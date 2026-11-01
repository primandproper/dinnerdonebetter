//
//  Branding.swift
//  ios
//

import DesignSystem
import Foundation

/// Core branding constants - change these to rebrand the project.
enum Branding {
  static let companyName = "Dinner Done Better"
  static let companyNameSlug = "dinnerdonebetter"
  static let publicDomain = "\(companyNameSlug).com"
  static let proEntitlementName = "\(companyName) Pro"

  // Keychain key prefix
  static let keychainPrefix = "com.\(companyNameSlug)"

  // URLs
  static let productionAPIDomain = "http-api.\(publicDomain)"
  static let productionWebDomain = "www.\(publicDomain)"
  static let productionGRPCDomain = "api.\(publicDomain)"
  static let productionMediaDomain = "media.\(publicDomain)"

  static let productionAPIURL = "https://\(productionAPIDomain)"
  static let productionWebURL = "https://\(productionWebDomain)"
  static let productionMediaURL = "https://\(productionMediaDomain)"

  // Design system theme, applied at the app root with `.dsTheme(_:)`. Only the platform-ui-swift
  // components follow it: the app's own views still read the static tokens
  // (`DSTheme.Colors.primary`), which always return `DSTheme.default`. Until those views read
  // `@Environment(\.dsTheme)`, changing this splits the look, so `BrandingThemeTests` pins it.
  static let theme = DSTheme.default
}
