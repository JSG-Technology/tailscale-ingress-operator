package main

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func createIngress(clientset *kubernetes.Clientset, service *v1.Service) {
	// Skip if service has no ports
	if len(service.Spec.Ports) == 0 {
		log.Printf("Skipping service %s: no ports defined", service.Name)
		return
	}

	ingressName := fmt.Sprintf("ingress-%s", service.Name)
	hostName := fmt.Sprintf("%s.yourdomain.com", service.Name)

	// Check if ingress already exists
	_, err := clientset.NetworkingV1().Ingresses(service.Namespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		// Ingress already exists, skip creation
		log.Printf("Ingress %s already exists, skipping creation", ingressName)
		return
	} else if !errors.IsNotFound(err) {
		// Unexpected error
		log.Printf("Error checking for existing ingress %s: %v", ingressName, err)
		return
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: service.Namespace,
			Annotations: map[string]string{
				"tailscale.com/enable":   "true",
				"tailscale.com/hostname": hostName,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: hostName,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *networkingv1.PathType { p := networkingv1.PathTypePrefix; return &p }(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: service.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: service.Spec.Ports[0].Port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = clientset.NetworkingV1().Ingresses(service.Namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create Ingress for service %s: %v", service.Name, err)
	} else {
		log.Printf("Ingress created: %s -> %s", service.Name, hostName)
	}
}

func main() {
	// Load in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to load cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Watch all namespaces
	factory := informers.NewSharedInformerFactory(clientset, time.Minute*10)
	serviceInformer := factory.Core().V1().Services().Informer()

	stopCh := make(chan struct{})
	defer runtime.HandleCrash()

	// Watch for new services
	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service, ok := obj.(*v1.Service)
			if !ok {
				log.Println("Could not parse Service object")
				return
			}
			log.Printf("New service detected: %s", service.Name)
			createIngress(clientset, service)
		},
	})

	log.Println("Starting Kubernetes Ingress Operator...")
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	<-stopCh
}
