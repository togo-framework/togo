// Package queue defines the job queue contract. Implementations live in their
// own repos (e.g. github.com/togo-framework/queue) and register a provider.
package queue

import "context"

// Handler processes a dispatched job payload.
type Handler func(ctx context.Context, payload any) error

// Queue dispatches named jobs to registered handlers.
type Queue interface {
	Handle(name string, h Handler)
	Dispatch(ctx context.Context, name string, payload any) error
}
