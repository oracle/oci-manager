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
package network

import (
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"

	"github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
)

// CreateOrUpdateVcn reconciles the virtual network resource
func CreateOrUpdateVcn(c clientset.Interface, network *cloudv1alpha1.Network, controllerRef *metav1.OwnerReference) (*v1alpha1.Vcn, bool, error) {

	vcn := &v1alpha1.Vcn{
		ObjectMeta: metav1.ObjectMeta{
			Name: network.Name,
			Labels: map[string]string{
				"network": network.Name,
			},
		},
		Spec: v1alpha1.VcnSpec{
			CompartmentRef: network.Namespace,
			CidrBlock:      network.Spec.CidrBlock,
			DNSLabel:       strings.Replace(network.Name, "-", "", -1),
		},
	}

	if controllerRef != nil {
		vcn.OwnerReferences = append(vcn.OwnerReferences, *controllerRef)
	}

	current, err := c.OcicoreV1alpha1().Vcns(network.Namespace).Get(vcn.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(vcn.Spec, current.Spec) && reflect.DeepEqual(vcn.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.Vcn)
		new.Spec = vcn.Spec
		new.Labels = vcn.Labels
		r, e := c.OcicoreV1alpha1().Vcns(network.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG virtual network create\n")
		r, e := c.OcicoreV1alpha1().Vcns(network.Namespace).Create(vcn)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateInternetGateway reconciles the internet gateway resource
func CreateOrUpdateInternetGateway(c clientset.Interface, network *cloudv1alpha1.Network, controllerRef *metav1.OwnerReference) (*v1alpha1.InternetGateway, bool, error) {

	internetgateway := &v1alpha1.InternetGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: network.Name,
			Labels: map[string]string{
				"network": network.Name,
			},
		},
		Spec: v1alpha1.InternetGatewaySpec{
			CompartmentRef: network.Namespace,
			VcnRef:         network.Name,
			IsEnabled:      true,
		},
	}

	if controllerRef != nil {
		internetgateway.OwnerReferences = append(internetgateway.OwnerReferences, *controllerRef)
	}

	current, err := c.OcicoreV1alpha1().InternetGatewaies(network.Namespace).Get(internetgateway.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(internetgateway.Spec, current.Spec) && reflect.DeepEqual(internetgateway.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.InternetGateway)
		new.Spec = internetgateway.Spec
		new.Labels = internetgateway.Labels
		r, e := c.OcicoreV1alpha1().InternetGatewaies(network.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG internet gateway create\n")
		r, e := c.OcicoreV1alpha1().InternetGatewaies(network.Namespace).Create(internetgateway)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateRouteTable reconciles the route table resource
func CreateOrUpdateRouteTable(c clientset.Interface, network *cloudv1alpha1.Network, controllerRef *metav1.OwnerReference) (*v1alpha1.RouteTable, bool, error) {

	routetable := &v1alpha1.RouteTable{
		ObjectMeta: metav1.ObjectMeta{
			Name: network.Name,
			Labels: map[string]string{
				"network": network.Name,
			},
		},
		Spec: v1alpha1.RouteTableSpec{
			CompartmentRef: network.Namespace,
			VcnRef:         network.Name,
			RouteRules: []v1alpha1.RouteRule{
				v1alpha1.RouteRule{
					CidrBlock:       "0.0.0.0/0",
					NetworkEntityID: network.Name,
				},
			},
		},
	}

	if controllerRef != nil {
		routetable.OwnerReferences = append(routetable.OwnerReferences, *controllerRef)
	}

	current, err := c.OcicoreV1alpha1().RouteTables(network.Namespace).Get(routetable.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(routetable.Spec, current.Spec) && reflect.DeepEqual(routetable.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.RouteTable)
		new.Spec = routetable.Spec
		new.Labels = routetable.Labels
		r, e := c.OcicoreV1alpha1().RouteTables(network.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG route table create\n")
		r, e := c.OcicoreV1alpha1().RouteTables(network.Namespace).Create(routetable)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// DeleteVcn deletes the virtual network resource
func DeleteVcn(c clientset.Interface, network *cloudv1alpha1.Network) (*v1alpha1.Vcn, error) {

	current, err := c.OcicoreV1alpha1().Vcns(network.Namespace).Get(network.Name, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG virtual network delete\n")
		if current.DeletionTimestamp == nil {
			if e := c.OcicoreV1alpha1().Vcns(network.Namespace).Delete(network.Name, &metav1.DeleteOptions{}); e != nil {
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

// DeleteInternetGateway deletes the internet gateway resource
func DeleteInternetGateway(c clientset.Interface, network *cloudv1alpha1.Network) (*v1alpha1.InternetGateway, error) {

	current, err := c.OcicoreV1alpha1().InternetGatewaies(network.Namespace).Get(network.Name, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG internet gateway delete\n")
		if current.DeletionTimestamp == nil {
			if e := c.OcicoreV1alpha1().InternetGatewaies(network.Namespace).Delete(network.Name, &metav1.DeleteOptions{}); e != nil {
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

// DeleteRouteTable deletes the route table resource
func DeleteRouteTable(c clientset.Interface, network *cloudv1alpha1.Network) (*v1alpha1.RouteTable, error) {

	current, err := c.OcicoreV1alpha1().RouteTables(network.Namespace).Get(network.Name, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG route table delete %#v\n", current)
		if current.DeletionTimestamp == nil {
			if e := c.OcicoreV1alpha1().RouteTables(network.Namespace).Delete(network.Name, &metav1.DeleteOptions{}); e != nil {
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
