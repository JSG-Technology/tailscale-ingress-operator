package main

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	// Potentially add other imports like metav1 if creating dummy objects
)

// TestCreateIngress is a placeholder for future tests.
// It demonstrates basic setup with a fake clientset.
func TestCreateIngress(t *testing.T) {
	// Initialize a fake clientset
	clientset := fake.NewSimpleClientset()

	// Example of how you might use the clientset in a test:
	// _, err := clientset.CoreV1().Pods("default").Get(context.TODO(), "my-pod", metav1.GetOptions{})
	// if err != nil && !errors.IsNotFound(err) {
	//  t.Fatalf("Expected not found error, got %v", err)
	// }
	// if !errors.IsNotFound(err) {
	//  t.Fatalf("Expected not found error, but got pod somehow: %v", err)
	// }

	// For now, we'll just check if the clientset is not nil
	if clientset == nil {
		t.Errorf("fake.NewSimpleClientset() returned nil")
	}

	// Future test logic will go here.
	// For example, creating an Ingress object and verifying it
	// using the fake clientset.
}

func TestCreateIngress_WithAnnotation_CreatesIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "test-service"
	serviceNamespace := "default"
	hostname := "my-custom-hostname"
	ingressName := serviceName + "-ingress"

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{
				"jsgtechnology.com/tailscale-autoingress": hostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	// We need to define createIngress or ensure it's available in the test scope.
	// For now, assuming createIngress is a function in 'main' package:
	// func createIngress(clientset kubernetes.Interface, service *corev1.Service) (*networkingv1.Ingress, error) { ... }
	// Since we don't have the actual createIngress function code yet, this test will fail
	// if createIngress is not defined or if its signature is different.
	// This test is written based on the assumption that createIngress will be available
	// and will attempt to create an Ingress object.

	// Simulate calling the actual createIngress function.
	// As the actual createIngress function is not available in this context,
	// we'll manually create an Ingress object here for the purpose of this test setup.
	// In a real scenario, you would call your `createIngress` function.
	// For now, let's assume createIngress would try to create something like this:
	// pt := networkingv1.PathTypePrefix // PathTypePrefix will be set by createIngress

	// Call the actual createIngress function (which is not yet implemented).
	// This function is expected to create an Ingress resource based on the service.
	_, err := createIngress(clientset, service)
	if err != nil {
		// If createIngress itself returns an error, we should check if that was expected.
		// For this test case, we expect successful creation.
		t.Fatalf("createIngress function returned an error: %v", err)
	}

	// Now, attempt to get the Ingress that should have been created by createIngress
	createdIngress, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress %s/%s: %v", serviceNamespace, ingressName, err)
	}

	// Verify Name
	if createdIngress.Name != ingressName {
		t.Errorf("Expected Ingress name %s, got %s", ingressName, createdIngress.Name)
	}

	// Verify Namespace
	if createdIngress.Namespace != serviceNamespace {
		t.Errorf("Expected Ingress namespace %s, got %s", serviceNamespace, createdIngress.Namespace)
	}

	// Verify IngressClassName (Note: The annotation is often used for older versions,
	// IngressClassName field is the standard way for K8s 1.18+)
	if createdIngress.Spec.IngressClassName == nil || *createdIngress.Spec.IngressClassName != "tailscale" {
		t.Errorf("Expected IngressClassName 'tailscale', got '%v'", createdIngress.Spec.IngressClassName)
	}
	// Also check annotation for compatibility, though IngressClassName is preferred
	if ann, ok := createdIngress.Annotations["kubernetes.io/ingress.class"]; !ok || ann != "tailscale" {
		t.Errorf("Expected Ingress annotation 'kubernetes.io/ingress.class: tailscale', got '%s'", ann)
	}


	// Verify TLS
	if len(createdIngress.Spec.TLS) != 1 {
		t.Fatalf("Expected 1 TLS entry, got %d", len(createdIngress.Spec.TLS))
	}
	if len(createdIngress.Spec.TLS[0].Hosts) != 1 {
		t.Fatalf("Expected 1 host in TLS entry, got %d", len(createdIngress.Spec.TLS[0].Hosts))
	}
	if createdIngress.Spec.TLS[0].Hosts[0] != hostname {
		t.Errorf("Expected TLS host '%s', got '%s'", hostname, createdIngress.Spec.TLS[0].Hosts[0])
	}

	// Verify Rules (basic check)
	if len(createdIngress.Spec.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(createdIngress.Spec.Rules))
	}
	if createdIngress.Spec.Rules[0].Host != hostname {
		t.Errorf("Expected rule host '%s', got '%s'", hostname, createdIngress.Spec.Rules[0].Host)
	}
	if createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name != serviceName {
		t.Errorf("Expected backend service name '%s', got '%s'", serviceName, createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name)
	}
	if createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number != service.Spec.Ports[0].Port {
		t.Errorf("Expected backend service port %d, got %d", service.Spec.Ports[0].Port, createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number)
	}
}

