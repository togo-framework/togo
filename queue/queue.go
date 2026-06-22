// Package queue provides a job queue abstraction. The default runs jobs
// in-process (async goroutines); a Redis/NATS-backed queue ships as a plugin.
package queue

import (
	"context"
	"sync"
)

// Handler processes a dispatched job payload.
type Handler func(ctx context.Context, payload any) error

// Queue dispatches named jobs to registered handlers.
type Queue interface {
	Handle(name string, h Handler)
	Dispatch(ctx context.Context, name string, payload any) error
}

type memory struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	onError  func(error)
}

// NewMemory returns an in-process queue. onError (optional) receives async errors.
func NewMemory(onError func(error)) Queue {
	return &memory{handlers: map[string]Handler{}, onError: onError}
}

func (m *memory) Handle(name string, h Handler) {
	m.mu.Lock()
	m.handlers[name] = h
	m.mu.Unlock()
}

// Dispatch runs the handler asynchronously. Unknown jobs are a no-op.
func (m *memory) Dispatch(ctx context.Context, name string, payload any) error {
	m.mu.RLock()
	h, ok := m.handlers[name]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	go func() {
		if err := h(context.WithoutCancel(ctx), payload); err != nil && m.onError != nil {
			m.onError(err)
		}
	}()
	return nil
}
