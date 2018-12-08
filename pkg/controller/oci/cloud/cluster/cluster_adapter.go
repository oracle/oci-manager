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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	mathrand "math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/kubernetes/cmd/kubeadm/app/phases/certs/pkiutil"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	cev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	idv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	cloudcommon "github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
	cloudcompute "github.com/oracle/oci-manager/pkg/controller/oci/cloud/compute"
)

const (
	ADMIN_USER = "kubernetes-admin"
)

type ClusterAdapter struct {
	subscribtions []schema.GroupVersionResource
	clientset     versioned.Interface
	kclientset    kubernetes.Interface
	lister        cache.GenericLister
	queue         workqueue.RateLimitingInterface
}

var controllerKind = cloudv1alpha1.SchemeGroupVersion.WithKind(cloudv1alpha1.ClusterKind)
var K8S_GROUPS = []string{MASTER_KIND, WORKER_KIND}

// init to register cluster cloud type
func init() {
	cloudcommon.RegisterCloudType(
		cloudv1alpha1.ClusterResourcePlural,
		cloudv1alpha1.ClusterKind,
		cloudv1alpha1.GroupName,
		&cloudv1alpha1.ClusterValidation,
		NewClusterAdapter,
	)
}

// factory method
func NewClusterAdapter(clientSet versioned.Interface, kubeclient kubernetes.Interface) cloudcommon.CloudTypeAdapter {
	na := ClusterAdapter{
		clientset:  clientSet,
		kclientset: kubeclient,
	}
	na.subscribtions = subscribe()
	mathrand.Seed(time.Now().UTC().UnixNano())
	return &na
}

// subscribe to resource events
func subscribe() []schema.GroupVersionResource {
	subs := make([]schema.GroupVersionResource, 0)

	subs = append(subs, cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.NetworkResourcePlural))
	subs = append(subs, cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.SecurityResourcePlural))

	// for managed / oke
	subs = append(subs, cev1alpha1.SchemeGroupVersion.WithResource(cev1alpha1.ClusterResourcePlural))
	subs = append(subs, cev1alpha1.SchemeGroupVersion.WithResource(cev1alpha1.NodePoolResourcePlural))
	// subs = append(subs, corev1alpha1.SchemeGroupVersion.WithResource(corev1alpha1.SubnetResourcePlural))

	// for non-managed / kubeadm
	subs = append(subs, cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.ComputeResourcePlural))
	subs = append(subs, cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.LoadBalancerResourcePlural))

	return subs
}

// set lister
func (a *ClusterAdapter) SetLister(lister cache.GenericLister) {
	a.lister = lister
}

// set queue
func (a *ClusterAdapter) SetQueue(q workqueue.RateLimitingInterface) {
	a.queue = q
}

// kind
func (a *ClusterAdapter) Kind() string {
	return cloudv1alpha1.ClusterKind
}

// resource
func (a *ClusterAdapter) Resource() string {
	return cloudv1alpha1.ClusterResourcePlural
}

// group version with resource
func (a *ClusterAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.ClusterResourcePlural)
}

// subscriptions
func (a *ClusterAdapter) Subscriptions() []schema.GroupVersionResource {
	return a.subscribtions
}

// object meta
func (a *ClusterAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*cloudv1alpha1.Cluster).ObjectMeta
}

// equivalent
func (a *ClusterAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	cluster1 := obj1.(*cloudv1alpha1.Cluster)
	cluster2 := obj2.(*cloudv1alpha1.Cluster)

	status_equal := reflect.DeepEqual(cluster1.Status, cluster2.Status)
	glog.Infof("Equivalent Status: %v", status_equal)
	spec_equal := reflect.DeepEqual(cluster1.Spec, cluster2.Spec)
	glog.Infof("Equivalent Spec: %v", spec_equal)

	equal := spec_equal && status_equal &&
		cluster1.Status.State != cloudv1alpha1.OperatorStatePending

	glog.Infof("Equivalent: %v", equal)

	return equal
}

