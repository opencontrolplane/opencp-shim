package core

import (
	// "log"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	opencpAPI *restful.WebService
)

type Core struct {
	Network   NetworkInterface
	Secret    SecretInterface
	APIServer APIInterface
}

func NewCore() *Core {
	return &Core{
		Network:   NewNetwork(),
		Secret:    NewSecret(),
		APIServer: NewAPIServer(),
	}
}

func (c Core) API() []*restful.WebService {
	api := new(restful.WebService).Path("/api").Consumes(restful.MIME_JSON, "application/yml").Produces(restful.MIME_JSON, "application/yml")
	api.Route(api.GET("").To(c.APIServer.APIServer).
		//Doc
		Doc("Available API versions").
		Metadata(restfulspec.KeyOpenAPITags, []string{"core"}).
		Operation("apierver").
		Writes(metav1.APIVersions{}).
		Returns(http.StatusOK, "OK", metav1.APIVersions{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	api.Route(api.GET("/v1").To(c.APIServer.ResourceList).
		//Doc
		Doc("get available resources").Operation("resourceList").
		Writes(metav1.APIResourceList{}).
		Returns(http.StatusOK, "OK", metav1.APIResourceList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))

	// Namespaces API
	api.Route(api.GET("/v1/namespaces").To(c.Network.List).
		//Doc
		Doc("list or watch objects of kind Namespace").Operation("listNamespace").
		Writes(corev1.NamespaceList{}).
		Returns(http.StatusOK, "OK", corev1.NamespaceList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	api.Route(api.GET("/v1/namespaces/{namespace}").To(c.Network.Get).
		//Doc
		Doc("Get a single Namespace").Operation("getNamespace").
		Param(api.PathParameter("namespace", "name of the Namespace").DataType("string")).
		Writes(corev1.Namespace{}).
		Returns(http.StatusOK, "OK", corev1.Namespace{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	api.Route(api.DELETE("/v1/namespaces/{namespace}").To(c.Network.Delete).
		//Doc
		Doc("Delete a single Namespace").Operation("deleteNamespace").
		Writes(corev1.Namespace{}).
		Param(api.PathParameter("namespace", "name of the Namespace").DataType("string")).
		Returns(http.StatusOK, "OK", corev1.Namespace{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	api.Route(api.POST("/v1/namespaces").To(c.Network.Create).
		//Doc
		Doc("list or watch objects of kind Namespace").Operation("createNamespace").
		Writes(corev1.Namespace{}).
		Returns(http.StatusOK, "OK", corev1.Namespace{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))

	// Events API
	api.Route(api.GET("/v1/namespaces/{namespace}/events").To(c.APIServer.Events).
		//Doc
		Doc("Get events for a object").Operation("listEventsForANamespaces").
		Param(api.PathParameter("namespace", "name of the Namespace").DataType("string")).
		Writes(corev1.EventList{}).
		Returns(http.StatusOK, "OK", corev1.EventList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))

	// Secrets API
	api.Route(api.GET("/v1/namespaces/{namespace}/secrets/{secret}").To(c.Secret.Get).
		//Doc
		Doc("Get a secret in a namespace").Operation("SecretByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"k8sIo_v1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		Param(opencpAPI.PathParameter("secret", "the secret name").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "k8s.io", Version: "v1", Kind: "Secret"}).
		Writes(corev1.Secret{}).
		Returns(http.StatusOK, "OK", corev1.Secret{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))

	return []*restful.WebService{api}
}

func (c Core) Version() []*restful.WebService {
	version := new(restful.WebService).Path("/version").Consumes(restful.MIME_JSON, "application/yml").Produces(restful.MIME_JSON, "application/yml")
	version.Route(version.GET("").To(c.APIServer.Version))

	return []*restful.WebService{version}
}
