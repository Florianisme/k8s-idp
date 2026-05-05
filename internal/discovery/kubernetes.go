package discovery

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"k8s-idp/internal/registry"
)

// KubernetesWatcher watches Pods with k8s-idp/enabled=true and syncs to the registry.
type KubernetesWatcher struct {
	client   kubernetes.Interface
	registry *registry.Registry
}

// NewKubernetesWatcher builds the watcher. Tries in-cluster config first, then KUBECONFIG.
func NewKubernetesWatcher(reg *registry.Registry) (*KubernetesWatcher, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfg, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("building kubernetes config: %w", err)
		}
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}
	return &KubernetesWatcher{client: client, registry: reg}, nil
}

// Start runs the informer until ctx is cancelled.
func (w *KubernetesWatcher) Start(ctx context.Context) {
	factory := informers.NewSharedInformerFactoryWithOptions(
		w.client, 0,
		informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
			opts.LabelSelector = "k8s-idp/enabled=true"
		}),
	)
	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if pod, ok := obj.(*corev1.Pod); ok {
				if svc, ok := serviceFromPod(pod); ok {
					w.registry.Add(svc)
				}
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			pod, ok := newObj.(*corev1.Pod)
			if !ok {
				return
			}
			if svc, ok := serviceFromPod(pod); ok {
				w.registry.Add(svc)
			} else {
				w.registry.Remove(podID(pod))
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					return
				}
				pod, ok = tombstone.Obj.(*corev1.Pod)
				if !ok {
					return
				}
			}
			w.registry.Remove(podID(pod))
		},
	})
	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced) {
		log.Println("warning: kubernetes informer cache sync timed out")
	}
	<-ctx.Done()
}

// podID returns the unique identifier for a pod.
func podID(pod *corev1.Pod) string {
	return pod.Namespace + "_" + pod.Name
}

// serviceFromPod converts a Pod into a Service. Returns false if the pod should be skipped.
func serviceFromPod(pod *corev1.Pod) (registry.Service, bool) {
	labels := pod.Labels
	if labels["k8s-idp/enabled"] != "true" {
		return registry.Service{}, false
	}
	if pod.Status.PodIP == "" {
		return registry.Service{}, false
	}
	name := labels["k8s-idp/name"]
	if name == "" {
		name = pod.Name
	}
	port := labels["k8s-idp/port"]
	if port == "" {
		port = "8080"
	}
	specPath := labels["k8s-idp/openapi-path"]
	return registry.Service{
		ID:          podID(pod),
		Name:        name,
		Description: labels["k8s-idp/description"],
		Owner:       labels["k8s-idp/owner"],
		SourceURL:   labels["k8s-idp/source-url"],
		HasSpec:     specPath != "",
		SpecPath:    specPath,
		SpecPort:    port,
		IP:          pod.Status.PodIP,
		Namespace:   pod.Namespace,
		PodName:     pod.Name,
	}, true
}