// delete cluster cloud event handler
func (a *ClusterAdapter) deleteOKE(cluster *cloudv1alpha1.Cluster, network *cloudv1alpha1.Network) (runtime.Object, error) {

	wait := false

	cecluster, err := DeleteCluster(a.clientset, cluster.Namespace, cluster.Name)
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cev1alpha1.ClusterKind, err.Error())
		return cluster, err
	}
	if cecluster != nil {
		err = errors.New("cluster resource is not deleted yet")
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cev1alpha1.ClusterKind, err.Error())
		wait = true
	}

	nodepool, err := DeleteNodePool(a.clientset, cluster.Namespace, cluster.Name)
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cev1alpha1.NodePoolKind, err.Error())
		return cluster, err
	}
	if nodepool != nil {
		err = errors.New("nodepool resource is not deleted yet")
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cev1alpha1.NodePoolKind, err.Error())
		wait = true
	}

	if network != nil {
		wait = true
		for key, netOctet := range network.Status.SubnetAllocationMap {

			var availabilityDomains []string
			if len(cluster.Status.AvailabilityZones) == 0 {

				availabilityDomains, err = cloudcommon.GetAvailabilityDomains(a.clientset, cluster.Namespace, cluster.Namespace)
				if err != nil {
					cloudcommon.SetCondition(&cluster.Status.OperatorStatus, "AvailabilityDomain", err.Error())
					return cluster, err
				}
			} else {
				availabilityDomains = cluster.Status.AvailabilityZones
			}

			for i, _ := range availabilityDomains {
				snOctet := netOctet + i + 1
				snName := key + "-" + strconv.Itoa(snOctet)
				glog.Infof("delete subnet: " + snName)
				sn, err := cloudcompute.DeleteSubnet(a.clientset, cluster.Namespace, snName)
				if err != nil {
					cloudcommon.SetCondition(&cluster.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
					return cluster, err
				}
				if sn != nil {
					err = errors.New("subnet resource is not deleted yet")
					cloudcommon.SetCondition(&cluster.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
				}

			}
		}

	}

	security, err := DeleteSecurity(a.clientset, cluster.Namespace, cluster.Name+"-lb")
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
		return cluster, err
	}
	if security != nil {
		err = errors.New("lb security resource is not deleted yet")
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
		wait = true
	}

	security, err = DeleteSecurity(a.clientset, cluster.Namespace, cluster.Name+"-node")
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
		return cluster, err
	}
	if security != nil {
		err = errors.New("node security resource is not deleted yet")
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
		wait = true
	}

	// rm policy only if everything else (ie subnets) is deleted - oke needs to rm the nodepool instances that blocks subnets
	if !wait {
		policy, err := DeletePolicy(a.clientset, cluster.Namespace, cluster.Name)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, idv1alpha1.PolicyKind, err.Error())
			return cluster, err
		}
		if policy != nil {
			err = errors.New("policy resource is not deleted yet")
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, idv1alpha1.PolicyKind, err.Error())
			wait = true
		}
	}

	// Remove finalizers
	if !wait && len(cluster.Finalizers) > 0 {
		cluster.SetFinalizers([]string{})
		return cluster, nil
	}

	return cluster, nil
}

