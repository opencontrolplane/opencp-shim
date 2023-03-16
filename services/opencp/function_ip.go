package opencp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	// "strings"

	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	"github.com/opencontrolplane/opencp-shim/pkg"

	opencpapi "github.com/opencontrolplane/opencp-spec/apis/v1alpha1"
	opencpgrpc "github.com/opencontrolplane/opencp-spec/grpc"

	restful "github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IPInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type IP struct {
}

func NewIP() IPInterface {
	return &IP{}
}

// IpAdress
func (p *IP) List(r *restful.Request, w *restful.Response) {
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

	var allIPs *opencpgrpc.IpList
	if len(allFields) > 0 {
		q := allFields["metadata.name"]
		ip, err := app.IP.GetIp(r.Request.Context(), &opencpgrpc.FilterOptions{
			Name: &q,
		})
		if err != nil {
			allIPs = &opencpgrpc.IpList{}
		}

		if ip != nil {
			allIPs.Items = append(allIPs.Items, ip)
		}
	} else {
		allIPs, err = app.IP.ListIp(r.Request.Context(), &opencpgrpc.FilterOptions{})
		if err != nil {
			respondStatus := pkg.RespondError(apiRequestInfo, "", "error listing ips", err)
			w.WriteAsJson(respondStatus)
			return
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, ip := range allIPs.Items {
			cell := metav1.TableRow{Cells: []interface{}{ip.Metadata.Name, ip.Metadata.UID, ip.Status.Ip, ip.Status.Assignedto.Name, ip.Status.Assignedto.Type}}
			tableRow = append(tableRow, cell)
		}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the instance"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the instance (from metadata)"},
				{Name: "IP", Type: "string", Format: "string", Description: "IP of the instance"},
				{Name: "Assigned to", Type: "string", Format: "string", Description: "Assigned to"},
				{Name: "Type", Type: "string", Format: "string", Description: "Type of resource"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	ipList := []opencpapi.IP{}
	for _, ip := range allIPs.Items {

		ipSpec := opencpapi.IPSpec{}
		ipStatus := opencpapi.IPStatus{}

		pkg.CopyTo(ip.Spec, &ipSpec)
		pkg.CopyTo(ip.Status, &ipStatus)

		singleIP := &opencpapi.IP{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IP",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *ip.Metadata,
			Spec:       ipSpec,
			Status:     ipStatus,
		}

		ipList = append(ipList, *singleIP)
	}

	list := opencpapi.IPList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IPList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: ipList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

func (p *IP) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Get all the networks again and return them
	ip, err := app.IP.GetIp(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondNotFound(apiRequestInfo))
		return
	}

	if ip == nil {
		respondStatus := pkg.RespondNotFound(apiRequestInfo)
		// print the request method and path
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		w.WriteAsJson(respondStatus)
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{ip.Metadata.Name, ip.Metadata.UID, ip.Status.Ip, ip.Status.Assignedto.Name, ip.Status.Assignedto.Type}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the instance"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the instance (from metadata)"},
				{Name: "IP", Type: "string", Format: "string", Description: "IP of the instance"},
				{Name: "Assigned to", Type: "string", Format: "string", Description: "Assigned to"},
				{Name: "Type", Type: "string", Format: "string", Description: "Type of resource"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	ipSpec := opencpapi.IPSpec{}
	ipStatus := opencpapi.IPStatus{}

	pkg.CopyTo(ip.Spec, &ipSpec)
	pkg.CopyTo(ip.Status, &ipStatus)

	ipRespond := &opencpapi.IP{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IP",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *ip.Metadata,
		Spec:       ipSpec,
		Status:     ipStatus,
	}

	// print the request method and path
	w.WriteAsJson(ipRespond)
}

func (p *IP) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	ip, err := app.IP.DeleteIp(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, apiRequestInfo.Name, "error deleting IP", err)
		w.WriteAsJson(respondStatus)
		return
	}

	if ip != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: "Success",
			Details: &metav1.StatusDetails{
				Name:  ip.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   ip.Metadata.UID,
			},
		}
	}

	if ip == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}
	// print the request method and path
	w.WriteAsJson(respondStatus)
}

func (p *IP) Create(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		// http.Error(w, "can't read body", http.StatusBadRequest)
		// return
	}

	ipOpenCP := &opencpgrpc.Ip{}
	err = json.Unmarshal(body, ipOpenCP)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	ip, err := app.IP.CreateIp(r.Request.Context(), ipOpenCP)
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, ipOpenCP.Metadata.Name, "error finding domain", err)
		w.WriteAsJson(respondStatus)
		return
	}

	ipSpec := opencpapi.IPSpec{}
	ipStatus := opencpapi.IPStatus{}

	pkg.CopyTo(ip.Spec, &ipSpec)
	pkg.CopyTo(ip.Status, &ipStatus)

	ipRespond := &opencpapi.IP{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IP",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *ip.Metadata,
		Spec:       ipSpec,
		Status:     ipStatus,
	}

	// print the request method and path
	w.WriteAsJson(ipRespond)
}
