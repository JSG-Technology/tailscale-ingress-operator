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
	annotationValue, ok := service.Annotations[AutoIngressAnnotation]
	if !ok {
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
	existingIngress, err := clientset.NetworkingV1().Ingresses(service.Namespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		// Ingress already exists, check if hostname has changed
		if len(existingIngress.Spec.TLS) > 0 && len(existingIngress.Spec.TLS[0].Hosts) > 0 {
			currentHostname := existingIngress.Spec.TLS[0].Hosts[0]
			newHostname := service.Name
			if annotationValue != "true" {
				newHostname = annotationValue
			}

			if currentHostname != newHostname {
				log.Printf("Hostname changed from %s to %s, recreating ingress", currentHostname, newHostname)
				deleteIngress(clientset, service)
			} else {
				// Hostname is the same, skip creation
				log.Printf("Ingress %s already exists with correct hostname, skipping creation", ingressName)
				return
			}
		} else {
			// Ingress exists but doesn't have hostname info, skip creation
			log.Printf("Ingress %s already exists, skipping creation", ingressName)
			return
		}
	} else if !errors.IsNotFound(err) {
		// Unexpected error
		log.Printf("Error checking for existing ingress %s: %v", ingressName, err)
		return
	}

	// Define the ingress class name
	ingressClassName := "tailscale"

	hostname := service.Name
	if annotationValue != "true" {
		hostname = annotationValue
	}

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
					Hosts: []string{hostname},
				},
			},
		},
	}

	_, err = clientset.NetworkingV1().Ingresses(service.Namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create Ingress for service %s: %v", service.Name, err)
	} else {
		log.Printf("Ingress created: %s -> %s", service.Name, hostname)
	}
}

func deleteIngress(clientset *kubernetes.Clientset, service *v1.Service) {
	ingressName := fmt.Sprintf("%s-ingress", service.Name)

	// Check if ingress exists before trying to delete it
	_, err := clientset.NetworkingV1().Ingresses(service.Namespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// Ingress doesn't exist, nothing to do
		return
	} else if err != nil {
		// Unexpected error
		log.Printf("Error checking for existing ingress %s: %v", ingressName, err)
		return
	}

	// Delete the ingress
	err = clientset.NetworkingV1().Ingresses(service.Namespace).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Failed to delete Ingress %s: %v", ingressName, err)
	} else {
		log.Printf("Ingress deleted: %s", ingressName)
	}
}

func handleService(clientset *kubernetes.Clientset, service *v1.Service, action string) {
	if annotationValue, ok := service.Annotations[AutoIngressAnnotation]; ok {
		log.Printf("%s service with %s annotation: %s (value: %s)", action, AutoIngressAnnotation, service.Name, annotationValue)
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

			// Check if the annotation was added, removed, or changed
			oldValue, oldHasAnnotation := oldService.Annotations[AutoIngressAnnotation]
			newValue, newHasAnnotation := newService.Annotations[AutoIngressAnnotation]

			if !oldHasAnnotation && newHasAnnotation {
				// Annotation was added
				handleService(clientset, newService, "Updated")
			} else if oldHasAnnotation && !newHasAnnotation {
				// Annotation was removed
				log.Printf("Annotation %s removed from service: %s", AutoIngressAnnotation, newService.Name)
				deleteIngress(clientset, newService)
			} else if oldHasAnnotation && newHasAnnotation && oldValue != newValue {
				// Annotation value changed
				log.Printf("Annotation %s value changed from %s to %s for service: %s", 
					AutoIngressAnnotation, oldValue, newValue, newService.Name)
				createIngress(clientset, newService)
			}
		},
		DeleteFunc: func(obj interface{}) {
			service, ok := obj.(*v1.Service)
			if !ok {
				// When a delete is observed, the object might be a DeletedFinalStateUnknown tombstone
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					log.Println("Could not parse deleted Service object")
					return
				}
				service, ok = tombstone.Obj.(*v1.Service)
				if !ok {
					log.Println("Could not parse Service from tombstone")
					return
				}
			}

			// Check if the deleted service had our annotation
			if _, ok := service.Annotations[AutoIngressAnnotation]; ok {
				log.Printf("Service with %s annotation deleted: %s", AutoIngressAnnotation, service.Name)
				deleteIngress(clientset, service)
			}
		},
	})

	log.Println("Starting Tailscale Ingress Operator...")
	log.Printf("Watching for services with '%s' annotation", AutoIngressAnnotation)
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	<-stopCh
}