// delete cluster cloud event handler
func (a *ClusterAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	cluster := obj.(*cloudv1alpha1.Cluster)
	glog.Infof("start cluster delete...")

	wait := false

	// Delete kubeconfig secret
	secret, err := DeleteSecret(a.kclientset, cluster.Namespace, cluster.Name)
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, "secret", err.Error())
		return cluster, err
	}
	if secret != nil {
		err = errors.New("secret resource is not deleted yet")
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, "secret", err.Error())
		wait = true
	}

	// Delete the network resource
	net, err := DeleteNetwork(a.clientset, cluster.Namespace, cluster.Name)
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.NetworkKind, err.Error())
		return cluster, err
	}
	if net != nil {
		err = errors.New("network resources are not deleted yet")
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.NetworkKind, err.Error())
		wait = true
	}

	if cluster.Spec.IsManaged {
		return a.deleteOKE(cluster, net)
	}

	_, err = DeleteLoadBalancer(a.clientset, cluster.Namespace, cluster.Name)
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.LoadBalancerKind, err.Error())
		return cluster, err
	}

	for _, group := range K8S_GROUPS {

		compute, err := DeleteCompute(a.clientset, cluster.Namespace, cluster.Name+"-"+group)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.ComputeKind, err.Error())
			return cluster, err
		}
		if compute != nil {
			err = errors.New(group + " compute resource is not deleted yet")
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.ComputeKind, err.Error())
			wait = true
		}

		security, err := DeleteSecurity(a.clientset, cluster.Namespace, cluster.Name+"-"+group)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
			return cluster, err
		}
		if security != nil {
			err = errors.New(group + " security resource is not deleted yet")
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
			wait = true
		}
	}

	// Remove finalizers
	if !wait && len(cluster.Finalizers) > 0 {
		cluster.SetFinalizers([]string{})
		return cluster, nil
	}

	return cluster, nil
}

// update cluster cloud event handler
func (a *ClusterAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	cluster := obj.(*cloudv1alpha1.Cluster)
	glog.Infof("start cluster update...")
	resultObj, e := a.clientset.CloudV1alpha1().Clusters(cluster.Namespace).Update(cluster)
	return resultObj, e
}

