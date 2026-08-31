package notifications

// NotificationKindMealPlanTask names the prep task reminder in the delivery metrics.
//
// It is no longer a routing key. The reminder used to travel as a message on the mobile
// notifications topic, tagged with this, and the consumer switched on it; it is now claimed from
// a work queue by the job that owns it, so the only thing left that reads this is a metric
// label.
const NotificationKindMealPlanTask = "meal_plan_task"
