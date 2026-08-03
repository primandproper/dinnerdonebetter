# User data bucket policy: no public binding at all.
#
# This bucket holds user data disclosure artifacts — one object per person containing everything
# the system knows about them. It was previously bound to the same allUsers/objectViewer policy
# the media bucket uses, which made every report world-readable to anyone who could guess or
# learn a report ID. Nothing reads these objects over HTTP: the API server reads them with this
# service account and returns the contents over an authenticated gRPC call, and the scheduler
# uses the same account to destroy them when they expire.
data "google_iam_policy" "user_data_policy" {
  binding {
    role = "roles/storage.objectAdmin"
    members = [
      "serviceAccount:workload-identity-sa@${local.gcp_project_id}.iam.gserviceaccount.com",
    ]
  }
}

# Media bucket policy: public read + workload-identity-sa write (for API uploads)
data "google_iam_policy" "api_media_policy" {
  binding {
    role = "roles/storage.objectViewer"
    members = [
      "allUsers",
    ]
  }
  binding {
    role = "roles/storage.objectAdmin"
    members = [
      "serviceAccount:workload-identity-sa@${local.gcp_project_id}.iam.gserviceaccount.com",
    ]
  }
}
