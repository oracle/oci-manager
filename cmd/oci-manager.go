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

package main

import (
	"context"
	"flag"
	"os"
	"reflect"
	"time"

	"github.com/golang/glog"
	"golang.org/x/time/rate"

	"k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ociauth "github.com/oracle/oci-go-sdk/common/auth"
	ociidentity "github.com/oracle/oci-go-sdk/identity"

	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	informers "github.com/oracle/oci-manager/pkg/client/informers/externalversions"

	"github.com/oracle/oci-manager/cmd/util"
	cloudcontroller "github.com/oracle/oci-manager/pkg/controller/oci/cloud"
	cloudcommon "github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"

	kubecontroller "github.com/oracle/oci-manager/pkg/controller/oci/kubernetes"
	kubecommon "github.com/oracle/oci-manager/pkg/controller/oci/kubernetes/common"

	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/cluster"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/compute"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/cpod"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/loadbalancer"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/network"
	"github.com/oracle/oci-manager/pkg/controller/oci/cloud/security"

	kubecore "github.com/oracle/oci-manager/pkg/controller/oci/kubernetes/core"

	"github.com/oracle/oci-manager/pkg/controller/oci/resources"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	"github.com/oracle/oci-manager/pkg/controller/oci/resources/ce"
	"github.com/oracle/oci-manager/pkg/controller/oci/resources/core"
	"github.com/oracle/oci-manager/pkg/controller/oci/resources/db"
	"github.com/oracle/oci-manager/pkg/controller/oci/resources/identity"
	"github.com/oracle/oci-manager/pkg/controller/oci/resources/lb"
)

var (
	Version      string
	kubeconfig   string
	ociconfig    string
	ipr          bool = false
	disableCloud bool = false
	resyncperiod int  = 60
	kubeclient   kubernetes.Interface
)

const (
	EnvPodNamespace = "OCIM_POD_NAMESPACE"
)

var registerdAdapters = []string{
	core.OciDomain, identity.OciDomain, lb.OciDomain, ce.OciDomain, db.OciDomain,
	cluster.CloudDomain,
	compute.CloudDomain,
	cpod.CloudDomain,
	loadbalancer.CloudDomain,
	network.CloudDomain,
	security.CloudDomain,
	kubecore.KubernetesDomain,
}

func main() {

	namespace := os.Getenv(EnvPodNamespace)

	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "kubeconfig file")
	flag.StringVar(&ociconfig, "ociconfig", ociconfig, "ociconfig file")
	flag.IntVar(&resyncperiod, "resync-seconds", resyncperiod, "full sync period in seconds")
	flag.BoolVar(&ipr, "ipr", false, "use instance principals")
	flag.BoolVar(&disableCloud, "disable-cloud", false, "disable cloud-abstraction controllers")

	flag.Set("logtostderr", "true")
	flag.Parse()

	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	if ociconfig == "" {
		ociconfig = os.Getenv("OCICONFIG")
	}

	if namespace == "" {
		namespace = "oci-system"
	}

	config := getKubeConfig()
	var err error
	kubeclient, err = kubernetes.NewForConfig(config)

	host, err := os.Hostname()
	glog.Infof("Got host %s", host)
	id := "oci-manager-" + host

	rl, err := resourcelock.New(resourcelock.ConfigMapsResourceLock,
		namespace,
		"oci-manager",
		kubeclient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: createRecorder(kubeclient, "oci-manager", namespace),
		})
	if err != nil {
		glog.Fatalf("error creating lock: %v", err)
	}

	leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 45 * time.Second,
		RenewDeadline: 30 * time.Second,
		RetryPeriod:   10 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				glog.Fatalf("leader election lost")
			},
		},
	})

	panic("unreachable")
}

func run(stopCh <-chan struct{}) {

	var (
		ocicfg ocisdkcommon.ConfigurationProvider
		err    error
	)

	//create CRD definitions
	config := getKubeConfig()
	createCrdDefinitions(config)

	// Create the oci resource client config using required ociconfig file.

	if ipr {
		ocicfg, err = ociauth.InstancePrincipalConfigurationProvider()
		if err != nil {
			glog.Errorf("Error creating ipr client: %v", err)
			os.Exit(1)
		}
		glog.Infof("Using instance principal client")

	} else if ociconfig != "" {
		ocicfg, err = ocisdkcommon.ConfigurationProviderFromFile(ociconfig, "")
		if err != nil {
			glog.Errorf("Error creating oci client from file: %v", err)
			os.Exit(1)
		}
		glog.Infof("Using oci client from config file %s", ociconfig)

	} else {
		ocicfg = ocisdkcommon.DefaultConfigProvider()
		glog.Infof("Using default client")
	}

	//check client config with compartment list
	checkCompartments(ocicfg)

	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	informersFactory := informers.NewSharedInformerFactory(clientset, time.Duration(resyncperiod)*time.Second)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeclient, time.Duration(resyncperiod)*time.Second)

	// Start cloud controllers first
	if !disableCloud {
		startCloudControllers(clientset, kubeclient, informersFactory, stopCh)
	}

	// Start resource controllers
	startResourceControllers(clientset, kubeclient, ocicfg, informersFactory, stopCh)

	// Start resource controllers
	startKubernetesControllers(clientset, kubeclient, kubeInformerFactory, stopCh)

	// Wait forever
	select {}
}

