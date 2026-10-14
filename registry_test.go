package togo

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// isolateRegistries empties the global plugin and provider registries for the
// duration of a test and restores them afterwards.
func isolateRegistries(t *testing.T) {
	t.Helper()
	savedPlugins, savedProviders := registry, providerRegistry
	registry, providerRegistry = nil, nil
	t.Cleanup(func() { registry, providerRegistry = savedPlugins, savedProviders })
}

// stubPlugin records its lifecycle calls into a shared log.
type stubPlugin struct {
	name        string
	priority    int
	registerErr error
	bootErr     error
	log         *[]string
}

func (p *stubPlugin) Name() string  { return p.name }
func (p *stubPlugin) Priority() int { return p.priority }
func (p *stubPlugin) Register(*Kernel) error {
	*p.log = append(*p.log, "register:"+p.name)
	return p.registerErr
}
func (p *stubPlugin) Boot(context.Context, *Kernel) error {
	*p.log = append(*p.log, "boot:"+p.name)
	return p.bootErr
}

func TestDiscoveredReturnsCopy(t *testing.T) {
	isolateRegistries(t)
	var log []string
	a := &stubPlugin{name: "a", log: &log}
	Register(a)
	got := Discovered()
	if len(got) != 1 || got[0] != a {
		t.Fatalf("Discovered() = %v, want [a]", got)
	}
	got[0] = nil
	if Discovered()[0] != a {
		t.Fatal("mutating Discovered() result changed the registry")
	}
}

func TestNewSeedsDiscoveredPlugins(t *testing.T) {
	isolateRegistries(t)
	var log []string
	Register(&stubPlugin{name: "auto", log: &log})
	k := New()
	k.Use(&stubPlugin{name: "manual", priority: -1, log: &log})
	var names []string
	for _, p := range k.Plugins() {
		names = append(names, p.Name())
	}
	if want := []string{"manual", "auto"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("Plugins() = %v, want %v", names, want)
	}
}

func TestKernelBoot(t *testing.T) {
	errReg, errBoot := errors.New("reg failed"), errors.New("boot failed")
	tests := []struct {
		name    string
		plugins []stubPlugin
		wantErr error
		wantMsg string
		wantLog []string
	}{
		{
			name:    "register all then boot all in priority order",
			plugins: []stubPlugin{{name: "feature", priority: 50}, {name: "db", priority: 0}, {name: "late", priority: 100}},
			wantLog: []string{"register:db", "register:feature", "register:late", "boot:db", "boot:feature", "boot:late"},
		},
		{
			name:    "register error stops before any boot",
			plugins: []stubPlugin{{name: "a", priority: 0}, {name: "bad", priority: 1, registerErr: errReg}, {name: "c", priority: 2}},
			wantErr: errReg,
			wantMsg: "register bad: reg failed",
			wantLog: []string{"register:a", "register:bad"},
		},
		{
			name:    "boot error stops later boots",
			plugins: []stubPlugin{{name: "a", priority: 0}, {name: "bad", priority: 1, bootErr: errBoot}, {name: "c", priority: 2}},
			wantErr: errBoot,
			wantMsg: "boot bad: boot failed",
			wantLog: []string{"register:a", "register:bad", "register:c", "boot:a", "boot:bad"},
		},
		{name: "no plugins", wantLog: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log []string
			k := newTestKernel(nil)
			for i := range tt.plugins {
				tt.plugins[i].log = &log
				k.Use(&tt.plugins[i])
			}
			err := k.Boot(context.Background())
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
					t.Fatalf("Boot err = %v, want %q wrapping %v", err, tt.wantMsg, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("Boot: %v", err)
			}
			if !reflect.DeepEqual(log, tt.wantLog) {
				t.Fatalf("lifecycle = %v, want %v", log, tt.wantLog)
			}
		})
	}
}

func TestKernelBootOnce(t *testing.T) {
	var log []string
	k := newTestKernel(nil)
	k.Use(&stubPlugin{name: "p", log: &log})
	for i := 0; i < 3; i++ {
		if err := k.Boot(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if want := []string{"register:p", "boot:p"}; !reflect.DeepEqual(log, want) {
		t.Fatalf("lifecycle after 3 Boot calls = %v, want %v", log, want)
	}
}

// A failed Boot must not mark the kernel booted, so a retry runs again.
func TestKernelBootRetryAfterFailure(t *testing.T) {
	var log []string
	p := &stubPlugin{name: "p", bootErr: errors.New("x"), log: &log}
	k := newTestKernel(nil)
	k.Use(p)
	if err := k.Boot(context.Background()); err == nil {
		t.Fatal("want error")
	}
	p.bootErr = nil
	if err := k.Boot(context.Background()); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if len(log) != 4 {
		t.Fatalf("lifecycle = %v, want two full rounds", log)
	}
}

func TestProvidersRunInPriorityOrder(t *testing.T) {
	isolateRegistries(t)
	var ran []string
	add := func(name string, prio int) {
		RegisterProviderFunc(name, prio, func(*Kernel) error { ran = append(ran, name); return nil })
	}
	add("queue", PriorityLate)
	add("log", PriorityCore)
	add("cache", PriorityService)
	add("storage", PriorityService)

	want := []string{"log", "cache", "storage", "queue"}
	if got := Providers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Providers() = %v, want %v", got, want)
	}
	New()
	if !reflect.DeepEqual(ran, want) {
		t.Fatalf("provider run order = %v, want %v", ran, want)
	}
}

// A later provider overrides a service set by an earlier one.
func TestProviderOverride(t *testing.T) {
	isolateRegistries(t)
	RegisterProviderFunc("plugin-cache", PriorityService+1, func(k *Kernel) error { k.Set("cache", "plugin"); return nil })
	RegisterProviderFunc("default-cache", PriorityService, func(k *Kernel) error { k.Set("cache", "default"); return nil })
	if v, _ := New().Get("cache"); v != "plugin" {
		t.Fatalf("cache = %v, want plugin", v)
	}
}

func TestFailingProviderIsLoggedAndSkipped(t *testing.T) {
	isolateRegistries(t)
	var buf bytes.Buffer
	ranAfter := false
	RegisterProviderFunc("swap-logger", PriorityCore, func(k *Kernel) error { k.Log = newTestKernel(&buf).Log; return nil })
	RegisterProviderFunc("broken", PriorityService, func(*Kernel) error { return errors.New("no redis") })
	RegisterProviderFunc("after", PriorityLate, func(*Kernel) error { ranAfter = true; return nil })
	New()
	if !ranAfter {
		t.Error("provider after the failing one did not run")
	}
	out := buf.String()
	for _, want := range []string{"provider failed", "provider=broken", "no redis"} {
		if !strings.Contains(out, want) {
			t.Fatalf("log = %q, missing %q", out, want)
		}
	}
}
