resource "google_pubsub_topic" "data_changes_topic" {
  name = "data_changes"
}

resource "google_pubsub_topic" "data_changes_deadletter_topic" {
  name = "data_changes_deadletter"
}

resource "google_pubsub_subscription" "data_changes_topic" {
  name  = google_pubsub_topic.data_changes_topic.name
  topic = google_pubsub_topic.data_changes_topic.id

  message_retention_duration = "604800s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.data_changes_deadletter_topic.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  enable_exactly_once_delivery = true
}

resource "google_pubsub_topic" "outbound_emails_topic" {
  name = "outbound_emails"
}

resource "google_pubsub_topic" "outbound_emails_deadletter_topic" {
  name = "outbound_emails_deadletter"
}

resource "google_pubsub_subscription" "outbound_emails_topic" {
  name  = google_pubsub_topic.outbound_emails_topic.name
  topic = google_pubsub_topic.outbound_emails_topic.id

  message_retention_duration = "604800s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.outbound_emails_deadletter_topic.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  enable_exactly_once_delivery = true
}

# The mail platform's identity, sign-in, password reset and waitlist doors queue. Every message
# carries the secret its mail delivers, so the async message handler's mail Drainer is the topic's
# only subscriber, and nothing is retained past what a retry needs: a day here, and a day on the
# dead_letter topic below for a mail the Drainer gave up on.
resource "google_pubsub_topic" "queued_mail_topic" {
  name = "queued_mail"
}

resource "google_pubsub_topic" "queued_mail_deadletter_topic" {
  name = "queued_mail_deadletter"
}

resource "google_pubsub_subscription" "queued_mail_topic" {
  name  = google_pubsub_topic.queued_mail_topic.name
  topic = google_pubsub_topic.queued_mail_topic.id

  message_retention_duration = "86400s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.queued_mail_deadletter_topic.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  enable_exactly_once_delivery = true
}

# Where the async message handler's pools dead-letter a message they gave up on: refused as
# unretryable, or failed past its attempts (pools.deadLetterTopicName). The publisher creates no
# topic, so without this one every dead letter is a failed publish.
#
# Queued mail dead-letters here with its payload, and so with the secret the mail delivers, so
# nothing is kept longer than the queued_mail subscription keeps it. The subscription is for an
# operator to pull from; nothing consumes it, so it is kept from expiring for inactivity.
resource "google_pubsub_topic" "dead_letter_topic" {
  name = "dead_letter"
}

resource "google_pubsub_subscription" "dead_letter_topic" {
  name  = google_pubsub_topic.dead_letter_topic.name
  topic = google_pubsub_topic.dead_letter_topic.id

  message_retention_duration = "86400s"
  retain_acked_messages      = false

  expiration_policy {
    ttl = ""
  }
}

resource "google_pubsub_topic" "mobile_notifications_topic" {
  name = "mobile_notifications"
}

resource "google_pubsub_topic" "mobile_notifications_deadletter_topic" {
  name = "mobile_notifications_deadletter"
}

resource "google_pubsub_subscription" "mobile_notifications_topic" {
  name  = google_pubsub_topic.mobile_notifications_topic.name
  topic = google_pubsub_topic.mobile_notifications_topic.id

  message_retention_duration = "604800s" # 7 days
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.mobile_notifications_deadletter_topic.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  enable_exactly_once_delivery = true
}

# There is no user data aggregation topic. A GDPR export is a request row a worker claims, not
# a message on a broker — see the dataprivacy adoption. Removing this destroys the topic and its
# subscription; anything still sitting in the dead-letter topic at that point was already
# undeliverable and is not carried across, since the code that would consume it is gone.

resource "google_pubsub_topic" "webhook_execution_requests_topic" {
  name = "webhook_execution_requests"
}

resource "google_pubsub_topic" "webhook_execution_requests_deadletter_topic" {
  name = "webhook_execution_requests_deadletter"
}

resource "google_pubsub_subscription" "webhook_execution_requests_topic" {
  name  = google_pubsub_topic.webhook_execution_requests_topic.name
  topic = google_pubsub_topic.webhook_execution_requests_topic.id

  message_retention_duration = "604800s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.webhook_execution_requests_deadletter_topic.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  enable_exactly_once_delivery = true
}