func TestCreateIngress_NoAnnotation_NoIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "no-annotation-service"
	serviceNamespace := "default"
	ingressName := serviceName + "-ingress"

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			// No "jsgtechnology.com/tailscale-autoingress" annotation
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	// Call the actual createIngress function (which is not yet implemented).
	// This function is expected to do nothing if the annotation is missing.
	ing, err := createIngress(clientset, service)

	// We expect createIngress to potentially return (nil, nil) or (nil, someError)
	// if it decides not to create an Ingress. Or it might create nothing and return (nil, nil).
	// The crucial part is verifying that no Ingress resource was actually created in the cluster.

	if err != nil {
		// If createIngress returns an error, and it's a specific "don't create" error, that's fine.
		// We log it for now, but the main check is the Get below.
		t.Logf("createIngress returned an error: %v (this might be expected)", err)
	}
	if ing != nil {
		t.Errorf("createIngress returned an Ingress object (%s) when no annotation was present, but should have returned nil", ing.Name)
	}

	// Attempt to get the Ingress that should NOT have been created.
	_, err = clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("Ingress %s/%s was found, but it should not have been created as the service lacked the required annotation.", serviceNamespace, ingressName)
	}

	// Check that the error is indeed a "NotFound" error.
	if !errors.IsNotFound(err) {
		t.Fatalf("Expected a 'NotFound' error when trying to get Ingress %s/%s, but got: %v", serviceNamespace, ingressName, err)
	}
}

func TestDeleteIngress_IngressExists_DeletesIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "delete-test-service"
	serviceNamespace := "default"
	ingressName := serviceName + "-ingress" // Convention for Ingress naming

	// 1. Create and add a networkingv1.Ingress object to the fake clientset
	existingIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: serviceNamespace,
		},
		Spec: networkingv1.IngressSpec{
			// Minimal valid spec
			Rules: []networkingv1.IngressRule{
				{
					Host: "example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									PathType: func() *networkingv1.PathType {
										pt := networkingv1.PathTypePrefix
										return &pt
									}(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{Number: 80},
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
	_, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Create(context.TODO(), existingIngress, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create initial Ingress for deletion test: %v", err)
	}

	// 2. Create a v1.Service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
		},
		Spec: corev1.ServiceSpec{ // Minimal spec for the service
			Ports: []corev1.ServicePort{{Port: 80}},
		},
	}

	// 3. Call deleteIngress
	// deleteIngress is not yet implemented. This call will cause a compile error
	// until deleteIngress(kubernetes.Interface, *corev1.Service) error is defined.
	err = deleteIngress(clientset, service)
	if err != nil {
		t.Fatalf("deleteIngress function returned an error: %v", err)
	}

	// 4. Verify that the Ingress object is now deleted
	_, err = clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("Ingress %s/%s was found, but it should have been deleted.", serviceNamespace, ingressName)
	}

	if !errors.IsNotFound(err) {
		t.Fatalf("Expected a 'NotFound' error when trying to get deleted Ingress %s/%s, but got: %v", serviceNamespace, ingressName, err)
	}
}

