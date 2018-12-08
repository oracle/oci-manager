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

package common

import (
	"errors"
	"github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Instance returns the instance object for the receiving oci resource
func Instance(clientset versioned.Interface, ns, name string) (instance *v1alpha1.Instance, err error) {

	instance, err = clientset.OcicoreV1alpha1().Instances(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return instance, err
	}
	return instance, nil
}

// Subnet returns the subnet resource for the receiving oci resource
func Subnet(clientset versioned.Interface, ns, name string) (subnet *v1alpha1.Subnet, err error) {

	subnet, err = clientset.OcicoreV1alpha1().Subnets(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return subnet, err
	}
	return subnet, nil
}

// SubnetId returns the oci id of the subnet for the receiving oci resource
func SubnetId(clientset versioned.Interface, ns, name string) (id string, err error) {

	subnet, err := clientset.OcicoreV1alpha1().Subnets(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return id, err
	}
	if subnet.Status.Resource == nil || *subnet.Status.Resource.Id == "" {
		return id, errors.New("Subnet resource is not created")
	}
	return *subnet.Status.Resource.Id, nil
}

// Vcn returns the vcn object for the receiving oci resource
func Vcn(clientset versioned.Interface, ns, name string) (vnet *v1alpha1.Vcn, err error) {

	vnet, err = clientset.OcicoreV1alpha1().Vcns(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return vnet, err
	}
	return vnet, nil
}

// VcnId returns the oci id of the vcn for the receiving oci resource
func VcnId(clientset versioned.Interface, ns, name string) (id string, err error) {

	vnet, err := clientset.OcicoreV1alpha1().Vcns(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return id, err
	}
	if vnet.Status.Resource == nil || *vnet.Status.Resource.Id == "" {
		return id, errors.New("Vcn resource is not created")
	}
	return *vnet.Status.Resource.Id, nil
}

// InternetGateway returns the internet gateway object for the receiving oci resource
func InternetGateway(clientset versioned.Interface, ns, name string) (ig *v1alpha1.InternetGateway, err error) {

	ig, err = clientset.OcicoreV1alpha1().InternetGatewaies(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return ig, err
	}
	return ig, nil
}

// InternetGatewayId returns the oci id of the internet gateway for the receiving oci resource
func InternetGatewayId(clientset versioned.Interface, ns, name string) (id string, err error) {

	ig, err := clientset.OcicoreV1alpha1().InternetGatewaies(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return id, err
	}
	if ig.Status.Resource == nil || *ig.Status.Resource.Id == "" {
		return id, errors.New("InternetGateway resource is not created")
	}
	return *ig.Status.Resource.Id, nil
}

// RouteTable returns the route table object for the receiving oci resource
func RouteTable(clientset versioned.Interface, ns, name string) (rt *v1alpha1.RouteTable, err error) {

	rt, err = clientset.OcicoreV1alpha1().RouteTables(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return rt, err
	}
	return rt, nil
}

// RouteTableId returns the oci id of the route table for the receiving oci resource
func RouteTableId(clientset versioned.Interface, ns, name string) (id string, err error) {

	rt, err := clientset.OcicoreV1alpha1().RouteTables(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return id, err
	}
	if rt.Status.Resource == nil || *rt.Status.Resource.Id == "" {
		return id, errors.New("RouteTable resource is not created")
	}
	return *rt.Status.Resource.Id, nil
}

// SecurityRuleSet returns the security rule set object for the receiving oci resource
func SecurityRuleSet(clientset versioned.Interface, ns, name string) (sl *v1alpha1.SecurityRuleSet, err error) {

	sl, err = clientset.OcicoreV1alpha1().SecurityRuleSets(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return sl, err
	}
	return sl, nil
}

// SecurityRuleSetId returns the oci id of the security rule set for the receiving oci resource
func SecurityRuleSetId(clientset versioned.Interface, ns, name string) (id string, err error) {

	sl, err := clientset.OcicoreV1alpha1().SecurityRuleSets(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return id, err
	}
	if sl.Status.Resource == nil || *sl.Status.Resource.Id == "" {
		return id, errors.New("SecurityRuleSet resource is not created")
	}
	return *sl.Status.Resource.Id, nil
}

// VolumeBackup returns the volume backup object for the receiving oci resource
func VolumeBackup(clientset versioned.Interface, ns, name string) (vb *v1alpha1.VolumeBackup, err error) {

	vb, err = clientset.OcicoreV1alpha1().VolumeBackups(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return vb, err
	}
	return vb, nil
}

// VolumeBackupId returns the oci id of the volume backup for the receiving oci resource
func VolumeBackupId(clientset versioned.Interface, ns, name string) (id string, err error) {

	vb, err := clientset.OcicoreV1alpha1().VolumeBackups(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return id, err
	}
	if vb.Status.Resource == nil || *vb.Status.Resource.Id == "" {
		return id, errors.New("VolumeBackup resource is not created")
	}
	return *vb.Status.Resource.Id, nil
}

// Volume returns the volume object for the receiving oci resource
func Volume(clientset versioned.Interface, ns, name string) (vol *v1alpha1.Volume, err error) {

	vol, err = clientset.OcicoreV1alpha1().Volumes(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return vol, err
	}
	return vol, nil
}

// VolumeId returns the oci id of the volume for the receiving oci resource
func VolumeId(clientset versioned.Interface, ns, name string) (id string, err error) {

	vol, err := clientset.OcicoreV1alpha1().Volumes(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return id, err
	}
	if vol.Status.Resource == nil || *vol.Status.Resource.Id == "" {
		return id, errors.New("Volume resource is not created")
	}
	return *vol.Status.Resource.Id, nil
}
