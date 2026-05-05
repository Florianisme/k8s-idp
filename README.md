# k8s-idp

An internal developer portal that auto-discovers services from Kubernetes Pods or Docker containers and renders their API documentation via Swagger UI.

## Features

- **Auto-discovery** — watches Kubernetes Pods or Docker containers for `k8s-idp/` labels; no manual service registration
- **API documentation** — renders OpenAPI/Swagger specs inline via Swagger UI
- **Service metadata** — shows name, description, owner, and source code link for each service
- **Single container** — Go backend embeds the Vue 3 frontend; one image, one port

## Quick Start

### Docker mode

```bash
docker run --rm -p 8080:8080 \
  -e MODE=docker \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/fme/k8s-idp:latest
```

Then label any running container to make it appear in the portal:

```bash
docker run -d \
  --label k8s-idp/enabled=true \
  --label "k8s-idp/name=My Service" \
  --label "k8s-idp/description=Does something useful" \
  --label k8s-idp/owner=my-team \
  --label k8s-idp/source-url=https://github.com/org/my-service \
  --label k8s-idp/openapi-path=/swagger/docs/openapi \
  my-service:latest
```

### Kubernetes mode

```bash
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/deployment.yaml
```

Label your pods:

```yaml
metadata:
  labels:
    k8s-idp/enabled: "true"
    k8s-idp/name: "Payment API"
    k8s-idp/description: "Handles all payment processing"
    k8s-idp/owner: "payments-team"
    k8s-idp/source-url: "https://github.com/org/payment-api"
    k8s-idp/openapi-path: "/swagger/docs/openapi"
```

## Labels

All labels use the `k8s-idp/` prefix and are applied directly to Pods (Kubernetes) or containers (Docker).

| Label | Required | Default | Description |
|---|---|---|---|
| `k8s-idp/enabled` | **yes** | — | Must be `"true"` to appear in the portal |
| `k8s-idp/name` | no | pod/container name | Human-readable display name |
| `k8s-idp/description` | no | — | Short description of the service |
| `k8s-idp/owner` | no | — | Team or person responsible |
| `k8s-idp/source-url` | no | — | Link to source code (GitHub, GitLab, etc.) |
| `k8s-idp/openapi-path` | no | — | Path where the service exposes its OpenAPI spec |
| `k8s-idp/port` | no | `8080` | Port to use when fetching the OpenAPI spec |

## Configuration

| Variable | Default | Description |
|---|---|---|
| `MODE` | `kubernetes` | Discovery mode: `kubernetes` or `docker` |
| `PORT` | `8080` | HTTP port the portal listens on |
| `KUBECONFIG` | auto | Path to kubeconfig (out-of-cluster Kubernetes) |
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon address |

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker or a local Kubernetes cluster (kind/minikube)

### Run locally

Terminal 1 — Go backend:

```bash
make build-frontend   # only needed once, or after frontend changes
MODE=docker go run .
```

Terminal 2 — Vite dev server (hot reload):

```bash
make dev-frontend
```

Open `http://localhost:5173`.

### Build & test

```bash
make test          # Go unit tests
make build         # build binary (runs frontend build first)
make docker-build  # build Docker image
```

### Project structure

```
k8s-idp/
├── main.go                        # entrypoint, go:embed
├── internal/
│   ├── api/handler.go             # HTTP API: list services + spec proxy
│   ├── discovery/
│   │   ├── kubernetes.go          # client-go informer watcher
│   │   └── docker.go              # Docker SDK event watcher
│   └── registry/registry.go      # thread-safe in-memory service store
├── frontend/src/
│   ├── App.vue                    # master-detail root layout
│   └── components/
│       ├── ServiceList.vue        # filterable sidebar list
│       ├── ServiceDetail.vue      # service metadata panel
│       └── SwaggerViewer.vue      # Swagger UI mount
├── k8s/
│   ├── rbac.yaml                  # ClusterRole for pod reads
│   └── deployment.yaml            # Deployment + Service
└── Dockerfile                     # multi-stage: node → go → alpine
```

## API

| Route | Description |
|---|---|
| `GET /api/services` | List all discovered services as JSON |
| `GET /api/services/{id}/spec` | Proxy OpenAPI spec from the service's pod/container |
| `GET /*` | Serve the Vue SPA |

Service IDs: `{namespace}_{podName}` for Kubernetes, first 12 chars of container ID for Docker.