// TestHandleService_UpdateFunc_AnnotationValueChanged_UpdatesIngress simulates an UpdateFunc where annotation value changes.
func TestHandleService_UpdateFunc_AnnotationValueChanged_UpdatesIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "update-value-change-service"
	serviceNamespace := "default"
	oldHostname := "old-hostname.example.com"
	newHostname := "new-hostname.example.com"
	ingressName := serviceName + "-ingress"
	servicePort := int32(80)
	pathTypePrefix := networkingv1.PathTypePrefix

	// 1. Define an oldService with the old annotation value
	oldService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{
				"jsgtechnology.com/tailscale-autoingress": oldHostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: servicePort, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	// 2. Pre-create a corresponding Ingress for oldService with oldHostname
	initialIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{"kubernetes.io/ingress.class": "tailscale"},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: func(s string) *string { return &s }("tailscale"),
			TLS:              []networkingv1.IngressTLS{{Hosts: []string{oldHostname}}},
			Rules: []networkingv1.IngressRule{
				{
					Host: oldHostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{Number: servicePort},
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
	_, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Create(context.TODO(), initialIngress, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create initial Ingress: %v", err)
	}

	// 3. Define a newService based on oldService, but with the annotation value changed
	newService := oldService.DeepCopy()
	newService.Annotations["jsgtechnology.com/tailscale-autoingress"] = newHostname

	// 4. Simulate the UpdateFunc condition: oldHasAnnotation && newHasAnnotation && oldValue != newValue
	// This should trigger createIngress(clientset, newService) internally within handleService.
	// handleService is not yet implemented.
	err = handleService(clientset, newService, "Updated_AnnotationValueChanged") // Event type string is illustrative
	if err != nil {
		t.Fatalf("handleService returned an error: %v", err)
	}

	// 5. Verify that the Ingress for the service is updated/recreated
	updatedIngress, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress %s/%s after handleService call: %v", serviceNamespace, ingressName, err)
	}

	// Verify Ingress properties reflect the newHostname
	if updatedIngress.Spec.IngressClassName == nil || *updatedIngress.Spec.IngressClassName != "tailscale" {
		t.Errorf("Expected IngressClassName 'tailscale', got '%v'", updatedIngress.Spec.IngressClassName)
	}
	if ann, ok := updatedIngress.Annotations["kubernetes.io/ingress.class"]; !ok || ann != "tailscale" {
		t.Errorf("Expected Ingress annotation 'kubernetes.io/ingress.class: tailscale', got '%s'", ann)
	}

	if len(updatedIngress.Spec.TLS) != 1 {
		t.Fatalf("Expected 1 TLS entry, got %d", len(updatedIngress.Spec.TLS))
	}
	if len(updatedIngress.Spec.TLS[0].Hosts) != 1 {
		t.Fatalf("Expected 1 host in TLS entry, got %d", len(updatedIngress.Spec.TLS[0].Hosts))
	}
	if updatedIngress.Spec.TLS[0].Hosts[0] != newHostname {
		t.Errorf("Expected TLS host to be '%s', got '%s'", newHostname, updatedIngress.Spec.TLS[0].Hosts[0])
	}

	if len(updatedIngress.Spec.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(updatedIngress.Spec.Rules))
	}
	rule := updatedIngress.Spec.Rules[0]
	if rule.Host != newHostname {
		t.Errorf("Expected rule host to be '%s', got '%s'", newHostname, rule.Host)
	}
	if rule.HTTP == nil || len(rule.HTTP.Paths) != 1 {
		t.Fatalf("Expected 1 HTTP path in rule, got %v", rule.HTTP)
	}
	path := rule.HTTP.Paths[0]
	if path.PathType == nil || *path.PathType != networkingv1.PathTypePrefix {
		t.Errorf("Expected PathType Prefix, got %v", path.PathType)
	}
	if path.Backend.Service.Name != serviceName {
		t.Errorf("Expected backend service name '%s', got '%s'", serviceName, path.Backend.Service.Name)
	}
	if path.Backend.Service.Port.Number != servicePort {
		t.Errorf("Expected backend service port %d, got %d", servicePort, path.Backend.Service.Port.Number)
	}
}

