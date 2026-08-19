# JWT Signing key
resource "random_string" "jwt_signing_key" {
  length  = 32
  special = false
}

# At-rest encryption key for user data disclosure artifacts.
#
# Changing this makes every artifact written under the old key permanently unreadable. That is
# survivable — artifacts expire in seven days and a subject can request a fresh one — but it is
# not a no-op, so this resource should not be tainted casually.
resource "random_string" "disclosure_artifact_encryption_key" {
  length  = 32
  special = false
}

# Admin webapp cookie encryption key (base64-encoded 32-byte key for SvelteKit COOKIE_ENCRYPTION_KEY)
resource "random_bytes" "admin_webapp_cookie_encryption_key" {
  length = 32
}

# Consumer webapp cookie encryption key (base64-encoded 32-byte key for SvelteKit COOKIE_ENCRYPTION_KEY)
resource "random_bytes" "consumer_webapp_cookie_encryption_key" {
  length = 32
}

### External API services ###

# Resend API key
variable "RESEND_API_KEY" {}

# PostHog API tokens: Project API Key (events + analytics); Personal API Key (feature flags API).
# Add both to Terraform Cloud: Workspace → Variables → Add variable.
variable "POSTHOG_API_KEY" {}
variable "POSTHOG_PERSONAL_API_KEY" {}

# Grafana Cloud API keys
# NOTE: the passwords are all effectively the same, but they maybe won't be one day? Who knows.
variable "GRAFANA_CLOUD_PROMETHEUS_USERNAME" {}
variable "GRAFANA_CLOUD_PROMETHEUS_PASSWORD" {}
variable "GRAFANA_CLOUD_LOKI_USERNAME" {}
variable "GRAFANA_CLOUD_LOKI_PASSWORD" {}
variable "GRAFANA_CLOUD_TEMPO_USERNAME" {}
variable "GRAFANA_CLOUD_TEMPO_PASSWORD" {}
variable "GRAFANA_CLOUD_PYROSCOPE_USERNAME" {}
variable "GRAFANA_CLOUD_PYROSCOPE_PASSWORD" {}

variable "APNS_KEY_ID" {}
variable "APNS_AUTH_KEY_P8" {
  sensitive = true
}
variable "APNS_TEAM_ID" {
  default = "K8R2Q5UWQS"
}
variable "APNS_PRODUCTION" {
  default     = "false"
  description = "Use APNs production environment (true) or sandbox (false). Sandbox for debug/TestFlight builds; production for App Store."
}

# Admin webapp config (cookie name and domain - required for admin webapp)
variable "ADMIN_WEBAPP_COOKIE_NAME" {
  default = "admin_webapp"
}

# Consumer webapp config (cookie name and domain - required for consumer webapp / root site)
variable "CONSUMER_WEBAPP_COOKIE_NAME" {
  default = "consumer_session"
}
