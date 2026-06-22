// Package cacheprovider registers the default in-memory cache. Blank-import to enable.
package cacheprovider

import (
	"github.com/togo-framework/togo"
	"github.com/togo-framework/togo/cache"
)

func init() {
	togo.RegisterProviderFunc("cache", togo.PriorityService, func(k *togo.Kernel) error {
		k.Cache = cache.NewMemory()
		return nil
	})
}
