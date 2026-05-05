package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMux(reg *registry.Registry) *http.ServeMux {
	mux := http.NewServeMux()
	api.NewHandler(reg).Register(mux)
	return mux
}

func TestListServices_Empty(t *testing.T) {
	mux := newMux(registry.New())
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/services", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []registry.Service
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty list, got %d", len(out))
	}
}

func TestListServices_WithService(t *testing.T) {
	reg := registry.New()
	reg.Add(registry.Service{ID: "ns_svc", Name: "Test"})
	mux := newMux(reg)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/services", nil))

	var out []registry.Service
	json.NewDecoder(w.Body).Decode(&out)
	if len(out) != 1 || out[0].Name != "Test" {
		t.Errorf("unexpected services: %+v", out)
	}
}

func TestProxySpec_ServiceNotFound(t *testing.T) {
	mux := newMux(registry.New())
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/services/missing/spec", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProxySpec_NoSpec(t *testing.T) {
	reg := registry.New()
	reg.Add(registry.Service{ID: "ns_svc", HasSpec: false})
	mux := newMux(reg)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/services/ns_svc/spec", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProxySpec_Proxies(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"openapi":"3.0.0"}`))
	}))
	defer upstream.Close()

	addr := upstream.Listener.Addr().String()
	lastColon := strings.LastIndex(addr, ":")
	host, port := addr[:lastColon], addr[lastColon+1:]

	reg := registry.New()
	reg.Add(registry.Service{
		ID: "ns_svc", HasSpec: true,
		IP: host, SpecPort: port, SpecPath: "/",
	})
	mux := newMux(reg)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/services/ns_svc/spec", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body != `{"openapi":"3.0.0"}` {
		t.Errorf("unexpected body: %s", body)
	}
}
