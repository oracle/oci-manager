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

package compute

import (
	b64 "encoding/base64"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"

	"fmt"
	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
)

const (
	IMAGE_ANNOTATION = "oci.oracle.com/instance.image"
	SHAPE_ANNOTATION = "oci.oracle.com/instance.shape"
)

// getInstanceName gets the name of compute's child instance with an ordinal index of ordinal
func getInstanceName(compute *cloudv1alpha1.Compute, ordinal int) string {
	return fmt.Sprintf("%s-%d", compute.Name, ordinal)
}

// CreateOrUpdateInstance reconciles the instance resource
func CreateOrUpdateInstance(c clientset.Interface,
	compute *cloudv1alpha1.Compute,
	controllerRef *metav1.OwnerReference,
	availabilityDomain *string,
	subnetRef *string,
	instanceName *string) (*v1alpha1.Instance, bool, error) {

	glog.Infof("CreateOrUpdateInstance: %s", *instanceName)

	var err error
	var image string
	var shape string

	if val, ok := compute.ObjectMeta.Annotations[IMAGE_ANNOTATION]; ok {
		image, err = common.GetImageSpecific(c, compute.Namespace, val)
		if err != nil {
			return nil, false, err
		}
	} else {
		image, err = common.GetImage(c, compute.Spec.Template.OsType, compute.Spec.Template.OsVersion, compute.Namespace)
		if err != nil {
			return nil, false, err
		}
	}
	glog.Infof("image: %s", image)

	if val, ok := compute.ObjectMeta.Annotations[SHAPE_ANNOTATION]; ok {
		shape, err = common.GetShapeSpecific(c, compute.Namespace, val)
		if err != nil {
			return nil, false, err
		}
	} else {
		shape, err = common.GetShapeByResourceRequirements(c, compute.Namespace, &compute.Spec.Template.Resources)
		if err != nil {
			return nil, false, err
		}
	}
	glog.Infof("shape: %s", shape)

	userData := ""
	if compute.Spec.Template.UserData.Shellscript != "" {
		userData = compute.Spec.Template.UserData.Shellscript
	} else if compute.Spec.Template.UserData.CloudConfig != "" {
		userData = "#cloud-config\n"
		userData += compute.Spec.Template.UserData.CloudConfig
	}

	instance := &v1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name: *instanceName,
			Labels: map[string]string{
				"compute": compute.Name,
			},
		},
		Spec: v1alpha1.InstanceSpec{
			AvailabilityDomain: *availabilityDomain,
			CompartmentRef:     compute.Namespace,
			Image:              image,
			Metadata: map[string]string{
				"ssh_authorized_keys": strings.Join(compute.Spec.Template.SshKeys, "\n"),
				"user_data":           b64.StdEncoding.EncodeToString([]byte(userData)),
			},
			Shape:     shape,
			SubnetRef: *subnetRef,
		},
	}

	if controllerRef != nil {
		instance.OwnerReferences = append(instance.OwnerReferences, *controllerRef)
	}

	current, err := c.OcicoreV1alpha1().Instances(compute.Namespace).Get(instance.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(instance.Spec, current.Spec) && reflect.DeepEqual(instance.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.Instance)
		new.Spec = instance.Spec
		new.Labels = instance.Labels
		r, e := c.OcicoreV1alpha1().Instances(compute.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG virtual compute create\n")
		r, e := c.OcicoreV1alpha1().Instances(compute.Namespace).Create(instance)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateSubnet reconciles the subnet resource
func CreateOrUpdateSubnet(c clientset.Interface,
	namespace string,
	ownerType string,
	ownerName string,
	networkName string,
	controllerRef *metav1.OwnerReference,
	uniqueName string,
	availabilityDomain string,
	cidrBlock string,
	securitySelector *map[string]string) (*v1alpha1.Subnet, bool, error) {

	fullName := ownerName + "-" + uniqueName

	glog.Infof("CreateOrUpdateSubnet: %s", fullName)
	dnsLabel := strings.Replace(fullName, "-", "", -1)

	selector := ""
	for k, v := range *securitySelector {
		if selector != "" {
			selector += ","
		}
		selector += k + "=" + v
	}
	listOptions := metav1.ListOptions{LabelSelector: selector}
	securities, err := c.CloudV1alpha1().Securities(namespace).List(listOptions)
	if err != nil {
		return nil, false, err
	}
	glog.Infof("Security selector: %s matched: %v", selector, securities.Items)

	securityRuleSets := []string{}
	for _, s := range securities.Items {
		securityRuleSets = append(securityRuleSets, networkName+"-"+s.Name)
	}

	subnet := &v1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name: fullName,
			Labels: map[string]string{
				ownerType:                 ownerName,
				cloudv1alpha1.NetworkKind: networkName,
			},
		},
		Spec: v1alpha1.SubnetSpec{
			CompartmentRef:      namespace,
			VcnRef:              networkName,
			RouteTableRef:       networkName,
			AvailabilityDomain:  availabilityDomain,
			CidrBlock:           cidrBlock,
			DisplayName:         fullName,
			DNSLabel:            dnsLabel,
			SecurityRuleSetRefs: securityRuleSets,
		},
	}

	if controllerRef != nil {
		subnet.OwnerReferences = append(subnet.OwnerReferences, *controllerRef)
	}

	current, err := c.OcicoreV1alpha1().Subnets(namespace).Get(fullName, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(subnet.Spec, current.Spec) && reflect.DeepEqual(subnet.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.Subnet)
		new.Spec = subnet.Spec
		new.Labels = subnet.Labels
		r, e := c.OcicoreV1alpha1().Subnets(namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		r, e := c.OcicoreV1alpha1().Subnets(namespace).Create(subnet)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// DeleteInstance deletes the instance resource
func DeleteInstance(c clientset.Interface, namespace string, instanceName string) (*v1alpha1.Instance, error) {

	current, err := c.OcicoreV1alpha1().Instances(namespace).Get(instanceName, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG virtual compute delete\n")
		if current.DeletionTimestamp == nil {
			if e := c.OcicoreV1alpha1().Instances(namespace).Delete(instanceName, &metav1.DeleteOptions{}); e != nil {
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

// DeleteSubnet deletes the subnet resource
func DeleteSubnet(c clientset.Interface, namespace string, name string) (*v1alpha1.Subnet, error) {

	current, err := c.OcicoreV1alpha1().Subnets(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OcicoreV1alpha1().Subnets(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
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
