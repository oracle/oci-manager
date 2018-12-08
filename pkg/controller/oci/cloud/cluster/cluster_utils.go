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

package cluster

import (
	"fmt"
	"github.com/golang/glog"
	"reflect"
	"strconv"
	"time"

	"k8s.io/api/core/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"

	ocice "github.com/oracle/oci-go-sdk/containerengine"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	cev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	idv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
	cloudcompute "github.com/oracle/oci-manager/pkg/controller/oci/cloud/compute"
)

const (
	MASTER_KIND = "master"
	WORKER_KIND = "worker"
	LB_SVC_KIND = "lb"
	NODE_KIND   = "node"

	ADMIN_RBAC = `kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kubernetes-admin-clusterrolebinding
subjects:
- kind: User
  name: kubernetes-admin
  apiGroup: ""
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""
`
	BASE_KUBEADM_INSTALL = `#!/bin/bash -x

iptables -F
sysctl net.bridge.bridge-nf-call-iptables=1

apt-get update
apt-get install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") $(lsb_release -cs) stable"
apt-get update && apt-get install -y docker-ce=$(apt-cache madison docker-ce | grep 17.03 | head -1 | awk '{print $3}')

apt-get update && apt-get install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF
apt-get update
apt-get install -y kubelet kubeadm kubectl
`

	// TCP for ssl passthru
	LB_PROTOCOL = "TCP"
	LB_PORT     = 443

	BACKEND_K8S_API_PORT = 6443
)

// const/vars needed for references
var POD_NETWORK = "10.244.0.0/16"
var SERVICE_NETWORK = "10.96.0.0/16"

var MASTER_RULES = []string{"0.0.0.0/0 tcp 22", "0.0.0.0/0 tcp 6443", "0.0.0.0/0 tcp 443"}
var WORKER_RULES = []string{"0.0.0.0/0 tcp 22", "0.0.0.0/0 tcp 10250"}

// getInstanceName gets the name of cluster's child instance with an ordinal index of ordinal
func getComputeName(cluster *cloudv1alpha1.Cluster, ordinal int) string {
	return fmt.Sprintf("%s-%d", cluster.Name, ordinal)
}

