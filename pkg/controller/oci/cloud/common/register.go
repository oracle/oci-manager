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

// CloudType configures kubernetes for a cloud workload type
type CloudType struct {
	GroupName      string
	Kind           string
	ResourcePlural string
	Validation     *apiextv1beta1.CustomResourceValidation
	AdapterFactory AdapterFactory
}

var typeRegistry = make(map[string]CloudType)

//Register cloud type adpater factory
func RegisterCloudType(plural, kind, groupName string, validation *apiextv1beta1.CustomResourceValidation, factory AdapterFactory) {
	_, ok := typeRegistry[kind]
	if ok {
		// TODO Is panicking ok given that this is part of a type-registration mechanism
		panic(fmt.Sprintf("Resource type %q has already been registered", kind))
	}
	typeRegistry[kind] = CloudType{
		Kind:           kind,
		ResourcePlural: plural,
		GroupName:      groupName,
		Validation:     validation,
		AdapterFactory: factory,
	}
}

// CloudTypes returns a mapping of kind (e.g. "network") to the
// type information required to configure its resource.
func CloudTypes() map[string]CloudType {
	//copy RequiredResources to avoid accidental mutation
	result := make(map[string]CloudType)
	for key, value := range typeRegistry {
		result[key] = value
	}
	return result
}
