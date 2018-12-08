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

package core

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	coregroup "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicore "github.com/oracle/oci-go-sdk/core"

	"context"
	"os"

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"regexp"
	"strings"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.SubnetKind,
		ocicorev1alpha1.SubnetResourcePlural,
		ocicorev1alpha1.SubnetControllerName,
		&ocicorev1alpha1.SubnetValidation,
		NewSubnetAdapter)
}

// SubnetAdapter implements the adapter interface for volume resource
type SubnetAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	vcnClient resourcescommon.VcnClientInterface
}

// NewSubnetAdapter creates a new adapter for subnet resource
func NewSubnetAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	sa := SubnetAdapter{}

	vcnClient, err := ocicore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}

	sa.vcnClient = &vcnClient
	sa.clientset = clientset
	sa.ctx = context.Background()

	return &sa

}

// Kind returns the resource kind string
func (a *SubnetAdapter) Kind() string {
	return ocicorev1alpha1.SubnetKind
}

// Resource returns the plural name of the resource type
func (a *SubnetAdapter) Resource() string {
	return ocicorev1alpha1.SubnetResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *SubnetAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.SubnetResourcePlural)
}

// ObjectType returns the subnet type for this adapter
func (a *SubnetAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.Subnet{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *SubnetAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.Subnet)
	return ok
}

// Copy returns a copy of a subnet object
func (a *SubnetAdapter) Copy(obj runtime.Object) runtime.Object {
	subnet := obj.(*ocicorev1alpha1.Subnet)
	return subnet.DeepCopyObject()
}

// Equivalent checks if two subnet objects are the same
func (a *SubnetAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	subnet1 := obj1.(*ocicorev1alpha1.Subnet)
	subnet2 := obj2.(*ocicorev1alpha1.Subnet)
	if subnet1.Status.Resource != nil {
		subnet1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if subnet2.Status.Resource != nil {
		subnet2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	return reflect.DeepEqual(subnet1, subnet2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *SubnetAdapter) IsResourceCompliant(obj runtime.Object) bool {
	subnet := obj.(*ocicorev1alpha1.Subnet)

	if subnet.Status.Resource == nil {
		return false
	}

	resource := subnet.Status.Resource

	if resource.LifecycleState == ocicore.SubnetLifecycleStateProvisioning ||
		resource.LifecycleState == ocicore.SubnetLifecycleStateTerminating {
		return true
	}

	if resource.LifecycleState == ocicore.SubnetLifecycleStateTerminated {
		return false
	}

	specDisplayName := resourcescommon.Display(subnet.Name, subnet.Spec.DisplayName)

	if *resource.DisplayName != *specDisplayName ||
		*resource.AvailabilityDomain != subnet.Spec.AvailabilityDomain ||
		*resource.CidrBlock != subnet.Spec.CidrBlock ||
		*resource.DnsLabel != subnet.Spec.DNSLabel ||
		*resource.ProhibitPublicIpOnVnic != false {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *SubnetAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	subnet1 := obj1.(*ocicorev1alpha1.Subnet)
	subnet2 := obj2.(*ocicorev1alpha1.Subnet)

	return subnet1.Status.Resource.LifecycleState != subnet2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *SubnetAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.Subnet).GetResourceID()
}

// ObjectMeta returns the object meta struct from the subnet object
func (a *SubnetAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.Subnet).ObjectMeta
}

// DependsOn returns a map of subnet dependencies (objects that the subnet depends on)
func (a *SubnetAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.Subnet).Spec.DependsOn
}

// Dependents returns a map of subnet dependents (objects that depend on the subnet)
func (a *SubnetAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.Subnet).Status.Dependents
}

