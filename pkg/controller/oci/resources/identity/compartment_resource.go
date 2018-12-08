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

package identity

import (
	"k8s.io/client-go/kubernetes"
	"reflect"
	"sort"
	"time"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	identitygroup "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com"
	ociidentityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicore "github.com/oracle/oci-go-sdk/core"
	ociidentity "github.com/oracle/oci-go-sdk/identity"

	"context"
	"os"

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	FullDeleteLabel = "FullDelete"
)

// Adapter functions
func init() {
	resourcescommon.RegisterResourceType(
		identitygroup.GroupName,
		ociidentityv1alpha1.CompartmentKind,
		ociidentityv1alpha1.CompartmentResourcePlural,
		ociidentityv1alpha1.CompartmentControllerName,
		NewCompartmentAdapter)
}

// CompartmentAdapter implements the adapter interface for compartment resource
type CompartmentAdapter struct {
	clientset            versioned.Interface
	ctx                  context.Context
	ociIdClient          resourcescommon.IdentityClientInterface
	ociCoreComputeClient resourcescommon.ComputeClientInterface
	tenancyId            string
}

// NewCompartmentAdapter creates a new adapter for compartment resource
func NewCompartmentAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	ca := CompartmentAdapter{}

	idClient, err := ociidentity.NewIdentityClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci IDENTITY client: %v", err)
		os.Exit(1)
	}

	computeClient, err := ocicore.NewComputeClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci COMPUTE client: %v", err)
		os.Exit(1)
	}

	ca.ociIdClient = &idClient
	ca.ociCoreComputeClient = &computeClient
	ca.clientset = clientset
	ca.tenancyId, _ = ociconfig.TenancyOCID()
	ca.ctx = context.Background()
	return &ca
}

// Kind returns the resource kind string
func (a *CompartmentAdapter) Kind() string {
	return ociidentityv1alpha1.CompartmentKind
}

// Resource returns the plural name of the resource type
func (a *CompartmentAdapter) Resource() string {
	return ociidentityv1alpha1.CompartmentResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *CompartmentAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ociidentityv1alpha1.SchemeGroupVersion.WithResource(ociidentityv1alpha1.CompartmentResourcePlural)
}

// ObjectType returns the compartment type for this adapter
func (a *CompartmentAdapter) ObjectType() runtime.Object {
	return &ociidentityv1alpha1.Compartment{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *CompartmentAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ociidentityv1alpha1.Compartment)
	return ok
}

// Copy returns a copy of a compartment object
func (a *CompartmentAdapter) Copy(obj runtime.Object) runtime.Object {
	compartment := obj.(*ociidentityv1alpha1.Compartment)
	return compartment.DeepCopyObject()
}

// Equivalent checks if two compartment objects are the same
func (a *CompartmentAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	compartment1 := obj1.(*ociidentityv1alpha1.Compartment)
	compartment2 := obj2.(*ociidentityv1alpha1.Compartment)

	if compartment1.Status.Resource != nil {
		compartment1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}

	if compartment2.Status.Resource != nil {
		compartment2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}

	if compartment2.Status.Images == nil || compartment2.Status.Shapes == nil {
		return false
	}

	return reflect.DeepEqual(compartment1.Status, compartment2.Status) && reflect.DeepEqual(compartment1.Spec, compartment2.Spec)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *CompartmentAdapter) IsResourceCompliant(obj runtime.Object) bool {
	compartment := obj.(*ociidentityv1alpha1.Compartment)
	if compartment.Status.Resource == nil {
		return false
	}

	resource := compartment.Status.Resource
	if resource.LifecycleState == ociidentity.CompartmentLifecycleStateCreating ||
		resource.LifecycleState == ociidentity.CompartmentLifecycleStateDeleting {
		return true
	}

	if resource.LifecycleState == ociidentity.CompartmentLifecycleStateDeleted ||
		resource.LifecycleState == ociidentity.CompartmentLifecycleStateInactive {
		return false
	}

	if *compartment.Status.Resource.Name != compartment.Name {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *CompartmentAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	compartment1 := obj1.(*ociidentityv1alpha1.Compartment)
	compartment2 := obj2.(*ociidentityv1alpha1.Compartment)

	if !reflect.DeepEqual(compartment1.Status.AvailabilityDomains, compartment2.Status.AvailabilityDomains) {
		return true
	}

	if !reflect.DeepEqual(compartment1.Status.Images, compartment2.Status.Images) {
		return true
	}

	if !reflect.DeepEqual(compartment1.Status.Shapes, compartment2.Status.Shapes) {
		return true
	}

	return compartment1.Status.Resource.LifecycleState != compartment2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *CompartmentAdapter) Id(obj runtime.Object) string {
	return obj.(*ociidentityv1alpha1.Compartment).GetResourceID()
}

// ObjectMeta returns the object meta struct from the compartment object
func (a *CompartmentAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ociidentityv1alpha1.Compartment).ObjectMeta
}

// DependsOn returns a map of compartment dependencies (objects that the compartment depends on)
func (a *CompartmentAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ociidentityv1alpha1.Compartment).Spec.DependsOn
}

// Dependents returns a map of compartment dependents (objects that depend on the compartment)
func (a *CompartmentAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ociidentityv1alpha1.Compartment).Status.Dependents
}

