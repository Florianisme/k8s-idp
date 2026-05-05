package registry_test

import (
	"fmt"
	"sync"
	"testing"
)

func TestAddAndList(t *testing.T) {
	r := registry.New()
	r.Add(registry.Service{ID: "ns_svc", Name: "My Service"})

	list := r.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 service, got %d", len(list))
	}
	if list[0].Name != "My Service" {
		t.Errorf("expected Name %q, got %q", "My Service", list[0].Name)
	}
}

func TestGet(t *testing.T) {
	r := registry.New()
	r.Add(registry.Service{ID: "ns_svc", Name: "X"})

	svc, ok := r.Get("ns_svc")
	if !ok {
		t.Fatal("expected to find service")
	}
	if svc.Name != "X" {
		t.Errorf("expected %q, got %q", "X", svc.Name)
	}

	_, ok = r.Get("missing")
	if ok {
		t.Error("expected not to find missing service")
	}
}

func TestRemove(t *testing.T) {
	r := registry.New()
	r.Add(registry.Service{ID: "a"})
	r.Add(registry.Service{ID: "b"})
	r.Remove("a")

	list := r.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 after remove, got %d", len(list))
	}
	if list[0].ID != "b" {
		t.Errorf("expected ID %q, got %q", "b", list[0].ID)
	}

	// Test removing a non-existent ID should not panic and not change state
	r.Remove("nonexistent")
	list = r.List()
	if len(list) != 1 {
		t.Errorf("expected 1 after removing non-existent, got %d", len(list))
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := registry.New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("svc-%d", i)
			r.Add(registry.Service{ID: id})
			r.List()
			r.Remove(id)
		}(i)
	}
	wg.Wait()

	// After all goroutines finish, the registry should be empty (each goroutine adds and then removes its service)
	list := r.List()
	if len(list) != 0 {
		t.Errorf("expected empty registry after all removes, got %d entries", len(list))
	}
}

func TestAddReplacesOnDuplicateID(t *testing.T) {
	r := registry.New()
	r.Add(registry.Service{ID: "x", Name: "first"})
	r.Add(registry.Service{ID: "x", Name: "second"})

	svc, ok := r.Get("x")
	if !ok {
		t.Fatal("expected service")
	}
	if svc.Name != "second" {
		t.Errorf("expected second value to win, got %q", svc.Name)
	}
	if list := r.List(); len(list) != 1 {
		t.Errorf("expected 1 entry after upsert, got %d", len(list))
	}
}