func TestDeleteIngress_NoIngressExists_NoAction(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "no-op-delete-service"
	serviceNamespace := "default"
	ingressName := serviceName + "-ingress" // Convention for Ingress naming

	// 1. Create a v1.Service object. No Ingress is added to the clientset.
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
		},
		Spec: corev1.ServiceSpec{ // Minimal spec for the service
			Ports: []corev1.ServicePort{{Port: 80}},
		},
	}

	// 2. Call deleteIngress
	// deleteIngress is not yet implemented. This call will cause a compile error
	// until deleteIngress(kubernetes.Interface, *corev1.Service) error is defined.
	// We expect it to be a no-op and not return an error.
	err := deleteIngress(clientset, service)
	if err != nil {
		t.Fatalf("deleteIngress function returned an error '%v', but expected no error when no Ingress exists.", err)
	}

	// 3. Verify that no Ingress was created or deleted (clientset should still not contain this Ingress)
	_, err = clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("Ingress %s/%s was found, but it should not exist as it was never created.", serviceNamespace, ingressName)
	}

	if !errors.IsNotFound(err) {
		t.Fatalf("Expected a 'NotFound' error when trying to get non-existent Ingress %s/%s, but got: %v", serviceNamespace, ingressName, err)
	}

	// Optionally, verify the list of ingresses is empty if no other ingresses are expected.
	ingressList, err := clientset.NetworkingV1().Ingresses(serviceNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list ingresses: %v", err)
	}
	if len(ingressList.Items) != 0 {
		t.Errorf("Expected 0 ingresses in namespace '%s', found %d", serviceNamespace, len(ingressList.Items))
	}
}

// TestHandleService_UpdateFunc_AnnotationAdded_CreatesIngress simulates an UpdateFunc where annotation is added.
func TestHandleService_UpdateFunc_AnnotationAdded_CreatesIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "update-test-service"
	serviceNamespace := "default"
	hostname := "updated-hostname.example.com"
	ingressName := serviceName + "-ingress"
	servicePort := int32(80)

	// 1. Define an oldService without the annotation
	oldService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			// No annotation initially
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: servicePort, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	// 2. Define a newService based on oldService, but with the annotation added
	newService := oldService.DeepCopy() // Start with a copy
	if newService.Annotations == nil {
		newService.Annotations = make(map[string]string)
	}
	newService.Annotations["jsgtechnology.com/tailscale-autoingress"] = hostname

	// 3. Simulate the UpdateFunc condition: !oldHasAnnotation && newHasAnnotation
	// This leads to calling handleService(clientset, newService, "Updated")
	// handleService is not yet implemented.
	err := handleService(clientset, newService, "Updated")
	if err != nil {
		t.Fatalf("handleService returned an error: %v", err)
	}

	// 4. Verify that an Ingress for the newService is created
	createdIngress, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress %s/%s after handleService call: %v", serviceNamespace, ingressName, err)
	}

	// 5. Verify Ingress properties
	if createdIngress.Name != ingressName {
		t.Errorf("Expected Ingress name %s, got %s", ingressName, createdIngress.Name)
	}
	if createdIngress.Namespace != serviceNamespace {
		t.Errorf("Expected Ingress namespace %s, got %s", serviceNamespace, createdIngress.Namespace)
	}

	// Verify IngressClassName and annotation
	if createdIngress.Spec.IngressClassName == nil || *createdIngress.Spec.IngressClassName != "tailscale" {
		t.Errorf("Expected IngressClassName 'tailscale', got '%v'", createdIngress.Spec.IngressClassName)
	}
	if ann, ok := createdIngress.Annotations["kubernetes.io/ingress.class"]; !ok || ann != "tailscale" {
		t.Errorf("Expected Ingress annotation 'kubernetes.io/ingress.class: tailscale', got '%s'", ann)
	}

	// Verify TLS
	if len(createdIngress.Spec.TLS) != 1 {
		t.Fatalf("Expected 1 TLS entry, got %d", len(createdIngress.Spec.TLS))
	}
	if len(createdIngress.Spec.TLS[0].Hosts) != 1 {
		t.Fatalf("Expected 1 host in TLS entry, got %d", len(createdIngress.Spec.TLS[0].Hosts))
	}
	if createdIngress.Spec.TLS[0].Hosts[0] != hostname {
		t.Errorf("Expected TLS host '%s', got '%s'", hostname, createdIngress.Spec.TLS[0].Hosts[0])
	}

	// Verify Rules
	if len(createdIngress.Spec.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(createdIngress.Spec.Rules))
	}
	rule := createdIngress.Spec.Rules[0]
	if rule.Host != hostname {
		t.Errorf("Expected rule host '%s', got '%s'", hostname, rule.Host)
	}
	if rule.HTTP == nil || len(rule.HTTP.Paths) != 1 {
		t.Fatalf("Expected 1 HTTP path in rule, got %v", rule.HTTP)
	}
	path := rule.HTTP.Paths[0]
	if path.PathType == nil || *path.PathType != networkingv1.PathTypePrefix {
		t.Errorf("Expected PathType Prefix, got %v", path.PathType)
	}
	if path.Backend.Service.Name != serviceName {
		t.Errorf("Expected backend service name '%s', got '%s'", serviceName, path.Backend.Service.Name)
	}
	if path.Backend.Service.Port.Number != servicePort {
		t.Errorf("Expected backend service port %d, got %d", servicePort, path.Backend.Service.Port.Number)
	}
}

