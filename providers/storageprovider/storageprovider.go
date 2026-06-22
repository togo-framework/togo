// Package storageprovider registers filesystem storage. Blank-import to enable.
package storageprovider

import (
	"github.com/togo-framework/togo"
	"github.com/togo-framework/togo/storage"
)

func init() {
	togo.RegisterProviderFunc("storage", togo.PriorityService, func(k *togo.Kernel) error {
		k.Storage = storage.NewFS(k.Config.StorageDir)
		return nil
	})
}
