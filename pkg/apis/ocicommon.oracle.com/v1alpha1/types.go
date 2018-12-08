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

package v1alpha1

import (
	"github.com/golang/glog"
	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

// ResourceStatus is a generic struct to store status information about any OCI resource
type ResourceStatus struct {
	State        ResourceState       `json:"state,omitempty"`
	ResetCounter int                 `json:"resetcounter,omitempty"`
	Message      string              `json:"message,omitempty"`
	Dependents   map[string][]string `json:"dependents,omitempty"`
}

// ResourceState stores state for any OCI resource
type ResourceState string

const (
	// ResourceStatePending is used when the resource is pending reconcilation
	ResourceStatePending ResourceState = "Pending"
	// ResourceStateCreated is used when the resource is created in OCI
	ResourceStateCreated ResourceState = "Created"
	// ResourceStateProcessed is used when the reconcilation completes
	ResourceStateProcessed ResourceState = "Processed"
	// ResourceStateError indicates if the resource reconcilation encountered an error
	ResourceStateError ResourceState = "Error"
)

// Regex constants used for object fields validation to ensure proper types and values
// Defined as constants because they are used multiple times in difference OCI resource types
const (
	AnyStringValidationRegex          = ".+"
	AvailabilityDomainValidationRegex = "^[a-zA-Z0-9]+\\:[a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]$"
	CidrValidationRegex               = "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\\/([0-9]|[1-2][0-9]|3[0-2]))$"
	DomainValidationRegex             = "^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\\.[a-zA-Z]{2,}$"
	HostnameValidationRegex           = "^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])$"
	Ipv4ValidationRegex               = "^$|^(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[0-9]{1,2})(\\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[0-9]{1,2})){3}$"
	LoadBalancerProtocolRegex         = "^HTTP$|^HTTP2$|^TCP$"
	NoOrAnyStringValidationRegex      = "^$|.+"

	ValidationTypeArray   = "array"
	ValidationTypeBoolean = "boolean"
	ValidationTypeInteger = "integer"
	ValidationTypeString  = "string"
)

// MetaDataValidation is variable used to construct the schema validation property for the CRDs
var MetaDataValidation = apiextv1beta1.JSONSchemaProps{
	Required: []string{"name"},
	Properties: map[string]apiextv1beta1.JSONSchemaProps{
		"name": {
			Type:    ValidationTypeString,
			Pattern: HostnameValidationRegex,
		},
		"namespace": {
			Type:    ValidationTypeString,
			Pattern: HostnameValidationRegex,
		},
	},
}

// Dependency is an array of explicit DependsOn relations between objects
type Dependency struct {
	DependsOn map[string]DependsOn `json:"dependson,omitempty"`
}

// DependsOn is user-defined explicit relationship between objects using selectors
type DependsOn struct {
	LabelSelector map[string]string `json:"labelselector,omitempty"`
	FieldSelector map[string]string `json:"fieldselector,omitempty"`
}

// HandleError updates the object with errors
func (s *ResourceStatus) HandleError(e error) error {

	if e != nil {
		if err, ok := ocisdkcommon.IsServiceError(e); ok {
			if err.GetCode() == "NotAuthorizedOrNotFound" {
				// assume deleted
				return nil
			}
		}
		glog.Errorf("OCI Error: %v", e)
		s.State = ResourceStateError
		s.Message = e.Error()
		return e
	}

	s.State = ResourceStateProcessed
	s.Message = "OK"

	return nil
}

// AddDependent adds a dependent to the resource status
func (s *ResourceStatus) AddDependent(kind string, obj runtime.Object) error {
	if obj != nil {

		if s.Dependents == nil {
			s.Dependents = make(map[string][]string)
		} else if objRegistered, _ := s.IsDependentRegistered(kind, obj); objRegistered {
			return nil
		}
		ref, err := cache.MetaNamespaceKeyFunc(obj)
		if err != nil {
			return err
		}
		if len(kind) > 0 {
			s.Dependents[kind] = append(s.Dependents[kind], ref)
		}
	}

	return nil
}

// IsDependentRegistered checks if a dependent exists in the resource status
func (s *ResourceStatus) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	if obj != nil {
		if s.Dependents == nil || len(s.Dependents) == 0 {
			return false, nil
		}

		dependentKey, err := cache.MetaNamespaceKeyFunc(obj)
		if err != nil {
			return false, err
		}

		if len(kind) > 0 {
			if refs, ok := s.Dependents[kind]; ok {
				for _, ref := range refs {
					if ref == dependentKey {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

// RemoveDependent removes a dependant from the resource status
func (s *ResourceStatus) RemoveDependent(kind string, obj runtime.Object) error {
	if obj != nil {
		if s.Dependents != nil {
			dependentKey, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return err
			}
			if refs, ok := s.Dependents[kind]; ok {
				newRefs := make([]string, 0)
				for _, ref := range refs {
					if ref != dependentKey {
						newRefs := append(newRefs, ref)
						s.Dependents[kind] = newRefs
						return nil
					}
				}
				if len(newRefs) > 0 {
					s.Dependents[kind] = newRefs
				} else {
					delete(s.Dependents, kind)
				}
			}
		}
	}
	return nil
}

// GetDependsOn is getter for DependsOn
func (d *Dependency) GetDependsOn() map[string]DependsOn {
	return d.DependsOn
}

// ObjectInterface is an interface for resource objects that supports dependencies
type ObjectInterface interface {
	AddDependent(kind string, obj runtime.Object) error
	RemoveDependent(kind string, obj runtime.Object) error
	IsDependentRegistered(kind string, obj runtime.Object) (bool, error)
	GetResourcePlural() string
	GetResourceID() string
	IsResource() bool
	GetGroupVersionResource() schema.GroupVersionResource
	GetResourceState() ResourceState
}
