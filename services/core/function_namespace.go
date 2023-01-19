package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/opencontrolplane/opencp-shim/pkg"
	clientv3 "go.etcd.io/etcd/client/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkInterface interface {
	List(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
}

type Network struct {
	EtcdClient *clientv3.Client
}

func NewNetwork(etcdClient *clientv3.Client) NetworkInterface {
	return &Network{EtcdClient: etcdClient}
}

// NetworkList Namespace functions (Civo Newtork)
func (n Network) List(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	log.Printf("%+v", apiRequestInfo)

	// Get all the networks
	allNetwork, err := app.Namespace.ListNamespace(r.Request.Context(), &opencpspec.FilterOptions{})
	if err != nil {
		log.Println(err)
	}

	// If the Header `Accept` is set with Table
	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, network := range allNetwork.Items {
			cell := metav1.TableRow{Cells: []interface{}{network.Metadata.Name, network.Metadata.UID, network.Status.Phase}}
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
				{Name: "State", Type: "string", Format: "string", Description: "State of the instance"},
			},
			Rows: tableRow,
		}
		w.WriteAsJson(list)
		return
	}

	networks := []corev1.Namespace{}
	for _, network := range allNetwork.Items {
		networks = append(networks, corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: *network.Metadata,
			Spec:       *network.Spec,
			Status:     *network.Status,
		})
	}
		

	networkList := corev1.NamespaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NamespaceList",
			APIVersion: "v1",
		},
		Items: networks,
	}
	// print the request method and path
	w.WriteAsJson(networkList)
}

func (n Network) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Get the network
	network, err := app.Namespace.GetNamespace(r.Request.Context(), &opencpspec.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		log.Println(err)
	}

	if network == nil {
		respondStatus := metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status:  "Failure",
			Reason:  metav1.StatusReasonNotFound,
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("namespaces %s not found", apiRequestInfo.Name),
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

	// If the Header `Accept` is set with Table
	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{{
			Cells: []interface{}{network.Metadata.Name, network.Metadata.UID, network.Status.Phase},
		}}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the instance"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the instance (from metadata)"},
				{Name: "State", Type: "string", Format: "string", Description: "State of the instance"},
			},
			Rows: tableRow,
		}
		w.WriteAsJson(list)
		return
	}

	if network != nil {
		networkRespond := corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: *network.Metadata,
			Spec:       *network.Spec,
			Status:     *network.Status,
		}
		w.WriteAsJson(networkRespond)
		return
	}

	// print the request method and path
	w.WriteAsJson(apiRequestInfo)
}

func (n Network) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Get the network
	network, err := app.Namespace.GetNamespace(r.Request.Context(), &opencpspec.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		log.Println(err)
	}

	// Delete the network
	uuidNetwork := string(network.Metadata.UID)
	_, err = app.Namespace.DeleteNamespace(r.Request.Context(), &opencpspec.FilterOptions{Id: &uuidNetwork})
	if err != nil {
		log.Println(err)
	}

	networkRespond := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: *network.Metadata,
		Spec: *network.Spec,
		Status: *network.Status,
	}

	// print the request method and path
	w.WriteAsJson(networkRespond)
}

func (n Network) Create(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	// resolver := pkg.RequestInfoResolver()
	// apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	// if err != nil {
	// 	log.Println(err)
	// }

	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		// http.Error(w, "can't read body", http.StatusBadRequest)
		// return
	}

	namespace := &opencpspec.Namespace{}
	err = json.Unmarshal(body, &namespace)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// err = etcdObjectReader.SetStoredCustomResource("cluster", namespace.Name, namespace.Annotations[corev1.LastAppliedConfigAnnotation])
	// if err != nil {
	// 	log.Println(err)
	// }

	// Create the network
	network, err := app.Namespace.CreateNamespace(r.Request.Context(), namespace)
	if err != nil {
		log.Println(err)
	}

	networkRespond := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: *network.Metadata,
		Spec:       *network.Spec,
		Status:     *network.Status,
	}

	// print the request method and path
	w.WriteAsJson(networkRespond)
}
