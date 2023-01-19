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

	// Get all the networks again and return them
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
			respondStatus := pkg.RespondError(apiRequestInfo, "error listing domains")
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

		// 	lasyApply, err := pkg.LastAppliedConfig(ctx, etcdObjectReader, "", &domain, "Domain")
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

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
			Spec: &domainSpec,
			Status: domainStatus,
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

	//Get LastAppliedConfig
	// lasyApply, err := pkg.LastAppliedConfig(ctx, etcdObjectReader, "", dnsDomain, "Domain")
	// if err != nil {
	// 	log.Println(err)
	// }

	// allRecords := []v1alpha1.DomainRecords{}
	// for _, record := range records {
	// 	allRecords = append(allRecords, v1alpha1.DomainRecords{
	// 		Name:     record.Name,
	// 		Value:    record.Value,
	// 		Type:     string(record.Type),
	// 		Priority: record.Priority,
	// 		TTL:      record.TTL,
	// 	})
	// }

	// domain := v1alpha1.Domain{
	// 	TypeMeta: metav1.TypeMeta{
	// 		Kind:       "Domain",
	// 		APIVersion: "opencp.io/v1alpha1",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: dnsDomain.Name,
	// 		// CreationTimestamp: metav1.Time{Time: vm.CreatedAt},
	// 		Annotations: map[string]string{
	// 			corev1.LastAppliedConfigAnnotation: lasyApply,
	// 		},
	// 		UID: types.UID(dnsDomain.ID),
	// 	},
	// 	Spec: v1alpha1.DomainSpec{
	// 		Records: allRecords,
	// 	},
	// 	Status: v1alpha1.DomainStatus{
	// 		State: "ready",
	// 	},
	// }

	domainSpec := opencpapi.DomainSpec{}
	domainStatus := &opencpapi.DomainStatus{}

	pkg.CopyTo(domain.Spec, &domainSpec)
	pkg.CopyTo(domain.Status, &domainStatus)

	domainsRespond := &opencpapi.Domain{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Domain",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *domain.Metadata,
		Spec: &domainSpec,
		Status: domainStatus,
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

	// err = etcdObjectReader.SetStoredCustomResource("", domainOpenCP.Name, domainOpenCP.Annotations[corev1.LastAppliedConfigAnnotation])
	// if err != nil {
	// 	log.Println(err)
	// }

	// Createn the domain
	domain, err := app.Domain.CreateDomain(r.Request.Context(), domainOpenCP)
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, "error finding domain")
		w.WriteAsJson(respondStatus)
		return
	}

	// // Create the records
	// if len(domainOpenCP.Spec.Records) > 0 {
	// 	for _, record := range domainOpenCP.Spec.Records {
	// 		recordConfig := &civogo.DNSRecordConfig{
	// 			Type:     civogo.DNSRecordType(record.Type),
	// 			Name:     record.Name,
	// 			Value:    record.Value,
	// 			Priority: record.Priority,
	// 			TTL:      record.TTL,
	// 		}
	// 		_, err := client.CreateDNSRecord(domain.ID, recordConfig)
	// 		if err != nil {
	// 			respondStatus := pkg.RespondError(apiRequestInfo, "error creating record")
	// 			w.WriteAsJson(respondStatus)
	// 			return
	// 		}
	// 	}
	// }

	// Get last applied config
	// lasyApply, err := etcdObjectReader.GetStoredCustomResource("", domainOpenCP.Name)
	// if err != nil {
	// 	if errors.Is(err, pkg.ErrNotFound) {
	// 		log.Println(err)
	// 		// Add to etcd
	// 	}
	// }

	// Get all the records
	// records, err := client.ListDNSRecords(domain.ID)
	// if err != nil {
	// 	respondStatus := pkg.RespondError(apiRequestInfo, "error finding records")
	// 	w.WriteAsJson(respondStatus)
	// 	return
	// }

	// allRecords := []v1alpha1.DomainRecords{}
	// for _, record := range records {
	// 	allRecords = append(allRecords, v1alpha1.DomainRecords{
	// 		Name:     record.Name,
	// 		Value:    record.Value,
	// 		Type:     string(record.Type),
	// 		Priority: record.Priority,
	// 		TTL:      record.TTL,
	// 	})
	// }

	// domainRespond := v1alpha1.Domain{
	// 	TypeMeta: metav1.TypeMeta{
	// 		Kind:       "Domain",
	// 		APIVersion: "opencp.io/v1alpha1",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      domain.Name,
	// 		Namespace: apiRequestInfo.Namespace,
	// 		Annotations: map[string]string{
	// 			corev1.LastAppliedConfigAnnotation: lasyApply,
	// 		},
	// 		// CreationTimestamp: metav1.Time{Time: firewall.CreatedAt},
	// 		UID: types.UID(domain.ID),
	// 	},
	// 	Spec: v1alpha1.DomainSpec{
	// 		Records: allRecords,
	// 	},
	// 	Status: v1alpha1.DomainStatus{
	// 		State: "ready",
	// 	},
	// }

	domainSpec := opencpapi.DomainSpec{}
	domainStatus := &opencpapi.DomainStatus{}

	pkg.CopyTo(domain.Spec, &domainSpec)
	pkg.CopyTo(domain.Status, &domainStatus)

	domainsRespond := &opencpapi.Domain{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Domain",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *domain.Metadata,
		Spec: &domainSpec,
		Status: domainStatus,
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
	q := apiRequestInfo.Name
	domain, err := app.Domain.GetDomain(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &q})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, "error finding domain")
		w.WriteAsJson(respondStatus)
		return
	}

	if domain != nil {
		_, err = app.Domain.DeleteDomain(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &domain.Metadata.Name})
		if err != nil {
			respondStatus = pkg.RespondError(apiRequestInfo, "error deleting domain")
			w.WriteAsJson(respondStatus)
			return
		}

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

		// Delete from etcd
		// _, err := etcdObjectReader.DeleteStoredCustomResource("", apiRequestInfo.Name)
		// if err != nil {
		// 	log.Println(err)
		// }
	}

	if domain == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}
	// print the request method and path
	w.WriteAsJson(respondStatus)
}