// reconcile - handles create and updates
func (a *ClusterAdapter) reconcileOKE(cluster *cloudv1alpha1.Cluster) (runtime.Object, error) {

	reconcileState := cloudv1alpha1.OperatorStateCreated
	controllerRef := cloudcommon.CreateControllerRef(cluster, controllerKind)

	_, _, err := CreateOrUpdatePolicy(
		a.clientset,
		cluster,
		controllerRef)
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, idv1alpha1.PolicyKind, err.Error())
		return cluster, nil
	} else {
		cloudcommon.RemoveCondition(&cluster.Status.OperatorStatus, idv1alpha1.PolicyKind)
	}

	_, _, err = CreateOrUpdateSecurity(
		a.clientset,
		cluster,
		controllerRef,
		"lb")
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
		return cluster, nil
	}

	_, _, err = CreateOrUpdateSecurity(
		a.clientset,
		cluster,
		controllerRef,
		"node")
	if err != nil {
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
		return cluster, nil
	} else {
		cloudcommon.RemoveCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind)
	}

	// 1-time get from compartment and randomize
	var availabilityDomains []string
	if len(cluster.Status.AvailabilityZones) == 0 {

		availabilityDomains, err := cloudcommon.GetAvailabilityDomains(a.clientset, cluster.Namespace, cluster.Namespace)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, "AvailabilityDomain", err.Error())
			return cluster, nil
		}
		cluster.Status.AvailabilityZones = availabilityDomains
		glog.Infof("ad set: %v", availabilityDomains)

	} else {
		availabilityDomains = cluster.Status.AvailabilityZones
	}

	adCount := len(availabilityDomains)

	allSubnetsReady := true

	// service loadbalancer subnets
	subnetKey := cluster.Name + "-lb"
	subnetOffset, err := cloudcommon.GetSubnetOffset(a.clientset, cluster.Namespace, cluster.Name, subnetKey)
	if err != nil {
		glog.Errorf("error getting subnet offset: %v", err)
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, "GetSubnetOffset", err.Error())
		return cluster, nil
	}

	network, err := a.clientset.CloudV1alpha1().Networks(cluster.Namespace).Get(cluster.Name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("error getting network: %v", err)
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, "GetNetwork", err.Error())
		return cluster, nil
	}

	networkOctets := strings.Split(network.Spec.CidrBlock, ".")

	lbSubnetMap := make(map[string]string, adCount)

	securitySelector := map[string]string{cloudv1alpha1.ClusterKind: cluster.Name, "lb": "true"}

	// Process the subnet and instance resources
	for i, availabilityDomain := range availabilityDomains {
		// can only be 2, not 1 or 3, for oke
		if i > 1 {
			break
		}
		subnetOctet := subnetOffset + i + 1
		cidrBlock := networkOctets[0] + "." + networkOctets[1] + "." + strconv.Itoa(subnetOctet) + ".0/24"
		subnetName := strconv.Itoa(subnetOctet)
		lbSubnetMap[availabilityDomain] = subnetKey + "-" + subnetName
		subnet, _, err := cloudcompute.CreateOrUpdateSubnet(
			a.clientset,
			cluster.Namespace,
			cloudv1alpha1.ComputeKind,
			subnetKey,
			cluster.Name,
			controllerRef,
			subnetName,
			availabilityDomain,
			cidrBlock,
			&securitySelector)

		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			return cluster, nil
		}

		if subnet.Status.State != common.ResourceStateProcessed || !subnet.IsResource() {
			msg := "subnet: " + subnet.Name + " not ready"
			glog.Infof(msg)
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, v1alpha1.SubnetKind, msg)
			reconcileState = cloudv1alpha1.OperatorStatePending
			allSubnetsReady = false
		}
	}

	// node subnets
	nodeSubnetMap := make(map[string]string, adCount)
	subnetKey = cluster.Name + "-node"
	subnetOffset, err = cloudcommon.GetSubnetOffset(a.clientset, cluster.Namespace, cluster.Name, subnetKey)
	if err != nil {
		glog.Errorf("error getting subnet offset: %v", err)
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, "GetSubnetOffset", err.Error())
		return cluster, nil
	}

	securitySelector = map[string]string{cloudv1alpha1.ClusterKind: cluster.Name, "node": "true"}

	// Process the subnet and instance resources

	for i, availabilityDomain := range availabilityDomains {
		subnetOctet := subnetOffset + i + 1
		cidrBlock := networkOctets[0] + "." + networkOctets[1] + "." + strconv.Itoa(subnetOctet) + ".0/24"
		subnetName := strconv.Itoa(subnetOctet)
		nodeSubnetMap[availabilityDomain] = subnetKey + "-" + subnetName
		subnet, _, err := cloudcompute.CreateOrUpdateSubnet(
			a.clientset,
			cluster.Namespace,
			cloudv1alpha1.ComputeKind,
			subnetKey,
			cluster.Name,
			controllerRef,
			subnetName,
			availabilityDomain,
			cidrBlock,
			&securitySelector)

		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			return cluster, nil
		}

		if subnet.Status.State != common.ResourceStateProcessed || !subnet.IsResource() {
			glog.Infof("subnet: %s not ready", subnet.Name)
			err = errors.New("subnet resources are not processed yet")
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			reconcileState = cloudv1alpha1.OperatorStatePending
			allSubnetsReady = false
		}
	}

	lbSubnetRefs := []string{}
	i := 1
	for _, v := range lbSubnetMap {
		// has to be exactly 2 subnets for oke not to error
		if i < 3 {
			lbSubnetRefs = append(lbSubnetRefs, v)
		}
		i++
	}
	nodeSubnetRefs := []string{}
	for _, v := range nodeSubnetMap {
		nodeSubnetRefs = append(nodeSubnetRefs, v)
	}

	if allSubnetsReady {

		ceCluster, _, err := CreateOrUpdateCluster(
			a.clientset,
			cluster,
			controllerRef,
			&lbSubnetRefs,
			&cluster.Name,
		)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
			return cluster, nil

		} else {

			kubeConfig := ""
			if ceCluster.Status.KubeConfig != nil {
				kubeConfig = *ceCluster.Status.KubeConfig
			}

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: cluster.Name,
				},
				Data: map[string][]byte{"kubeconfig": []byte(kubeConfig)},
			}

			cm, err := a.kclientset.CoreV1().Secrets(cluster.Namespace).Get(cluster.Name, metav1.GetOptions{})
			if err != nil {
				glog.Infof("err getting secret: %v", err)
			}
			if apierrors.IsNotFound(err) {
				glog.Infof("create secret for kubeconfig")
				cm, err = a.kclientset.CoreV1().Secrets(cluster.Namespace).Create(secret)
				if err != nil {
					glog.Infof("err creating secret: %v", err)
				}

			} else {
				if string(secret.Data["kubeconfig"]) != string(cm.Data["kubeconfig"]) {
					glog.Infof("update secret: %v to %v", cm, secret)
					cm, err = a.kclientset.CoreV1().Secrets(cluster.Namespace).Update(secret)
				} else {
					glog.Infof("matching kubeconfig, no update")
				}

			}
		}

		_, _, err = CreateOrUpdateNodePool(
			a.clientset,
			cluster,
			controllerRef,
			&nodeSubnetRefs,
			&cluster.Name,
		)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
			return cluster, nil
		}

	}

	// Everything is done. Update the State, reset the Conditions and return
	if reconcileState == cloudv1alpha1.OperatorStateCreated {
		cluster.Status.State = cloudv1alpha1.OperatorStateCreated
		cluster.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
	} else {
		cluster.Status.State = cloudv1alpha1.OperatorStatePending
	}
	glog.Infof("thru oke reconcile")

	return cluster, nil
}

