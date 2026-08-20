//
//  OAuth2ExchangeTests.swift
//  iosTests
//
//  Covers the wire format of the OAuth 2.1 authorization code exchange: the parts the
//  authorization server refuses outright if the app gets them wrong.
//

import Foundation
@testable import ios
import Testing

struct OAuth2ExchangeTests {
    // MARK: - Helpers

    /// Query parameters of a request's URL, as a dictionary.
    private func queryParameters(of request: URLRequest) -> [String: String] {
        guard let url = request.url,
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
            let items = components.queryItems
        else {
            return [:]
        }
        return items.reduce(into: [:]) { $0[$1.name] = $1.value }
    }

    /// Form parameters of a request's `application/x-www-form-urlencoded` body.
    private func bodyParameters(of request: URLRequest) -> [String: String] {
        guard let body = request.httpBody, let encoded = String(data: body, encoding: .utf8) else {
            return [:]
        }
        return encoded.split(separator: "&").reduce(into: [:]) { result, pair in
            let halves = pair.split(separator: "=", maxSplits: 1)
            guard halves.count == 2 else { return }
            let name = String(halves[0]).removingPercentEncoding ?? String(halves[0])
            let value = String(halves[1]).removingPercentEncoding ?? String(halves[1])
            result[name] = value
        }
    }

    // MARK: - PKCE

    @Test("PKCE derives the RFC 7636 Appendix B challenge from its verifier")
    func testPKCEMatchesRFCVector() {
        let pkce = PKCE(verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk")

        #expect(pkce.challenge == "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM")
    }

    @Test("PKCE challenge method is S256, the only one the server accepts")
    func testPKCEChallengeMethodIsS256() {
        #expect(PKCE.challengeMethod == "S256")
    }

    @Test("base64url encoding is unpadded and uses the URL-safe alphabet")
    func testBase64URLEncoding() {
        let encoded = PKCE.base64URLEncoded(Data((0..<32).map { UInt8($0) }))

        #expect(encoded == "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8")
        #expect(!encoded.contains("="))
        #expect(!encoded.contains("+"))
        #expect(!encoded.contains("/"))
    }

    @Test("Generated verifiers are unpadded base64url within the RFC 7636 length range")
    func testGeneratedVerifierShape() {
        let verifier = PKCE().verifier

        #expect(verifier.count >= 43)
        #expect(verifier.count <= 128)
        #expect(!verifier.contains("="))
        #expect(!verifier.contains("+"))
        #expect(!verifier.contains("/"))
    }

    @Test("Each PKCE instance generates a distinct verifier")
    func testGeneratedVerifiersAreDistinct() {
        let verifiers = Set((0..<16).map { _ in PKCE().verifier })

        #expect(verifiers.count == 16)
    }

    @Test("A generated verifier's challenge is not the verifier itself")
    func testGeneratedChallengeIsHashed() {
        let pkce = PKCE()

        #expect(pkce.challenge != pkce.verifier)
        #expect(pkce.challenge.count == 43)
    }

    // MARK: - Authorization request

    private func authorizationRequest(
        challenge: String = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
    ) -> URLRequest? {
        OAuth2Exchange.authorizationRequest(
            authorizeURL: "https://api.example.com/authorize",
            bearerToken: "jwt-token",
            clientID: "client-id",
            redirectURI: "https://api.example.com",
            state: "state-value",
            challenge: challenge
        )
    }

    @Test("Authorization request is a POST, so the server authenticates instead of rendering a form")
    func testAuthorizationRequestIsPOST() throws {
        let request = try #require(authorizationRequest())

        #expect(request.httpMethod == "POST")
        #expect(request.value(forHTTPHeaderField: "Authorization") == "Bearer jwt-token")
        #expect(
            request.value(forHTTPHeaderField: "Content-Type") == "application/x-www-form-urlencoded")
    }

    @Test("Authorization parameters stay in the query string, not the body")
    func testAuthorizationParametersAreInTheQuery() throws {
        let request = try #require(authorizationRequest())
        let query = queryParameters(of: request)

        #expect(request.httpBody == nil)
        #expect(query["response_type"] == "code")
        #expect(query["client_id"] == "client-id")
        #expect(query["redirect_uri"] == "https://api.example.com")
        #expect(query["state"] == "state-value")
    }

