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
package security

import (
	"reflect"
	"strconv"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"

	"github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"

	ocicore "github.com/oracle/oci-go-sdk/core"
)

// ParseIngressRule converts a string from the spec to IngressSecurityRule type
func ParseIngressRule(s string) ocicore.IngressSecurityRule {
	var source string
	var min int
	var max int

	x := strings.Fields(s)

	// Each rule must have at least 2 fields for source and protocol
	if len(x) < 2 {
		return ocicore.IngressSecurityRule{}
	}

	source = x[0]

	if len(x) > 2 {
		if s, err := strconv.Atoi(x[2]); err == nil {
			min = s
		}
		max = min
	}
	if len(x) > 3 {
		if s, err := strconv.Atoi(x[3]); err == nil {
			max = s
		}
	}

	// https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml
	// https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
	var protocol string
	switch p := strings.ToLower(x[1]); p {
	case "icmp":
		protocol = "1"
		return ocicore.IngressSecurityRule{
			Source:   &source,
			Protocol: &protocol,
		}
	case "tcp":
		protocol = "6"
		return ocicore.IngressSecurityRule{
			Source:   &source,
			Protocol: &protocol,
			TcpOptions: &ocicore.TcpOptions{
				DestinationPortRange: &ocicore.PortRange{
					Min: &min,
					Max: &max,
				},
			},
		}
	case "udp":
		protocol = "17"
		return ocicore.IngressSecurityRule{
			Source:   &source,
			Protocol: &protocol,
			UdpOptions: &ocicore.UdpOptions{
				DestinationPortRange: &ocicore.PortRange{
					Min: &min,
					Max: &max,
				},
			},
		}
	default:
		return ocicore.IngressSecurityRule{}
	}

}

// ParseEgressRule converts a string from the spec to EgressSecurityRule type
func ParseEgressRule(s string) ocicore.EgressSecurityRule {
	var destination string
	var min int
	var max int

	x := strings.Fields(s)

	// Each rule must have at least 2 fields for source and protocol
	if len(x) < 2 {
		return ocicore.EgressSecurityRule{}
	}

	destination = x[0]

	if len(x) > 2 {
		if s, err := strconv.Atoi(x[2]); err == nil {
			min = s
		}
		max = min
	}
	if len(x) > 3 {
		if s, err := strconv.Atoi(x[3]); err == nil {
			max = s
		}
	}

	// https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml
	// https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
	var protocol string
	switch p := strings.ToLower(x[1]); p {
	case "icmp":
		protocol = "1"
		return ocicore.EgressSecurityRule{
			Destination: &destination,
			Protocol:    &protocol,
		}
	case "tcp":
		protocol = "6"
		return ocicore.EgressSecurityRule{
			Destination: &destination,
			Protocol:    &protocol,
			TcpOptions: &ocicore.TcpOptions{
				DestinationPortRange: &ocicore.PortRange{
					Min: &min,
					Max: &max,
				},
			},
		}
	case "udp":
		protocol = "17"
		return ocicore.EgressSecurityRule{
			Destination: &destination,
			Protocol:    &protocol,
			UdpOptions: &ocicore.UdpOptions{
				DestinationPortRange: &ocicore.PortRange{
					Min: &min,
					Max: &max,
				},
			},
		}
	default:
		return ocicore.EgressSecurityRule{}
	}

}

// CreateOrUpdateSecurityRuleSet reconciles the security list resource
func CreateOrUpdateSecurityRuleSet(c clientset.Interface, security *cloudv1alpha1.Security, controllerRef *metav1.OwnerReference, vncRef string) (*v1alpha1.SecurityRuleSet, bool, error) {

	// ingress
	ingressSecurityRules := []ocicore.IngressSecurityRule{}
	if security.Spec.Ingress != nil {
		for _, s := range security.Spec.Ingress {
			ingressSecurityRules = append(ingressSecurityRules, ParseIngressRule(s))
		}
	}
	// default - allow ssh
	if len(ingressSecurityRules) == 0 {
		source := "0.0.0.0/0"
		protocol := "tcp"
		sshPort := 22
		defaultIngress := ocicore.IngressSecurityRule{
			Source:   &source,
			Protocol: &protocol,
			TcpOptions: &ocicore.TcpOptions{
				DestinationPortRange: &ocicore.PortRange{
					Min: &sshPort,
					Max: &sshPort,
				},
			},
		}
		ingressSecurityRules = append(ingressSecurityRules, defaultIngress)
	}

	// egress
	egressSecurityRules := []ocicore.EgressSecurityRule{}
	if security.Spec.Egress != nil {
		for _, s := range security.Spec.Egress {
			egressSecurityRules = append(egressSecurityRules, ParseEgressRule(s))
		}
	}
	// default - allow all
	if len(egressSecurityRules) == 0 {
		destination := "0.0.0.0/0"
		protocol := "all"
		defaultEgress := ocicore.EgressSecurityRule{
			Destination: &destination,
			Protocol:    &protocol,
		}
		egressSecurityRules = append(egressSecurityRules, defaultEgress)
	}

	securityRuleSetName := vncRef + "-" + security.Name

	securityruleset := &v1alpha1.SecurityRuleSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: securityRuleSetName,
			Labels: map[string]string{
				"security": security.Name,
			},
		},
		Spec: v1alpha1.SecurityRuleSetSpec{
			CompartmentRef:       security.Namespace,
			VcnRef:               vncRef,
			EgressSecurityRules:  egressSecurityRules,
			IngressSecurityRules: ingressSecurityRules,
		},
	}

	if controllerRef != nil {
		securityruleset.OwnerReferences = append(securityruleset.OwnerReferences, *controllerRef)
	}

	current, err := c.OcicoreV1alpha1().SecurityRuleSets(security.Namespace).Get(securityruleset.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(securityruleset.Spec, current.Spec) && reflect.DeepEqual(securityruleset.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*v1alpha1.SecurityRuleSet)
		new.Spec = securityruleset.Spec
		new.Labels = securityruleset.Labels
		r, e := c.OcicoreV1alpha1().SecurityRuleSets(security.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {

		r, e := c.OcicoreV1alpha1().SecurityRuleSets(security.Namespace).Create(securityruleset)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// DeleteSecurityRuleSet deletes the security list resource
func DeleteSecurityRuleSet(c clientset.Interface, security *cloudv1alpha1.Security, securityRuleSetName string) (*v1alpha1.SecurityRuleSet, error) {
	current, err := c.OcicoreV1alpha1().SecurityRuleSets(security.Namespace).Get(securityRuleSetName, metav1.GetOptions{})
	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OcicoreV1alpha1().SecurityRuleSets(security.Namespace).Delete(securityRuleSetName, &metav1.DeleteOptions{}); e != nil {
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
