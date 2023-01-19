package core

import (
	// "log"
	"net/http"

	// "git.civo.com/alejandro/api-v3/pkg"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// opencpspec "github.com/opencontrolplane/opencp-spec/v1alpha1"
)

type Core struct {
	// EtcdClient *clientv3.Client
	Network   NetworkInterface
	// Secret    SecretInterface
	APIServer APIInterface
}

func NewCore(etcdClient *clientv3.Client) *Core {
	// DB Init
	// db, err := pkg.NewSecretService()
	// if err != nil {
	// 	log.Println(err)
	// }

	return &Core{
		Network:   NewNetwork(etcdClient),
		// Secret:    NewSecret(etcdClient, db),
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
	// api.Route(api.GET("/v1/secrets").To(c.Secret.List).
	// 	//Doc
	// 	Doc("list or watch objects of kind Secret").Operation("listSecretForAllNamespaces").
	// 	Writes(corev1.SecretList{}).
	// 	Returns(http.StatusOK, "OK", corev1.SecretList{}).
	// 	Returns(http.StatusUnauthorized, "Unauthorized", nil))
	// api.Route(api.GET("/v1/namespaces/{namespace}/secrets").To(c.Secret.List).
	// 	//Doc
	// 	Doc("list or watch objects of kind Secret").Operation("listSecretForAllNamespaces").
	// 	Param(api.PathParameter("namespace", "name of the Namespace").DataType("string")).
	// 	Writes(corev1.SecretList{}).
	// 	Returns(http.StatusOK, "OK", corev1.SecretList{}).
	// 	Returns(http.StatusUnauthorized, "Unauthorized", nil))
	// api.Route(api.GET("/v1/namespaces/{namespace}/secrets/{secret}").To(c.Secret.Get).
	// 	//Doc
	// 	Doc("Get a single Secret").Operation("getSecret").
	// 	Param(api.PathParameter("namespace", "name of the Namespace").DataType("string")).
	// 	Param(api.PathParameter("secret", "name of the Secret").DataType("string")).
	// 	Writes(corev1.Secret{}).
	// 	Returns(http.StatusOK, "OK", corev1.Secret{}).
	// 	Returns(http.StatusUnauthorized, "Unauthorized", nil))
		
	return []*restful.WebService{api}
}

func (c Core) Version() []*restful.WebService {
	version := new(restful.WebService).Path("/version").Consumes(restful.MIME_JSON, "application/yml").Produces(restful.MIME_JSON, "application/yml")
	version.Route(version.GET("").To(c.APIServer.Version))

	return []*restful.WebService{version}
}
