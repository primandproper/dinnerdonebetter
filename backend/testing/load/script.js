/**
 * k6 read-only load test for the Dinner Done Better API.
 *
 * Auth flow (same as mobile/web apps):
 * 1. LoginForToken (gRPC) with username/password → JWT
 * 2. OAuth2 authorize with Bearer JWT + client creds → auth code
 * 3. OAuth2 token exchange (code + client creds) → access token
 * 4. Use access token as Bearer for gRPC read calls
 *
 * Env vars:
 *   K6_GRPC_TARGET, K6_HTTP_BASE - API URLs (default: production)
 *   K6_OAUTH2_CLIENT_ID, K6_OAUTH2_CLIENT_SECRET - Load test OAuth2 client
 *   K6_LOADTEST_USERNAME, K6_LOADTEST_PASSWORD - Load test user
 *
 * Run: k6 run --vus 10 --duration 5m testing/load/script.js
 * (from backend/) or backend/testing/load/script.js (from repo root)
 */

import grpc from "k6/net/grpc";
import http from "k6/http";
import { check } from "k6";

const grpcTarget =
  __ENV.K6_GRPC_TARGET || "api.dinnerdonebetter.com:443";
const httpBase =
  __ENV.K6_HTTP_BASE || "https://http-api.dinnerdonebetter.com";
const usePlaintext = __ENV.K6_GRPC_PLAINTEXT === "true";
const clientId = __ENV.K6_OAUTH2_CLIENT_ID;
const clientSecret = __ENV.K6_OAUTH2_CLIENT_SECRET;
const username = __ENV.K6_LOADTEST_USERNAME;
const password = __ENV.K6_LOADTEST_PASSWORD;

const redirectUri = httpBase;

