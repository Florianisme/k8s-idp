package discovery

import (
	"testing"

	"github.com/docker/docker/api/types"
)

func TestServiceFromContainer_FullLabels(t *testing.T) {
	c := types.Container{
		ID:    "abc1234567890xyz",
		Names: []string{"/my-service"},
		Labels: map[string]string{
			"k8s-idp/enabled":      "true",
			"k8s-idp/name":         "My Service",
			"k8s-idp/description":  "A service",
			"k8s-idp/owner":        "my-team",
			"k8s-idp/source-url":   "https://github.com/org/svc",
			"k8s-idp/openapi-path": "/api/openapi.json",
			"k8s-idp/port":         "3000",
		},
	}
	svc, ok := serviceFromContainer(c, "172.17.0.2")
	if !ok {
		t.Fatal("expected service")
	}
	if svc.ID != "abc123456789" {
		t.Errorf("ID: got %q, want first 12 chars", svc.ID)
	}
	if svc.Name != "My Service" {
		t.Errorf("Name: got %q", svc.Name)
	}
	if svc.SpecPort != "3000" {
		t.Errorf("SpecPort: got %q", svc.SpecPort)
	}
	if svc.IP != "172.17.0.2" {
		t.Errorf("IP: got %q", svc.IP)
	}
	if svc.ContainerName != "my-service" {
		t.Errorf("ContainerName: got %q", svc.ContainerName)
	}
	if !svc.HasSpec {
		t.Error("expected HasSpec=true")
	}
}

func TestServiceFromContainer_DefaultNameAndPort(t *testing.T) {
	c := types.Container{
		ID:    "aabbccddeeff1122",
		Names: []string{"/fallback-name"},
		Labels: map[string]string{
			"k8s-idp/enabled":      "true",
			"k8s-idp/openapi-path": "/spec",
		},
	}
	svc, ok := serviceFromContainer(c, "10.0.0.1")
	if !ok {
		t.Fatal("expected service")
	}
	if svc.Name != "fallback-name" {
		t.Errorf("expected name from container name, got %q", svc.Name)
	}
	if svc.SpecPort != "8080" {
		t.Errorf("expected default port 8080, got %q", svc.SpecPort)
	}
}

func TestServiceFromContainer_NotEnabled(t *testing.T) {
	c := types.Container{
		ID:     "aabbccddeeff",
		Names:  []string{"/svc"},
		Labels: map[string]string{"k8s-idp/enabled": "false"},
	}
	_, ok := serviceFromContainer(c, "10.0.0.1")
	if ok {
		t.Error("expected no service")
	}
}

func TestServiceFromContainer_NoIP(t *testing.T) {
	c := types.Container{
		ID:     "aabbccddeeff",
		Names:  []string{"/svc"},
		Labels: map[string]string{"k8s-idp/enabled": "true"},
	}
	_, ok := serviceFromContainer(c, "")
	if ok {
		t.Error("expected no service when no IP")
	}
}
