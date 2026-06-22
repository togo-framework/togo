package togo

import (
	"context"
	"sort"
	"sync"
)

// HookFunc handles a fired event. Returning an error stops the chain.
type HookFunc func(ctx context.Context, payload any) error

type listener struct {
	priority int
	fn       HookFunc
}

// Hooks is a priority-ordered event bus for lifecycle events (resource CRUD,
// auth, plugin load). Listeners run in ascending priority (0 first).
type Hooks struct {
	mu        sync.RWMutex
	listeners map[string][]listener
}

func newHooks() *Hooks { return &Hooks{listeners: map[string][]listener{}} }

// On registers a listener for an event at the given priority (0–100).
func (h *Hooks) On(event string, priority int, fn HookFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ls := append(h.listeners[event], listener{priority: priority, fn: fn})
	sort.SliceStable(ls, func(i, j int) bool { return ls[i].priority < ls[j].priority })
	h.listeners[event] = ls
}

// Fire invokes every listener for an event in priority order, stopping on the
// first error.
func (h *Hooks) Fire(ctx context.Context, event string, payload any) error {
	h.mu.RLock()
	ls := append([]listener(nil), h.listeners[event]...)
	h.mu.RUnlock()
	for _, l := range ls {
		if err := l.fn(ctx, payload); err != nil {
			return err
		}
	}
	return nil
}
