package discovery

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceFromPod_FullLabels(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "payment-api-abc", Namespace: "production",
			Labels: map[string]string{
				"k8s-idp/enabled":      "true",
				"k8s-idp/name":         "Payment API",
				"k8s-idp/description":  "Payment processing",
				"k8s-idp/owner":        "payments-team",
				"k8s-idp/source-url":   "https://github.com/org/payment",
				"k8s-idp/openapi-path": "/openapi.json",
				"k8s-idp/port":         "9090",
			},
		},
		Status: corev1.PodStatus{PodIP: "10.0.0.1"},
	}

	svc, ok := serviceFromPod(pod)
	if !ok {
		t.Fatal("expected service to be created")
	}
	if svc.ID != "production_payment-api-abc" {
		t.Errorf("ID: got %q", svc.ID)
	}
	if svc.Name != "Payment API" {
		t.Errorf("Name: got %q", svc.Name)
	}
	if svc.SpecPort != "9090" {
		t.Errorf("SpecPort: got %q", svc.SpecPort)
	}
	if svc.IP != "10.0.0.1" {
		t.Errorf("IP: got %q", svc.IP)
	}
	if !svc.HasSpec {
		t.Error("expected HasSpec=true")
	}
	if svc.Namespace != "production" {
		t.Errorf("Namespace: got %q", svc.Namespace)
	}
}

func TestServiceFromPod_DefaultNameAndPort(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod", Namespace: "default",
			Labels: map[string]string{
				"k8s-idp/enabled":      "true",
				"k8s-idp/openapi-path": "/spec",
			},
		},
		Status: corev1.PodStatus{PodIP: "10.0.0.2"},
	}
	svc, ok := serviceFromPod(pod)
	if !ok {
		t.Fatal("expected service")
	}
	if svc.Name != "my-pod" {
		t.Errorf("expected name to fall back to pod name, got %q", svc.Name)
	}
	if svc.SpecPort != "8080" {
		t.Errorf("expected default port 8080, got %q", svc.SpecPort)
	}
}

func TestServiceFromPod_NotEnabled(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "svc", Namespace: "ns",
			Labels: map[string]string{"k8s-idp/enabled": "false"},
		},
		Status: corev1.PodStatus{PodIP: "10.0.0.1"},
	}
	_, ok := serviceFromPod(pod)
	if ok {
		t.Error("expected service NOT to be created")
	}
}

func TestServiceFromPod_NoPodIP(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "svc", Namespace: "ns",
			Labels: map[string]string{"k8s-idp/enabled": "true"},
		},
		Status: corev1.PodStatus{PodIP: ""},
	}
	_, ok := serviceFromPod(pod)
	if ok {
		t.Error("expected no service when pod has no IP")
	}
}