// TestHandleService_UpdateFunc_AnnotationRemoved_DeletesIngress simulates an UpdateFunc where annotation is removed.
func TestHandleService_UpdateFunc_AnnotationRemoved_DeletesIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "update-remove-annot-service"
	serviceNamespace := "default"
	hostname := "initial-hostname.example.com"
	ingressName := serviceName + "-ingress"
	servicePort := int32(80)
	pathTypePrefix := networkingv1.PathTypePrefix

	// 1. Define an oldService with the annotation
	oldService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{
				"jsgtechnology.com/tailscale-autoingress": hostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: servicePort, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	// 2. Pre-create a corresponding Ingress for oldService
	initialIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{"kubernetes.io/ingress.class": "tailscale"},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: func(s string) *string { return &s }("tailscale"),
			TLS:              []networkingv1.IngressTLS{{Hosts: []string{hostname}}},
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{Number: servicePort},
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
	_, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Create(context.TODO(), initialIngress, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create initial Ingress: %v", err)
	}

	// 3. Define a newService based on oldService, but with the annotation removed
	newService := oldService.DeepCopy()
	delete(newService.Annotations, "jsgtechnology.com/tailscale-autoingress")

	// 4. Simulate the UpdateFunc condition: oldHasAnnotation && !newHasAnnotation
	// This should trigger deleteIngress(clientset, newService) internally within handleService.
	// handleService is not yet implemented.
	err = handleService(clientset, newService, "Updated_AnnotationRemoved") // Event type string is illustrative
	if err != nil {
		// Depending on how handleService/deleteIngress are implemented,
		// an error might be acceptable if it's specifically about "Ingress not found"
		// if deleteIngress was called twice or something. But for a clean removal,
		// no error is expected from handleService itself.
		t.Fatalf("handleService returned an error: %v", err)
	}

	// 5. Verify that the Ingress for the service is deleted
	_, err = clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("Ingress %s/%s was found, but it should have been deleted.", serviceNamespace, ingressName)
	}

	if !errors.IsNotFound(err) {
		t.Fatalf("Expected a 'NotFound' error when trying to get deleted Ingress %s/%s, but got: %v", serviceNamespace, ingressName, err)
	}
}

