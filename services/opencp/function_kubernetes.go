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

type KubernetesInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type Kubernetes struct {
}

func NewKubernetes() KubernetesInterface {
	return &Kubernetes{}
}

// KubernetesList - List all the kubernetes clusters
func (k *Kubernetes) List(r *restful.Request, w *restful.Response) {
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
	var kubernetesClusterList *opencpgrpc.KubernetesClusterList
	if len(allFields) > 0 {
		nameVm := allFields["metadata.name"]
		kubernetesClusterList, err = app.KubernetesCluster.ListKubernetesCluster(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &nameVm})
		if err != nil {
			log.Println(err)
		}
	} else {
		q := apiRequestInfo.Namespace
		kubernetesClusterList, err = app.KubernetesCluster.ListKubernetesCluster(r.Request.Context(), &opencpgrpc.FilterOptions{Namespace: &q})
		if err != nil {
			log.Println(err)
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, cluster := range kubernetesClusterList.Items {
			cell := metav1.TableRow{
				Cells: []interface{}{cluster.Metadata.Name, cluster.Metadata.UID, len(cluster.Spec.Pools), cluster.Status.PublicIP, cluster.Status.State, pkg.TimeDiff(cluster.Metadata.CreationTimestamp.Time)},
				Object: runtime.RawExtension{
					Object: &metav1.PartialObjectMetadata{
						TypeMeta: metav1.TypeMeta{
							Kind:       "VirtualMachine",
							APIVersion: "opencp.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      cluster.Metadata.Name,
							UID:       cluster.Metadata.UID,
							Namespace: cluster.Metadata.Namespace,
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
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the instance"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the cluster (from metadata)"},
				{Name: "Pools", Type: "string", Format: "string", Description: "Pool count (from spec)"},
				{Name: "Public IP", Type: "string", Format: "string", Description: "Public IP of the instance"},
				{Name: "State", Type: "string", Format: "string", Description: "State of the Cluster"},
				{Name: "Age", Type: "string", Format: "date-time", Description: "Time running"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	k8sList := []v1alpha1.KubernetesCluster{}
	for _, k8s := range kubernetesClusterList.Items {
		var emptyKuberntesSpec v1alpha1.KubernetesClusterSpec
		var emptyKubernetesStatus v1alpha1.KubernetesClusterStatus

		pkg.CopyTo(k8s.Spec, &emptyKuberntesSpec)
		pkg.CopyTo(k8s.Status, &emptyKubernetesStatus)

		k8s := v1alpha1.KubernetesCluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       "KubernetesCluster",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *k8s.Metadata,
			Spec:       emptyKuberntesSpec,
			Status:     emptyKubernetesStatus,
		}

		k8sList = append(k8sList, k8s)
	}

	list := v1alpha1.KuberenetesClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubernetesClusterList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: k8sList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// KubernetesGet get a kubernetes cluster
func (k *Kubernetes) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	cluster, err := app.KubernetesCluster.GetKubernetesCluster(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name, Namespace: &apiRequestInfo.Namespace})
	if err != nil {
		log.Println(err)
		w.WriteAsJson(pkg.RespondNotFound(apiRequestInfo))
		return
	}

	if cluster == nil {
		respondStatus := metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status:  "Failure",
			Reason:  metav1.StatusReasonNotFound,
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("kubernetes cluster %s not found", apiRequestInfo.Name),
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
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{cluster.Metadata.Name, cluster.Metadata.UID, len(cluster.Spec.Pools), cluster.Status.PublicIP, cluster.Status.State, pkg.TimeDiff(cluster.Metadata.CreationTimestamp.Time)}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the instance"},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the cluster (from metadata)"},
				{Name: "Pools", Type: "string", Format: "string", Description: "Pool count (from spec)"},
				{Name: "Public IP", Type: "string", Format: "string", Description: "Public IP of the instance"},
				{Name: "State", Type: "string", Format: "string", Description: "State of the Cluster"},
				{Name: "Age", Type: "string", Format: "date-time", Description: "Time running"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	var emptyKubernetesSpec v1alpha1.KubernetesClusterSpec
	var emptyKubernetesStatus v1alpha1.KubernetesClusterStatus

	pkg.CopyTo(cluster.Spec, &emptyKubernetesSpec)
	pkg.CopyTo(cluster.Status, &emptyKubernetesStatus)

	K3sCluster := v1alpha1.KubernetesCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubernetesCluster",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *cluster.Metadata,
		Spec:       emptyKubernetesSpec,
		Status:     emptyKubernetesStatus,
	}

	// print the request method and path
	w.WriteAsJson(K3sCluster)
}

// KubernetesCreate create a kubernetes cluster
func (k *Kubernetes) Create(r *restful.Request, w *restful.Response) {
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

	kubernetesCluster := opencpgrpc.KubernetesCluster{}
	err = json.Unmarshal(body, &kubernetesCluster)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Create the cluster
	cluster, err := app.KubernetesCluster.CreateKubernetesCluster(r.Request.Context(), &kubernetesCluster)
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, kubernetesCluster.Metadata.Name, "error creating the kubernetes cluster", err)
		w.WriteAsJson(respondStatus)
		return
	}

	var emptyKubernetesSpec v1alpha1.KubernetesClusterSpec
	var emptyKubernetesStatus v1alpha1.KubernetesClusterStatus

	pkg.CopyTo(cluster.Spec, &emptyKubernetesSpec)
	pkg.CopyTo(cluster.Status, &emptyKubernetesStatus)

	// Return the cluster
	k3sCluster := v1alpha1.KubernetesCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubernetesCluster",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *cluster.Metadata,
		Spec:       emptyKubernetesSpec,
		Status:     emptyKubernetesStatus,
	}

	w.WriteAsJson(k3sCluster)
}

// KubernetesUpdate update a kubernetes cluster
func (k *Kubernetes) Update(r *restful.Request, w *restful.Response) {
	// Not implemented
	// a json respond is returned
	json := `{"message": "Not implemented"}`
	w.WriteAsJson(json)
}

// KubernetesDelete delete a kubernetes cluster
func (k *Kubernetes) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	// Send to delete the cluster
	cluster, err := app.KubernetesCluster.DeleteKubernetesCluster(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &apiRequestInfo.Name})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, apiRequestInfo.Name, "error deleteing the kubernetes cluster", err)
		w.WriteAsJson(respondStatus)
		return
	}

	if cluster != nil {
		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: metav1.StatusSuccess,
			Details: &metav1.StatusDetails{
				Name:  cluster.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   cluster.Metadata.UID,
			},
		}
	}

	if cluster == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}

	// print the request method and path
	w.WriteAsJson(respondStatus)
}