// CreateOrUpdateCompute reconciles the compute resource
func CreateOrUpdateCompute(c clientset.Interface,
	cluster *cloudv1alpha1.Cluster,
	controllerRef *metav1.OwnerReference,
	kind string,
	extraSans string) (*cloudv1alpha1.Compute, bool, error) {

	glog.Infof("CreateOrUpdateCompute: %s %s", kind, cluster.Name)

	computeName := cluster.Name + "-" + kind
	replicas := 1
	sshKeys := []string{}
	userData := BASE_KUBEADM_INSTALL

	// the lb subnets will be on .10/16, masters on .20/16
	endpoint := cluster.Name + "-master-0." + cluster.Name + "master21." + cluster.Name + ".oraclevcn.com"

	labels := map[string]string{
		cloudv1alpha1.ClusterKind: cluster.Name,
		kind:                      "true",
	}

	if kind == MASTER_KIND {
		replicas = cluster.Spec.Master.Replicas
		sshKeys = cluster.Spec.Master.Template.SshKeys

		// ca cert and key
		userData += "mkdir -p /etc/kubernetes/pki\n"

		userData += "cat <<EOF >/etc/kubernetes/pki/ca.crt\n"
		userData += cluster.Spec.CA.Certificate
		userData += "EOF\n"

		userData += "cat <<EOF >/etc/kubernetes/pki/ca.key\n"
		userData += cluster.Spec.CA.Key
		userData += "EOF\n"

		// kubernetes-admin user cluster-admin role binding
		userData += "cat <<EOF >/etc/kubernetes/kubernetes-admin-rbac.yaml\n"
		userData += ADMIN_RBAC
		userData += "EOF\n"

		userData += "kubeadm init" +
			" --apiserver-cert-extra-sans=" + extraSans + "," + endpoint +
			" --pod-network-cidr=" + POD_NETWORK +
			" --token=" + cluster.Spec.Token + "\n"

		userData += "export KUBECONFIG=/etc/kubernetes/admin.conf\n"
		userData += "kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/v0.10.0/Documentation/kube-flannel.yml\n"

		// allow cluster-admin role to kubernetes-admin user with client certs signed by ca cert
		userData += "kubectl apply -f /etc/kubernetes/kubernetes-admin-rbac.yaml\n"

	} else if kind == WORKER_KIND {
		replicas = cluster.Spec.Worker.Replicas
		sshKeys = cluster.Spec.Master.Template.SshKeys
		userData += "kubeadm join --token=" + cluster.Spec.Token + " --discovery-token-unsafe-skip-ca-verification " +
			endpoint + ":" + strconv.Itoa(BACKEND_K8S_API_PORT)

	}

	compute := &cloudv1alpha1.Compute{
		ObjectMeta: metav1.ObjectMeta{
			Name:   computeName,
			Labels: labels,
		},
		Spec: cloudv1alpha1.ComputeSpec{
			Network:  cluster.Name,
			Replicas: replicas,
			Template: cloudv1alpha1.Template{
				UserData: cloudv1alpha1.UserData{
					Shellscript: userData,
				},
				OsType:    "ubuntu",
				OsVersion: "16.04",
				SshKeys:   sshKeys,
			},
			SecuritySelector: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
				kind:                      "true",
			},
		},
	}

	if val, ok := cluster.Annotations[cloudcompute.IMAGE_ANNOTATION]; ok {
		compute.Annotations = make(map[string]string)
		compute.Annotations[cloudcompute.IMAGE_ANNOTATION] = val
	}

	if controllerRef != nil {
		compute.OwnerReferences = append(compute.OwnerReferences, *controllerRef)
	}

	current, err := c.CloudV1alpha1().Computes(cluster.Namespace).Get(compute.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(compute.Spec, current.Spec) && reflect.DeepEqual(compute.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*cloudv1alpha1.Compute)
		new.Spec = compute.Spec
		new.Labels = compute.Labels
		r, e := c.CloudV1alpha1().Computes(cluster.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {

		compute.Status.State = cloudv1alpha1.OperatorStatePending
		r, e := c.CloudV1alpha1().Computes(cluster.Namespace).Create(compute)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateNetwork reconciles the network resource
func CreateOrUpdateNetwork(c clientset.Interface,
	cluster *cloudv1alpha1.Cluster,
	controllerRef *metav1.OwnerReference) (*cloudv1alpha1.Network, bool, error) {

	netName := cluster.Name

	glog.Infof("CreateOrUpdateNetwork: %s", netName)

	net := &cloudv1alpha1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: netName,
			Labels: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
			},
		},
		Spec: cloudv1alpha1.NetworkSpec{
			CidrBlock: "10.0.0.0/16",
		},
	}

	if controllerRef != nil {
		net.OwnerReferences = append(net.OwnerReferences, *controllerRef)
	}

	current, err := c.CloudV1alpha1().Networks(cluster.Namespace).Get(netName, metav1.GetOptions{})

	if err == nil {
		return current, false, nil
	} else if apierrors.IsNotFound(err) {
		glog.Infof("CreateOrUpdateNetwork: not found. creating")
		r, e := c.CloudV1alpha1().Networks(cluster.Namespace).Create(net)
		return r, true, e
	} else {
		glog.Infof("CreateOrUpdateNetwork: unknown err: %v", err)
		return nil, false, err
	}
}

