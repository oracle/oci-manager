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
	"k8s.io/apimachinery/pkg/runtime"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// ResourceType is a common type for all oci resources
type KubeContollerType struct {
	ControllerName string
	GroupName      string
	Kind           string
	ResourcePlural string
	Type           runtime.Object
	Validation     *apiextv1beta1.CustomResourceValidation
	AdapterFactory AdapterFactory
}

var typeRegistry = make(map[string]KubeContollerType)

// 	RegisterResourceType(controllerName string, runtime.Object(), factory AdapterFactory)
func RegisterKubernetesType(groupName, kind, resourcePlural, controllerName string, object runtime.Object, factory AdapterFactory) {
	RegisterKubernetesTypeWithValidation(groupName, kind, resourcePlural, controllerName, object, nil, factory)
}

// RegisterResourceTypeWithValidation returns a resource type struct with proper validation
func RegisterKubernetesTypeWithValidation(groupName, kind, resourcePlural, controllerName string,
	objectType runtime.Object,

	validation *apiextv1beta1.CustomResourceValidation,
	factory AdapterFactory) {

	_, ok := typeRegistry[controllerName]
	if ok {
		// TODO Is panicking ok given that this is part of a type-registration mechanism
		panic(fmt.Sprintf("Resource type %q has already been registered", controllerName))
	}
	typeRegistry[controllerName] = KubeContollerType{
		ControllerName: controllerName,
		GroupName:      groupName,
		Kind:           kind,
		ResourcePlural: resourcePlural,
		Validation:     validation,
		Type:           objectType,
		AdapterFactory: factory,
	}
}

// ResourceTypes returns a mapping of kind (e.g. "namespace") to the
// type information required to configure its resource.
func KubernetesTypes() map[string]KubeContollerType {
	// TODO copy RequiredResources to avoid accidental mutation
	result := make(map[string]KubeContollerType)
	for key, value := range typeRegistry {
		result[key] = value
	}
	return result
}
