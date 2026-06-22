// Package realtime defines the realtime broker contract. Implementations live in
// their own repos (e.g. github.com/togo-framework/realtime) and register a provider.
package realtime

import "net/http"

// Broker fans out server-push messages to connected clients.
type Broker interface {
	Publish(event, data string)
	Handler() http.HandlerFunc
}
