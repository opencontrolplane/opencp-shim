package opencp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"strings"

	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	"github.com/opencontrolplane/opencp-shim/pkg"
	opencpapi "github.com/opencontrolplane/opencp-spec/apis/v1alpha1"
	opencpgrpc "github.com/opencontrolplane/opencp-spec/grpc"

	restful "github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/types"
)

type DomainInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type Domain struct {
}

func NewDomain() DomainInterface {
	return &Domain{}
}

// DomainList - List of domains
func (d *Domain) List(r *restful.Request, w *restful.Response) {
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

	// Get all the domains again and return them
	var allDomains *opencpgrpc.DomainList
	if len(allFields) > 0 {
		q := allFields["metadata.name"]
		domain, err := app.Domain.GetDomain(r.Request.Context(), &opencpgrpc.FilterOptions{
			Name: &q,
		})
		if err != nil {
			allDomains = &opencpgrpc.DomainList{}
		}

		if domain != nil {
			allDomains.Items = append(allDomains.Items, domain)
		}
	} else {
		allDomains, err = app.Domain.ListDomains(r.Request.Context(), &opencpgrpc.FilterOptions{})
		if err != nil {
			respondStatus := pkg.RespondError(apiRequestInfo, "", "error listing domains", err)
			w.WriteAsJson(respondStatus)
			return
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, domain := range allDomains.Items {
			cell := metav1.TableRow{Cells: []interface{}{domain.Metadata.Name, domain.Metadata.UID, len(domain.Spec.Records), domain.Status.State}}
			tableRow = append(tableRow, cell)
		}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the Domain"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the Domain (from metadata)"},
				{Name: "Total records", Type: "string", Format: "string", Description: "The total of records (from metadata)"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the domain"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	domainList := []opencpapi.Domain{}
	for _, domain := range allDomains.Items {

		domainSpec := opencpapi.DomainSpec{}
		domainStatus := &opencpapi.DomainStatus{}

		pkg.CopyTo(domain.Spec, &domainSpec)
		pkg.CopyTo(domain.Status, &domainStatus)

		singleDomain := &opencpapi.Domain{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Domain",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *domain.Metadata,
			Spec:       &domainSpec,
			Status:     domainStatus,
		}

		domainList = append(domainList, *singleDomain)
	}

	list := opencpapi.DomainList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DomainList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: domainList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// DomainGet - Get a domain
func (d *Domain) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	q := apiRequestInfo.Name
	domain, err := app.Domain.GetDomain(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &q})
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondNotFound(apiRequestInfo))
		return
	}

	if domain == nil {
		respondStatus := pkg.RespondNotFound(apiRequestInfo)
		// print the request method and path
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		w.WriteAsJson(respondStatus)
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{domain.Metadata.Name, domain.Metadata.UID, len(domain.Spec.Records), domain.Status.State}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the Domain"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the Domain (from metadata)"},
				{Name: "Total records", Type: "string", Format: "string", Description: "The total of records (from metadata)"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the domain"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	domainSpec := opencpapi.DomainSpec{}
	domainStatus := opencpapi.DomainStatus{}

	pkg.CopyTo(domain.Spec, &domainSpec)
	pkg.CopyTo(domain.Status, &domainStatus)

	domainsRespond := &opencpapi.Domain{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Domain",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *domain.Metadata,
		Spec:       &domainSpec,
		Status:     &domainStatus,
	}

	// print the request method and path
	w.WriteAsJson(domainsRespond)
}

// DomainCreate - Create a domain
func (d *Domain) Create(r *restful.Request, w *restful.Response) {
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

	domainOpenCP := &opencpgrpc.Domain{}
	err = json.Unmarshal(body, &domainOpenCP)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Createn the domain
	domain, err := app.Domain.CreateDomain(r.Request.Context(), domainOpenCP)
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, domainOpenCP.Metadata.Name, "error finding domain", err)
		w.WriteAsJson(respondStatus)
		return
	}

	domainSpec := opencpapi.DomainSpec{}
	domainStatus := opencpapi.DomainStatus{}

	pkg.CopyTo(domain.Spec, &domainSpec)
	pkg.CopyTo(domain.Status, &domainStatus)

	domainsRespond := &opencpapi.Domain{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Domain",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *domain.Metadata,
		Spec:       &domainSpec,
		Status:     &domainStatus,
	}

	// print the request method and path
	w.WriteAsJson(domainsRespond)
}

// DomainDelete - Delete a domain
func (d *Domain) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	domain, err := app.Domain.DeleteDomain(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, apiRequestInfo.Name, "error deleting domain", err)
		w.WriteAsJson(respondStatus)
		return
	}

	if domain != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: "Success",
			Details: &metav1.StatusDetails{
				Name:  domain.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   domain.Metadata.UID,
			},
		}
	}

	if domain == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}
	// print the request method and path
	w.WriteAsJson(respondStatus)
}
