package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"k8s-idp/internal/registry"
)

// Handler serves the /api routes.
type Handler struct {
	registry *registry.Registry
	client   *http.Client
}

// NewHandler creates a Handler backed by the given registry.
func NewHandler(reg *registry.Registry) *Handler {
	return &Handler{
		registry: reg,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Register mounts the API routes onto mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/services", h.listServices)
	mux.HandleFunc("GET /api/services/{id}/spec", h.proxySpec)
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {
	services := h.registry.List()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		log.Printf("listServices: encode: %v", err)
	}
}

func (h *Handler) proxySpec(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	svc, ok := h.registry.Get(id)
	if !ok {
		http.Error(w, `{"error":"service not found"}`, http.StatusNotFound)
		return
	}
	if !svc.HasSpec {
		http.Error(w, `{"error":"no OpenAPI spec configured"}`, http.StatusNotFound)
		return
	}
	target := fmt.Sprintf("http://%s:%s%s", svc.IP, svc.SpecPort, svc.SpecPath)
	resp, err := h.client.Get(target)
	if err != nil {
		http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("proxySpec: copy body: %v", err)
	}
}
