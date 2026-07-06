//
//  CurrentUserService.swift
//  ios
//
//  TTL cache for the current user (`getSelf`), built on platform-swift's in-memory Cache.
//

import Cache
import Foundation
import Observability
import SwiftProtobuf

enum CurrentUserError: Error { case noUser }

/// Caches the current user (`getSelf`) with a short TTL so the several call sites that fetch it
/// don't each round-trip. SwiftProtobuf messages aren't `Codable`, so the serialized proto is
/// stored as `Data`. Session-scoped: cleared on logout / session invalidation.
actor CurrentUserService {
  static let shared = CurrentUserService()

  private let cache: InMemoryCache<Data>
  private static let key = "self"

  init(ttl: Duration = .seconds(120)) {
    cache = InMemoryCache<Data>(
      expiry: ttl, maxEntries: 1, name: "current_user_cache",
      pillars: PlatformServices.shared.pillars)
  }

  /// Returns the current user — from cache when fresh, else via `getSelf` (cached on success).
  func currentUser(using authManager: AuthenticationManager, forceRefresh: Bool = false)
    async throws -> Identity_User
  {
    if !forceRefresh, let data = try? await cache.get(Self.key),
      let user = try? Identity_User(serializedData: data)
    {
      return user
    }
    let response = try await authManager.authenticatedCall("getSelf", idempotent: true) {
      client, metadata, options in
      try await client.auth.getSelf(Auth_GetSelfRequest(), metadata: metadata, options: options)
    }
    guard response.hasResult else { throw CurrentUserError.noUser }
    let user = response.result
    if let data = try? user.serializedData() {
      try? await cache.set(Self.key, to: data)
    }
    return user
  }

  /// Drops the cached user. Call on logout / session invalidation.
  func clear() async {
    try? await cache.delete(Self.key)
  }
}
