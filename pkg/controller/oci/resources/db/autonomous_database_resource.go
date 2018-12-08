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

package db

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	dbgroup "github.com/oracle/oci-manager/pkg/apis/ocidb.oracle.com"
	ocidbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocidb.oracle.com/v1alpha1"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocidb "github.com/oracle/oci-go-sdk/database"

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		dbgroup.GroupName,
		ocidbv1alpha1.AutonomousDatabaseKind,
		ocidbv1alpha1.AutonomousDatabaseResourcePlural,
		ocidbv1alpha1.AutonomousDatabaseControllerName,
		&ocidbv1alpha1.AutonomousDatabaseValidation,
		NewAutonomousDatabaseAdapter)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// AutonomousDatabaseAdapter implements the adapter interface for autonomousdatabase resource
type AutonomousDatabaseAdapter struct {
	clientset  versioned.Interface
	kubeclient kubernetes.Interface
	ctx        context.Context
	dbClient   resourcescommon.DatabaseClientInterface
	seededRand *rand.Rand
}

// NewAutonomousDatabaseAdapter creates a new adapter for autonomousdatabase resource
func NewAutonomousDatabaseAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	ada := AutonomousDatabaseAdapter{
		clientset:  clientset,
		kubeclient: kubeclient,
		ctx:        context.Background(),
		seededRand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	dbClient, err := ocidb.NewDatabaseClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci db client: %v", err)
		os.Exit(1)
	}
	ada.dbClient = &dbClient
	return &ada
}

// Kind returns the resource kind string
func (a *AutonomousDatabaseAdapter) Kind() string {
	return ocidbv1alpha1.AutonomousDatabaseKind
}

// Resource returns the plural name of the resource type
func (a *AutonomousDatabaseAdapter) Resource() string {
	return ocidbv1alpha1.AutonomousDatabaseResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *AutonomousDatabaseAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocidbv1alpha1.SchemeGroupVersion.WithResource(ocidbv1alpha1.AutonomousDatabaseResourcePlural)
}

// ObjectType returns the autonomousdatabase type for this adapter
func (a *AutonomousDatabaseAdapter) ObjectType() runtime.Object {
	return &ocidbv1alpha1.AutonomousDatabase{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *AutonomousDatabaseAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocidbv1alpha1.AutonomousDatabase)
	return ok
}

// Copy returns a copy of a autonomousdatabase object
func (a *AutonomousDatabaseAdapter) Copy(obj runtime.Object) runtime.Object {
	AutonomousDatabase := obj.(*ocidbv1alpha1.AutonomousDatabase)
	return AutonomousDatabase.DeepCopyObject()
}

// Equivalent checks if two autonomousdatabase objects are the same
func (a *AutonomousDatabaseAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	return true
}

// IsResourceCompliant
func (a *AutonomousDatabaseAdapter) IsResourceCompliant(obj runtime.Object) bool {
	adb := obj.(*ocidbv1alpha1.AutonomousDatabase)
	if adb.Status.Resource == nil {
		return false
	}

	resource := adb.Status.Resource

	if resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateScaleInProgress ||
		resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateBackupInProgress ||
		resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateRestoreInProgress ||
		resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateProvisioning ||
		resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateStopping {
		return true
	}

	if resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateStopped ||
		resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateUnavailable ||
		resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateRestoreFailed ||
		resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateTerminated {
		return false
	}

	specDisplayName := resourcescommon.Display(adb.Name, adb.Spec.DisplayName)

	if *adb.Status.Resource.CpuCoreCount != *adb.Spec.CpuCoreCount ||
		*adb.Status.Resource.DisplayName != *specDisplayName ||
		*adb.Status.Resource.DataStorageSizeInTBs != *adb.Spec.DataStorageSizeInTBs {
		return false
	}
	return true
}

