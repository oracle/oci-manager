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

package cpod

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/golang/glog"
	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	"github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
)

// DeleteInstance deletes the virtual compute resource
func DeleteCpodInstance(c clientset.Interface, cpod *cloudv1alpha1.Cpod) (*v1alpha1.Instance, error) {

	inst, err := c.OcicoreV1alpha1().Instances(cpod.Namespace).Get(cpod.Name, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG virtual compute delete\n")
		if inst.DeletionTimestamp == nil {
			glog.Infof("Deleting oci instnce %s", inst.Name)
			if e := c.OcicoreV1alpha1().Instances(cpod.Namespace).Delete(cpod.Name, &metav1.DeleteOptions{}); e != nil {
				return inst, e
			}
		}
		return inst, nil
	} else if apierrors.IsNotFound(err) {
		return nil, nil
	} else {
		return nil, err
	}

}

func getInstanceName(cpod *cloudv1alpha1.Cpod) string {
	return cpod.Name
}

func CreateOrUpdateCpodInstance(c clientset.Interface,
	cpod *cloudv1alpha1.Cpod,
	controllerRef *metav1.OwnerReference) (*v1alpha1.Instance, bool, error) {

	// TODO: add logic to use explicit image and shape from annotation
	image, err := common.GetImage(c, DefaultCpodInstanceOsType, DefaultCpodInstanceOSVersion, cpod.Namespace)
	if err != nil {
		return nil, false, err
	}

	var cpuLimit = resource.MustParse("0")
	var memLimit = resource.MustParse("0")
	for _, container := range cpod.Spec.Containers {
		if container.Resources.Limits != nil {
			cpuLimit.Add(*container.Resources.Limits.Cpu())
			memLimit.Add(*container.Resources.Limits.Memory())
		}
	}

	cpuInt64, _ := cpuLimit.AsInt64()
	memInt64, _ := memLimit.AsInt64()
	var shape string
	if cpuInt64 == 0 || memInt64 == 0 {
		shape = DefaultShape
	} else {
		shape, err = common.GetShape(c, int(cpuInt64), int(memInt64), true, cpod.Namespace)
	}

	if err != nil {
		return nil, false, err
	}

	/*
		availabilityDomain := cpod.Annotations[common.ADAnnotationName]
		if availabilityDomain == "" {
			return nil, false, fmt.Errorf("%v annotation must be specified on CPOD %v definition", common.ADAnnotationName, cpod.Name)
		}
	*/

	subnetRef := cpod.Annotations[common.SubnetAnnotationName]
	if subnetRef == "" {
		return nil, false, fmt.Errorf("%v annotation must be specified on CPOD %v definition", common.SubnetAnnotationName, cpod.Name)
	}

	subnet, err := c.OcicoreV1alpha1().Subnets(cpod.Namespace).Get(subnetRef, metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}
	if !subnet.IsResource() {
		return nil, false, fmt.Errorf("Subnet %s is not ready for CPOD %s", subnetRef, cpod.Name)
	}

	availabilityDomain := subnet.Status.Resource.AvailabilityDomain

	userData := getUserData(cpod)

	instance := &v1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name: getInstanceName(cpod),
			Labels: map[string]string{
				"cpod": cpod.Name,
			},
		},
		Spec: v1alpha1.InstanceSpec{
			AvailabilityDomain: *availabilityDomain,
			CompartmentRef:     cpod.Namespace,
			Image:              image,
			Metadata: map[string]string{
				"user_data": b64.StdEncoding.EncodeToString([]byte(userData)),
			},
			Shape:     shape,
			SubnetRef: subnetRef,
		},
	}
	//	"ssh_authorized_keys": strings.Join(cpod.Spec.SshKeys, "\n"),
	if cpod.Spec.SshKeys != nil && len(cpod.Spec.SshKeys) > 0 {
		instance.Spec.Metadata["ssh_authorized_keys"] = strings.Join(cpod.Spec.SshKeys, "\n")
	}
	if controllerRef != nil {
		instance.OwnerReferences = append(instance.OwnerReferences, *controllerRef)
	}

	current, err := c.OcicoreV1alpha1().Instances(cpod.Namespace).Get(instance.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(instance.Spec, current.Spec) && reflect.DeepEqual(instance.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.Instance)
		new.Spec = instance.Spec
		new.Labels = instance.Labels
		r, e := c.OcicoreV1alpha1().Instances(cpod.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG virtual compute create\n")
		r, e := c.OcicoreV1alpha1().Instances(cpod.Namespace).Create(instance)
		glog.Infof("Created oci instnce %s", instance.Name)
		return r, true, e
	} else {
		return nil, false, err
	}

}

func getUserData(cpod *cloudv1alpha1.Cpod) string {

	var data = userDataDockerTemplate
	for _, container := range cpod.Spec.Containers {
		parts := strings.Split(container.Image, "/")
		containerRef := parts[1] + "/" + parts[2]
		containerPort := container.Ports[0].ContainerPort
		hostPort := DefaultCpodHostPort
		if container.Ports[0].HostPort != 0 {
			hostPort = fmt.Sprint(container.Ports[0].HostPort)
		}
		r := strings.NewReplacer(
			"{docker-image}", container.Image,
			"{container-name}", container.Name,
			"{container-port}", fmt.Sprint(containerPort),
			"{container}", containerRef,
			"{host-port}", hostPort,
		)
		cntrRun := r.Replace(dockerContainerRun)
		data = data + cntrRun
	}

	return data
}

var userDataDockerTemplate string = `#!/bin/bash -x
yum install -y yum-utils device-mapper-persistent-data lvm2
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum install -y docker-ce-18.03.1.ce
systemctl start docker
systemctl enable docker
`

var dockerContainerRun string = `docker pull {docker-image}
docker run --name {container-name} -d -p {host-port}:{container-port} {container}
`