func startCloudControllers(clientSet clientset.Interface, kubeclient kubernetes.Interface, resourceIFactory informers.SharedInformerFactory, stopChan <-chan struct{}) {

	workQueues := make(map[string]workqueue.RateLimitingInterface)
	cloudInformersFactory := informers.NewSharedInformerFactory(clientSet, time.Duration(resyncperiod)*time.Second)
	for kind, cloudType := range cloudcommon.CloudTypes() {
		glog.Infof("Starting cloud controller for %s\n", kind)
		rateLimiter := workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(2*time.Second, 1000*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(float64(10)), 100)},
		)
		workQueues[kind] = workqueue.NewRateLimitingQueue(rateLimiter)

		controller := cloudcontroller.New(cloudType.AdapterFactory, clientSet, kubeclient, cloudInformersFactory, resourceIFactory, workQueues)
		controller.Run(stopChan)
		time.Sleep(3 * time.Second)
	}

}

func startResourceControllers(clientset clientset.Interface, kubeclient kubernetes.Interface, ocicfg ocisdkcommon.ConfigurationProvider, informersFactory informers.SharedInformerFactory, stopCh <-chan struct{}) {

	adapterSpecificArgs := make(map[string]interface{})
	workQueues := make(map[string]workqueue.RateLimitingInterface)
	controllers := make(map[string]*resources.Controller)
	for kind, ocitype := range resourcescommon.ResourceTypes() {
		glog.Infof("Starting resource controller for %s/%s\n", ocitype.GroupName, kind)
		rateLimiter := workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(2*time.Second, 1000*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(float64(10)), 100)},
		)
		workQueues[kind] = workqueue.NewRateLimitingQueue(rateLimiter)

		controllers[kind] = resources.Start(clientset, kubeclient, ocicfg, informersFactory, stopCh, ocitype.AdapterFactory, adapterSpecificArgs, workQueues)
		time.Sleep(5 * time.Second)
	}

}

func startKubernetesControllers(clientset clientset.Interface, kubeclient kubernetes.Interface, informersFactory kubeinformers.SharedInformerFactory, stopCh <-chan struct{}) {

	adapterSpecificArgs := make(map[string]interface{})
	workQueues := make(map[string]workqueue.RateLimitingInterface)
	controllers := make(map[string]*kubecontroller.Controller)
	for key, kubetype := range kubecommon.KubernetesTypes() {
		glog.Infof("Starting kubernetes controller for %s\n", key)
		rateLimiter := workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(2*time.Second, 1000*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(float64(10)), 100)},
		)
		objectType := reflect.TypeOf(kubetype.Type).String()
		workQueues[objectType] = workqueue.NewRateLimitingQueue(rateLimiter)

		controllers[key] = kubecontroller.Start(clientset, kubeclient, kubetype.Type, kubetype.AdapterFactory, adapterSpecificArgs, informersFactory, stopCh, workQueues)
		time.Sleep(5 * time.Second)
	}

}

func getKubeConfig() (config *rest.Config) {
	var err error
	// Create the kube object client config. Use kubeconfig if given, otherwise assume in-cluster.
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Errorf("Error creating kube client config, please specify valid --kubeconfig flag or KUBECONFIG env variable: %v", err)
		os.Exit(1)
	}
	return
}

func createCrdDefinitions(config *rest.Config) {
	kubeclient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	// Create resource definitions
	for _, ocitype := range resourcescommon.ResourceTypes() {
		_, err = util.CreateResourceDefinition(kubeclient, ocitype.ResourcePlural, ocitype.Kind, ocitype.GroupName, ocitype.Validation)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			panic(err)
		}
	}

	for _, cloudtype := range cloudcommon.CloudTypes() {
		_, err = util.CreateResourceDefinition(kubeclient, cloudtype.ResourcePlural, cloudtype.Kind, cloudtype.GroupName, cloudtype.Validation)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			panic(err)
		}
	}

}

func checkCompartments(ocicfg ocisdkcommon.ConfigurationProvider) error {

	client, err := ociidentity.NewIdentityClientWithConfigurationProvider(ocicfg)

	if err != nil {
		glog.Errorf("Error creating oci client: %v", err)
		os.Exit(1)
	}

	tenancyID, err := ocicfg.TenancyOCID()
	_, err = client.ListCompartments(context.Background(), ociidentity.ListCompartmentsRequest{CompartmentId: &tenancyID})

	if err != nil {
		glog.Errorf("Error querying compartment: %v", err)
		os.Exit(1)
	}

	return nil
}

func createRecorder(kubecli kubernetes.Interface, name, namespace string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	//eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubecli.Core().RESTClient()).Events(namespace)})
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: name})
}
