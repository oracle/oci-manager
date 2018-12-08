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
	"sort"
	"strconv"
	"strings"

	"github.com/golang/glog"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ADAnnotationName     string = "oci.oracle.com.ad"
	SubnetAnnotationName string = "oci.oracle.com.subnet"
)

// Get list of oci availability domains
func GetAvailabilityDomains(c clientset.Interface, namespace string, compartmentRef string) ([]string, error) {
	compartment, err := c.OciidentityV1alpha1().Compartments(namespace).Get(compartmentRef, metav1.GetOptions{})
	if err != nil {
		return []string{}, err
	}
	return compartment.Status.AvailabilityDomains, nil
}

// Get OCI instance image
func GetImage(c clientset.Interface, osType string, osVersion string, namespace string) (string, error) {
	os := strings.ToLower(osType) + "-" + osVersion
	match := ""
	compartment, err := c.OciidentityV1alpha1().Compartments(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return match, err
	}

	keys := make([]string, len(compartment.Status.Images))
	for i, _ := range compartment.Status.Images {
		keys = append(keys, i)
	}
	sort.Sort(sort.StringSlice(keys))
	for _, val := range keys {
		lowerVal := strings.ToLower(val)
		if strings.Contains(lowerVal, os) {
			match = val
			// break - comment/continue loop instead of reverse sorting then break'ing
		}
	}
	if match == "" {
		return match, errors.New("no image matching os type and version")
	}

	return match, nil
}

// Get OCI instance shape
func GetShape(c clientset.Interface, cpuCores int, ramGbytes int, isVirtual bool, namespace string) (string, error) {
	shape := "VM"
	if !isVirtual {
		shape = "BM"
	}
	shape += ".Standard"
	if cpuCores > 0 {
		shape += strconv.Itoa(cpuCores)
	} else {
		shape += "1"
	}

	if ramGbytes > 0 {
		shape += "." + strconv.Itoa(ramGbytes)
	} else {
		shape += ".1"
	}

	match := ""
	compartment, err := c.OciidentityV1alpha1().Compartments(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return match, err
	}

	for _, s := range compartment.Status.Shapes {
		if shape == s {
			match = shape
		}
	}
	if match == "" {
		return match, errors.New("no shape matching cpu ram and isVirtual")
	}

	return match, nil
}

// Get OCI instance shape
func GetShapeByResourceRequirements(c clientset.Interface, namespace string, requirements *v1.ResourceRequirements) (string, error) {
	shape := "VM.Standard"

	if requirements.Limits.Cpu().Value() > 0 {
		shape += requirements.Limits.Cpu().String()
	} else {
		shape += "1"
	}

	if requirements.Limits.Memory().Value() > 0 {
		shape += "." + requirements.Limits.Memory().String()
	} else {
		shape += ".1"
	}

	match := ""
	compartment, err := c.OciidentityV1alpha1().Compartments(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return match, err
	}

	for _, s := range compartment.Status.Shapes {
		if shape == s {
			match = shape
		}
	}
	if match == "" {
		return match, errors.New("no shape matching cpu ram")
	}

	return match, nil
}

// Get specific OCI image
func GetImageSpecific(c clientset.Interface, namespace string, img string) (string, error) {
	match := ""
	compartment, err := c.OciidentityV1alpha1().Compartments(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return match, err
	}
	for i, _ := range compartment.Status.Images {
		if i == img {
			match = i
		}
	}
	if match == "" {
		return match, errors.New("no image matching: " + img)
	}
	return match, nil
}

// Get specific OCI shape
func GetShapeSpecific(c clientset.Interface, namespace string, shape string) (string, error) {
	match := ""
	compartment, err := c.OciidentityV1alpha1().Compartments(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return match, err
	}
	for _, s := range compartment.Status.Shapes {
		if shape == s {
			match = s
		}
	}
	if match == "" {
		return match, errors.New("no shape matching: " + shape)
	}
	return match, nil
}

func isAvailable(a int, list *[]int) bool {
	for _, b := range *list {
		if b == a {
			return false
		}
	}
	return true
}

// get subnet offset - used by compute and loadbalancer (key)
func GetSubnetOffset(c clientset.Interface, namespace string, networkName string, key string) (int, error) {

	network, err := c.CloudV1alpha1().Networks(namespace).Get(networkName, metav1.GetOptions{})
	if err != nil {
		return -1, err
	}
	if network.Status.SubnetAllocationMap == nil {
		network.Status.SubnetAllocationMap = make(map[string]int)
	}
	var subnetOffset int
	if val, ok := network.Status.SubnetAllocationMap[key]; ok {
		subnetOffset = val
		if err != nil {
			return -1, err
		}
	} else {
		allocated := make([]int, len(network.Status.SubnetAllocationMap))
		for _, val := range network.Status.SubnetAllocationMap {
			allocated = append(allocated, val)
		}

		// 10, 20 ... 250 ...allows for 25 compute on a network, w/ 9 az (subnet /24) per compute
		i := 10
		for i < 260 {
			if isAvailable(i, &allocated) {
				network.Status.SubnetAllocationMap[key] = i
				_, err := c.CloudV1alpha1().Networks(namespace).Update(network)
				if err != nil {
					return -1, err
				}
				subnetOffset = i
				break
			}
			i += 10
		}
		if i == 260 {
			msg := "all /24 used in: " + network.Spec.CidrBlock
			glog.Errorf(msg)
			return -100, errors.New(msg)
		}
	}
	glog.Infof("GetSubnetOffset subnet %s %v", key, subnetOffset)
	return subnetOffset, nil
}
