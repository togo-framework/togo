// Package queueprovider registers the in-process job queue. Blank-import to enable.
package queueprovider

import (
	"github.com/togo-framework/togo"
	"github.com/togo-framework/togo/queue"
)

func init() {
	togo.RegisterProviderFunc("queue", togo.PriorityLate, func(k *togo.Kernel) error {
		k.Queue = queue.NewMemory(func(err error) { k.Log.Error("queue job failed", "err", err) })
		return nil
	})
}