// CreateOrUpdateCluster reconciles the ce cluster resource
func CreateOrUpdateCluster(c clientset.Interface,
	cluster *cloudv1alpha1.Cluster,
	controllerRef *metav1.OwnerReference,
	serviceLbSubnetRefs *[]string,
	clusterName *string) (*cev1alpha1.Cluster, bool, error) {

	glog.Infof("CreateOrUpdateCluster: %s", *clusterName)

	var err error

	options := &ocice.ClusterCreateOptions{
		KubernetesNetworkConfig: &ocice.KubernetesNetworkConfig{
			PodsCidr:     &POD_NETWORK,
			ServicesCidr: &SERVICE_NETWORK,
		},
	}

	version := "v" + cluster.Spec.Version

	clusterObject := &cev1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: *clusterName,
			Labels: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
			},
		},
		Spec: cev1alpha1.ClusterSpec{
			CompartmentRef:      cluster.Namespace,
			KubernetesVersion:   &version,
			VcnRef:              cluster.Name,
			ServiceLbSubnetRefs: *serviceLbSubnetRefs,
			Options:             options,
		},
	}

	if controllerRef != nil {
		clusterObject.OwnerReferences = append(clusterObject.OwnerReferences, *controllerRef)
	}

	current, err := c.OciceV1alpha1().Clusters(cluster.Namespace).Get(clusterObject.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(clusterObject.Spec, current.Spec) && reflect.DeepEqual(clusterObject.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*cev1alpha1.Cluster)
		new.Spec = clusterObject.Spec
		new.Labels = clusterObject.Labels
		r, e := c.OciceV1alpha1().Clusters(cluster.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG virtual compute create\n")
		r, e := c.OciceV1alpha1().Clusters(cluster.Namespace).Create(clusterObject)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateNodePool reconciles the nodepool resource
func CreateOrUpdateNodePool(c clientset.Interface,
	cluster *cloudv1alpha1.Cluster,
	controllerRef *metav1.OwnerReference,
	nodeSubnetRefs *[]string,
	clusterName *string) (*cev1alpha1.NodePool, bool, error) {

	glog.Infof("CreateOrUpdateNodePool: %s", *clusterName)

	qtyPerSubnet := cluster.Spec.Worker.Replicas / 3

	shape, err := common.GetShapeByResourceRequirements(c, cluster.Namespace, &cluster.Spec.Worker.Template.Resources)

	// OKE limits image to Oracle-Linux 7.4 or 7.5
	image := cluster.Spec.Worker.Template.OsType + "-" + cluster.Spec.Worker.Template.OsVersion

	sshKey := ""
	if cluster.Spec.Worker.Template.SshKeys != nil {
		sshKey = cluster.Spec.Worker.Template.SshKeys[0]
	}

	version := "v" + cluster.Spec.Version

	nodePoolObject := &cev1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: *clusterName,
			Labels: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
			},
		},
		Spec: cev1alpha1.NodePoolSpec{
			CompartmentRef:    cluster.Namespace,
			KubernetesVersion: &version,
			QuantityPerSubnet: &qtyPerSubnet,
			NodeImageName:     &image,
			NodeShape:         &shape,
			ClusterRef:        cluster.Name,
			SubnetRefs:        *nodeSubnetRefs,
			SshPublicKey:      &sshKey,
		},
	}

	if controllerRef != nil {
		nodePoolObject.OwnerReferences = append(nodePoolObject.OwnerReferences, *controllerRef)
	}

	current, err := c.OciceV1alpha1().NodePools(cluster.Namespace).Get(nodePoolObject.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(nodePoolObject.Spec, current.Spec) && reflect.DeepEqual(nodePoolObject.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*cev1alpha1.NodePool)
		new.Spec = nodePoolObject.Spec
		new.Labels = nodePoolObject.Labels
		r, e := c.OciceV1alpha1().NodePools(cluster.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG virtual compute create\n")
		r, e := c.OciceV1alpha1().NodePools(cluster.Namespace).Create(nodePoolObject)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdatePolicy reconciles the policy resource
func CreateOrUpdatePolicy(c clientset.Interface,
	cluster *cloudv1alpha1.Cluster,
	controllerRef *metav1.OwnerReference) (*idv1alpha1.Policy, bool, error) {

	glog.Infof("CreateOrUpdatePolicy: %s", cluster.Name)

	desc := "oci-manager cloud cluster for oke"

	// used to get the root compartment id
	compartment, err := c.OciidentityV1alpha1().Compartments(cluster.Namespace).Get(cluster.Namespace, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Could not get compartment: %s err: %v", cluster.Name, err)
	}

	if compartment.Status.Resource == nil || compartment.Status.Resource.CompartmentId == nil {
		return nil, false, fmt.Errorf("compartment not populated yet")
	}

	policyObject := &idv1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster.Name,
			Labels: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
			},
		},
		Spec: idv1alpha1.PolicySpec{
			CompartmentRef: *compartment.Status.Resource.CompartmentId,
			Description:    &desc,
			Statements:     []string{"allow service OKE to manage all-resources in tenancy"},
		},
	}

	if controllerRef != nil {
		policyObject.OwnerReferences = append(policyObject.OwnerReferences, *controllerRef)
	}

	current, err := c.OciidentityV1alpha1().Policies(cluster.Namespace).Get(policyObject.Name, metav1.GetOptions{})

	if err == nil {
		if reflect.DeepEqual(policyObject.Spec, current.Spec) && reflect.DeepEqual(policyObject.Labels, current.Labels) {
			return current, false, nil
		}
		new := current.DeepCopyObject().(*idv1alpha1.Policy)
		new.Spec = policyObject.Spec
		new.Labels = policyObject.Labels
		r, e := c.OciidentityV1alpha1().Policies(cluster.Namespace).Update(new)
		return r, true, e
	} else if apierrors.IsNotFound(err) {
		// fmt.Printf("DEBUG virtual compute create\n")
		r, e := c.OciidentityV1alpha1().Policies(cluster.Namespace).Create(policyObject)
		return r, true, e
	} else {
		return nil, false, err
	}

}

