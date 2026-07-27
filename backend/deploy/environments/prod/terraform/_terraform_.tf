terraform {
  required_version = "1.14.5"

  backend "remote" {
    organization = "dinnerdonebetter" # must be literal; backend blocks don't support interpolation

    workspaces {
      name = "prod-backend"
    }
  }
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.6.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "5.18.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "7.23.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "2.7.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "3.0.1"
    }
    algolia = {
      source  = "philippe-vandermoere/algolia"
      version = "0.7.0"
    }
    grafana = {
      source  = "grafana/grafana"
      version = "4.28.0"
    }
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "1.27.0"
    }
  }
}
