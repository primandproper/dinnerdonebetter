locals {
  k8s_namespace = "prod"
}

provider "kubernetes" {
  config_path    = "./terraform_kubeconfig"
  config_context = "${local.k8s_namespace}_context"
}

data "google_container_cluster" "prod_cluster" {
  name     = "prod"
  location = local.gcp_region
}

resource "kubernetes_namespace_v1" "prod" {
  metadata {
    name = local.k8s_namespace
    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [data.google_container_cluster.prod_cluster]
}

# Kubernetes secrets

# APNs .p8 key mounted as file for push notifications (async message handler)
resource "kubernetes_secret_v1" "apns_credentials" {
  metadata {
    name      = "apns-credentials"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  data = {
    # Key becomes filename when mounted; value is p8 content (Terraform base64-encodes automatically)
    "apns-auth-key.p8" = var.APNS_AUTH_KEY_P8
  }
}

resource "kubernetes_secret_v1" "cloudflare_api_key" {
  metadata {
    name      = "cloudflare-api-key"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  data = {
    "token" = var.CLOUDFLARE_API_TOKEN
  }
}

resource "kubernetes_config_map_v1" "pubsub_topics" {
  metadata {
    name      = "pubsub-topic-names"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  data = {
    data_changes               = google_pubsub_topic.data_changes_topic.id
    outbound_emails            = google_pubsub_topic.outbound_emails_topic.id
    search_index_requests      = google_pubsub_topic.search_index_requests_topic.id
    mobile_notifications       = google_pubsub_topic.mobile_notifications_topic.id
    user_data_aggregator       = google_pubsub_topic.user_data_aggregator_topic.id
    webhook_execution_requests = google_pubsub_topic.webhook_execution_requests_topic.id
  }
}

resource "kubernetes_secret_v1" "api_service_config" {
  metadata {
    name      = "api-service-config"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  # Per-service database passwords are merged in from database_users.tf, so adding a
  # service user does not require touching this block.
  data = merge(local.database_password_secret_data, {
    OAUTH2_TOKEN_ENCRYPTION_KEY        = random_string.oauth2_token_encryption_key.result
    DISCLOSURE_ARTIFACT_ENCRYPTION_KEY = random_string.disclosure_artifact_encryption_key.result
    JWT_SIGNING_KEY                    = base64encode(random_string.jwt_signing_key.result)
    DATABASE_HOST                      = google_sql_database_instance.prod.private_ip_address
    POSTHOG_API_KEY                    = var.POSTHOG_API_KEY
    POSTHOG_PERSONAL_API_KEY           = var.POSTHOG_PERSONAL_API_KEY
    ALGOLIA_APPLICATION_ID             = var.ALGOLIA_APPLICATION_ID
    ALGOLIA_API_KEY                    = var.ALGOLIA_API_KEY
    PUSH_NOTIFICATIONS_PROVIDER        = "apns_fcm"
    PUSH_NOTIFICATIONS_APNS_KEY_ID     = var.APNS_KEY_ID
    PUSH_NOTIFICATIONS_APNS_TEAM_ID    = var.APNS_TEAM_ID
    PUSH_NOTIFICATIONS_APNS_BUNDLE_ID  = local.ios_bundle_id
    PUSH_NOTIFICATIONS_APNS_PRODUCTION = var.APNS_PRODUCTION
    RESEND_API_KEY                     = var.RESEND_API_KEY
  })
}

# this is the sort of resource that should probably ideally live in the infra folder, but it's here for now
# because I haven't yet wanted to fuss with figuring out how to manage the code that creates the cluster
# alongside the code that creates resources in that cluster.
# Admin webapp and MCP server: OAuth2 credentials + cookie config
# Maps to env vars DINNER_DONE_BETTER_API_SERVICE_OAUTH2_API_CLIENT_ID / _SECRET
resource "kubernetes_secret_v1" "admin_webapp_config" {
  metadata {
    name      = local.k8s_admin_webapp_cfg
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  data = {
    OAUTH2_CLIENT_ID      = var.ADMIN_WEBAPP_OAUTH2_CLIENT_ID
    OAUTH2_CLIENT_SECRET  = var.ADMIN_WEBAPP_OAUTH2_CLIENT_SECRET
    COOKIE_NAME           = var.ADMIN_WEBAPP_COOKIE_NAME
    COOKIE_ENCRYPTION_KEY = random_bytes.admin_webapp_cookie_encryption_key.base64
    COOKIE_DOMAIN         = local.admin_domain
  }
}

# Consumer webapp (root site): OAuth2 credentials + cookie config
# Note: ConfigMap dinner-done-better-consumer-webapp-config (from kustomize) holds config.json;
# this Secret holds sensitive values injected via env vars.
resource "kubernetes_secret_v1" "consumer_webapp_config" {
  metadata {
    name      = local.k8s_consumer_webapp
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  data = {
    OAUTH2_CLIENT_ID      = var.CONSUMER_WEBAPP_OAUTH2_CLIENT_ID
    OAUTH2_CLIENT_SECRET  = var.CONSUMER_WEBAPP_OAUTH2_CLIENT_SECRET
    COOKIE_NAME           = var.CONSUMER_WEBAPP_COOKIE_NAME
    COOKIE_ENCRYPTION_KEY = random_bytes.consumer_webapp_cookie_encryption_key.base64
    COOKIE_DOMAIN         = local.public_domain
  }
}

# MCP server OAuth2 credentials for DINNER_DONE_BETTER_API_SERVICE_OAUTH2_API_CLIENT_ID / _SECRET
resource "kubernetes_secret_v1" "mcp_server_config" {
  metadata {
    name      = "mcp-server-config"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  data = {
    OAUTH2_CLIENT_ID     = var.MCP_SERVICE_OAUTH2_CLIENT_ID
    OAUTH2_CLIENT_SECRET = var.MCP_SERVICE_OAUTH2_CLIENT_SECRET
  }
}

# this is the sort of resource that should probably ideally live in the infra folder, but it's here for now
# because I haven't yet wanted to fuss with figuring out how to manage the code that creates the cluster
# alongside the code that creates resources in that cluster.
resource "kubernetes_secret_v1" "grafana_cloud_creds" {
  metadata {
    name      = "grafana-cloud-creds"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace_v1.prod]

  data = {
    GRAFANA_CLOUD_PROMETHEUS_USERNAME = var.GRAFANA_CLOUD_PROMETHEUS_USERNAME
    GRAFANA_CLOUD_PROMETHEUS_PASSWORD = var.GRAFANA_CLOUD_PROMETHEUS_PASSWORD
    GRAFANA_CLOUD_LOKI_USERNAME       = var.GRAFANA_CLOUD_LOKI_USERNAME
    GRAFANA_CLOUD_LOKI_PASSWORD       = var.GRAFANA_CLOUD_LOKI_PASSWORD
    GRAFANA_CLOUD_TEMPO_USERNAME      = var.GRAFANA_CLOUD_TEMPO_USERNAME
    GRAFANA_CLOUD_TEMPO_PASSWORD      = var.GRAFANA_CLOUD_TEMPO_PASSWORD
    GRAFANA_CLOUD_PYROSCOPE_USERNAME  = var.GRAFANA_CLOUD_PYROSCOPE_USERNAME
    GRAFANA_CLOUD_PYROSCOPE_PASSWORD  = var.GRAFANA_CLOUD_PYROSCOPE_PASSWORD
  }
}