    @Test("Authorization request sends an S256 challenge, never plain")
    func testAuthorizationRequestSendsS256Challenge() throws {
        let request = try #require(authorizationRequest())
        let query = queryParameters(of: request)

        #expect(query["code_challenge"] == "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM")
        #expect(query["code_challenge_method"] == "S256")
    }

    // MARK: - Token request

    private func tokenRequest(clientSecret: String = "client-secret") -> URLRequest? {
        OAuth2Exchange.tokenRequest(
            tokenURL: "https://api.example.com/token",
            clientID: "client-id",
            clientSecret: clientSecret,
            code: "auth-code",
            codeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
            redirectURI: "https://api.example.com"
        )
    }

    @Test("Token request posts the code alongside its verifier and redirect URI")
    func testTokenRequestBody() throws {
        let request = try #require(tokenRequest())
        let body = bodyParameters(of: request)

        #expect(request.httpMethod == "POST")
        #expect(
            request.value(forHTTPHeaderField: "Content-Type") == "application/x-www-form-urlencoded")
        #expect(body["grant_type"] == "authorization_code")
        #expect(body["code"] == "auth-code")
        #expect(body["code_verifier"] == "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk")
        #expect(body["redirect_uri"] == "https://api.example.com")
        #expect(body["client_id"] == "client-id")
        #expect(body["client_secret"] == "client-secret")
    }

    @Test("Token request escapes secrets that contain form-significant characters")
    func testTokenRequestEscapesSecret() throws {
        let request = try #require(tokenRequest(clientSecret: "a+b&c=d/e"))

        #expect(bodyParameters(of: request)["client_secret"] == "a+b&c=d/e")
    }

    @Test("Refresh request presents the stored refresh token")
    func testRefreshRequestBody() throws {
        let request = try #require(
            OAuth2Exchange.refreshRequest(
                tokenURL: "https://api.example.com/token",
                clientID: "client-id",
                clientSecret: "client-secret",
                refreshToken: "refresh-token"
            ))
        let body = bodyParameters(of: request)

        #expect(request.httpMethod == "POST")
        #expect(body["grant_type"] == "refresh_token")
        #expect(body["refresh_token"] == "refresh-token")
        #expect(body["client_id"] == "client-id")
        #expect(body["client_secret"] == "client-secret")
    }

    // MARK: - Form encoding

    @Test("Form encoding escapes every character that carries meaning in a body")
    func testFormEncodingEscapesReservedCharacters() {
        let encoded = OAuth2Exchange.formEncoded(["key": "a+b&c=d/e f"])

        #expect(encoded == "key=a%2Bb%26c%3Dd%2Fe%20f")
    }

    @Test("Form encoding sorts by key, so a body is a function of its parameters")
    func testFormEncodingIsSorted() {
        let encoded = OAuth2Exchange.formEncoded(["b": "2", "a": "1", "c": "3"])

        #expect(encoded == "a=1&b=2&c=3")
    }

    // MARK: - Access token lifetime

    @Test("Assumed access token lifetime matches the server's 15 minutes")
    func testDefaultAccessTokenLifetime() {
        #expect(OAuth2Exchange.defaultAccessTokenLifetime == 900)
    }

    // MARK: - Endpoints

    @Test("OAuth2 endpoints point at the authorization server's paths")
    func testOAuth2Endpoints() {
        #expect(APIConfiguration.oauth2AuthorizeURL == "\(APIConfiguration.serverURL)/authorize")
        #expect(APIConfiguration.oauth2TokenURL == "\(APIConfiguration.serverURL)/token")
        #expect(!APIConfiguration.oauth2AuthorizeURL.contains("/oauth2/"))
        #expect(!APIConfiguration.oauth2TokenURL.contains("/oauth2/"))
    }

    @Test("Redirect URI is the server URL exactly, since matching is byte for byte")
    func testRedirectURIMatchesServerURL() {
        #expect(APIConfiguration.oauth2RedirectURI == APIConfiguration.serverURL)
        #expect(!APIConfiguration.oauth2RedirectURI.hasSuffix("/"))
    }
}
