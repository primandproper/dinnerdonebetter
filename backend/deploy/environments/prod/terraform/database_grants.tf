# Privileges for the per-service database users created in database_users.tf.
#
# google_sql_user creates a role but grants it nothing, so without this file every
# service user could connect and then read nothing. These grants previously lived in
# schema migration 00020_service_users.sql, which had to create the roles itself
# (with a PL/pgSQL loop, since Postgres has no CREATE ROLE IF NOT EXISTS) purely so
# the grants would not fail in localdev and CI, where no Terraform runs. Role and
# privilege management now live together here, and the migrations know nothing about
# service users.
#
# Unlike the other providers here, this one opens a real connection to Postgres during
# plan and apply. It connects through the Cloud SQL connector (scheme "gcppostgres")
# rather than a raw IP, which is what keeps it workable from a Terraform Cloud run:
# the connector authenticates over the Cloud SQL Admin API, so the instance needs no
# authorized_networks entry — and it has none. The prerequisites are instead that the
# Cloud SQL Admin API is enabled and the workspace's GCP service account holds
# roles/cloudsql.client.
#
# FIRST APPLY: this provider is configured from resources in this same configuration
# (the instance's connection name, the generated api_db_user password), which Terraform
# cannot resolve while those resources do not yet exist. On a greenfield apply, create
# the instance and users first — `terraform apply -target=google_sql_user.database_user`
# — then apply normally to add the grants. Subsequent applies are single-pass.
#
# None of this has been exercised against a live instance, because nothing is deployed.
# Treat it as unverified until that first apply.
#
# All services get full access. The goal is per-service identity in pg_stat_activity,
# not privilege isolation; this preserves the migrations' original GRANT ALL. If real
# least-privilege is ever wanted, this is the file to narrow — per-service privilege
# lists here would be the whole change.

locals {
  # Every service role from database_users.tf gets the same grants, so adding a user
  # there is all it takes to have it provisioned here too. Values are read off the
  # google_sql_user resources rather than the username strings, which is what makes each
  # grant depend on its role already existing without needing a depends_on block.
  service_database_grantees = {
    for username in local.service_database_usernames :
    username => google_sql_user.database_user[username].name
  }
}

# Connects as api_db_user rather than a superuser: it owns the tables the migrations
# create, and an owner can both grant on its objects and set its own default
# privileges, so no elevated credentials are needed here.
provider "postgresql" {
  scheme    = "gcppostgres"
  host      = google_sql_database_instance.prod.connection_name
  username  = local.api_database_username
  password  = random_password.database_user[local.api_database_username].result
  superuser = false
}

# Connecting is not implied by table privileges: a role needs CONNECT on the database
# and USAGE on the schema before any of the grants below matter. Postgres 15 dropped
# the implicit grant of schema privileges to PUBLIC and this instance runs Postgres 17,
# so both must be explicit.
resource "postgresql_grant" "service_users_connect" {
  for_each = local.service_database_grantees

  role        = each.value
  database    = google_sql_database.api_database.name
  object_type = "database"
  privileges  = ["CONNECT"]
}

resource "postgresql_grant" "service_users_schema_usage" {
  for_each = local.service_database_grantees

  role        = each.value
  database    = google_sql_database.api_database.name
  schema      = "public"
  object_type = "schema"
  privileges  = ["USAGE"]
}

# Future tables and sequences. Default privileges are recorded per creating role, so
# these are scoped to api_db_user — the role the API server runs migrations as. With
# them in place before migrations run, every table a migration creates carries these
# grants at creation, which is what removes the ordering dependency that kept the
# grants in the migrations to begin with.
resource "postgresql_default_privileges" "service_users_future_tables" {
  for_each = local.service_database_grantees

  role        = each.value
  database    = google_sql_database.api_database.name
  schema      = "public"
  owner       = local.api_database_username
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"]
}

resource "postgresql_default_privileges" "service_users_future_sequences" {
  for_each = local.service_database_grantees

  role        = each.value
  database    = google_sql_database.api_database.name
  schema      = "public"
  owner       = local.api_database_username
  object_type = "sequence"
  privileges  = ["SELECT", "UPDATE", "USAGE"]
}

# Tables and sequences that already exist. Default privileges only ever apply to
# objects created after they are set, so a database migrated before this file existed
# needs these to catch up. On a database provisioned from scratch they are a no-op.
resource "postgresql_grant" "service_users_existing_tables" {
  for_each = local.service_database_grantees

  role        = each.value
  database    = google_sql_database.api_database.name
  schema      = "public"
  object_type = "table"
  objects     = [] # empty means every table in the schema
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"]
}

resource "postgresql_grant" "service_users_existing_sequences" {
  for_each = local.service_database_grantees

  role        = each.value
  database    = google_sql_database.api_database.name
  schema      = "public"
  object_type = "sequence"
  objects     = [] # empty means every sequence in the schema
  privileges  = ["SELECT", "UPDATE", "USAGE"]
}
