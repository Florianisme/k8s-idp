# Developer Portal Design Spec

## Stack
- Frontend: Vue 3 + Vuetify 3 (Vite)
- Backend: Go 1.22
- Container: Multi-stage Dockerfile → alpine (~25MB)
- OpenAPI rendering: Swagger UI

## Labels (k8s-idp/ prefix, on Pods or containers)
| Label | Required | Default | Description |
|---|---|---|---|
| k8s-idp/enabled | yes | — | Must be "true" |
| k8s-idp/name | no | pod/container name | Display name |
| k8s-idp/description | no | — | Short description |
| k8s-idp/owner | no | — | Team or person |
| k8s-idp/source-url | no | — | Source code URL |
| k8s-idp/openapi-path | no | — | Path to OpenAPI spec |
| k8s-idp/port | no | 8080 | Port to fetch spec from |

## Service ID format
- Kubernetes: `{namespace}_{podName}` (underscore separator, URL-safe)
- Docker: first 12 chars of container ID

## API
- GET /api/services — list all services
- GET /api/services/{id}/spec — proxy OpenAPI spec
- GET /* — serve embedded Vue SPA

## Environment Variables
| Var | Default | Description |
|---|---|---|
| MODE | kubernetes | kubernetes or docker |
| PORT | 8080 | HTTP listen port |
| KUBECONFIG | auto | Kubeconfig path |
| DOCKER_HOST | unix:///var/run/docker.sock | Docker socket |