// CreateObject creates the subnet object
func (a *SubnetAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Subnet)
	return a.clientset.OcicoreV1alpha1().Subnets(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the subnet object
func (a *SubnetAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Subnet)
	return a.clientset.OcicoreV1alpha1().Subnets(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the subnet object
func (a *SubnetAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.Subnet)
	return a.clientset.OcicoreV1alpha1().Subnets(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the subnet depends on
func (a *SubnetAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Subnet)
	deps := make([]runtime.Object, 0)

	for _, securityrulesetRef := range object.Spec.SecurityRuleSetRefs {
		if !resourcescommon.IsOcid(securityrulesetRef) {
			securityruleset, err := resourcescommon.SecurityRuleSet(a.clientset, object.ObjectMeta.Namespace, securityrulesetRef)
			if err != nil {
				return nil, err
			}
			deps = append(deps, securityruleset)
		}
	}

	if !resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	if !resourcescommon.IsOcid(object.Spec.VcnRef) {
		virtualnetwork, err := resourcescommon.Vcn(a.clientset, object.ObjectMeta.Namespace, object.Spec.VcnRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, virtualnetwork)
	}

	if !resourcescommon.IsOcid(object.Spec.RouteTableRef) {
		routetable, err := resourcescommon.RouteTable(a.clientset, object.ObjectMeta.Namespace, object.Spec.RouteTableRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, routetable)
	}

	return deps, nil
}

// Create creates the subnet resource in oci
func (a *SubnetAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		object           = obj.(*ocicorev1alpha1.Subnet)
		compartmentId    string
		virtualnetworkId string
		routetableId     string
		err              error
	)

	if resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartmentId = object.Spec.CompartmentRef
	} else {
		compartmentId, err = resourcescommon.CompartmentId(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
	}

	if resourcescommon.IsOcid(object.Spec.VcnRef) {
		virtualnetworkId = object.Spec.VcnRef
	} else {
		virtualnetworkId, err = resourcescommon.VcnId(a.clientset, object.ObjectMeta.Namespace, object.Spec.VcnRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
	}

	if resourcescommon.IsOcid(object.Spec.RouteTableRef) {
		routetableId = object.Spec.RouteTableRef
	} else {
		routetableId, err = resourcescommon.RouteTableId(a.clientset, object.ObjectMeta.Namespace, object.Spec.RouteTableRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
	}

	var securityrulesetList []string
	for _, securityrulesetRef := range object.Spec.SecurityRuleSetRefs {
		var securityrulesetId string
		if resourcescommon.IsOcid(securityrulesetRef) {
			securityrulesetId = securityrulesetRef
		} else {
			securityrulesetId, err = resourcescommon.SecurityRuleSetId(a.clientset, object.ObjectMeta.Namespace, securityrulesetRef)
			if err != nil {
				return object, object.Status.HandleError(err)
			}
		}
		securityrulesetList = append(securityrulesetList, securityrulesetId)
	}

	// create a new RouteTable
	request := ocicore.CreateSubnetRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.VcnId = ocisdkcommon.String(virtualnetworkId)
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)
	request.AvailabilityDomain = ocisdkcommon.String(object.Spec.AvailabilityDomain)
	request.CidrBlock = ocisdkcommon.String(object.Spec.CidrBlock)
	//request.DhcpOptionsId = ocisdkcommon.String("")
	request.DnsLabel = ocisdkcommon.String(object.Spec.DNSLabel)
	request.ProhibitPublicIpOnVnic = ocisdkcommon.Bool(false)
	request.RouteTableId = ocisdkcommon.String(routetableId)
	request.SecurityListIds = securityrulesetList

	// TODO: figure out why first successful create tx doesnt set status and subsequent fail due to duplicate create request w/ same opc token
	// until then, have this workaround/handling of overlap response
	//request.OpcRetryToken = ocisdkcommon.String(string(object.UID))
	//glog.Infof("Subnet: %s OpcRetryToken: %s", object.Name, string(object.UID))

	r, err := a.vcnClient.CreateSubnet(a.ctx, request)

	//     message: 'Service error:InvalidParameter. The requested CIDR 10.0.13.0/24 is invalid:
	// subnet ocid1.subnet.oc1.phx.aaaaaaaaivzyp4bbjuselwkrcutubz7igavxqui5te3rixtgrcswt3hnba5q
	// with CIDR 10.0.13.0/24 overlaps with this CIDR.. http status code: 400'
	if err != nil && strings.Contains(err.Error(), "overlap") {
		re := regexp.MustCompile("ocid[a-z0-9\\.]*")
		ocid := re.FindString(err.Error())

		glog.Infof("get subnet after overlap: %s", ocid)

		getRequest := ocicore.GetSubnetRequest{
			SubnetId: &ocid,
		}
		getResp, err := a.vcnClient.GetSubnet(a.ctx, getRequest)
		if err != nil {
			glog.Errorf("get subnet after overlap error: %v", err)
			return object, object.Status.HandleError(err)
		}
		if *getResp.DisplayName == object.Name || *getResp.DisplayName == object.Spec.DisplayName {
			glog.Infof("overlap name matches - setting resource")
			return object.SetResource(&getResp.Subnet), nil
		}

	}

	return object.SetResource(&r.Subnet), object.Status.HandleError(err)
}

// Delete deletes the subnet resource in oci
func (a *SubnetAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Subnet)

	request := ocicore.DeleteSubnetRequest{
		SubnetId: object.Status.Resource.Id,
	}

	_, e := a.vcnClient.DeleteSubnet(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the subnet resource from oci
func (a *SubnetAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Subnet)

	request := ocicore.GetSubnetRequest{
		SubnetId: object.Status.Resource.Id,
	}

	e := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		r, e := a.vcnClient.GetSubnet(a.ctx, request)
		if e != nil {
			return false, e
		}
		if r.LifecycleState != ocicore.SubnetLifecycleStateProvisioning {
			object.SetResource(&r.Subnet)
			return true, nil
		}
		return false, e
	})

	return object, object.Status.HandleError(e)
}

// Update updates the subnet resource in oci
func (a *SubnetAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Subnet)

	request := ocicore.UpdateSubnetRequest{
		SubnetId: object.Status.Resource.Id,
	}

	if object.Status.Resource.LifecycleState != ocicore.SubnetLifecycleStateAvailable {
		return object, errors.New(string(object.Status.Resource.LifecycleState))
	}

	r, e := a.vcnClient.UpdateSubnet(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.Subnet), object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the subnet resource in the subnet object
func (a *SubnetAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
