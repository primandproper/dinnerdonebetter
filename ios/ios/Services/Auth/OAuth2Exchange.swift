//
//  OAuth2Exchange.swift
//  ios
//
//  The wire format of the OAuth 2.1 authorization code exchange, kept apart from
//  AuthenticationManager so each step's request can be asserted directly in tests.
//

import CryptoKit
import Foundation

/// A PKCE (RFC 7636) verifier and the S256 challenge derived from it.
///
/// The authorization server requires a challenge and accepts `S256` only. An absent
/// `code_challenge_method` is refused rather than defaulted, because RFC 7636 defaults it to
/// `plain`.
struct PKCE {
  /// The only challenge method the authorization server accepts.
  static let challengeMethod = "S256"

  /// The high-entropy secret, replayed on the token request to prove the code is ours.
  let verifier: String

  /// `base64url(SHA256(verifier))`, sent on the authorization request.
  var challenge: String {
    Self.base64URLEncoded(Data(SHA256.hash(data: Data(verifier.utf8))))
  }

  /// Generates a fresh verifier: 32 random bytes, base64url-encoded to 43 characters, the low
  /// end of the 43–128 RFC 7636 allows.
  init() {
    let bytes = SymmetricKey(size: .bits256).withUnsafeBytes { Data($0) }
    self.verifier = Self.base64URLEncoded(bytes)
  }

  /// Wraps a caller-supplied verifier, so tests can pin the RFC 7636 vector.
  init(verifier: String) {
    self.verifier = verifier
  }

  /// base64url per RFC 4648 §5, unpadded, as RFC 7636 §4.2 requires.
  static func base64URLEncoded(_ data: Data) -> String {
    data.base64EncodedString()
      .replacingOccurrences(of: "+", with: "-")
      .replacingOccurrences(of: "/", with: "_")
      .replacingOccurrences(of: "=", with: "")
  }
}

/// Builds the HTTP requests of the authorization code exchange against the OAuth 2.1 server.
enum OAuth2Exchange {
  /// Lifetime assumed when a token response omits `expires_in`. Access tokens are 15 minutes;
  /// assuming longer would leave the app presenting a dead token instead of refreshing.
  static let defaultAccessTokenLifetime: TimeInterval = 900

  /// The authorization request that issues a code.
  ///
  /// POST, not GET: a `GET /authorize` renders an HTML login form — the answer for a browser
  /// arriving without a session — and only a POST runs the authenticator that reads the bearer
  /// token. The authorization parameters stay in the query string either way, so nothing moves
  /// into a body.
  ///
  /// - Parameter redirectURI: matched byte for byte against the client's registered URIs, so it
  ///   must be one of them exactly, trailing slash included.
  static func authorizationRequest(
    authorizeURL: String,
    bearerToken: String,
    clientID: String,
    redirectURI: String,
    state: String,
    challenge: String
  ) -> URLRequest? {
    guard var components = URLComponents(string: authorizeURL) else { return nil }

    components.queryItems = [
      URLQueryItem(name: "response_type", value: "code"),
      URLQueryItem(name: "client_id", value: clientID),
      URLQueryItem(name: "redirect_uri", value: redirectURI),
      URLQueryItem(name: "state", value: state),
      URLQueryItem(name: "code_challenge", value: challenge),
      URLQueryItem(name: "code_challenge_method", value: PKCE.challengeMethod),
    ]

    guard let url = components.url else { return nil }

    var request = URLRequest(url: url)
    request.httpMethod = "POST"
    request.setValue("Bearer \(bearerToken)", forHTTPHeaderField: "Authorization")
    request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")

    return request
  }

  /// The request that redeems an authorization code. `redirect_uri` is checked again here,
  /// byte for byte against the one the code was issued for, and `code_verifier` against the
  /// challenge that came with it.
  static func tokenRequest(
    tokenURL: String,
    clientID: String,
    clientSecret: String,
    code: String,
    codeVerifier: String,
    redirectURI: String
  ) -> URLRequest? {
    formPost(
      url: tokenURL,
      parameters: [
        "grant_type": "authorization_code",
        "code": code,
        "code_verifier": codeVerifier,
        "redirect_uri": redirectURI,
        "client_id": clientID,
        "client_secret": clientSecret,
      ]
    )
  }

  /// The request that trades a refresh token for a new pair. Refresh tokens rotate: the token
  /// in the response replaces the stored one, and presenting a spent one revokes the whole
  /// family and forces a full re-login.
  static func refreshRequest(
    tokenURL: String,
    clientID: String,
    clientSecret: String,
    refreshToken: String
  ) -> URLRequest? {
    formPost(
      url: tokenURL,
      parameters: [
        "grant_type": "refresh_token",
        "refresh_token": refreshToken,
        "client_id": clientID,
        "client_secret": clientSecret,
      ]
    )
  }

  /// Encodes an `application/x-www-form-urlencoded` body. Keys are sorted so the body is a
  /// function of its parameters alone.
  static func formEncoded(_ parameters: [String: String]) -> String {
    parameters
      .sorted { $0.key < $1.key }
      .map { "\(percentEncoded($0.key))=\(percentEncoded($0.value))" }
      .joined(separator: "&")
  }

  private static func formPost(url: String, parameters: [String: String]) -> URLRequest? {
    guard let url = URL(string: url) else { return nil }

    var request = URLRequest(url: url)
    request.httpMethod = "POST"
    request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
    request.httpBody = formEncoded(parameters).data(using: .utf8)

    return request
  }

  /// Everything outside RFC 3986's unreserved set is escaped. `.urlQueryAllowed` would leave
  /// `+`, `&`, `=` and `/` alone, and all four carry meaning in a form body — a client secret
  /// containing one would arrive mangled or split into another parameter.
  private static let unreservedCharacters = CharacterSet(
    charactersIn: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~")

  private static func percentEncoded(_ value: String) -> String {
    value.addingPercentEncoding(withAllowedCharacters: unreservedCharacters) ?? value
  }
}
