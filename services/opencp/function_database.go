package opencp

import (
	"encoding/json"
	// "errors"
	"io"
	"log"
	"strings"

	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	"github.com/opencontrolplane/opencp-shim/pkg"
	"github.com/opencontrolplane/opencp-spec/apis/v1alpha1"
	opencpgrpc "github.com/opencontrolplane/opencp-spec/grpc"

	restful "github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/types"
)

type DatabaseInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type Database struct {
}

func NewDatabase() DatabaseInterface {
	return &Database{}
}

// List - List of Database
func (d *Database) List(r *restful.Request, w *restful.Response) {
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

	// Get all the networks again and return them
	var allDatabase *opencpgrpc.DatabaseList
	if len(allFields) > 0 {
		nameDatabase := allFields["metadata.name"]
		allDatabase, err = app.Database.ListDatabase(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &nameDatabase})
		if err != nil {
			log.Println(err)
		}
	} else {
		q := apiRequestInfo.Namespace
		allDatabase, err = app.Database.ListDatabase(r.Request.Context(), &opencpgrpc.FilterOptions{Namespace: &q})
		if err != nil {
			log.Println(err)
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, db := range allDatabase.Items {
			cell := metav1.TableRow{
				Cells: []interface{}{db.Metadata.Name, db.Metadata.UID, db.Spec.Nodes, db.Spec.Size, db.Spec.Engine, db.Spec.EngineVersion, db.Status.State},
				Object: runtime.RawExtension{
					Object: &metav1.PartialObjectMetadata{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Database",
							APIVersion: "opencp.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      db.Metadata.Name,
							UID:       db.Metadata.UID,
							Namespace: db.Metadata.Namespace,
						},
					},
				},
			}
			tableRow = append(tableRow, cell)
		}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the Database"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the Database (from metadata)"},
				{Name: "Nodes", Type: "string", Format: "string", Description: "Number of nodes"},
				{Name: "Size", Type: "string", Format: "string", Description: "Size of the Database"},
				{Name: "Engine", Type: "string", Format: "string", Description: "Engine of the Database"},
				{Name: "Engine Version", Type: "string", Format: "string", Description: "Engine Version of the Database"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the Database"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	databaseList := []v1alpha1.Database{}
	for _, db := range allDatabase.Items {

		var emptyDatabaseSpec v1alpha1.DatabaseSpec
		var emptyDatabaseStatus v1alpha1.DatabaseStatus

		pkg.CopyTo(db.Spec, &emptyDatabaseSpec)
		pkg.CopyTo(db.Status, &emptyDatabaseStatus)

		singleDatabase := v1alpha1.Database{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Database",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *db.Metadata,
			Spec:       emptyDatabaseSpec,
			Status:     emptyDatabaseStatus,
		}

		databaseList = append(databaseList, singleDatabase)
	}

	list := v1alpha1.DatabaseList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DatabaseList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: databaseList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// Get - Get a Database
func (d *Database) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	db, err := app.Database.GetDatabase(r.Request.Context(), &opencpgrpc.FilterOptions{Namespace: &apiRequestInfo.Namespace, Name: &apiRequestInfo.Name})
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondNotFound(apiRequestInfo))
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{db.Metadata.Name, db.Metadata.UID, db.Spec.Nodes, db.Spec.Size, db.Spec.Engine, db.Spec.EngineVersion, db.Status.State}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the Database"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the Database (from metadata)"},
				{Name: "Nodes", Type: "string", Format: "string", Description: "Number of nodes"},
				{Name: "Size", Type: "string", Format: "string", Description: "Size of the Database"},
				{Name: "Engine", Type: "string", Format: "string", Description: "Engine of the Database"},
				{Name: "Engine Version", Type: "string", Format: "string", Description: "Engine Version of the Database"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the Database"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	var emptyDatabaseSpec v1alpha1.DatabaseSpec
	var emptyDatabaseStatus v1alpha1.DatabaseStatus

	pkg.CopyTo(db.Spec, &emptyDatabaseSpec)
	pkg.CopyTo(db.Status, &emptyDatabaseStatus)

	database := v1alpha1.Database{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Database",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *db.Metadata,
		Spec:       emptyDatabaseSpec,
		Status:     emptyDatabaseStatus,
	}

	// print the request method and path
	w.WriteAsJson(database)
}

// Create - Create a Database
func (d *Database) Create(r *restful.Request, w *restful.Response) {
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

	databaseOpenCP := opencpgrpc.Database{}
	err = json.Unmarshal(body, &databaseOpenCP)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Create the database
	db, err := app.Database.CreateDatabase(r.Request.Context(), &databaseOpenCP)
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondError(apiRequestInfo, databaseOpenCP.Metadata.Name, "error creating database", err))
		return
	}

	var emptyDatabaseSpec v1alpha1.DatabaseSpec
	var emptyDatabaseStatus v1alpha1.DatabaseStatus

	pkg.CopyTo(db.Spec, &emptyDatabaseSpec)
	pkg.CopyTo(db.Status, &emptyDatabaseStatus)

	database := v1alpha1.Database{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Database",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *db.Metadata,
		Spec:       emptyDatabaseSpec,
		Status:     emptyDatabaseStatus,
	}

	w.WriteAsJson(database)
}

// Delete - Delete a Database
func (d *Database) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	db, err := app.Database.DeleteDatabase(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, apiRequestInfo.Name, "error deleteing the database", err)
		w.WriteAsJson(respondStatus)
		return
	}

	if db != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: "Success",
			Details: &metav1.StatusDetails{
				Name:  db.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   db.Metadata.UID,
			},
		}
	}

	if db == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}

	// print the request method and path
	w.WriteAsJson(respondStatus)
}