// TestHandleService_AddFunc_WithAnnotation_CreatesIngress simulates an AddFunc event.
func TestHandleService_AddFunc_WithAnnotation_CreatesIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "add-test-service"
	serviceNamespace := "default"
	hostname := "added-hostname.example.com"
	ingressName := serviceName + "-ingress"
	servicePort := int32(80)

	// 1. Define a v1.Service object with the annotation and a port
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{
				"jsgtechnology.com/tailscale-autoingress": hostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: servicePort, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	// 2. Simulate an "add" event by calling handleService
	// handleService is not yet implemented. This call will cause a compile error
	// until handleService(kubernetes.Interface, *corev1.Service, string) error is defined.
	// The "New" string is a placeholder for how an event type might be passed.
	err := handleService(clientset, service, "New")
	if err != nil {
		t.Fatalf("handleService returned an error: %v", err)
	}

	// 3. Verify that an Ingress for the service is created
	createdIngress, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress %s/%s after handleService call: %v", serviceNamespace, ingressName, err)
	}

	// 4. Verify Ingress properties
	if createdIngress.Name != ingressName {
		t.Errorf("Expected Ingress name %s, got %s", ingressName, createdIngress.Name)
	}
	if createdIngress.Namespace != serviceNamespace {
		t.Errorf("Expected Ingress namespace %s, got %s", serviceNamespace, createdIngress.Namespace)
	}

	// Verify IngressClassName and annotation
	if createdIngress.Spec.IngressClassName == nil || *createdIngress.Spec.IngressClassName != "tailscale" {
		t.Errorf("Expected IngressClassName 'tailscale', got '%v'", createdIngress.Spec.IngressClassName)
	}
	if ann, ok := createdIngress.Annotations["kubernetes.io/ingress.class"]; !ok || ann != "tailscale" {
		// This check assumes createIngress also sets the annotation for compatibility.
		t.Errorf("Expected Ingress annotation 'kubernetes.io/ingress.class: tailscale', got '%s'", ann)
	}

	// Verify TLS
	if len(createdIngress.Spec.TLS) != 1 {
		t.Fatalf("Expected 1 TLS entry, got %d", len(createdIngress.Spec.TLS))
	}
	if len(createdIngress.Spec.TLS[0].Hosts) != 1 {
		t.Fatalf("Expected 1 host in TLS entry, got %d", len(createdIngress.Spec.TLS[0].Hosts))
	}
	if createdIngress.Spec.TLS[0].Hosts[0] != hostname {
		t.Errorf("Expected TLS host '%s', got '%s'", hostname, createdIngress.Spec.TLS[0].Hosts[0])
	}

	// Verify Rules
	if len(createdIngress.Spec.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(createdIngress.Spec.Rules))
	}
	rule := createdIngress.Spec.Rules[0]
	if rule.Host != hostname {
		t.Errorf("Expected rule host '%s', got '%s'", hostname, rule.Host)
	}
	if rule.HTTP == nil || len(rule.HTTP.Paths) != 1 {
		t.Fatalf("Expected 1 HTTP path in rule, got %v", rule.HTTP)
	}
	path := rule.HTTP.Paths[0]
	if path.PathType == nil || *path.PathType != networkingv1.PathTypePrefix {
		t.Errorf("Expected PathType Prefix, got %v", path.PathType)
	}
	if path.Backend.Service.Name != serviceName {
		t.Errorf("Expected backend service name '%s', got '%s'", serviceName, path.Backend.Service.Name)
	}
	if path.Backend.Service.Port.Number != servicePort {
		t.Errorf("Expected backend service port %d, got %d", servicePort, path.Backend.Service.Port.Number)
	}
}

// TestHandleService_AddFunc_WithoutAnnotation_NoIngress simulates an AddFunc event for a service without the annotation.
func TestHandleService_AddFunc_WithoutAnnotation_NoIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "add-no-annotation-service"
	serviceNamespace := "default"
	ingressName := serviceName + "-ingress"
	servicePort := int32(80)

	// 1. Define a v1.Service object WITHOUT the annotation but with a port
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			// Annotation "jsgtechnology.com/tailscale-autoingress" is deliberately omitted
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: servicePort, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	// 2. Simulate an "add" event by calling handleService
	// handleService is not yet implemented.
	err := handleService(clientset, service, "New")
	if err != nil {
		// Depending on handleService's design, it might return an error if it
		// decides not to act, or it might return nil. For this test, we assume
		// it shouldn't error out fatally, but rather just not create an Ingress.
		// If it's designed to return a specific error for "no action needed",
		// that could be checked here. For now, we just log it.
		t.Logf("handleService returned an error: %v (this might be expected if it signals no action)", err)
	}

	// 3. Verify that NO Ingress for the service is created
	_, err = clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("Ingress %s/%s was found, but it should NOT have been created as the service lacked the annotation.", serviceNamespace, ingressName)
	}

	if !errors.IsNotFound(err) {
		t.Fatalf("Expected a 'NotFound' error when trying to get Ingress %s/%s, but got: %v", serviceNamespace, ingressName, err)
	}

	// 4. Optionally, verify the list of ingresses is empty
	ingressList, listErr := clientset.NetworkingV1().Ingresses(serviceNamespace).List(context.TODO(), metav1.ListOptions{})
	if listErr != nil {
		t.Fatalf("Failed to list ingresses: %v", listErr)
	}
	if len(ingressList.Items) != 0 {
		t.Errorf("Expected 0 ingresses in namespace '%s', found %d", serviceNamespace, len(ingressList.Items))
	}
}

