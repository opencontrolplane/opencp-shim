package core

import (
	"log"
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	"github.com/opencontrolplane/opencp-shim/pkg"
	opencpgrpc "github.com/opencontrolplane/opencp-spec/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretInterface interface {
	Get(r *restful.Request, w *restful.Response)
}

type Secret struct {
}

func NewSecret() SecretInterface {
	return &Secret{}
}

// Get is the function that will return the secret
func (s *Secret) Get(r *restful.Request, w *restful.Response) {
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

	coreSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Metadata.Name,
			Namespace: cluster.Metadata.Namespace,
			CreationTimestamp: metav1.Time{
				Time: cluster.Metadata.CreationTimestamp.Time,
			},
		},
		Data: map[string][]byte{
			"kubeconfig": []byte(cluster.Spec.Kubeconfig),
		},
		Type: "Opaque",
	}

	if coreSecret.Name == "" {
		respondStatus := pkg.RespondNotFound(apiRequestInfo)
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		w.WriteAsJson(respondStatus)
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{coreSecret.Name, coreSecret.Type, len(coreSecret.Data), pkg.TimeDiff(cluster.Metadata.CreationTimestamp.Time)}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names", Priority: 0},
				{Name: "Type", Type: "string", Format: "", Description: "Used to facilitate programmatic handling of secret data.", Priority: 0},
				{Name: "Data", Type: "string", Format: "", Description: "Data contains the secret data. Each key must consist of alphanumeric characters, '-', '_' or '.'. The serialized form of the secret data is a base64 encoded string, representing the arbitrary (possibly non-string) data value here. Described in https://tools.ietf.org/html/rfc4648#section-4", Priority: 0},
				{Name: "Age", Type: "string", Format: "", Description: "CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata", Priority: 0},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	// print the request method and path
	w.WriteAsJson(coreSecret)
}