// reconcile - handles create and updates
func (a *ClusterAdapter) Reconcile(obj runtime.Object) (runtime.Object, error) {
	cluster := obj.(*cloudv1alpha1.Cluster)

	controllerRef := cloudcommon.CreateControllerRef(cluster, controllerKind)
	reconcileState := cloudv1alpha1.OperatorStateCreated

	_, created, err := CreateOrUpdateNetwork(
		a.clientset,
		cluster,
		controllerRef)
	if err != nil {
		glog.Infof("CreateOrUpdateNetwork err: %v", err)
		cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.NetworkKind, err.Error())
		return cluster, nil
	}

	if cluster.Spec.IsManaged {
		return a.reconcileOKE(cluster)
	}

	lbIp := ""
	for _, group := range K8S_GROUPS {

		_, created, err = CreateOrUpdateSecurity(
			a.clientset,
			cluster,
			controllerRef,
			group)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.SecurityKind, err.Error())
			return cluster, nil
		}

		if group == MASTER_KIND {

			lb, created, err := CreateOrUpdateLoadBalancer(
				a.clientset,
				cluster,
				controllerRef)
			if err != nil {
				cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.LoadBalancerKind, err.Error())
				return cluster, nil
			}
			if created || lb.Status.IPAddress == "" {
				cluster.Status.State = cloudv1alpha1.OperatorStatePending
				cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.LoadBalancerKind, "waiting for lb ip address")
				return cluster, nil
			}

			lbIp = lb.Status.IPAddress
		}

		_, created, err = CreateOrUpdateCompute(
			a.clientset,
			cluster,
			controllerRef,
			group,
			lbIp)
		if err != nil {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.ComputeKind, err.Error())
			return cluster, nil
		}
		if created {
			cloudcommon.SetCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.ComputeKind, "pending compute")
			reconcileState = cloudv1alpha1.OperatorStatePending
		} else {
			cloudcommon.RemoveCondition(&cluster.Status.OperatorStatus, cloudv1alpha1.ComputeKind)
		}

	}

	cpb, _ := pem.Decode([]byte(cluster.Spec.CA.Certificate))
	caCert, e := x509.ParseCertificate(cpb.Bytes)
	if e != nil {
		glog.Errorf("parse cert err: %v", e)
	}

	kpb, _ := pem.Decode([]byte(cluster.Spec.CA.Key))
	caKey, e := x509.ParsePKCS1PrivateKey(kpb.Bytes)
	if e != nil {
		glog.Errorf("parse key err: %v", e)
	}

	// otherwise, create a client certs
	clientCertConfig := certutil.Config{
		CommonName:   ADMIN_USER,
		Organization: []string{"admin"},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCert, clientKey, err := pkiutil.NewCertAndKey(caCert, caKey, clientCertConfig)
	if err != nil {
		glog.Errorf("failure while creating %s client certificate: %v", ADMIN_USER, err)
		return nil, nil
	}

	// create a kubeconfig with the client certs
	config := kubeconfigutil.CreateWithCerts(
		"https://"+lbIp,
		cluster.Name,
		ADMIN_USER,
		[]byte(cluster.Spec.CA.Certificate),
		certutil.EncodePrivateKeyPEM(clientKey),
		certutil.EncodeCertPEM(clientCert))

	// simplify context name so federation-v2 join doesn't fail with: Invalid value: "cluster-user@cluster": a DNS-1123 subdomain ...
	dirtyContext := config.CurrentContext
	cleanContext := cluster.Name
	config.Contexts[cleanContext] = config.Contexts[dirtyContext]
	config.CurrentContext = cleanContext
	delete(config.Contexts, dirtyContext)

	configBytes, err := clientcmd.Write(*config)
	if err != nil {
		return nil, fmt.Errorf("failure while clientcmd.Write(*config): %v", err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster.Name,
		},
		Data: map[string][]byte{"kubeconfig": configBytes},
	}

	cm, err := a.kclientset.CoreV1().Secrets(cluster.Namespace).Get(cluster.Name, metav1.GetOptions{})
	if err != nil {
		glog.Infof("err getting configmap: %v", err)
	}
	if apierrors.IsNotFound(err) {
		glog.Infof("create secret for kubeconfig")
		cm, err = a.kclientset.CoreV1().Secrets(cluster.Namespace).Create(secret)
		if err != nil {
			glog.Infof("err creating configmap: %v", err)
		}

	} else {
		if string(secret.Data["kubeconfig"]) != string(cm.Data["kubeconfig"]) {
			glog.Infof("update secret: %s", cm.Name)
			cm, err = a.kclientset.CoreV1().Secrets(cluster.Namespace).Update(secret)
		} else {
			glog.Infof("matching kubeconfig, no update")
		}

	}

	// Everything is done. Update the State, reset the Conditions and return
	if reconcileState == cloudv1alpha1.OperatorStateCreated {
		cluster.Status.State = cloudv1alpha1.OperatorStateCreated
		if cluster.Status.Conditions != nil {
			cluster.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
		}
	} else {
		cluster.Status.State = reconcileState
	}
	glog.Infof("thru reconcile")
	return cluster, nil

}

