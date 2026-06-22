// Package realtimeprovider registers the SSE realtime broker. Blank-import to enable.
package realtimeprovider

import (
	"github.com/togo-framework/togo"
	"github.com/togo-framework/togo/realtime"
)

func init() {
	togo.RegisterProviderFunc("realtime", togo.PriorityService, func(k *togo.Kernel) error {
		k.Realtime = realtime.NewBroker()
		return nil
	})
}