func TestCreateIngress_CorrectIngressExists_NoAction(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "test-service"
	serviceNamespace := "default"
	hostname := "correct-hostname"
	ingressName := serviceName + "-ingress"
	initialResourceVersion := "1" // Used to check if the Ingress was modified
	servicePort := int32(80)
	pathTypePrefix := networkingv1.PathTypePrefix

	// 1. Create an initial, correctly configured Ingress object
	existingIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ingressName,
			Namespace:       serviceNamespace,
			ResourceVersion: initialResourceVersion,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "tailscale",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: func(s string) *string { return &s }("tailscale"),
			TLS: []networkingv1.IngressTLS{
				{Hosts: []string{hostname}},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{Number: servicePort},
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
	_, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Create(context.TODO(), existingIngress, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create initial Ingress: %v", err)
	}

	// 2. Create a Service object that matches the Ingress configuration
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{
				"jsgtechnology.com/tailscale-autoingress": hostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: servicePort, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	// 3. Call createIngress
	// This should ideally be a no-op as the Ingress is already correct.
	returnedIngress, err := createIngress(clientset, service)
	if err != nil {
		t.Fatalf("createIngress function returned an error: %v", err)
	}

	// It's possible createIngress returns the existing object if found and correct
	if returnedIngress != nil && returnedIngress.Name != ingressName {
		t.Errorf("createIngress returned an Ingress with name %s, expected %s or nil", returnedIngress.Name, ingressName)
	}


	// 4. Verify the Ingress object was not modified
	retrievedIngress, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress %s/%s: %v", serviceNamespace, ingressName, err)
	}

	if retrievedIngress.ResourceVersion != initialResourceVersion {
		t.Errorf("Expected Ingress ResourceVersion to be '%s' (unchanged), got '%s'", initialResourceVersion, retrievedIngress.ResourceVersion)
	}

	// Sanity checks for other fields
	if retrievedIngress.Spec.IngressClassName == nil || *retrievedIngress.Spec.IngressClassName != "tailscale" {
		t.Errorf("Expected IngressClassName 'tailscale', got '%v'", retrievedIngress.Spec.IngressClassName)
	}
	if len(retrievedIngress.Spec.TLS) != 1 || retrievedIngress.Spec.TLS[0].Hosts[0] != hostname {
		t.Errorf("Expected TLS host '%s', got '%s'", hostname, retrievedIngress.Spec.TLS[0].Hosts[0])
	}
	if len(retrievedIngress.Spec.Rules) != 1 || retrievedIngress.Spec.Rules[0].Host != hostname {
		t.Errorf("Expected rule host '%s', got '%s'", hostname, retrievedIngress.Spec.Rules[0].Host)
	}
}

func TestCreateIngress_HostnameChanged_RecreatesIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "test-service"
	serviceNamespace := "default"
	oldHostname := "old-hostname"
	newHostname := "new-hostname"
	ingressName := serviceName + "-ingress" // Assuming Ingress name is derived from service name
	servicePort := int32(80)
	pathTypePrefix := networkingv1.PathTypePrefix

	// 1. Create an initial Ingress object with oldHostname
	initialIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{ // ensure class annotation for consistency
				"kubernetes.io/ingress.class": "tailscale",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: func(s string) *string { return &s }("tailscale"),
			TLS: []networkingv1.IngressTLS{
				{Hosts: []string{oldHostname}},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: oldHostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{Number: servicePort},
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
	_, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Create(context.TODO(), initialIngress, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create initial Ingress: %v", err)
	}

	// 2. Create a Service object with the newHostname annotation
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{
				"jsgtechnology.com/tailscale-autoingress": newHostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: servicePort, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	// 3. Call createIngress
	// Assuming createIngress will find the existing Ingress by name and update it,
	// or delete and recreate it.
	_, err = createIngress(clientset, service)
	if err != nil {
		t.Fatalf("createIngress function returned an error: %v", err)
	}

	// 4. Verify the Ingress object
	updatedIngress, err := clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get Ingress %s/%s after update: %v", serviceNamespace, ingressName, err)
	}

	// Verify IngressClassName
	if updatedIngress.Spec.IngressClassName == nil || *updatedIngress.Spec.IngressClassName != "tailscale" {
		t.Errorf("Expected IngressClassName 'tailscale', got '%v'", updatedIngress.Spec.IngressClassName)
	}
	if ann, ok := updatedIngress.Annotations["kubernetes.io/ingress.class"]; !ok || ann != "tailscale" {
		t.Errorf("Expected Ingress annotation 'kubernetes.io/ingress.class: tailscale', got '%s'", ann)
	}


	// Verify TLS hostname updated
	if len(updatedIngress.Spec.TLS) != 1 {
		t.Fatalf("Expected 1 TLS entry, got %d", len(updatedIngress.Spec.TLS))
	}
	if len(updatedIngress.Spec.TLS[0].Hosts) != 1 {
		t.Fatalf("Expected 1 host in TLS entry, got %d", len(updatedIngress.Spec.TLS[0].Hosts))
	}
	if updatedIngress.Spec.TLS[0].Hosts[0] != newHostname {
		t.Errorf("Expected TLS host to be updated to '%s', got '%s'", newHostname, updatedIngress.Spec.TLS[0].Hosts[0])
	}

	// Verify Rules updated
	if len(updatedIngress.Spec.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(updatedIngress.Spec.Rules))
	}
	if updatedIngress.Spec.Rules[0].Host != newHostname {
		t.Errorf("Expected rule host to be updated to '%s', got '%s'", newHostname, updatedIngress.Spec.Rules[0].Host)
	}
	if updatedIngress.Spec.Rules[0].HTTP == nil || len(updatedIngress.Spec.Rules[0].HTTP.Paths) != 1 {
		t.Fatalf("Expected 1 HTTP path in rule, got %v", updatedIngress.Spec.Rules[0].HTTP)
	}
	path := updatedIngress.Spec.Rules[0].HTTP.Paths[0]
	if path.Backend.Service.Name != serviceName {
		t.Errorf("Expected backend service name '%s', got '%s'", serviceName, path.Backend.Service.Name)
	}
	if path.Backend.Service.Port.Number != servicePort {
		t.Errorf("Expected backend service port %d, got %d", servicePort, path.Backend.Service.Port.Number)
	}
}

func TestCreateIngress_NoPorts_NoIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	serviceName := "no-ports-service"
	serviceNamespace := "default"
	hostname := "my-custom-hostname"
	ingressName := serviceName + "-ingress"

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Annotations: map[string]string{
				"jsgtechnology.com/tailscale-autoingress": hostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{}, // Empty list of ports
		},
	}

	// Call the actual createIngress function (which is not yet implemented).
	// This function is expected to do nothing if the service has no ports.
	ing, err := createIngress(clientset, service)

	if err != nil {
		// Log the error, as it might be an explicit error from createIngress
		// indicating why it didn't create the Ingress (e.g., "service has no ports").
		t.Logf("createIngress returned an error: %v (this might be expected)", err)
	}
	if ing != nil {
		t.Errorf("createIngress returned an Ingress object (%s) when service had no ports, but should have returned nil", ing.Name)
	}

	// Attempt to get the Ingress that should NOT have been created.
	_, err = clientset.NetworkingV1().Ingresses(serviceNamespace).Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("Ingress %s/%s was found, but it should not have been created as the service has no ports.", serviceNamespace, ingressName)
	}

	// Check that the error is indeed a "NotFound" error.
	if !errors.IsNotFound(err) {
		t.Fatalf("Expected a 'NotFound' error when trying to get Ingress %s/%s, but got: %v", serviceNamespace, ingressName, err)
	}
}
