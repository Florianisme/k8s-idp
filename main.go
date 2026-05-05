package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s-idp/internal/api"
	"k8s-idp/internal/discovery"
	"k8s-idp/internal/registry"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

func main() {
	mode := getEnv("MODE", "kubernetes")
	port := getEnv("PORT", "8080")

	reg := registry.New()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch mode {
	case "kubernetes":
		w, err := discovery.NewKubernetesWatcher(reg)
		if err != nil {
			log.Fatalf("kubernetes watcher: %v", err)
		}
		go w.Start(ctx)
	case "docker":
		w, err := discovery.NewDockerWatcher(reg)
		if err != nil {
			log.Fatalf("docker watcher: %v", err)
		}
		go w.Start(ctx)
	default:
		log.Fatalf("unknown MODE %q — must be 'kubernetes' or 'docker'", mode)
	}

	mux := http.NewServeMux()
	api.NewHandler(reg).Register(mux)

	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("embedding frontend: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// For SPA routing: serve index.html for paths that don't match a real file.
		if _, err := fs.Stat(distFS, r.URL.Path[1:]); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux}
	log.Printf("k8s-idp listening on :%s (MODE=%s)", port, mode)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
