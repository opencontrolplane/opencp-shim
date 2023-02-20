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

type ObjectStorageInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type ObjectStorage struct {
}

func NewObjectStorage() ObjectStorageInterface {
	return &ObjectStorage{}
}

// List - List all onject storage
func (s *ObjectStorage) List(r *restful.Request, w *restful.Response) {
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
	var allObjectStorage *opencpgrpc.ObjectStorageList
	if len(allFields) > 0 {
		q := allFields["metadata.name"]
		objStorage, err := app.ObjectStorage.GetObjectStorage(r.Request.Context(), &opencpgrpc.FilterOptions{
			Name: &q,
		})
		if err != nil {
			allObjectStorage = &opencpgrpc.ObjectStorageList{}
		}

		if objStorage != nil {
			allObjectStorage.Items = append(allObjectStorage.Items, objStorage)
		}
	} else {
		allObjectStorage, err = app.ObjectStorage.ListObjectStorage(r.Request.Context(), &opencpgrpc.FilterOptions{})
		if err != nil {
			respondStatus := pkg.RespondError(apiRequestInfo, "error listing object storage")
			w.WriteAsJson(respondStatus)
			return
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, objectstorage := range allObjectStorage.Items {
			tableRow = append(tableRow, metav1.TableRow{
				Cells: []interface{}{objectstorage.Metadata.Name, objectstorage.Metadata.UID, objectstorage.Spec.Size, objectstorage.Status.State},
			})
		}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the ObjectStorage", Priority: 0},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the ObjectStorage (from metadata)"},
				{Name: "Size", Type: "integer", Format: "string", Description: "The size of the object storage (from spec)"},
				{Name: "Status", Type: "string", Format: "string", Description: "Status of the ObjectStorage"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	objectstorageList := []v1alpha1.ObjectStorage{}
	for _, objectstorage := range allObjectStorage.Items {
		objStorageSpec := v1alpha1.ObjectStorageSpec{}
		objStorageStatus := v1alpha1.ObjectStorageStatus{}

		pkg.CopyTo(objectstorage.Spec, &objStorageSpec)
		pkg.CopyTo(objectstorage.Status, &objStorageStatus)

		obstorage := v1alpha1.ObjectStorage{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ObjectStorage",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *objectstorage.Metadata,
			Spec:       objStorageSpec,
			Status:     objStorageStatus,
		}

		objectstorageList = append(objectstorageList, obstorage)
	}

	list := v1alpha1.ObjectStorageList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ObjectStorageList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: objectstorageList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// Get - Get a object storage
func (s *ObjectStorage) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	objectstorage, err := app.ObjectStorage.GetObjectStorage(r.Request.Context(), &opencpgrpc.FilterOptions{
		Name: &apiRequestInfo.Name,
	})
	if err != nil {
		// print the log
		log.Println(err)
	}

	if objectstorage == nil {
		respondStatus := pkg.RespondNotFound(apiRequestInfo)
		// print the request method and path
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		w.WriteAsJson(respondStatus)
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{objectstorage.Metadata.Name, objectstorage.Metadata.UID, objectstorage.Spec.Size, objectstorage.Status.State}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the ObjectStorage", Priority: 0},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the ObjectStorage (from metadata)"},
				{Name: "Size", Type: "integer", Format: "string", Description: "The size of the object storage (from metadata)"},
				{Name: "Status", Type: "string", Format: "string", Description: "Status of the ObjectStorage"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	objStorageSpec := v1alpha1.ObjectStorageSpec{}
	objStorageStatus := v1alpha1.ObjectStorageStatus{}

	pkg.CopyTo(objectstorage.Spec, &objStorageSpec)
	pkg.CopyTo(objectstorage.Status, &objStorageStatus)

	objectstorageReturn := v1alpha1.ObjectStorage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ObjectStorage",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *objectstorage.Metadata,
		Spec:       objStorageSpec,
		Status:     objStorageStatus,
	}

	// print the request method and path
	w.WriteAsJson(objectstorageReturn)
}

// DomainCreate - Create a domain
func (s *ObjectStorage) Create(r *restful.Request, w *restful.Response) {
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

	objectstorageOpenCP := opencpgrpc.ObjectStorage{}
	err = json.Unmarshal(body, &objectstorageOpenCP)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Createn the ObjectStore
	objectstorage, err := app.ObjectStorage.CreateObjectStorage(r.Request.Context(), &objectstorageOpenCP)
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, "error creating the object storage")
		w.WriteAsJson(respondStatus)
		return
	}

	objStorageSpec := v1alpha1.ObjectStorageSpec{}
	objStorageStatus := v1alpha1.ObjectStorageStatus{}

	pkg.CopyTo(objectstorage.Spec, &objStorageSpec)
	pkg.CopyTo(objectstorage.Status, &objStorageStatus)

	obstorageRespond := v1alpha1.ObjectStorage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ObjectStorage",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *objectstorage.Metadata,
		Spec:       objStorageSpec,
		Status:     objStorageStatus,
	}
	w.WriteAsJson(obstorageRespond)
}

// Delete - Delete a sshkey
func (s *ObjectStorage) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	objectstorage, err := app.ObjectStorage.DeleteObjectStorage(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, "error deleting ObjectStore")
		w.WriteAsJson(respondStatus)
		return
	}

	if objectstorage != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: "Success",
			Details: &metav1.StatusDetails{
				Name:  objectstorage.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   objectstorage.Metadata.UID,
			},
		}
	}

	if objectstorage == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}
	// print the request method and path
	w.WriteAsJson(respondStatus)
}
