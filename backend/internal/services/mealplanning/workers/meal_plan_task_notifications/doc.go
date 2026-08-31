/*
Package mealplantasknotifications sends each meal plan prep task's reminder push exactly once,
driven by a platform-go workqueue.

# What this replaces

It used to be two halves that never agreed on when a task was done. A scheduled job queried for
"every incomplete task with notification_sent_at IS NULL whose event has not started" and
published one message per task onto the mobile notifications topic; the async message consumer
sent the pushes and stamped notification_sent_at. The stamp was the only record that a task had
been handled, and it landed in a different process, in a different transaction, some time after
the job had already moved on.

That is a claim/complete queue with the claim left out, and it failed the way one always does:

  - Nothing was leased, so every tick republished every task the consumer had not yet stamped.
  - Nothing counted attempts, so a task whose notification could not be built — an account that
    no longer exists, a context row that went missing — was rediscovered, republished and
    re-failed on every tick until its event started. The failure was a log line each time and
    nothing else.
  - Nothing bounded a pass, so the first tick after a backlog published all of it at once.
  - Nothing recorded why a task had failed, or how long the backlog had been waiting.

# What is here instead

One queue, keyed by meal plan task ID, and one job that both fills and drains it:

	discover -> Enqueue -> Claim (leased) -> send -> mark sent -> Complete

The lease is the piece that was missing. A task claimed by this pass is invisible to the next
one, and a pass that dies mid-batch loses its leases rather than its work — the tasks come back
when the leases lapse.

# Why the send happens under the lease

Because Complete has to mean the same thing the discovery query means. notification_sent_at is
what makes a task stop being discoverable, so if the queue item completed at publish time and
the stamp landed later in another process, the next discovery would find the task still unsent,
re-enqueue it, and restart the completed item — the original bug with a table in front of it.

So the work under the lease runs to the stamp: fan the push out to the recipients' devices, mark
the task notified, complete. The two facts are then the same fact, and re-enqueueing a task that
is genuinely still owed a notification is exactly the right thing for it to do.

MealPlanTaskNotificationHasBeenSent is still checked first. A lease can lapse while its holder
is merely slow, so two passes can briefly hold one task, and the read is what keeps that from
being a second push.

# Delivery still belongs to somebody else

The push itself goes through notifications/push, which the async message consumer also uses for
the notifications that genuinely are message-driven. This package decides which tasks are owed a
reminder and what it says; it does not know what a device token is.
*/
package mealplantasknotifications
