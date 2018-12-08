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

package util

import (
	"fmt"
	"github.com/golang/glog"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilErrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"reflect"
	"time"
)

func CreateResourceDefinition(clientset apiextensionsclient.Interface, plural, kind, groupName string, validation *apiextensionsv1beta1.CustomResourceValidation) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	var crdname = plural + "." + groupName
	var schemeGroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha1"}

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdname,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   groupName,
			Version: schemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: plural,
				Kind:   kind,
			},
			Validation: validation,
		},
	}
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			found, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdname, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			updCrd := found.DeepCopy()
			updCrd.Spec.Group = groupName
			updCrd.Spec.Version = schemeGroupVersion.Version
			updCrd.Spec.Scope = apiextensionsv1beta1.NamespaceScoped
			updCrd.Spec.Validation = validation
			updCrd.Spec.Names.Plural = plural
			updCrd.Spec.Names.Kind = kind

			if !reflect.DeepEqual(found.Spec, updCrd.Spec) {
				crd.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion
				_, updErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Update(updCrd)
				if updErr != nil {
					glog.Errorf("Error updating crd %s: %v", crd.Name, updErr)
					return nil, err
				}

				glog.Infof("Updated resource definition for %s/%s\n", groupName, kind)
			}
		} else {
			return nil, err
		}
	} else {
		glog.Infof("Created resource definition for %s/%s\n", groupName, kind)
	}

	// wait for CRD being established
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdname, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					fmt.Printf("CRD %s is established.\n", crdname)
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, err
	})
	if err != nil {
		deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(plural, nil)
		if deleteErr != nil {
			return nil, utilErrors.NewAggregate([]error{err, deleteErr})
		}
		return nil, err
	}
	return crd, nil
}
