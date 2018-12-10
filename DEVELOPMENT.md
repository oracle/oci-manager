# Building

To test and build oci-manager run:
```bash
make
```

To run the oci-manager from your local development environment run:

```bash
export KUBECONFIG=~/.kube/config
export OCICONFIG=~/.oci/config
make run
```

To deploy and run oci-manager in your Kubernetes development cluster run:
```bash
export OCICONFIG=~/.oci/config
export DOCKER_REGISTRY=k8sfed/oci-manager
make deploy
```

To create a docker image run:
```bash
make image
```


# Extending

The oci-manager can be extended and versioned. We reuse some kubernetes code-gen to simplify.  
In this doc I will cover how to add a group and type for new set of resources along with the 
kubernetes controllers (aka adapters).

## Create a Group 

A group is a domain or set of types that map to a client in the [oci sdk](https://github.com/oracle/oci-go-sdk).  
In this document I will use the [database](https://github.com/oracle/oci-go-sdk/tree/master/database) group.

Steps:

1: Create apis dir with v1alpha1 version subdir by:
```bash
cd pkg/apis
mkdir ocidb.oracle.com/v1alpha1
cd ocidb.oracle.com
```

2: Create a `register.go` in top level of dir:
```go
package ocidb_oracle_com

const (
	// GroupName used for all resources in this package
	GroupName = "ocidb.oracle.com"
)
```

3: Create doc.go for the kubernetes code-gen:
```go
// +k8s:deepcopy-gen=package,register

// Package v1alpha1 is the v1alpha1 version of the API.
// +groupName=ocidb.oracle.com
package v1alpha1
```

4: Update hack/update-codegen.sh then run (from project root):
```bash
hack/update-codegen.sh
``` 

5: Create Controller (Adapter) resource group:

```bash
cd pkg/controller/oci/resource
mkdir db
cd db
```

6: Add register.go containing something like:
```go
package db

// OciDomain is unique domain string for all resources in db package
const (
	OciDomain = "db"
)
```

7: Register in oci-manger.go cmd

I am adding this step to the Create Group sequence because I will only need to be done once for the group, 
but nothing will be registered until a new Type and Adapter is adding in the following section of this document.

```go
// include
"github.com/oracle/oci-manager/pkg/controller/oci/resources/db"

// few lines down (note the db.OciDomain)
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

```


### Create a Type and Adapter

Most of the time one would just need to add a type to an existing group.  In this example will use the OCI AutonomousDatabase.

Steps:

1: Create a new `autonomous_database`_types.go in:

```bash
cd pkg/apis/ocidb.oracle.com/v1alpha1
```

This types file is used to code-gen kubernetes informers and clientset.
In the future we plan on automating this with some templating and code-gen of our own, 
but for now it would be simplest to copy an existing types file.

The Spec is used as the desired state. For example here is the AutonomousDatabaseSpec that is used for crud operations against OCI.

```go
// AutonomousDatabaseSpec describes a AutonomousDatabase spec
type AutonomousDatabaseSpec struct {
	CompartmentRef string `json:"compartmentRef"`

	// The number of CPU cores to be made available to the database.
	CpuCoreCount *int `mandatory:"true" json:"cpuCoreCount"`

	// The quantity of data in the database, in terabytes.
	DataStorageSizeInTBs *int `mandatory:"true" json:"dataStorageSizeInTBs"`

	// The user-friendly name for the Autonomous Database. The name does not have to be unique.
	DisplayName string `mandatory:"false" json:"displayName"`

	// The Oracle license model that applies to the Oracle Autonomous Database. The default is BRING_YOUR_OWN_LICENSE.
	LicenseModel ocidb.AutonomousDatabaseLicenseModelEnum `mandatory:"false" json:"licenseModel,omitempty"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	common.Dependency
}
```

The Status is used to model the actual state. It places an oci sdk typed value in the `resource`.

```go
// AutonomousDatabaseStatus describes a AutonomousDatabase status
type AutonomousDatabaseStatus struct {
	common.ResourceStatus

	Resource *AutonomousDatabaseResource `json:"resource,omitempty"`
}

// AutonomousDatabaseResource describes a AutonomousDatabase resource from oci
type AutonomousDatabaseResource struct {
	*ocidb.AutonomousDatabase
}
```
The *ocidb above references the oci sdk from import section:
```go
ocidb "github.com/oracle/oci-go-sdk/database"
```

2: Register the type and its list type in the register.go
```go
// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&AutonomousDatabase{},
		&AutonomousDatabaseList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
```

3: Create OCI client interface

Grep func from sdk’s client - grep func vendor/github.com/oracle/oci-go-sdk/database/database_client.go
… put public ones into pkg/controller/oci/resource/common/ociclient.go

4: Create the fake OCI Client

For unit tests, implement fake api for the corresponding client in pkg/controller/oci/resources/common/ociclient.go.

Example fake clients (and where they should go) can be found in pkg/controller/oci/resources/fake

5: Create an Adapter (aka Resource Controller)

A Resource Adapter implements a ResourceTypeAdapter defined in (pkg/controller/oci/resource/common/adapter.go):
```go
type ResourceTypeAdapter interface {
	Kind() string
	Resource() string
	GroupVersionWithResource() schema.GroupVersionResource
	ObjectType() runtime.Object
	IsExpectedType(obj interface{}) bool
	Copy(obj runtime.Object) runtime.Object
	Equivalent(obj1, obj2 runtime.Object) bool
	Id(obj runtime.Object) string
	ObjectMeta(obj runtime.Object) *metav1.ObjectMeta
	DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn
	Dependents(obj runtime.Object) map[string][]string
	DependsOnRefs(obj runtime.Object) ([]runtime.Object, error)

	// Operations target the resource service apis
	Create(obj runtime.Object) (runtime.Object, error)
	Delete(obj runtime.Object) (runtime.Object, error)
	Get(obj runtime.Object) (runtime.Object, error)
	Update(obj runtime.Object) (runtime.Object, error)

	// Operations target CRDs
	CreateObject(obj runtime.Object) (runtime.Object, error)
	UpdateObject(obj runtime.Object) (runtime.Object, error)
	DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error
	UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error)
}
```
Until code-gen is added, copy an existing resource controller (and test) and replace with your type and logic.


# Build, Test and Run

make will run fmt and test. You can run within your IDE's debugger if you set the KUBECONFIG and OCICONFIG env var.  
Or just recreate your statefulset using the yaml you used in [setup.md](../docs/setup.md) with the image you created 
from steps at the top of this document.