// Read-only endpoints to exercise (no data written).
// request can be a static object or a function (data) => request for context-dependent calls.
const READ_ENDPOINTS = [
  // Valid enums (shared)
  { method: "mealplanning.MealPlanningService/GetValidInstruments", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidIngredients", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidVessels", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidPreparations", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidIngredientGroups", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidMeasurementUnits", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/SearchForValidIngredients", request: { filter: {}, query: "chicken", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/SearchForValidInstruments", request: { filter: {}, query: "knife", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/SearchForValidVessels", request: { filter: {}, query: "pan", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/SearchForValidPreparations", request: { filter: {}, query: "dice", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/SearchForValidMeasurementUnits", request: { filter: {}, query: "cup", use_search_service: false } },
  // Mealplanning – account/user-scoped
  { method: "mealplanning.MealPlanningService/GetRecipes", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetRecipes", request: { filter: { max_response_size: 10 } } },
  { method: "mealplanning.MealPlanningService/GetRecipes", request: { filter: { max_response_size: 50 } } },
  { method: "mealplanning.MealPlanningService/GetMealPlansForAccount", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetMeals", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetMeals", request: { filter: { max_response_size: 10 } } },
  { method: "mealplanning.MealPlanningService/GetMeals", request: { filter: { max_response_size: 50 } } },
  { method: "mealplanning.MealPlanningService/GetRecipeLists", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetMealLists", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/SearchForRecipes", request: { filter: {}, query: "chicken", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/SearchForMealEligibleRecipes", request: { filter: {}, query: "pasta" } },
  { method: "mealplanning.MealPlanningService/SearchForMeals", request: { filter: {}, query: "dinner", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/GetAccountInstrumentOwnerships", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetUserIngredientPreferences", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/SearchForRecipesWithInstrumentOwnership", request: { filter: {}, query: "salad" } },
  { method: "mealplanning.MealPlanningService/SearchForValidIngredientGroups", request: { filter: {}, query: "dairy", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/SearchForValidIngredientStates", request: { filter: {}, query: "raw", use_search_service: false } },
  { method: "mealplanning.MealPlanningService/GetValidPrepTaskConfigs", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidIngredientStates", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidPreparationInstruments", request: { filter: {} } },
  { method: "mealplanning.MealPlanningService/GetValidPreparationVessels", request: { filter: {} } },
  // Webhooks
  { method: "webhooks.WebhooksService/GetWebhooks", request: { filter: {} } },
  { method: "webhooks.WebhooksService/GetWebhookTriggerEvents", request: { filter: {} } },
  // Identity – accounts and invitations (user-scoped; no admin-only GetUsersForAccount/SearchForUsers)
  { method: "identity.IdentityService/GetAccounts", request: { filter: {} } },
  { method: "identity.IdentityService/GetReceivedAccountInvitations", request: { filter: {} } },
  { method: "identity.IdentityService/GetSentAccountInvitations", request: { filter: {} } },
  // Settings
  { method: "settings.SettingsService/GetServiceSettingConfigurationsForAccount", request: { filter: {} } },
  { method: "settings.SettingsService/GetServiceSettingConfigurationsForUser", request: { filter: {} } },
  { method: "settings.SettingsService/GetServiceSettings", request: { filter: {} } },
  { method: "settings.SettingsService/SearchForServiceSettings", request: { filter: {}, query: "notification" } },
  // Audit
  { method: "audit.AuditService/GetAuditLogEntriesForAccount", request: (d) => ({ filter: {}, account_id: d.accountId }) },
  { method: "audit.AuditService/GetAuditLogEntriesForUser", request: (d) => ({ filter: {}, user_id: d.userId }) },
  // Notifications
  { method: "notifications.UserNotificationsService/GetUserNotifications", request: { filter: {} } },
  { method: "notifications.UserNotificationsService/GetUserDeviceTokens", request: { filter: {} } },
  // Issue reports
  { method: "issue_reports.IssueReportsService/GetIssueReports", request: { filter: {} } },
  { method: "issue_reports.IssueReportsService/GetIssueReportsByStatus", request: { status: "open", filter: {} } },
  { method: "issue_reports.IssueReportsService/GetIssueReportsBySubjectType", request: { subject_type: "recipes", filter: {} } },
  // Auth
  { method: "auth.AuthService/GetAuthStatus", request: {} },
  { method: "auth.AuthService/GetActiveAccount", request: {} },
  { method: "auth.AuthService/GetSelf", request: {} },
];

function randomElement(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

/**
 * Fetch JWT via LoginForToken (gRPC, unauthenticated).
 */
function fetchLoginToken() {
  const client = new grpc.Client();
  client.connect(grpcTarget, { plaintext: usePlaintext, reflect: true });

  const response = client.invoke("auth.AuthService/LoginForToken", {
    input: {
      username: username,
      password: password,
    },
  });

  client.close();

  if (!response || response.status !== grpc.StatusOK) {
    throw new Error(
      `LoginForToken failed: ${response ? response.status : "no response"}`
    );
  }

  // Protobuf JSON may use snake_case or camelCase depending on k6/grpc version
  const result = response.message?.result;
  const token =
    result?.access_token || result?.accessToken;
  if (!token) {
    const msg = JSON.stringify(response.message || response);
    throw new Error(
      `LoginForToken: no access_token in response. Status=${response.status}. Response: ${msg}`
    );
  }
  return token;
}

/**
 * Exchange JWT for OAuth2 tokens via authorize + token flow.
 */
function exchangeForOAuth2Token(jwt) {
  const state = `k6-${Date.now()}-${Math.random().toString(36).slice(2)}`;
  // PKCE requires code_verifier/code_challenge to be 43-128 chars
  const codeVerifier =
    `k6-${Date.now()}-${Math.random().toString(36).slice(2)}` +
    Math.random().toString(36).slice(2) +
    Math.random().toString(36).slice(2);

  const authParams = [
    "client_id=" + encodeURIComponent(clientId),
    "redirect_uri=" + encodeURIComponent(redirectUri),
    "response_type=code",
    "scope=anything",
    "state=" + encodeURIComponent(state),
    "code_challenge_method=plain",
    "code_challenge=" + encodeURIComponent(codeVerifier),
  ].join("&");

  const authUrl = `${httpBase}/oauth2/authorize?${authParams}`;
  const authRes = http.get(authUrl, {
    headers: { Authorization: `Bearer ${jwt}` },
    redirects: 0,
  });

  if (authRes.status !== 302 && authRes.status !== 301) {
    throw new Error(
      `OAuth2 authorize failed: ${authRes.status} ${authRes.body}. Request had client_id=${clientId ? "set" : "MISSING"} redirect_uri=${redirectUri ? "set" : "MISSING"}`
    );
  }

  const location = authRes.headers.Location;
  if (!location) {
    throw new Error("OAuth2 authorize: no Location header");
  }

  const codeMatch = location.match(/[?&]code=([^&]+)/);
  const code = codeMatch ? codeMatch[1] : null;
  if (!code) {
    throw new Error(`OAuth2 authorize: no code in redirect: ${location}`);
  }

  const tokenRes = http.post(
    `${httpBase}/oauth2/token`,
    {
      grant_type: "authorization_code",
      code: code,
      redirect_uri: redirectUri,
      client_id: clientId,
      client_secret: clientSecret,
      code_verifier: codeVerifier,
    },
    {
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    }
  );

  if (tokenRes.status !== 200) {
    throw new Error(
      `OAuth2 token exchange failed: ${tokenRes.status} ${tokenRes.body}`
    );
  }

  const body = tokenRes.json();
  const accessToken = body.access_token;
  if (!accessToken) {
    throw new Error(
      `OAuth2 token: no access_token in response: ${JSON.stringify(body)}`
    );
  }
  return accessToken;
}

function fetchAccountAndUserIds(accessToken) {
  const client = new grpc.Client();
  client.connect(grpcTarget, { plaintext: usePlaintext, reflect: true });

  const accountRes = client.invoke("auth.AuthService/GetActiveAccount", {}, {
    metadata: { authorization: `Bearer ${accessToken}` },
  });
  const selfRes = client.invoke("auth.AuthService/GetSelf", {}, {
    metadata: { authorization: `Bearer ${accessToken}` },
  });

  client.close();

  const accountId = accountRes?.status === grpc.StatusOK && accountRes?.message?.result
    ? (accountRes.message.result.id || accountRes.message.result.Id || "")
    : "";
  const userId = selfRes?.status === grpc.StatusOK && selfRes?.message?.result
    ? (selfRes.message.result.id || selfRes.message.result.Id || "")
    : "";

  return { accountId, userId };
}

export function setup() {
  if (!clientId || !clientSecret || !username || !password) {
    throw new Error(
      "Missing required env: K6_OAUTH2_CLIENT_ID, K6_OAUTH2_CLIENT_SECRET, K6_LOADTEST_USERNAME, K6_LOADTEST_PASSWORD"
    );
  }

  const jwt = fetchLoginToken();
  const accessToken = exchangeForOAuth2Token(jwt);
  const { accountId, userId } = fetchAccountAndUserIds(accessToken);
  return { accessToken, accountId, userId };
}

export default function (data) {
  const client = new grpc.Client();
  client.connect(grpcTarget, { plaintext: usePlaintext, reflect: true });

  const endpoint = randomElement(READ_ENDPOINTS);
  const request =
    typeof endpoint.request === "function"
      ? endpoint.request(data)
      : endpoint.request;
  const response = client.invoke(endpoint.method, request, {
    metadata: { authorization: `Bearer ${data.accessToken}` },
  });

  check(response, {
    "status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  if (response && response.status !== grpc.StatusOK) {
    console.warn(
      `gRPC ${endpoint.method} failed: ${response.status} ${JSON.stringify(response.message)}`
    );
  }

  client.close();
}
