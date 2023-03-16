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
)

type FirewallInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type Firewall struct {
}

func NewFirewall() FirewallInterface {
	return &Firewall{}
}

// VirtualMachineList - List of VirtualMachine
func (f *Firewall) List(r *restful.Request, w *restful.Response) {
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
	var allFirewall *opencpgrpc.FirewallList
	if len(allFields) > 0 {
		nameFirewall := allFields["metadata.name"]
		allFirewall, err = app.Firewall.ListFirewall(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &nameFirewall})
		if err != nil {
			log.Println(err)
		}
	} else {
		q := apiRequestInfo.Namespace
		allFirewall, err = app.Firewall.ListFirewall(r.Request.Context(), &opencpgrpc.FilterOptions{Namespace: &q})
		if err != nil {
			log.Println(err)
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, fw := range allFirewall.Items {
			cell := metav1.TableRow{
				Cells: []interface{}{fw.Metadata.Name, fw.Metadata.UID, fw.Status.TotalRules, fw.Status.State},
				Object: runtime.RawExtension{
					Object: &metav1.PartialObjectMetadata{
						TypeMeta: metav1.TypeMeta{
							Kind:       "VirtualMachine",
							APIVersion: "opencp.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      fw.Metadata.Name,
							UID:       fw.Metadata.UID,
							Namespace: fw.Metadata.Namespace,
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
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the firewall"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the firewall (from metadata)"},
				{Name: "Total rules", Type: "string", Format: "string", Description: "The total of rule inside the firewall (from metadata)"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the instance"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	fwList := []v1alpha1.Firewall{}
	for _, fw := range allFirewall.Items {
		var emptyFirewallSpec v1alpha1.FirewallSpec
		var emptyFirewallStatus v1alpha1.FirewallStatus

		pkg.CopyTo(fw.Spec, &emptyFirewallSpec)
		pkg.CopyTo(fw.Status, &emptyFirewallStatus)

		firewall := v1alpha1.Firewall{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Firewall",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *fw.Metadata,
			Spec:       emptyFirewallSpec,
			Status:     emptyFirewallStatus,
		}

		fwList = append(fwList, firewall)
	}

	list := v1alpha1.FirewallList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FirewallList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: fwList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// FirewallGet get a firewall
func (f *Firewall) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	fw, err := app.Firewall.GetFirewall(r.Request.Context(), &opencpgrpc.FilterOptions{Namespace: &apiRequestInfo.Namespace, Name: &apiRequestInfo.Name})
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondNotFound(apiRequestInfo))
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{fw.Metadata.Name, fw.Metadata.UID, fw.Status.TotalRules, fw.Status.State}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the firewall"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the firewall (from metadata)"},
				{Name: "Total rules", Type: "string", Format: "string", Description: "The total of rule inside the firewall (from metadata)"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the instance"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	emptyFirewallSpec := v1alpha1.FirewallSpec{}
	emptyFirewallStatus := v1alpha1.FirewallStatus{}

	pkg.CopyTo(fw.Spec, &emptyFirewallSpec)
	pkg.CopyTo(fw.Status, &emptyFirewallStatus)

	firewall := v1alpha1.Firewall{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Firewall",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *fw.Metadata,
		Spec:       emptyFirewallSpec,
		Status:     emptyFirewallStatus,
	}

	// print the request method and path
	w.WriteAsJson(firewall)
}

// FirewallCreate create a firewall
func (f *Firewall) Create(r *restful.Request, w *restful.Response) {
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

	firewallOpenCP := opencpgrpc.Firewall{}
	err = json.Unmarshal(body, &firewallOpenCP)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Create the firewall
	fw, err := app.Firewall.CreateFirewall(r.Request.Context(), &firewallOpenCP)
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondError(apiRequestInfo, firewallOpenCP.Metadata.Name, "error creating firewall", err))
		return
	}

	// Return the firewall
	emptyFirewallSpec := v1alpha1.FirewallSpec{}
	emptyFirewallStatus := v1alpha1.FirewallStatus{}

	pkg.CopyTo(fw.Spec, &emptyFirewallSpec)
	pkg.CopyTo(fw.Status, &emptyFirewallStatus)

	firewallRespond := v1alpha1.Firewall{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Firewall",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *fw.Metadata,
		Spec:       emptyFirewallSpec,
		Status:     emptyFirewallStatus,
	}

	w.WriteAsJson(firewallRespond)
}

// FirewallDelete delete a firewall
func (f *Firewall) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	fw, err := app.Firewall.DeleteFirewall(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, apiRequestInfo.Name, "error deleteing the firewall", err)
		w.WriteAsJson(respondStatus)
		return
	}

	if fw != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: "Success",
			Details: &metav1.StatusDetails{
				Name:  fw.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   fw.Metadata.UID,
			},
		}
	}

	if fw == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}
	// print the request method and path
	w.WriteAsJson(respondStatus)
}