// IsResourceStatusChanged checks if two autonomousdatabase objects are the same
func (a *AutonomousDatabaseAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	ad1 := obj1.(*ocidbv1alpha1.AutonomousDatabase)
	ad2 := obj2.(*ocidbv1alpha1.AutonomousDatabase)

	if ad1.Status.Resource.LifecycleState != ad2.Status.Resource.LifecycleState {
		return true
	}

	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *AutonomousDatabaseAdapter) Id(obj runtime.Object) string {
	return obj.(*ocidbv1alpha1.AutonomousDatabase).GetResourceID()
}

// ObjectMeta returns the object meta struct from the autonomousdatabase object
func (a *AutonomousDatabaseAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocidbv1alpha1.AutonomousDatabase).ObjectMeta
}

// DependsOn returns a map of autonomousdatabase dependencies (objects that the autonomousdatabase depends on)
func (a *AutonomousDatabaseAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocidbv1alpha1.AutonomousDatabase).Spec.DependsOn
}

// Dependents returns a map of autonomousdatabase dependents (objects that depend on the autonomousdatabase)
func (a *AutonomousDatabaseAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocidbv1alpha1.AutonomousDatabase).Status.Dependents
}

// CreateObject creates the autonomousdatabase object
func (a *AutonomousDatabaseAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocidbv1alpha1.AutonomousDatabase)
	return a.clientset.OcidbV1alpha1().AutonomousDatabases(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the autonomousdatabase object
func (a *AutonomousDatabaseAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocidbv1alpha1.AutonomousDatabase)
	return a.clientset.OcidbV1alpha1().AutonomousDatabases(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the autonomousdatabase object
func (a *AutonomousDatabaseAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocidbv1alpha1.AutonomousDatabase)
	return a.clientset.OcidbV1alpha1().AutonomousDatabases(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the autonomousdatabase depends on
func (a *AutonomousDatabaseAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var object = obj.(*ocidbv1alpha1.AutonomousDatabase)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	return deps, nil
}

// get limited set of random char from charset const
func (a *AutonomousDatabaseAdapter) RandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[a.seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (a *AutonomousDatabaseAdapter) RandomString(length int) string {
	return a.RandomStringWithCharset(length, charset)
}

// Create creates the autonomousdatabase resource in oci
func (a *AutonomousDatabaseAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		db            = obj.(*ocidbv1alpha1.AutonomousDatabase)
		compartmentId string
		err           error
	)

	if resourcescommon.IsOcid(db.Spec.CompartmentRef) {
		compartmentId = db.Spec.CompartmentRef
	} else {
		compartmentId, err = resourcescommon.CompartmentId(a.clientset, db.ObjectMeta.Namespace, db.Spec.CompartmentRef)
		if err != nil {
			return db, db.Status.HandleError(err)
		}
	}

	existingSecret, err := a.kubeclient.CoreV1().Secrets(db.Namespace).Get(db.Name, metav1.GetOptions{})
	if err != nil {
		glog.V(4).Infof("err getting secret: %v", err)
	}
	var adminPassword string
	if apierrors.IsNotFound(err) {
		glog.Infof("create secret for admin password")
		// Password must be 12 to 30 characters and contain at least one uppercase letter, one lowercase letter, and one number.
		// The password cannot contain the double quote (") character or the username "admin"
		adminPassword = a.RandomString(30)
		secretData := make(map[string][]byte, 0)
		secretData["password"] = []byte(adminPassword)
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: db.Name,
			},
			Data: secretData,
		}

		existingSecret, err = a.kubeclient.CoreV1().Secrets(db.Namespace).Create(secret)
		if err != nil {
			glog.Infof("err creating secret: %v", err)
			return db, db.Status.HandleError(err)
		}
	} else {
		adminPassword = string(existingSecret.Data["password"])
	}

	// create a new AutonomousDatabase
	request := ocidb.CreateAutonomousDatabaseRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.AdminPassword = &adminPassword
	request.CpuCoreCount = db.Spec.CpuCoreCount
	request.DataStorageSizeInTBs = db.Spec.DataStorageSizeInTBs
	request.DbName = &db.Name
	request.DisplayName = resourcescommon.Display(db.Name, db.Spec.DisplayName)
	request.OpcRetryToken = ocisdkcommon.String(string(db.UID))
	glog.Infof("AutonomousDatabase: %s OpcRetryToken: %s", db.Name, *request.OpcRetryToken)

	r, err := a.dbClient.CreateAutonomousDatabase(a.ctx, request)
	if err != nil {
		return db, db.Status.HandleError(err)
	}

	return db.SetResource(&r.AutonomousDatabase), db.Status.HandleError(err)
}

// Unzip will decompress a zip archive, moving all files and folders
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}
		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)

		} else {
			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}
			_, err = io.Copy(outFile, rc)
			// Close the file without defer to close before next iteration of loop
			outFile.Close()
			if err != nil {
				return filenames, err
			}
		}
	}
	return filenames, nil
}