// callback for resource
func (a *ClusterAdapter) CallbackForResource(resource schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			a.ignoreAddEvent(obj)
		},
		UpdateFunc: func(old, obj interface{}) {
			a.processUpdateOrDeleteEvent(obj)
		},
		DeleteFunc: func(obj interface{}) {
			a.processUpdateOrDeleteEvent(obj)
		},
	}

	return handlers
}

// resolve controller reference
func (a *ClusterAdapter) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *cloudv1alpha1.Cluster {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	obj, err := a.lister.ByNamespace(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	cluster := obj.(*cloudv1alpha1.Cluster)

	if cluster.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return cluster
}

// ignore add event
func (a *ClusterAdapter) ignoreAddEvent(obj interface{}) {
	glog.V(4).Infof("Got add event: %v", obj)
}

// handle update or delete events
func (a *ClusterAdapter) processUpdateOrDeleteEvent(obj interface{}) {
	object := obj.(metav1.Object)
	if controllerRef := cloudcommon.GetControllerOf(object); controllerRef != nil {
		cluster := a.resolveControllerRef(object.GetNamespace(), controllerRef)
		if cluster == nil {
			return
		}
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cluster)
		if err != nil {
			glog.Errorf("Cluster deletion state error %v", err)
			return
		}
		glog.V(4).Infof("Cluster %s received update event for %s %s\n", key, reflect.TypeOf(object).String(), object.GetName())
		a.queue.Add(key)
		//a.Reconcile(cluster)
		return
	}
}
