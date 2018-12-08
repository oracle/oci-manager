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
	"fmt"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// ResourceType is a common type for all oci resources
type ResourceType struct {
	GroupName      string
	Kind           string
	ResourcePlural string
	ControllerName string
	Validation     *apiextv1beta1.CustomResourceValidation
	AdapterFactory AdapterFactory
}

var typeRegistry = make(map[string]ResourceType)

// RegisterResourceType registers the resource as an adapter
func RegisterResourceType(groupName, kind, resourcePlural, controllerName string, factory AdapterFactory) {
	RegisterResourceTypeWithValidation(groupName, kind, resourcePlural, controllerName, nil, factory)
}

// RegisterResourceTypeWithValidation returns a resource type struct with proper validation
func RegisterResourceTypeWithValidation(groupName, kind, resourcePlural string, controllerName string, validation *apiextv1beta1.CustomResourceValidation, factory AdapterFactory) {
	_, ok := typeRegistry[kind]
	if ok {
		// TODO Is panicking ok given that this is part of a type-registration mechanism
		panic(fmt.Sprintf("Resource type %q has already been registered", kind))
	}
	typeRegistry[kind] = ResourceType{
		GroupName:      groupName,
		Kind:           kind,
		ResourcePlural: resourcePlural,
		ControllerName: controllerName,
		Validation:     validation,
		AdapterFactory: factory,
	}
}

// ResourceTypes returns a mapping of kind (e.g. "compartment") to the
// type information required to configure its resource.
func ResourceTypes() map[string]ResourceType {
	// TODO copy RequiredResources to avoid accidental mutation
	result := make(map[string]ResourceType)
	for key, value := range typeRegistry {
		result[key] = value
	}
	return result
}