// use db name to fetch wallet zip and put into secret
func (a *AutonomousDatabaseAdapter) FetchWalletToSecret(db *ocidbv1alpha1.AutonomousDatabase) error {

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: db.Name + "-wallet",
		},
	}

	_, err := a.kubeclient.CoreV1().Secrets(db.Namespace).Get(secret.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {

		baseDest := "/tmp/" + db.Name
		zipFile := baseDest + ".zip"
		passwordSecret, err := a.kubeclient.CoreV1().Secrets(db.Namespace).Get(db.Name, metav1.GetOptions{})
		adminPassword := string(passwordSecret.Data["password"])

		walletRequest := ocidb.GenerateAutonomousDatabaseWalletRequest{
			AutonomousDatabaseId: db.Status.Resource.Id,
			GenerateAutonomousDatabaseWalletDetails: ocidb.GenerateAutonomousDatabaseWalletDetails{
				Password: &adminPassword,
			},
		}
		walletResponse, err := a.dbClient.GenerateAutonomousDatabaseWallet(a.ctx, walletRequest)
		if err != nil {
			glog.Errorf("generate wallet err: %v", err)
			return err
		}
		if walletResponse.Content != nil {
			reader := walletResponse.Content
			defer reader.Close()

			walletbytes, err := ioutil.ReadAll(reader)
			if err != nil {
				glog.Errorf("ioutil wallet read err: %v", err)
				return err
			}

			if err := ioutil.WriteFile(zipFile, walletbytes, 07440); err != nil {
				glog.Errorf("ioutil wallet write err: %v", err)
				return err
			}
			files, err := Unzip(zipFile, baseDest)
			if err != nil {
				glog.Errorf("unzip error: %v", err)
				return err
			}
			glog.V(5).Infof("unzipped: " + strings.Join(files, " "))

			walletData := make(map[string][]byte, 0)
			for _, file := range files {
				simpleFile := strings.Replace(file, baseDest+"/", "", 1)
				fileContent, err := ioutil.ReadFile(file)
				if err != nil {
					glog.Errorf("could not read file: %s err: %v", file, err)
					return err
				}
				walletData[simpleFile] = fileContent
			}

			secret.Data = walletData

			_, err = a.kubeclient.CoreV1().Secrets(db.Namespace).Create(secret)
			if err != nil {
				glog.Errorf("could not create wallet secret err: %v", err)
				return err
			}
			glog.Infof("created secret: %s", secret.Name)

			err = rmDir(baseDest)
			if err != nil {
				glog.Errorf("could not rm dir: %s err: %v", baseDest, err)
				return err
			}

			err = os.Remove(zipFile)
			if err != nil {
				glog.Errorf("could not rm file: %s err: %v", zipFile, err)
				return err
			}

		}
	}
	return nil
}

