package mealplantasknotifications

import (
	"github.com/primandproper/platform-go/v13/workqueue"
)

// Item is one leased meal plan task, and Stats is the queue's shape. They are aliases rather
// than types of their own so a *workqueue.Queue[string] satisfies Queue without an adapter,
// while the worker's own signatures stay free of the generic.
type (
	Item  = workqueue.Item[string]
	Stats = workqueue.Stats
)

// TaskQueue is this worker's work queue, wrapped in a type of its own.
//
// The wrapper exists because the operations tier is also built on workqueue and also registers a
// *workqueue.Queue[string]: two unnamed providers of one type collide in the container, and —
// worse than colliding — whichever survived would be handed to whoever asked. The two queues are
// rows in the same table, told apart by Config.Name, so nothing about them is interchangeable
// and nothing should be able to receive one in place of the other.
//
// It embeds the queue rather than restating it: everything the worker drives, and the Close that
// shutdown owes the queue's batching goroutine, is promoted.
type TaskQueue struct {
	*workqueue.Queue[string]
}