// CreateOrUpdateLoadBalancer reconciles the LoadBalancer resource
func CreateOrUpdateLoadBalancer(c clientset.Interface,
	cluster *cloudv1alpha1.Cluster,
	controllerRef *metav1.OwnerReference) (*cloudv1alpha1.LoadBalancer, bool, error) {

	glog.Infof("CreateOrUpdateLoadBalancer: %s", cluster.Name)

	listeners := []cloudv1alpha1.Listener{}
	listener := cloudv1alpha1.Listener{
		Port:     LB_PORT,
		Protocol: LB_PROTOCOL,
	}
	listeners = append(listeners, listener)

	healthCheck := cloudv1alpha1.HealthCheck{
		Protocol: LB_PROTOCOL,
		Port:     BACKEND_K8S_API_PORT,
	}

	lb := &cloudv1alpha1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster.Name,
			Labels: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
			},
		},
		Spec: cloudv1alpha1.LoadBalancerSpec{
			ComputeSelector: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
				MASTER_KIND:               "true",
			},
			BackendPort: BACKEND_K8S_API_PORT,
			SecuritySelector: map[string]string{
				cloudv1alpha1.ClusterKind: cluster.Name,
				MASTER_KIND:               "true",
			},
			Listeners:   listeners,
			HealthCheck: healthCheck,
			IsPrivate:   false,
		},
		Status: cloudv1alpha1.LoadBalancerStatus{
			Network: cluster.Name,
		},
	}

	if controllerRef != nil {
		lb.OwnerReferences = append(lb.OwnerReferences, *controllerRef)
	}

	lb.Spec.Listeners = listeners

	current, err := c.CloudV1alpha1().LoadBalancers(cluster.Namespace).Get(cluster.Name, metav1.GetOptions{})
	if err == nil {
		if current.Status.IPAddress != "" {
			return current, false, nil
		}

		e := wait.PollImmediate(30*time.Second, 600*time.Second, func() (bool, error) {
			current, err = c.CloudV1alpha1().LoadBalancers(cluster.Namespace).Get(cluster.Name, metav1.GetOptions{})
			if err != nil {
				glog.Errorf("CreateOrUpdateLoadBalancer get ip error: %v", err)
				return false, err
			}
			glog.Infof("CreateOrUpdateLoadBalancer ip: %s", current.Status.IPAddress)

			if current.Status.IPAddress != "" {
				glog.Infof("breaking out of wait loop for ip")
				return true, nil
			}
			return false, nil
		})

		if e != nil {
			glog.Errorf("error from trying to get lb ip: %v", e)
		}

		return current, false, nil

	} else if apierrors.IsNotFound(err) {
		glog.Infof("lb not found, creating")
		_, e := c.CloudV1alpha1().LoadBalancers(cluster.Namespace).Create(lb)
		return lb, true, e
	} else {
		return nil, false, err
	}
}

