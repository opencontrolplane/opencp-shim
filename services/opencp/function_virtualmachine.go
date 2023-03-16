package opencp

import (
	"encoding/json"
	"fmt"
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
	"k8s.io/apimachinery/pkg/runtime"
)

type VirtualMachineInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type VirtualMachine struct {
	// EtcdClient *clientv3.Client
}

func NewVirtualMachine() VirtualMachineInterface {
	return &VirtualMachine{}
}

// VirtualMachineList - List of VirtualMachine
func (v *VirtualMachine) List(r *restful.Request, w *restful.Response) {
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
	var virtualMachineList *opencpgrpc.VirtualMachineList
	if len(allFields) > 0 {
		nameVm := allFields["metadata.name"]
		virtualMachineList, err = app.VirtualMachine.ListVirtualMachine(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &nameVm})
		if err != nil {
			log.Println(err)
		}
	} else {
		q := apiRequestInfo.Namespace
		virtualMachineList, err = app.VirtualMachine.ListVirtualMachine(r.Request.Context(), &opencpgrpc.FilterOptions{Namespace: &q})
		if err != nil {
			log.Println(err)
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, vm := range virtualMachineList.Items {

			// _, err = pkg.LastAppliedConfig(ctx, etcdObjectReader, network.Label, &vm, "VirtualMachine")
			// if err != nil {
			// 	log.Println(err)
			// }

			cell := metav1.TableRow{
				Cells: []interface{}{vm.Metadata.Name, vm.Metadata.UID, vm.Spec.Size, vm.Status.PublicIP, vm.Status.PrivateIP, vm.Status.State},
				Object: runtime.RawExtension{
					Object: &metav1.PartialObjectMetadata{
						TypeMeta: metav1.TypeMeta{
							Kind:       "VirtualMachine",
							APIVersion: "opencp.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      vm.Metadata.Name,
							UID:       vm.Metadata.UID,
							Namespace: vm.Metadata.Namespace,
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
				{Name: "Hostname", Type: "string", Format: "name", Description: "Hostname of the instance"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the instance (from metadata)"},
				{Name: "Size", Type: "string", Format: "string", Description: "Size of the instance (from metadata)"},
				// {Name: "CPU", Type: "string", Format: "string", Description: "CPU of the instance (from metadata)"},
				// {Name: "RAM", Type: "string", Format: "string", Description: "Ram of the instance"},
				// {Name: "SSD", Type: "string", Format: "string", Description: "SSD Disk of the instance"},
				{Name: "Public IP", Type: "date", Format: "date", Description: "Public IP of the instance"},
				{Name: "Private IP", Type: "date", Format: "date", Description: "Private IP of the instance"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the instance"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	vmList := []v1alpha1.VirtualMachine{}
	for _, vm := range virtualMachineList.Items {
		// lasyApply, err := pkg.LastAppliedConfig(ctx, etcdObjectReader, network.Label, &vm, "VirtualMachine")
		// if err != nil {
		// 	log.Println(err)
		// }
		var emptyVirtualMachineSpec v1alpha1.VirtualMachineSpec
		var emptyVirtualMachineStatus v1alpha1.VirtualMachineStatus

		pkg.CopyTo(vm.Spec, &emptyVirtualMachineSpec)
		pkg.CopyTo(vm.Status, &emptyVirtualMachineStatus)

		vm := v1alpha1.VirtualMachine{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VirtualMachine",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *vm.Metadata,
			Spec:       emptyVirtualMachineSpec,
			Status:     emptyVirtualMachineStatus,
		}

		vmList = append(vmList, vm)
	}

	list := v1alpha1.VirtualMachineList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachineList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: vmList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// VirtualMachineGet get a virtual machine
func (v *VirtualMachine) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	virtualMachine, err := app.VirtualMachine.GetVirtualMachine(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name, Namespace: &apiRequestInfo.Namespace})
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondNotFound(apiRequestInfo))
		return
	}

	if virtualMachine == nil {
		respondStatus := metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status:  "Failure",
			Reason:  metav1.StatusReasonNotFound,
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Virtual Machine %s not found", apiRequestInfo.Name),
			Details: &metav1.StatusDetails{
				Name:  apiRequestInfo.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
			},
		}

		// print the request method and path
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		w.WriteAsJson(respondStatus)
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{{Cells: []interface{}{virtualMachine.Metadata.Name, virtualMachine.Metadata.UID, virtualMachine.Spec.Size, virtualMachine.Status.PublicIP, virtualMachine.Status.PrivateIP, virtualMachine.Status.State}}}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Hostname", Type: "string", Format: "name", Description: "Hostname of the instance"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the instance (from metadata)"},
				{Name: "Size", Type: "string", Format: "string", Description: "Size of the instance (from metadata)"},
				{Name: "Public IP", Type: "date", Format: "date", Description: "Public IP of the instance"},
				{Name: "Private IP", Type: "date", Format: "date", Description: "Private IP of the instance"},
				{Name: "Status", Type: "date", Format: "date", Description: "Status of the instance"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	var virtualMachineSpec v1alpha1.VirtualMachineSpec
	var virtualMachineStatus v1alpha1.VirtualMachineStatus

	pkg.CopyTo(virtualMachine.Spec, &virtualMachineSpec)
	pkg.CopyTo(virtualMachine.Status, &virtualMachineStatus)

	virtualMachineRespond := v1alpha1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *virtualMachine.Metadata,
		Spec:       virtualMachineSpec,
		Status:     virtualMachineStatus,
	}

	// print the request method and path
	w.WriteAsJson(virtualMachineRespond)
}

// VirtualMachineCreate create a new virtual machine
func (v *VirtualMachine) Create(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	// resolver := pkg.RequestInfoResolver()
	// apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	// if err != nil {
	// 	log.Println(err)
	// }

	// Real all the body of the request and unmarshal it
	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}

	virtualMachine := &opencpgrpc.VirtualMachine{}
	err = json.Unmarshal(body, &virtualMachine)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// err = etcdObjectReader.SetStoredCustomResource(virtualMachine.Namespace, virtualMachine.Name, virtualMachine.Annotations[corev1.LastAppliedConfigAnnotation])
	// if err != nil {
	// 	log.Println(err)
	// }

	// getDiskImage, err := client.FindDiskImage(virtualMachine.Spec.Image)
	// if err != nil {
	// 	log.Println(err)
	// }

	// // Create the VM
	// vm := &civogo.InstanceConfig{
	// 	Hostname:         virtualMachine.Name,
	// 	ReverseDNS:       virtualMachine.Name,
	// 	Size:             virtualMachine.Spec.Size,
	// 	Region:           client.Region,
	// 	PublicIPRequired: strconv.FormatBool(virtualMachine.Spec.Ipv4),
	// 	NetworkID:        getNetwork.ID,
	// 	TemplateID:       getDiskImage.ID,
	// 	Script:           virtualMachine.Spec.UserScript,
	// 	Tags:             virtualMachine.Spec.Tags,
	// }

	// // Check the firewall
	// if virtualMachine.Spec.Firewall != "" {
	// 	getFirewall, err := client.FindFirewall(virtualMachine.Spec.Firewall)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	vm.FirewallID = getFirewall.ID
	// }

	// if virtualMachine.Spec.Auth.User != "" {
	// 	vm.InitialUser = virtualMachine.Spec.Auth.User
	// }

	// if virtualMachine.Spec.Auth.SSHKey != "" {
	// 	vm.SSHKeyID = virtualMachine.Spec.Auth.SSHKey
	// }

	virtualMachine, err = app.VirtualMachine.CreateVirtualMachine(r.Request.Context(), virtualMachine)
	if err != nil {
		log.Println(err)
	}

	// // Get last applied config
	// lasyApply, err := etcdObjectReader.GetStoredCustomResource(getNetwork.Label, instance.Hostname)
	// if err != nil {
	// 	if errors.Is(err, pkg.ErrNotFound) {
	// 		log.Println(err)
	// 		// Add to etcd
	// 	}
	// }

	var virtualMachineSpec v1alpha1.VirtualMachineSpec
	var virtualMachineStatus v1alpha1.VirtualMachineStatus

	pkg.CopyTo(virtualMachine.Spec, &virtualMachineSpec)
	pkg.CopyTo(virtualMachine.Status, &virtualMachineStatus)

	virtualMachineRespond := v1alpha1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *virtualMachine.Metadata,
		Spec:       virtualMachineSpec,
		Status:     virtualMachineStatus,
	}

	// w.WriteAsJson(virtualMachineRespond)
	w.WriteAsJson(virtualMachineRespond)
}

// VirtualMachineDelete delete a kubernetes cluster
func (v *VirtualMachine) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	// Send to delete the virtual machine
	virtualMachine, err := app.VirtualMachine.DeleteVirtualMachine(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, apiRequestInfo.Name, "error deleting the virtual machine", err)
		w.WriteAsJson(respondStatus)
		return
	}

	if virtualMachine != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: metav1.StatusSuccess,
			Details: &metav1.StatusDetails{
				Name:  virtualMachine.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   virtualMachine.Metadata.UID,
			},
		}
	}

	if virtualMachine == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
		return
	}

	// Respond with the status
	w.WriteAsJson(respondStatus)
}