// CreateObject creates the compartment object
func (a *CompartmentAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Compartment)
	return a.clientset.OciidentityV1alpha1().Compartments(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the compartment object
func (a *CompartmentAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Compartment)
	return a.clientset.OciidentityV1alpha1().Compartments(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the compartment object
func (a *CompartmentAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ociidentityv1alpha1.Compartment)
	return a.clientset.OciidentityV1alpha1().Compartments(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the compartment depends on
func (a *CompartmentAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	return make([]runtime.Object, 0), nil
}

// Shapes returns the oci shapes available in this compartment
func (a *CompartmentAdapter) Shapes(obj runtime.Object) []string {
	var object = obj.(*ociidentityv1alpha1.Compartment)
	shapes := make(map[string]string)

	request := ocicore.ListShapesRequest{
		CompartmentId: object.Status.Resource.Id,
	}

	r, err := a.ociCoreComputeClient.ListShapes(a.ctx, request)

	if r.Items == nil || len(r.Items) == 0 || err != nil {
		glog.Errorf("Invalid response from ListShapes, error: %v", err)
		return make([]string, 0)
	}

	for _, ociShape := range r.Items {
		shapes[*(ociShape.Shape)] = "found"
	}
	keys := make([]string, len(shapes))
	i := 0
	for k := range shapes {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// Images returns the oci images available in this compartment
func (a *CompartmentAdapter) Images(obj runtime.Object) map[string]string {
	var object = obj.(*ociidentityv1alpha1.Compartment)
	images := make(map[string]string)

	request := ocicore.ListImagesRequest{
		CompartmentId: object.Status.Resource.Id,
	}

	r, err := a.ociCoreComputeClient.ListImages(a.ctx, request)

	if r.Items == nil || len(r.Items) == 0 || err != nil {
		glog.Errorf("Invalid response from ListImages, error: %v", err)
		return images
	}

	for _, ociImage := range r.Items {
		images[*(ociImage.DisplayName)] = *(ociImage.Id)
	}
	return images
}

// AvailabilityDomains returns the oci availability domains available in this compartment
func (a *CompartmentAdapter) AvailabilityDomains(obj runtime.Object) []string {
	var object = obj.(*ociidentityv1alpha1.Compartment)
	availabilityDomains := []string{}

	request := ociidentity.ListAvailabilityDomainsRequest{
		CompartmentId: object.Status.Resource.Id,
	}

	r, err := a.ociIdClient.ListAvailabilityDomains(a.ctx, request)

	if r.Items == nil || len(r.Items) == 0 || err != nil {
		glog.Errorf("Invalid response from ListAvailabilityDomain, error: %v", err)
		return availabilityDomains
	}

	for _, ociAvailabilityDomain := range r.Items {
		availabilityDomains = append(availabilityDomains, *(ociAvailabilityDomain.Name))
	}
	return availabilityDomains
}

// Create creates the compartment resource in oci
func (a *CompartmentAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var compartment = obj.(*ociidentityv1alpha1.Compartment)

	response, e := a.ociIdClient.ListCompartments(a.ctx, ociidentity.ListCompartmentsRequest{CompartmentId: &a.tenancyId})

	if e != nil {
		glog.Errorf("Error querying compartment: %v", e)
		return nil, compartment.Status.HandleError(e)
	}

	for _, r := range response.Items {
		if *(r.Name) == compartment.Name {
			return compartment.SetResource(&r), compartment.Status.HandleError(e)
		}
	}

	// Create compartment not supported yet
	createRequest := ociidentity.CreateCompartmentRequest{
		OpcRetryToken: ocisdkcommon.String(string(compartment.UID)),
		CreateCompartmentDetails: ociidentity.CreateCompartmentDetails{
			Name:          &compartment.Name,
			Description:   &compartment.Spec.Description,
			CompartmentId: &a.tenancyId,
		},
	}
	resp, e := a.ociIdClient.CreateCompartment(a.ctx, createRequest)
	if e != nil {
		return compartment, e
	}
	// if created (vs discovered) allow deletion
	labels := make(map[string]string, 0)
	labels[FullDeleteLabel] = "true"
	compartment.Labels = labels

	// wait 5s else get shapes / images will return 404
	time.Sleep(5 * time.Second)

	return compartment.SetResource(&resp.Compartment), e
}

// Delete deletes the compartment resource in oci
func (a *CompartmentAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Compartment)

	if val, ok := object.Labels[FullDeleteLabel]; ok {
		if val == "true" {
			request := ociidentity.DeleteCompartmentRequest{
				CompartmentId: object.Status.Resource.Id,
			}
			_, e := a.ociIdClient.DeleteCompartment(a.ctx, request)

			if e == nil && object.Status.Resource != nil {
				object.Status.Resource.Id = ocisdkcommon.String("")
			}

		} else {
			glog.Infof("not deleting compartment in oci due to missing label: %s", FullDeleteLabel)
		}
	} else {
		glog.Infof("compartment delete, but no ForceDelete label")
	}

	return object, nil
}

// Get retrieves the compartment resource from oci
func (a *CompartmentAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Compartment)

	r, e := a.ociIdClient.GetCompartment(a.ctx, ociidentity.GetCompartmentRequest{CompartmentId: object.Status.Resource.Id})
	if e != nil {
		return object, object.Status.HandleError(e)
	}

	object.Status.Shapes = a.Shapes(object)
	object.Status.Images = a.Images(object)
	object.Status.AvailabilityDomains = a.AvailabilityDomains(object)

	return object.SetResource(&r.Compartment), object.Status.HandleError(e)
}

// Update updates the compartment resource in oci - NOT IMPLEMENTED!
func (a *CompartmentAdapter) Update(obj runtime.Object) (runtime.Object, error) {

	// Update compartment not implemented
	return a.Get(obj)
}

// UpdateForResource calls a common UpdateForResource method to update the compartment resource in the compartment object
func (a *CompartmentAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