// CreateOrUpdateSecurity reconciles the security resource
func CreateOrUpdateSecurity(c clientset.Interface,
	cluster *cloudv1alpha1.Cluster,
	controllerRef *metav1.OwnerReference, kind string) (*cloudv1alpha1.Security, bool, error) {

	secName := cluster.Name
	var ingressRules []string

	labels := map[string]string{
		cloudv1alpha1.ClusterKind: cluster.Name,
	}

	secName += "-" + kind
	labels[kind] = "true"

	switch kind {
	case LB_SVC_KIND:
		ingressRules = MASTER_RULES
	case NODE_KIND:
		ingressRules = MASTER_RULES
	case MASTER_KIND:
		ingressRules = MASTER_RULES
	case WORKER_KIND:
		ingressRules = WORKER_RULES
	}

	glog.Infof("CreateOrUpdateSecurity: %s", secName)

	networkSelector := make(map[string]string)
	networkSelector[cloudv1alpha1.ClusterKind] = cluster.Name

	sec := &cloudv1alpha1.Security{
		ObjectMeta: metav1.ObjectMeta{
			Name:   secName,
			Labels: labels,
		},
		Spec: cloudv1alpha1.SecuritySpec{
			Ingress:         ingressRules,
			NetworkSelector: networkSelector,
		},
	}

	if controllerRef != nil {
		sec.OwnerReferences = append(sec.OwnerReferences, *controllerRef)
	}

	current, err := c.CloudV1alpha1().Securities(cluster.Namespace).Get(secName, metav1.GetOptions{})
	if err == nil {
		return current, false, nil
	} else if apierrors.IsNotFound(err) {
		r, e := c.CloudV1alpha1().Securities(cluster.Namespace).Create(sec)
		glog.Infof("created security: %s", sec.Name)
		return r, true, e
	} else {
		glog.Infof("unknown security err: %v", err)
		return nil, false, err
	}
}

// DeleteLoadBalancer deletes the loadbalancer resource
func DeleteLoadBalancer(c clientset.Interface, namespace string, lbName string) (*cloudv1alpha1.LoadBalancer, error) {

	current, err := c.CloudV1alpha1().LoadBalancers(namespace).Get(lbName, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG virtual cluster delete\n")
		if current.DeletionTimestamp == nil {
			if e := c.CloudV1alpha1().LoadBalancers(namespace).Delete(lbName, &metav1.DeleteOptions{}); e != nil {
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

// DeleteCompute deletes the compute resource
func DeleteCompute(c clientset.Interface, namespace string, instanceName string) (*cloudv1alpha1.Compute, error) {

	current, err := c.CloudV1alpha1().Computes(namespace).Get(instanceName, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG virtual cluster delete\n")
		if current.DeletionTimestamp == nil {
			if e := c.CloudV1alpha1().Computes(namespace).Delete(instanceName, &metav1.DeleteOptions{}); e != nil {
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

// DeleteSecurity deletes the security resource
func DeleteSecurity(c clientset.Interface, namespace string, securityName string) (*cloudv1alpha1.Security, error) {

	current, err := c.CloudV1alpha1().Securities(namespace).Get(securityName, metav1.GetOptions{})

	if err == nil {
		// fmt.Printf("DEBUG virtual cluster delete\n")
		if current.DeletionTimestamp == nil {
			if e := c.CloudV1alpha1().Securities(namespace).Delete(securityName, &metav1.DeleteOptions{}); e != nil {
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

// DeleteNetwork deletes the subnet resource
func DeleteNetwork(c clientset.Interface, namespace string, name string) (*cloudv1alpha1.Network, error) {

	current, err := c.CloudV1alpha1().Networks(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.CloudV1alpha1().Networks(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
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

// DeletePolicy deletes the policy resource
func DeletePolicy(c clientset.Interface, namespace string, name string) (*idv1alpha1.Policy, error) {

	current, err := c.OciidentityV1alpha1().Policies(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OciidentityV1alpha1().Policies(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
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

// DeleteCluster deletes the cluster resource
func DeleteCluster(c clientset.Interface, namespace string, name string) (*cev1alpha1.Cluster, error) {

	current, err := c.OciceV1alpha1().Clusters(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OciceV1alpha1().Clusters(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
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

// DeleteNodePool deletes the nodepool resource
func DeleteNodePool(c clientset.Interface, namespace string, name string) (*cev1alpha1.NodePool, error) {

	current, err := c.OciceV1alpha1().NodePools(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.OciceV1alpha1().NodePools(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
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

// DeleteSecret deletes the secret resource
func DeleteSecret(c kubernetes.Interface, namespace string, name string) (*v1.Secret, error) {

	current, err := c.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		if current.DeletionTimestamp == nil {
			if e := c.CoreV1().Secrets(namespace).Delete(name, &metav1.DeleteOptions{}); e != nil {
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
