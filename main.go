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

// Annotation key for services that should have auto-generated ingresses
const AutoIngressAnnotation = "jsgtechnology.com/tailscale-autoingress"

func createIngress(clientset *kubernetes.Clientset, service *v1.Service) {
	// Check if service has the required annotation
	if _, ok := service.Annotations[AutoIngressAnnotation]; !ok {
		// Skip services without the annotation
		return
	}

	// Skip if service has no ports
	if len(service.Spec.Ports) == 0 {
		log.Printf("Skipping service %s: no ports defined", service.Name)
		return
	}

	ingressName := fmt.Sprintf("%s-ingress", service.Name)

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

	// Define the ingress class name
	ingressClassName := "tailscale"

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: service.Namespace,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			DefaultBackend: &networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: service.Name,
					Port: networkingv1.ServiceBackendPort{
						Number: service.Spec.Ports[0].Port,
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: []string{service.Name},
				},
			},
		},
	}

	_, err = clientset.NetworkingV1().Ingresses(service.Namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create Ingress for service %s: %v", service.Name, err)
	} else {
		log.Printf("Ingress created: %s -> %s", service.Name, service.Name)
	}
}

func handleService(clientset *kubernetes.Clientset, service *v1.Service, action string) {
	if _, ok := service.Annotations[AutoIngressAnnotation]; ok {
		log.Printf("%s service with %s annotation: %s", action, AutoIngressAnnotation, service.Name)
		createIngress(clientset, service)
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

	// Watch for new and updated services
	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service, ok := obj.(*v1.Service)
			if !ok {
				log.Println("Could not parse Service object")
				return
			}
			handleService(clientset, service, "New")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldService, ok := oldObj.(*v1.Service)
			if !ok {
				log.Println("Could not parse old Service object")
				return
			}

			newService, ok := newObj.(*v1.Service)
			if !ok {
				log.Println("Could not parse new Service object")
				return
			}

			// Check if the annotation was added in this update
			_, oldHasAnnotation := oldService.Annotations[AutoIngressAnnotation]
			_, newHasAnnotation := newService.Annotations[AutoIngressAnnotation]

			if !oldHasAnnotation && newHasAnnotation {
				handleService(clientset, newService, "Updated")
			}
		},
	})

	log.Println("Starting Tailscale Ingress Operator...")
	log.Printf("Watching for services with '%s' annotation", AutoIngressAnnotation)
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	<-stopCh
}
