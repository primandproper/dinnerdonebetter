# Per-service database roles.
#
# TO ADD A SERVICE: put its username in service_database_usernames below. That is the
# whole change — the password, the Cloud SQL user, the Kubernetes secret entry, and
# every grant in database_grants.tf are all derived from this list.
#
# Two things still have to be done by hand for a new service, because neither lives in
# Terraform: point the service's config at the username (deploy/environments/prod/
# kustomize/configs/*.json) and add a kustomize patch wiring DATABASE_<NAME>_PASSWORD
# from the secret into the pod (deploy/environments/prod/kustomize/patches/).
#
# The username is used verbatim as the Postgres role name, so it must be a valid
# unquoted identifier: lowercase, no dashes.

locals {
  # The role the API server runs migrations as. It owns every table those migrations
  # create, which is why it never appears as a grantee in database_grants.tf — an owner
  # needs no grant to reach its own objects.
  api_database_username = "api_db_user"

  # Every other service's role. Keep sorted; the order has no effect.
  service_database_usernames = toset([
    "async_message_handler",
    "db_cleaner",
    "mcp_server",
    "meal_plan_finalizer",
    "meal_plan_grocery_list_initializer",
    "meal_plan_task_creator",
    "mobile_notification_scheduler",
    "queue_test",
    "search_data_index_scheduler",
  ])

  database_usernames = setunion(local.service_database_usernames, [local.api_database_username])

  # Consumed by the application secret in kubernetes.tf. Keys are DATABASE_<USERNAME>_
  # PASSWORD, matching what the kustomize patches read; api_db_user is the one
  # exception, having been wired as DATABASE_API_PASSWORD before this was generated.
  database_password_secret_data = merge(
    {
      DATABASE_API_PASSWORD = random_password.database_user[local.api_database_username].result
    },
    {
      for username in local.service_database_usernames :
      "DATABASE_${upper(username)}_PASSWORD" => random_password.database_user[username].result
    },
  )
}

resource "random_password" "database_user" {
  for_each = local.database_usernames

  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "database_user" {
  for_each = local.database_usernames

  name     = each.value
  instance = google_sql_database_instance.prod.name
  password = random_password.database_user[each.key].result
}
