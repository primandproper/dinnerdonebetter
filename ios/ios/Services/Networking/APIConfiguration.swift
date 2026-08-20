//
//  APIConfiguration.swift
//  ios
//
//  Created by Jeffrey Dorrycott on 12/8/25.
//

import Foundation

/// Represents the different deployment environments for the app.
enum AppEnvironment: String, CaseIterable {
  case local
  case production

  var displayName: String {
    switch self {
    case .local: return "Local"
    case .production: return "Production"
    }
  }

  var subtitle: String {
    switch self {
    case .local: return "localhost"
    case .production: return Branding.publicDomain
    }
  }

  var iconName: String {
    switch self {
    case .local: return "laptopcomputer"
    case .production: return "globe"
    }
  }
}

struct APIConfiguration {
  private static let environmentKey = "selectedEnvironment"

  /// Current environment, persisted across launches via UserDefaults.
  /// Defaults to `.local` in DEBUG builds and `.production` in release builds
  /// if no selection has been saved.
  static var currentEnvironment: AppEnvironment {
    get {
      if let saved = UserDefaults.standard.string(forKey: environmentKey),
        let env = AppEnvironment(rawValue: saved)
      {
        return env
      }
      return .production
    }
    set {
      let oldValue = currentEnvironment
      UserDefaults.standard.set(newValue.rawValue, forKey: environmentKey)
      if oldValue != newValue {
        NotificationCenter.default.post(name: .environmentDidChange, object: newValue)
      }
    }
  }

  /// The base URL for API requests (HTTP/REST endpoints)
  static var serverURL: String {
    switch currentEnvironment {
    case .local:
      return "http://localhost:8000"
    case .production:
      return Branding.productionAPIURL
    }
  }

  /// The web URL used for Universal Links
  /// This should match the domain in your entitlements file
  static var webURL: String {
    switch currentEnvironment {
    case .local:
      return "http://localhost:3000"
    case .production:
      return Branding.productionWebURL
    }
  }

  /// The gRPC host for the current environment
  static var grpcHost: String {
    switch currentEnvironment {
    case .local:
      return "0.0.0.0"
    case .production:
      return Branding.productionGRPCDomain
    }
  }

  /// The gRPC port for the current environment
  static var grpcPort: Int {
    switch currentEnvironment {
    case .local:
      return 8001
    case .production:
      return 443
    }
  }

  /// Whether the gRPC connection should use TLS.
  /// Local development uses plaintext; remote environments use TLS.
  static var grpcUsesTLS: Bool {
    switch currentEnvironment {
    case .local:
      return false
    case .production:
      return true
    }
  }

  /// Whether search requests should use Algolia (production) or database (local).
  /// Production has Algolia indices; local dev does not.
  static var useSearchService: Bool {
    currentEnvironment == .production
  }

  /// Base URL where uploaded media is hosted (no bucket/path).
  static var mediaURLPrefix: String {
    switch currentEnvironment {
    case .local:
      return "http://localhost:8000/uploads"
    case .production:
      return Branding.productionMediaURL
    }
  }

  /// Constructs the full URL for a media object from its bucket and storage path.
  /// - Parameters:
  ///   - storagePath: The object's path within the bucket (e.g. "userid/fileid/name.jpg")
  ///   - bucket: The bucket name (e.g. "avatars")
  static func mediaURL(forStoragePath storagePath: String, bucket: String) -> URL? {
    guard !storagePath.isEmpty, !bucket.isEmpty else { return nil }
    let prefix = mediaURLPrefix
    let path = "\(prefix)/\(bucket)/\(storagePath)"
    return URL(string: path)
  }

  /// Constructs the full URL string for a media object from its bucket and storage path.
  static func mediaURLString(forStoragePath storagePath: String, bucket: String) -> String? {
    guard !storagePath.isEmpty, !bucket.isEmpty else { return nil }
    let prefix = mediaURLPrefix
    return "\(prefix)/\(bucket)/\(storagePath)"
  }

  // OAuth2 Configuration — values injected via Secrets.xcconfig -> Info.plist at build time
  static var oauth2ClientID: String {
    Bundle.main.infoDictionary?["OAuth2ClientID"] as? String ?? ""
  }

  static var oauth2ClientSecret: String {
    Bundle.main.infoDictionary?["OAuth2ClientSecret"] as? String ?? ""
  }

  // OAuth2 endpoints, as served by the OAuth 2.1 authorization server.
  static var oauth2AuthorizeURL: String {
    return "\(serverURL)/authorize"
  }

  static var oauth2TokenURL: String {
    return "\(serverURL)/token"
  }

  /// The redirect URI sent at `/authorize` and again at `/token`.
  ///
  /// Nothing listens here: the authorization code is read off the `Location` header of the 302
  /// rather than followed. It is matched byte for byte against the client's registered URIs,
  /// so it has to be `serverURL` exactly — the address `ddb-bootstrap init --api-server-url`
  /// registers for every first-party client.
  static var oauth2RedirectURI: String {
    return serverURL
  }
}

extension Notification.Name {
  static let environmentDidChange = Notification.Name("environmentDidChange")
}
