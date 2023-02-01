package opencp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	"github.com/opencontrolplane/opencp-shim/pkg"
	"github.com/opencontrolplane/opencp-spec/apis/v1alpha1"
	opencpgrpc "github.com/opencontrolplane/opencp-spec/grpc"

	restful "github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


type ObjectStorageCredentialInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type ObjectStorageCredential struct {
}

func NewObjectStorageCredential() ObjectStorageCredentialInterface {
	return &ObjectStorageCredential{}
}

// List - List all the object storage credential
func (s *ObjectStorageCredential) List(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Check if we need filter the list
	fileds := r.QueryParameter("fieldSelector")
	allFields := make(map[string]string)
	if fileds != "" {
		filedsList := strings.Split(fileds, ",")
		for _, field := range filedsList {
			fieldSplit := strings.Split(field, "=")
			allFields[fieldSplit[0]] = fieldSplit[1]
		}
	}

	// Get all the onbject storage again and return them
	var allObjectStorageCredential *opencpgrpc.ObjectStorageCredentialList
	if len(allFields) > 0 {
		q := allFields["metadata.name"]
		objStorageCredential, err := app.ObjectStorageCredential.GetObjectStorageCredential(r.Request.Context(), &opencpgrpc.FilterOptions{
			Name: &q,
		})
		if err != nil {
			allObjectStorageCredential = &opencpgrpc.ObjectStorageCredentialList{}
		}

		if objStorageCredential != nil {
			allObjectStorageCredential.Items = append(allObjectStorageCredential.Items, objStorageCredential)
		}
	} else {
		allObjectStorageCredential, err = app.ObjectStorageCredential.ListObjectStorageCredential(r.Request.Context(), &opencpgrpc.FilterOptions{})
		if err != nil {
			respondStatus := pkg.RespondError(apiRequestInfo, "error listing object storage credential")
			w.WriteAsJson(respondStatus)
			return
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, objectstorageCredential := range allObjectStorageCredential.Items {
			tableRow = append(tableRow, metav1.TableRow{
				Cells: []interface{}{objectstorageCredential.Metadata.Name, objectstorageCredential.Metadata.UID, objectstorageCredential.Spec.Accesskey, objectstorageCredential.Status.State},
			})
		}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the ObjectStorage Credential", Priority: 0},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the ObjectStorage Credential (from metadata)"},
				{Name: "Access Key", Type: "string", Format: "string", Description: "The access key (from metadata)"},
				{Name: "Status", Type: "string", Format: "string", Description: "Status of the ObjectStorage Credential (from metadata)"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	objectstorageCredentialList := []v1alpha1.ObjectStorageCredential{}
	for _, objectstorage := range allObjectStorageCredential.Items {

		objStorageCredentialSpec := v1alpha1.ObjectStorageCredentialSpec{}
		objStorageCredentialStatus := v1alpha1.ObjectStorageCredentialStatus{}

		pkg.CopyTo(objectstorage.Spec, &objStorageCredentialSpec)
		pkg.CopyTo(objectstorage.Status, &objStorageCredentialStatus)

		obstorageCredential := v1alpha1.ObjectStorageCredential{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ObjectStorageCredential",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *objectstorage.Metadata,
			Spec: objStorageCredentialSpec,
			Status: objStorageCredentialStatus,
		}

		objectstorageCredentialList = append(objectstorageCredentialList, obstorageCredential)
	}

	list := v1alpha1.ObjectStorageCredentialList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ObjectStorageCredentialList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: objectstorageCredentialList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// Get - Get a the object storage credential
func (s *ObjectStorageCredential) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	objectstorageCredential, err := app.ObjectStorageCredential.GetObjectStorageCredential(r.Request.Context(), &opencpgrpc.FilterOptions{
		Name: &apiRequestInfo.Name,
	})
	if err != nil {
		// print the log
		log.Println(err)
	}

	if objectstorageCredential == nil {
		respondStatus := pkg.RespondNotFound(apiRequestInfo)
		// print the request method and path
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		w.WriteAsJson(respondStatus)
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{objectstorageCredential.Metadata.Name, objectstorageCredential.Metadata.UID, objectstorageCredential.Spec.Accesskey, objectstorageCredential.Status.State}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the ObjectStorage Credential", Priority: 0},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the ObjectStorage Credential (from metadata)"},
				{Name: "Access Key", Type: "string", Format: "string", Description: "The access key (from metadata)"},
				{Name: "Status", Type: "string", Format: "string", Description: "Status of the ObjectStorage Credential (from metadata)"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	objStorageCredentialSpec := v1alpha1.ObjectStorageCredentialSpec{}
	objStorageCredentialStatus := v1alpha1.ObjectStorageCredentialStatus{}

	pkg.CopyTo(objectstorageCredential.Spec, &objStorageCredentialSpec)
	pkg.CopyTo(objectstorageCredential.Status, &objStorageCredentialStatus)

	obs := v1alpha1.ObjectStorageCredential{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ObjectStorageCredential",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *objectstorageCredential.Metadata,
		Spec: objStorageCredentialSpec,
		Status: objStorageCredentialStatus,
	}

	// print the request method and path
	w.WriteAsJson(obs)
}

// DomainCreate - Create a domain
func (s *ObjectStorageCredential) Create(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Real all the body of the request and unmarshal it
	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}

	objectstorageCredentialOpenCP := opencpgrpc.ObjectStorageCredential{}
	err = json.Unmarshal(body, &objectstorageCredentialOpenCP)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Createn the ObjectStore
	objectstorageCredential, err := app.ObjectStorageCredential.CreateObjectStorageCredential(r.Request.Context(), &objectstorageCredentialOpenCP)
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, "error creating the object storage credential")
		w.WriteAsJson(respondStatus)
		return
	}

	objStorageCredentialSpec := v1alpha1.ObjectStorageCredentialSpec{}
	objStorageCredentialStatus := v1alpha1.ObjectStorageCredentialStatus{}

	pkg.CopyTo(objectstorageCredential.Spec, &objStorageCredentialSpec)
	pkg.CopyTo(objectstorageCredential.Status, &objStorageCredentialStatus)

	obs := v1alpha1.ObjectStorageCredential{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ObjectStorageCredential",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *objectstorageCredential.Metadata,
		Spec: objStorageCredentialSpec,
		Status: objStorageCredentialStatus,
	}

	// print the request method and path
	w.WriteAsJson(obs)
}

// Delete - Delete a sshkey
func (s *ObjectStorageCredential) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	objectstorageCredential, err := app.ObjectStorageCredential.DeleteObjectStorageCredential(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, "error deleting Object Store Credential")
		w.WriteAsJson(respondStatus)
		return
	}

	if objectstorageCredential != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: "Success",
			Details: &metav1.StatusDetails{
				Name:  objectstorageCredential.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   objectstorageCredential.Metadata.UID,
			},
		}
	}

	if objectstorageCredential == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}
	// print the request method and path
	w.WriteAsJson(respondStatus)
}
