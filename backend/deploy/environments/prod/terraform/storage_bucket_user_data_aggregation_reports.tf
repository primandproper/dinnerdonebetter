resource "google_storage_bucket" "user_data_storage" {
  provider                    = google
  name                        = "${local.gcp_project_id}-userdata"
  location                    = "US"
  uniform_bucket_level_access = false
  force_destroy               = true

  # Versioning off: the reaper deleting an artifact has to mean the artifact is gone. With
  # versioning on, a delete only writes a tombstone and the object survives as a noncurrent
  # version, so the expiry this bucket exists to honor would not actually happen.
  versioning {
    enabled = false
  }

  cors {
    origin          = ["https://${local.public_domain}"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  # Backstop, not the mechanism. The reaper deletes each artifact when its disclosure expires
  # after seven days; this catches anything the reaper never got to, and is deliberately shorter
  # than it was so a missed object does not sit for a month.
  lifecycle_rule {
    condition {
      age = 14
    }
    action {
      type = "Delete"
    }
  }

  # Purges the noncurrent versions that accumulated while versioning was on. Without this,
  # turning versioning off leaves every previously "deleted" artifact intact.
  lifecycle_rule {
    condition {
      with_state = "ARCHIVED"
    }
    action {
      type = "Delete"
    }
  }
}

resource "google_storage_bucket_iam_policy" "user_data_policy" {
  bucket      = google_storage_bucket.user_data_storage.name
  policy_data = data.google_iam_policy.user_data_policy.policy_data
}

# Domain-named bucket for user data. Requires Search Console domain verification.
resource "google_storage_bucket" "user_data_domain" {
  provider                    = google
  name                        = local.userdata_domain
  location                    = "US"
  uniform_bucket_level_access = false
  force_destroy               = true

  # Versioning off: the reaper deleting an artifact has to mean the artifact is gone. With
  # versioning on, a delete only writes a tombstone and the object survives as a noncurrent
  # version, so the expiry this bucket exists to honor would not actually happen.
  versioning {
    enabled = false
  }

  cors {
    origin          = ["https://${local.public_domain}"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  # Backstop, not the mechanism. The reaper deletes each artifact when its disclosure expires
  # after seven days; this catches anything the reaper never got to, and is deliberately shorter
  # than it was so a missed object does not sit for a month.
  lifecycle_rule {
    condition {
      age = 14
    }
    action {
      type = "Delete"
    }
  }

  # Purges the noncurrent versions that accumulated while versioning was on. Without this,
  # turning versioning off leaves every previously "deleted" artifact intact.
  lifecycle_rule {
    condition {
      with_state = "ARCHIVED"
    }
    action {
      type = "Delete"
    }
  }
}

resource "google_storage_bucket_iam_policy" "user_data_domain_policy" {
  bucket      = google_storage_bucket.user_data_domain.name
  policy_data = data.google_iam_policy.user_data_policy.policy_data
}
