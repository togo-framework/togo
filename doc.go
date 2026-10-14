// Package togo is the microkernel of the togo framework (https://to-go.dev).
//
// The kernel is deliberately thin: configuration, a priority-ordered hook/event
// bus, a plugin loader and registry, a database handle, an SQL dialect registry
// and HTTP server bootstrap. Every capability — REST, GraphQL, auth, dashboard,
// resources, cache, queue, storage — ships as a [Plugin] or [Provider] in its
// own module and registers itself from an init function, so an application opts
// into features by blank-importing them.
//
// A minimal application:
//
//	func main() {
//		k := togo.New()
//		defer k.Close()
//		k.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
//			w.Write([]byte("ok"))
//		})
//		log.Fatal(k.Serve(context.Background()))
//	}
//
// Configuration is read from environment variables by [LoadConfig]; SQLite is
// the default database driver.
package togo
