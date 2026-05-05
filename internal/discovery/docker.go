package discovery

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"k8s-idp/internal/registry"
)

// DockerWatcher watches containers with k8s-idp/enabled=true and syncs to the registry.
type DockerWatcher struct {
	client   *client.Client
	registry *registry.Registry
}

// NewDockerWatcher creates a DockerWatcher using DOCKER_HOST or the default socket.
func NewDockerWatcher(reg *registry.Registry) (*DockerWatcher, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &DockerWatcher{client: cli, registry: reg}, nil
}

// Start loads existing containers then watches events until ctx is cancelled.
func (w *DockerWatcher) Start(ctx context.Context) {
	if err := w.loadRunning(ctx); err != nil {
		log.Printf("docker: initial load error: %v", err)
	}

	f := filters.NewArgs(
		filters.Arg("type", "container"),
		filters.Arg("event", "start"),
		filters.Arg("event", "die"),
		filters.Arg("event", "stop"),
		filters.Arg("label", "k8s-idp/enabled=true"),
	)
	eventCh, errCh := w.client.Events(ctx, types.EventsOptions{Filters: f})
	for {
		select {
		case ev := <-eventCh:
			switch ev.Action {
			case "start":
				w.addContainer(ctx, ev.Actor.ID)
			case "die", "stop":
				id := ev.Actor.ID
				if len(id) > 12 {
					id = id[:12]
				}
				w.registry.Remove(id)
			}
		case err := <-errCh:
			if err != nil && ctx.Err() == nil {
				log.Printf("docker: event stream error: %v", err)
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

func (w *DockerWatcher) loadRunning(ctx context.Context) error {
	f := filters.NewArgs(
		filters.Arg("label", "k8s-idp/enabled=true"),
		filters.Arg("status", "running"),
	)
	containers, err := w.client.ContainerList(ctx, types.ContainerListOptions{Filters: f})
	if err != nil {
		return err
	}
	for _, c := range containers {
		ip := firstIP(c.NetworkSettings.Networks)
		if svc, ok := serviceFromContainer(c, ip); ok {
			w.registry.Add(svc)
		}
	}
	return nil
}

func (w *DockerWatcher) addContainer(ctx context.Context, id string) {
	info, err := w.client.ContainerInspect(ctx, id)
	if err != nil {
		log.Printf("docker: inspect %s: %v", id, err)
		return
	}
	ip := firstIP(info.NetworkSettings.Networks)
	c := types.Container{
		ID:     info.ID,
		Names:  []string{info.Name},
		Labels: info.Config.Labels,
	}
	if svc, ok := serviceFromContainer(c, ip); ok {
		w.registry.Add(svc)
	}
}

func firstIP(networks map[string]*networktypes.EndpointSettings) string {
	for _, n := range networks {
		if n != nil && n.IPAddress != "" {
			return n.IPAddress
		}
	}
	return ""
}

// serviceFromContainer converts a Docker container into a Service.
func serviceFromContainer(c types.Container, ip string) (registry.Service, bool) {
	labels := c.Labels
	if labels["k8s-idp/enabled"] != "true" {
		return registry.Service{}, false
	}
	if ip == "" {
		return registry.Service{}, false
	}
	name := labels["k8s-idp/name"]
	if name == "" && len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}
	port := labels["k8s-idp/port"]
	if port == "" {
		port = "8080"
	}
	specPath := labels["k8s-idp/openapi-path"]
	shortID := c.ID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	containerName := ""
	if len(c.Names) > 0 {
		containerName = strings.TrimPrefix(c.Names[0], "/")
	}
	return registry.Service{
		ID:            shortID,
		Name:          name,
		Description:   labels["k8s-idp/description"],
		Owner:         labels["k8s-idp/owner"],
		SourceURL:     labels["k8s-idp/source-url"],
		HasSpec:       specPath != "",
		SpecPath:      specPath,
		SpecPort:      port,
		IP:            ip,
		ContainerName: containerName,
		ContainerID:   shortID,
	}, true
}
