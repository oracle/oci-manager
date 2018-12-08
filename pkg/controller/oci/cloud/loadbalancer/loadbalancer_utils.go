/*
Copyright 2018 Oracle and/or its affiliates. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package loadbalancer

import (
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	// corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"strconv"
)

// CreateOrUpdateLoadBalancer reconciles the loadbalancer resource
func CreateOrUpdateLoadBalancer(c clientset.Interface,
	controllerRef *metav1.OwnerReference,
	lb *cloudv1alpha1.LoadBalancer,
	subnets []string) (*v1alpha1.LoadBalancer, bool, error) {

	lbResource := &v1alpha1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name: lb.Name,
			Labels: map[string]string{
				cloudv1alpha1.LoadBalancerKind: lb.Name,
			},
		},
		Spec: v1alpha1.LoadBalancerSpec{
			CompartmentRef: lb.Namespace,
			IsPrivate:      lb.Spec.IsPrivate,
			Shape:          lb.Spec.BandwidthMbps,
			SubnetRefs:     subnets,
		},
	}

	if controllerRef != nil {
		lbResource.OwnerReferences = append(lbResource.OwnerReferences, *controllerRef)
	}

	current, err := c.OcilbV1alpha1().LoadBalancers(lb.Namespace).Get(lbResource.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(lbResource.Spec, current.Spec) && reflect.DeepEqual(lbResource.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.LoadBalancer)
		new.Spec = lbResource.Spec
		new.Labels = lbResource.Labels
		r, e := c.OcilbV1alpha1().LoadBalancers(lb.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		r, e := c.OcilbV1alpha1().LoadBalancers(lb.Namespace).Create(lbResource)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateIBackendSet reconciles the backendset resource
func CreateOrUpdateBackendSet(c clientset.Interface,
	controllerRef *metav1.OwnerReference,
	lb *cloudv1alpha1.LoadBalancer) (*v1alpha1.BackendSet, bool, error) {

	interval := 5000
	if lb.Spec.HealthCheck.IntervalInMillis != 0 {
		interval = lb.Spec.HealthCheck.IntervalInMillis
	}
	timeout := 3000
	if lb.Spec.HealthCheck.TimeoutInMillis != 0 {
		timeout = lb.Spec.HealthCheck.TimeoutInMillis
	}
	retries := 3
	if lb.Spec.HealthCheck.Retries != 0 {
		retries = lb.Spec.HealthCheck.Retries
	}
	returnCode := 200
	if lb.Spec.HealthCheck.ReturnCode != 0 {
		returnCode = lb.Spec.HealthCheck.ReturnCode
	}

	var protocol string
	if lb.Spec.HealthCheck.Protocol != "" {
		protocol = lb.Spec.HealthCheck.Protocol
	} else {
		for _, l := range lb.Spec.Listeners {
			protocol = l.Protocol
		}
	}
	if protocol == "" {
		protocol = "http"
	}

	healthChecker := &v1alpha1.HealthChecker{
		URLPath:          lb.Spec.HealthCheck.URLPath,
		Port:             lb.Spec.HealthCheck.Port,
		IntervalInMillis: interval,
		ReturnCode:       returnCode,
		TimeoutInMillis:  timeout,
		Retries:          retries,
		Protocol:         protocol,
	}

	sessionPersistenceConfig := &v1alpha1.SessionPersistenceConfiguration{}
	if lb.Spec.SessionPersistenceCookie != "" {
		sessionPersistenceConfig.CookieName = lb.Spec.SessionPersistenceCookie
	} else {
		sessionPersistenceConfig = nil
	}

	bs := &v1alpha1.BackendSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: lb.Name,
			Labels: map[string]string{
				cloudv1alpha1.LoadBalancerKind: lb.Name,
			},
		},
		Spec: v1alpha1.BackendSetSpec{
			LoadBalancerRef:          lb.Name,
			HealthChecker:            healthChecker,
			Policy:                   lb.Spec.BalanceMode,
			SessionPersistenceConfig: sessionPersistenceConfig,
		},
	}

	if controllerRef != nil {
		bs.OwnerReferences = append(bs.OwnerReferences, *controllerRef)
	}

	current, err := c.OcilbV1alpha1().BackendSets(lb.Namespace).Get(bs.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(bs.Spec, current.Spec) && reflect.DeepEqual(bs.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.BackendSet)
		new.Spec = bs.Spec
		new.Labels = bs.Labels
		r, e := c.OcilbV1alpha1().BackendSets(lb.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		r, e := c.OcilbV1alpha1().BackendSets(lb.Namespace).Create(bs)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateBackend reconciles the backend resource
func CreateOrUpdateBackend(c clientset.Interface,
	controllerRef *metav1.OwnerReference,
	lb *cloudv1alpha1.LoadBalancer,
	instanceName string,
	weight int) (*v1alpha1.Backend, bool, error) {

	backend := &v1alpha1.Backend{
		ObjectMeta: metav1.ObjectMeta{
			Name: lb.Name + "-" + instanceName,
			Labels: map[string]string{
				cloudv1alpha1.LoadBalancerKind: lb.Name,
			},
		},
		Spec: v1alpha1.BackendSpec{
			BackendSetRef:   lb.Name,
			InstanceRef:     instanceName,
			LoadBalancerRef: lb.Name,
			Port:            lb.Spec.BackendPort,
			Weight:          weight,
		},
	}

	if controllerRef != nil {
		backend.OwnerReferences = append(backend.OwnerReferences, *controllerRef)
	}

	current, err := c.OcilbV1alpha1().Backends(lb.Namespace).Get(backend.Name, metav1.GetOptions{})
	// this is populated by resource controller
	backend.Spec.IPAddress = current.Spec.IPAddress

	if err == nil {
		if reflect.DeepEqual(backend.Spec, current.Spec) && reflect.DeepEqual(backend.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.Backend)
		new.Spec = backend.Spec
		new.Labels = backend.Labels
		r, e := c.OcilbV1alpha1().Backends(lb.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		r, e := c.OcilbV1alpha1().Backends(lb.Namespace).Create(backend)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateCertificate reconciles the certificate resource
func CreateOrUpdateCertificate(c clientset.Interface,
	controllerRef *metav1.OwnerReference,
	lb *cloudv1alpha1.LoadBalancer,
	listener *cloudv1alpha1.Listener) (*v1alpha1.Certificate, bool, error) {

	cert := &v1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name: lb.Name + strconv.Itoa(listener.Port),
			Labels: map[string]string{
				cloudv1alpha1.LoadBalancerKind: lb.Name,
			},
		},
		Spec: v1alpha1.CertificateSpec{
			LoadBalancerRef:   lb.Name,
			PublicCertificate: listener.SSLCertificate.Certificate,
			Passphrase:        listener.SSLCertificate.Passphrase,
			PrivateKey:        listener.SSLCertificate.PrivateKey,
			CACertificate:     listener.SSLCertificate.CACertificate,
		},
	}

	if controllerRef != nil {
		cert.OwnerReferences = append(cert.OwnerReferences, *controllerRef)
	}

	current, err := c.OcilbV1alpha1().Certificates(lb.Namespace).Get(cert.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(cert.Spec, current.Spec) && reflect.DeepEqual(cert.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.Certificate)
		new.Spec = cert.Spec
		new.Labels = cert.Labels
		r, e := c.OcilbV1alpha1().Certificates(lb.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		r, e := c.OcilbV1alpha1().Certificates(lb.Namespace).Create(cert)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateListener reconciles the listener resource
func CreateOrUpdateListener(c clientset.Interface,
	controllerRef *metav1.OwnerReference,
	lb *cloudv1alpha1.LoadBalancer,
	listener *cloudv1alpha1.Listener) (*v1alpha1.Listener, bool, error) {

	idleTimeoutSec := 60
	if listener.IdleTimeoutSec != 0 {
		idleTimeoutSec = listener.IdleTimeoutSec
	}

	protocol := "HTTP"
	if listener.Protocol != "" {
		protocol = listener.Protocol
	}

	listenerName := lb.Name + strconv.Itoa(listener.Port)

	listenerResource := &v1alpha1.Listener{
		ObjectMeta: metav1.ObjectMeta{
			Name: listenerName,
			Labels: map[string]string{
				cloudv1alpha1.LoadBalancerKind: lb.Name,
			},
		},
		Spec: v1alpha1.ListenerSpec{
			LoadBalancerRef:       lb.Name,
			Port:                  listener.Port,
			Protocol:              protocol,
			DefaultBackendSetName: lb.Name,
			IdleTimeout:           int64(idleTimeoutSec),
		},
	}

	if listener.SSLCertificate.Certificate != "" {
		listenerResource.Spec.CertificateRef = listenerName
	}

	if controllerRef != nil {
		listenerResource.OwnerReferences = append(listenerResource.OwnerReferences, *controllerRef)
	}

	current, err := c.OcilbV1alpha1().Listeners(lb.Namespace).Get(listenerResource.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(listenerResource.Spec, current.Spec) && reflect.DeepEqual(listenerResource.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.Listener)
		new.Spec = listenerResource.Spec
		new.Labels = listenerResource.Labels
		r, e := c.OcilbV1alpha1().Listeners(lb.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		r, e := c.OcilbV1alpha1().Listeners(lb.Namespace).Create(listenerResource)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// DeleteLoadBalancer deletes the loadBalancer resource
func DeleteLoadBalancer(c clientset.Interface, lb *cloudv1alpha1.LoadBalancer) (*v1alpha1.LoadBalancer, error) {
	current, err := c.OcilbV1alpha1().LoadBalancers(lb.Namespace).Get(lb.Name, metav1.GetOptions{})
	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OcilbV1alpha1().LoadBalancers(lb.Namespace).Delete(lb.Name, &metav1.DeleteOptions{}); e != nil {
				return current, e
			}
		}
		return current, nil
	} else if apierrors.IsNotFound(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

// DeleteListener deletes the listener resource
func DeleteListener(c clientset.Interface, namespace string, name string) (*v1alpha1.Listener, error) {
	current, err := c.OcilbV1alpha1().Listeners(namespace).Get(name, metav1.GetOptions{})
	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OcilbV1alpha1().Listeners(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
				return current, e
			}
		}
		return current, nil
	} else if apierrors.IsNotFound(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

// DeleteBackend deletes the backend resource
func DeleteBackend(c clientset.Interface, namespace string, name string) (*v1alpha1.Backend, error) {
	current, err := c.OcilbV1alpha1().Backends(namespace).Get(name, metav1.GetOptions{})
	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OcilbV1alpha1().Backends(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
				return current, e
			}
		}
		return current, nil
	} else if apierrors.IsNotFound(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

// DeleteBackendSet deletes the backendset resource
func DeleteBackendSet(c clientset.Interface, lb *cloudv1alpha1.LoadBalancer) (*v1alpha1.BackendSet, error) {
	current, err := c.OcilbV1alpha1().BackendSets(lb.Namespace).Get(lb.Name, metav1.GetOptions{})
	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OcilbV1alpha1().BackendSets(lb.Namespace).Delete(lb.Name, &metav1.DeleteOptions{}); e != nil {
				return current, e
			}
		}
		return current, nil
	} else if apierrors.IsNotFound(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

// DeleteCertificate deletes the certificate resource
func DeleteCertificate(c clientset.Interface, lb *cloudv1alpha1.LoadBalancer) (*v1alpha1.Certificate, error) {
	current, err := c.OcilbV1alpha1().Certificates(lb.Namespace).Get(lb.Name, metav1.GetOptions{})
	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OcilbV1alpha1().Certificates(lb.Namespace).Delete(lb.Name, &metav1.DeleteOptions{}); e != nil {
				return current, e
			}
		}
		return current, nil
	} else if apierrors.IsNotFound(err) {
		return nil, nil
	} else {
		return nil, err
	}
}