// remove directory
func rmDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete deletes the autonomousdatabase resource in oci
func (a *AutonomousDatabaseAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	db := obj.(*ocidbv1alpha1.AutonomousDatabase)

	err := a.kubeclient.CoreV1().Secrets(db.Namespace).Delete(db.Name, &metav1.DeleteOptions{})
	if err != nil {
		glog.Errorf("error deleting secret: %s err: %v", db.Name, err)
	}
	err = a.kubeclient.CoreV1().Secrets(db.Namespace).Delete(db.Name+"-wallet", &metav1.DeleteOptions{})
	if err != nil {
		glog.Errorf("error deleting secret: %s err: %v", db.Name+"-wallet", err)
	}

	baseDest := "/tmp/" + db.Name
	err = os.RemoveAll(baseDest)
	if err != nil {
		glog.Errorf("error deleting wallet directory: %s err: %v", db.Name, err)
	}
	err = os.Remove(baseDest + ".zip")
	if err != nil {
		glog.Errorf("error deleting wallet zip file: %s err: %v", db.Name, err)
	}

	request := ocidb.DeleteAutonomousDatabaseRequest{
		AutonomousDatabaseId: db.Status.Resource.Id,
	}
	_, e := a.dbClient.DeleteAutonomousDatabase(a.ctx, request)
	if e == nil && db.Status.Resource != nil {
		db.Status.Resource.Id = ocisdkcommon.String("")
	}
	return db, db.Status.HandleError(e)
}

// Get retrieves the autonomousdatabase resource from oci
func (a *AutonomousDatabaseAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var db = obj.(*ocidbv1alpha1.AutonomousDatabase)

	request := ocidb.GetAutonomousDatabaseRequest{
		AutonomousDatabaseId: db.Status.Resource.Id,
	}

	r, e := a.dbClient.GetAutonomousDatabase(a.ctx, request)
	if e == nil {
		if r.AutonomousDatabase.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateAvailable {
			err := a.FetchWalletToSecret(db)
			if err != nil {
				return db, db.Status.HandleError(err)
			}
		} else {
			glog.V(4).Infof("skipping database wallet fetch due to database not in available state")
		}
	} else {
		return db, db.Status.HandleError(e)
	}

	return db.SetResource(&r.AutonomousDatabase), db.Status.HandleError(e)
}

// Update updates the autonomousdatabase resource in oci
func (a *AutonomousDatabaseAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	db := obj.(*ocidbv1alpha1.AutonomousDatabase)

	// skip update if the database in not in provisioning state
	if db.Status.Resource.LifecycleState == ocidb.AutonomousDatabaseLifecycleStateProvisioning {
		glog.V(4).Infof("skipping database update due to provisioning state")
		return nil, nil
	}

	// do an extra Get since the resource API doesn't like idempotent updates
	current := ocidb.GetAutonomousDatabaseRequest{
		AutonomousDatabaseId: db.Status.Resource.Id,
	}

	c, e := a.dbClient.GetAutonomousDatabase(a.ctx, current)
	if e == nil {
		if *c.AutonomousDatabase.CpuCoreCount == *db.Spec.CpuCoreCount &&
			*c.AutonomousDatabase.DataStorageSizeInTBs == *db.Spec.DataStorageSizeInTBs {
			glog.V(4).Infof("skipping database update because scaling parameters did not change")
			return db.SetResource(&c.AutonomousDatabase), db.Status.HandleError(e)
		}
	} else {
		glog.V(4).Infof("skipping database update due to inability to verify current scaling parameters")
	}

	request := ocidb.UpdateAutonomousDatabaseRequest{
		AutonomousDatabaseId: db.Status.Resource.Id,
		UpdateAutonomousDatabaseDetails: ocidb.UpdateAutonomousDatabaseDetails{
			DisplayName:          resourcescommon.Display(db.Name, db.Spec.DisplayName),
			DataStorageSizeInTBs: db.Spec.DataStorageSizeInTBs,
			CpuCoreCount:         db.Spec.CpuCoreCount,
		},
	}

	r, e := a.dbClient.UpdateAutonomousDatabase(a.ctx, request)

	if e != nil {
		return db, db.Status.HandleError(e)
	}

	return db.SetResource(&r.AutonomousDatabase), db.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the autonomousdatabase resource in the autonomousdatabase object
func (a *AutonomousDatabaseAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
