package togo

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestHooksFireOrder(t *testing.T) {
	tests := []struct {
		name       string
		priorities []int // registration order; listener i records label i
		want       []int // expected run order of labels
	}{
		{name: "ascending priority", priorities: []int{50, 0, 100, 10}, want: []int{1, 3, 0, 2}},
		{name: "equal priority keeps registration order", priorities: []int{5, 5, 5}, want: []int{0, 1, 2}},
		{name: "mixed ties", priorities: []int{10, 0, 10, 0}, want: []int{1, 3, 0, 2}},
		{name: "no listeners", priorities: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHooks()
			var got []int
			for i, p := range tt.priorities {
				h.On("evt", p, func(context.Context, any) error { got = append(got, i); return nil })
			}
			h.On("other", 0, func(context.Context, any) error { t.Error("listener for another event ran"); return nil })
			if err := h.Fire(context.Background(), "evt", nil); err != nil {
				t.Fatalf("Fire: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("run order = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHooksFireStopsOnError(t *testing.T) {
	h := newHooks()
	errStop := errors.New("stop")
	var ran []string
	h.On("save", 0, func(context.Context, any) error { ran = append(ran, "first"); return nil })
	h.On("save", 1, func(context.Context, any) error { ran = append(ran, "second"); return errStop })
	h.On("save", 2, func(context.Context, any) error { ran = append(ran, "third"); return nil })

	if err := h.Fire(context.Background(), "save", nil); !errors.Is(err, errStop) {
		t.Fatalf("Fire err = %v, want %v", err, errStop)
	}
	if want := []string{"first", "second"}; !reflect.DeepEqual(ran, want) {
		t.Fatalf("ran = %v, want %v", ran, want)
	}
}

func TestHooksFirePassesContextAndPayload(t *testing.T) {
	type key struct{}
	h := newHooks()
	ctx := context.WithValue(context.Background(), key{}, "v")
	var gotCtx, gotPayload any
	h.On("e", 0, func(c context.Context, p any) error { gotCtx, gotPayload = c.Value(key{}), p; return nil })
	_ = h.Fire(ctx, "e", 123)
	if gotCtx != "v" || gotPayload != 123 {
		t.Fatalf("listener got ctx value %v, payload %v", gotCtx, gotPayload)
	}
}

// A listener registered while an event is firing must not run in that same
// Fire (Fire snapshots listeners) and must not deadlock.
func TestHooksOnDuringFire(t *testing.T) {
	h := newHooks()
	calls := 0
	h.On("e", 0, func(context.Context, any) error {
		calls++
		h.On("e", 1, func(context.Context, any) error { calls += 10; return nil })
		return nil
	})
	_ = h.Fire(context.Background(), "e", nil)
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestHooksZeroValueUsable(t *testing.T) {
	var h Hooks
	ran := false
	h.On("e", 0, func(context.Context, any) error { ran = true; return nil })
	if err := h.Fire(context.Background(), "e", nil); err != nil || !ran {
		t.Fatalf("zero-value Hooks: ran=%v err=%v", ran, err)
	}
}

func TestHooksConcurrent(t *testing.T) {
	h := newHooks()
	var mu sync.Mutex
	n := 0
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			h.On("e", 0, func(context.Context, any) error { mu.Lock(); n++; mu.Unlock(); return nil })
		}()
		go func() { defer wg.Done(); _ = h.Fire(context.Background(), "e", nil) }()
	}
	wg.Wait()
	n = 0
	_ = h.Fire(context.Background(), "e", nil)
	if n != 50 {
		t.Fatalf("after 50 concurrent On calls, Fire ran %d listeners", n)
	}
}
