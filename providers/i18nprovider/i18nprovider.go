// Package i18nprovider registers JSON-keyed translations. Blank-import to enable.
package i18nprovider

import (
	"github.com/togo-framework/togo"
	"github.com/togo-framework/togo/i18n"
)

func init() {
	togo.RegisterProviderFunc("i18n", togo.PriorityService, func(k *togo.Kernel) error {
		k.I18n = i18n.Load(k.Config.LocaleDir, k.Config.Locale)
		return nil
	})
}
