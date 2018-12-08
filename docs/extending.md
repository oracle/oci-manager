# Extending oci-manager Overview

The oci-manager can be extended and versioned. We reuse some kubernetes code-gen to simplify.  
In this doc I will cover how to add a group and type for new set of resources along with the 
kubernetes controllers (aka adapters).

## Create a Group and Types

### Group

Create apis dir with v1alpha1 version subdir by:
```bash
cd pkg/apis
mkdir somedomain.oracle.com/v1alpha1
cd somedomain.oracle.com
```

create a `register.go` in top level of dir:
```go
package somedomain_oracle_com

const (
	// GroupName used for all resources in this package
	GroupName = "somedomain.oracle.com"
)
```

### Type

Next create new `resource`_types.go in the v1alpha1 subdir:

```bash
cd v1alpha1
```

This types file is used to code-gen kubernetes informers and clientset.
In the future we plan on automating this with some templating and code-gen of our own, 
but for now it would be simplest to copy an existing types file like: 
pkg/apis/ocidb.oracle.com/autonomous_database_types.go

The Spec struct is used as the desired state. For example here is the AutonomousDatabaseSpec that is used for crud operations against OCI.

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

The Status struct is used to model the actual state. I places an oci sdk typed value in the `resource`.

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

Create some more files for the kubernetes code-gen
- docs.go
- register.go adding your new type and list type

Update hack/update-codegen.sh then run (from project root):
```bash
hack/update-codegen.sh
``` 

# Create a Controller (aka Adapter)

```bash
cd pkg/controller/oci/resource
mkdir somedomain
cd somedomain
```

add register.go containing something like:
```go
package somedomain

// OciDomain is unique domain string for all resources in db package
const (
	OciDomain = "somedomain"
)
```

## Create client interface
Grep func from sdk’s client - grep func vendor/github.com/oracle/oci-go-sdk/foo/bar_client.go
… put public ones into common/ociclient.go

## Create Resource Controller

A Resource controller implements a ResourceTypeAdapter defined in (pkg/controller/oci/resource/common/adapter.go):
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


# Register in oci-manger.go
In this example will add db resource adapters:
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

# Fake OCI Client

Add a fake client if the sdk has a separate client for it.

Implement fake api for the corresponding client in pkg/controller/oci/resources/common/ociclient.go.

Example fake clients (and where they should go) can be found in pkg/controller/oci/resources/fake

# Build, Test and Run

make will run fmt and test. You can run within your IDE's debugger if you set the KUBECONFIG and OCICONFIG env var.  Or just recreate your statefulset using the yaml you used in [setup.md](setup.md)