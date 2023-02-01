package opencp

import (
	// "log"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/opencontrolplane/opencp-spec/apis/v1alpha1"

	// clientv3 "go.etcd.io/etcd/client/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	opencpAPI *restful.WebService
)

type OpenCP struct {
	Kubernetes              KubernetesInterface
	Domain                  DomainInterface
	Firewall                FirewallInterface
	IP                      IPInterface
	VirtualMachine          VirtualMachineInterface
	SSHKey                  SSHKeyInterface
	ObjectStorage           ObjectStorageInterface
	ObjectStorageCredential ObjectStorageCredentialInterface
	Database                DatabaseInterface
}

func init() {
	opencpAPI = new(restful.WebService).Path("/apis/opencp.io/v1alpha1").Consumes(restful.MIME_JSON, "application/yaml").Produces(restful.MIME_JSON, "application/yaml")
}

func NewOpenCP() *OpenCP {
	return &OpenCP{
		Kubernetes:              NewKubernetes(),
		Domain:                  NewDomain(),
		Firewall:                NewFirewall(),
		IP:                      NewIP(),
		VirtualMachine:          NewVirtualMachine(),
		SSHKey:                  NewSSHKey(),
		ObjectStorage:           NewObjectStorage(),
		ObjectStorageCredential: NewObjectStorageCredential(),
		Database:                NewDatabase(),
	}
}

func (c *OpenCP) OpenCP() []*restful.WebService {

	// API Resource List
	opencpAPI.Route(opencpAPI.GET("").To(c.apiResourceListv1alpha1))

	c.KubernetesHandler()
	c.VirtualMachineHandler()
	c.IPHandler()
	c.FirewallHandler()
	c.DomainHandler()
	c.SSHKeyHandler()
	c.ObjectStorageHandler()
	c.ObjectStorageCredentialHandler()
	c.DatabaseHandler()

	return []*restful.WebService{opencpAPI}
}

// Kubernetes API Resource List
func (c *OpenCP) KubernetesHandler() {
	// Kuberenetes API
	opencpAPI.Route(opencpAPI.GET("/kubernetesclusters").To(c.Kubernetes.List).
		// Doc
		Doc("List all kubernetes cluster").Operation("KubernetesList").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Writes(v1alpha1.KuberenetesClusterList{}).
		Returns(http.StatusOK, "OK", v1alpha1.KuberenetesClusterList{}).
		// add extra metadata to the response
		AddExtension("x-kubernetes-action", "list").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "KuberenetesCluster"}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/kubernetesclusters").To(c.Kubernetes.List).
		// Doc
		Doc("List a kubernetes cluster in a namespace").Operation("KubernetesListByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		AddExtension("x-kubernetes-action", "list").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "KuberenetesCluster"}).
		Writes(v1alpha1.KuberenetesClusterList{}).
		Returns(http.StatusOK, "OK", v1alpha1.KuberenetesClusterList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/kubernetesclusters/{clustername}").To(c.Kubernetes.Get).
		// Doc
		Doc("Get a kubernetes cluster in a namespace").Operation("KubernetesGetByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		Param(opencpAPI.PathParameter("clustername", "name of the kubernetes cluster").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "KuberenetesCluster"}).
		Writes(v1alpha1.KubernetesCluster{}).
		Returns(http.StatusOK, "OK", v1alpha1.KubernetesCluster{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/namespaces/{namespace}/kubernetesclusters").To(c.Kubernetes.Create).
		// Doc
		Doc("Create a kubernetes cluster in a namespace").Operation("KubernetesCreateByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "KuberenetesCluster"}).
		Writes(v1alpha1.KubernetesCluster{}).
		Returns(http.StatusOK, "OK", v1alpha1.KubernetesCluster{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.PATCH("/namespaces/{namespace}/kubernetesclusters/{clustername}").To(c.Kubernetes.Update).
		// Doc
		Consumes(string(types.JSONPatchType), string(types.MergePatchType)).
		Doc("Update a kubernetes cluster in a namespace").Operation("KubernetesUpdateByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		Param(opencpAPI.PathParameter("clustername", "name of the kubernetes cluster").DataType("string")).
		AddExtension("x-kubernetes-action", "patch").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "KuberenetesCluster"}).
		Writes(v1alpha1.KubernetesCluster{}).
		Returns(http.StatusOK, "OK", v1alpha1.KubernetesCluster{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/namespaces/{namespace}/kubernetesclusters/{clustername}").To(c.Kubernetes.Delete).
		// Doc
		Doc("Delete a kubernetes cluster in a namespace").Operation("KubernetesDeleteByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		Param(opencpAPI.PathParameter("clustername", "name of the kubernetes cluster").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "KuberenetesCluster"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// IP API Resource List
func (c *OpenCP) IPHandler() {
	// IP
	opencpAPI.Route(opencpAPI.GET("/ips").To(c.IP.List).
		// Doc
		Doc("get all civo ip").Operation("IPList").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Writes(v1alpha1.IPList{}).
		Returns(http.StatusOK, "OK", v1alpha1.IPList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/ips/{ipname}").To(c.IP.Get).
		// Doc
		Doc("create civo IP").Operation("IPGet").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Writes(v1alpha1.IP{}).
		Param(opencpAPI.PathParameter("ipname", "name of the ip").DataType("string")).
		Returns(http.StatusOK, "OK", v1alpha1.IP{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/ips/{ipname}").To(c.IP.Delete).
		// Doc
		Doc("delete civo IP").Operation("IPDelete").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Writes(metav1.Status{}).
		Param(opencpAPI.PathParameter("ipname", "name of the ip").DataType("string")).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/ips").To(c.IP.Create).
		// Doc
		Doc("Create a civo IP").Operation("IPCreate").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Writes(v1alpha1.IP{}).
		Returns(http.StatusOK, "OK", v1alpha1.IP{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// VirtualMachine Resource List
func (c *OpenCP) VirtualMachineHandler() {
	// VirtualMachine API
	opencpAPI.Route(opencpAPI.GET("/virtualmachines").To(c.VirtualMachine.List).
		// Doc
		Doc("List all Virtual Machine").Operation("VirtualMachineList").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("VirtualMachineList").
		Writes(v1alpha1.VirtualMachineList{}).
		Returns(http.StatusOK, "OK", v1alpha1.VirtualMachineList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/virtualmachines").To(c.VirtualMachine.List).
		// Doc
		Doc("List Virtual Machine in a namespace").Operation("VirtualMachineListByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("VirtualMachineListNamespace").
		Param(opencpAPI.PathParameter("namespace", "namespace of the virtual machine cluster").DataType("string")).
		Writes(v1alpha1.VirtualMachineList{}).
		Returns(http.StatusOK, "OK", v1alpha1.VirtualMachineList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/virtualmachines/{virtualmachine}").To(c.VirtualMachine.Get).
		// Doc
		Doc("Get a virtual machone in a namespace").Operation("VirtualMachineGetByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		Param(opencpAPI.PathParameter("virtualmachine", "name of the virtual machine").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "VirtualMachine"}).
		Writes(v1alpha1.VirtualMachine{}).
		Returns(http.StatusOK, "OK", v1alpha1.VirtualMachine{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/namespaces/{namespace}/virtualmachines").To(c.VirtualMachine.Create).
		// Doc
		Doc("Create a virtual machine in a namespace").Operation("VirtualMachineCreateByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the virtual machine").DataType("string")).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "VirtualMachine"}).
		Writes(v1alpha1.VirtualMachine{}).
		Returns(http.StatusOK, "OK", v1alpha1.VirtualMachine{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/namespaces/{namespace}/virtualmachines/{virtualmachine}").To(c.VirtualMachine.Delete).
		// Doc
		Doc("Delete a virtual machine in a namespace").Operation("VirtualMachineDeleteByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the kubernetes cluster").DataType("string")).
		Param(opencpAPI.PathParameter("virtualmachine", "name of the virtual machine").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "VirtualMachine"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// Firewall Resource List
func (c *OpenCP) FirewallHandler() {
	opencpAPI.Route(opencpAPI.GET("/firewalls").To(c.Firewall.List).
		// Doc
		Doc("List all Firewalls").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("FirewallsList").
		Writes(v1alpha1.FirewallList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Firewall"}).
		Returns(http.StatusOK, "OK", v1alpha1.FirewallList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/firewalls").To(c.Firewall.List).
		// Doc
		Doc("List all Firewalls in a namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("FirewallsListNamespace").
		Param(opencpAPI.PathParameter("namespace", "namespace of the firewall").DataType("string")).
		Writes(v1alpha1.FirewallList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Firewall"}).
		Returns(http.StatusOK, "OK", v1alpha1.FirewallList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/firewalls/{firewall}").To(c.Firewall.Get).
		// Doc
		Doc("Get a firewall in a namespace").Operation("FirewallsGetByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the firewall").DataType("string")).
		Param(opencpAPI.PathParameter("firewall", "name of the firewall").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Firewall"}).
		Writes(v1alpha1.Firewall{}).
		Returns(http.StatusOK, "OK", v1alpha1.Firewall{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/namespaces/{namespace}/firewalls").To(c.Firewall.Create).
		// Doc
		Doc("Create a firewall in a namespace").Operation("FirewallCreateByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the firewall").DataType("string")).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Firewall"}).
		Writes(v1alpha1.Firewall{}).
		Returns(http.StatusOK, "OK", v1alpha1.Firewall{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/namespaces/{namespace}/firewalls/{firewall}").To(c.Firewall.Delete).
		// Doc
		Doc("Delete a firewall in a namespace").Operation("FirewallDeleteByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the firewall").DataType("string")).
		Param(opencpAPI.PathParameter("firewall", "name of firewall").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Firewall"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// Domain Resource List
func (c *OpenCP) DomainHandler() {
	opencpAPI.Route(opencpAPI.GET("/domains").To(c.Domain.List).
		// Doc
		Doc("List all domains").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("DomainsList").
		Writes(v1alpha1.DomainList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Domain"}).
		Returns(http.StatusOK, "OK", v1alpha1.DomainList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/domains/{domain}").To(c.Domain.Get).
		// Doc
		Doc("Get a domain").Operation("DomainsGet").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("domain", "name of the domain").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Domain"}).
		Writes(v1alpha1.Domain{}).
		Returns(http.StatusOK, "OK", v1alpha1.Domain{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/domains").To(c.Domain.Create).
		// Doc
		Doc("Create a domain").Operation("DomainCreate").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Domain"}).
		Writes(v1alpha1.Domain{}).
		Returns(http.StatusOK, "OK", v1alpha1.Domain{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/domains/{domain}").To(c.Domain.Delete).
		// Doc
		Doc("Delete a domain").Operation("DomainDelete").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("domain", "name of domain").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Domain"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// SSHKey Resource List
func (c *OpenCP) SSHKeyHandler() {
	opencpAPI.Route(opencpAPI.GET("/sshkeys").To(c.SSHKey.List).
		// Doc
		Doc("List all ssh key").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("SSHKeyList").
		Writes(v1alpha1.SSHKeyList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "SSHKey"}).
		Returns(http.StatusOK, "OK", v1alpha1.SSHKeyList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/sshkeys/{sshkey}").To(c.SSHKey.Get).
		// Doc
		Doc("Get a ssh key").Operation("SSHKeyGet").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("sshkey", "name of the ssh key").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "SSHKey"}).
		Writes(v1alpha1.SSHKey{}).
		Returns(http.StatusOK, "OK", v1alpha1.SSHKey{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/sshkeys").To(c.SSHKey.Create).
		// Doc
		Doc("Create a ssh key").Operation("SSHKeyCreate").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "SSHKey"}).
		Writes(v1alpha1.SSHKey{}).
		Returns(http.StatusOK, "OK", v1alpha1.SSHKey{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/sshkeys/{sshkey}").To(c.SSHKey.Delete).
		// Doc
		Doc("Delete a ssh key").Operation("SSHKeyDelete").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("sshkey", "name of ssh key").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "SSHKey"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// ObjectStorage Resource List
func (c *OpenCP) ObjectStorageHandler() {
	opencpAPI.Route(opencpAPI.GET("/objectstorages").To(c.ObjectStorage.List).
		// Doc
		Doc("List all object storages").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("ObjectStorageList").
		Writes(v1alpha1.ObjectStorageList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorage"}).
		Returns(http.StatusOK, "OK", v1alpha1.ObjectStorageList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/objectstorages/{objectstorage}").To(c.ObjectStorage.Get).
		// Doc
		Doc("Get a ObjectStorage").Operation("ObjectStorageGet").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("objectstorage", "name of the objectstorage").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorage"}).
		Writes(v1alpha1.ObjectStorage{}).
		Returns(http.StatusOK, "OK", v1alpha1.ObjectStorage{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/objectstorages").To(c.ObjectStorage.Create).
		// Doc
		Doc("Create a ObjectStorage").Operation("ObjectStorageCreate").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorage"}).
		Writes(v1alpha1.ObjectStorage{}).
		Returns(http.StatusOK, "OK", v1alpha1.ObjectStorage{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/objectstorages/{objectstorage}").To(c.ObjectStorage.Delete).
		// Doc
		Doc("Delete a ObjectStorage").Operation("ObjectStorageDelete").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("objectstorage", "name of objectstorage").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorage"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// ObjectStorage Credential Resource List
func (c *OpenCP) ObjectStorageCredentialHandler() {
	opencpAPI.Route(opencpAPI.GET("/objectstoragecredentials").To(c.ObjectStorageCredential.List).
		// Doc
		Doc("List all object storages credentials").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("ObjectStorageCredentialList").
		Writes(v1alpha1.ObjectStorageCredentialList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorageCredential"}).
		Returns(http.StatusOK, "OK", v1alpha1.ObjectStorageCredentialList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/objectstoragecredentials/{credential}").To(c.ObjectStorageCredential.Get).
		// Doc
		Doc("Get a ObjectStorage Credential").Operation("ObjectStorageCredentialGet").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("credential", "name of the objectstorage credential").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorageCredential"}).
		Writes(v1alpha1.ObjectStorageCredential{}).
		Returns(http.StatusOK, "OK", v1alpha1.ObjectStorageCredential{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/objectstoragecredentials").To(c.ObjectStorageCredential.Create).
		// Doc
		Doc("Create a ObjectStorage Credential").Operation("ObjectStorageCredentialCreate").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorageCredential"}).
		Writes(v1alpha1.ObjectStorageCredential{}).
		Returns(http.StatusOK, "OK", v1alpha1.ObjectStorageCredential{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/objectstoragecredentials/{credential}").To(c.ObjectStorageCredential.Delete).
		// Doc
		Doc("Delete a ObjectStorage Credential").Operation("ObjectStorageCredentialDelete").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("credential", "name of objectstorage credential").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "ObjectStorageCredential"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}

// Database Resource List
func (c *OpenCP) DatabaseHandler() {
	opencpAPI.Route(opencpAPI.GET("/databases").To(c.Database.List).
		// Doc
		Doc("List all databases").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("DatabasesList").
		Writes(v1alpha1.DatabaseList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Database"}).
		Returns(http.StatusOK, "OK", v1alpha1.DatabaseList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/databases").To(c.Database.List).
		// Doc
		Doc("List all databases by namespaces").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Operation("DatabasesNamespacesList").
		Writes(v1alpha1.DatabaseList{}).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Database"}).
		Returns(http.StatusOK, "OK", v1alpha1.DatabaseList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.GET("/namespaces/{namespace}/databases/{database}").To(c.Database.Get).
		// Doc
		Doc("Get a database in a namespace").Operation("DatabaseGetByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the database").DataType("string")).
		Param(opencpAPI.PathParameter("database", "name of the database").DataType("string")).
		AddExtension("x-kubernetes-action", "get").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Database"}).
		Writes(v1alpha1.Database{}).
		Returns(http.StatusOK, "OK", v1alpha1.Database{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.POST("/namespaces/{namespace}/databases").To(c.Database.Create).
		// Doc
		Doc("Create a database in a namespace").Operation("DatabaseCreateByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the database").DataType("string")).
		AddExtension("x-kubernetes-action", "post").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Database"}).
		Writes(v1alpha1.Database{}).
		Returns(http.StatusOK, "OK", v1alpha1.Database{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	opencpAPI.Route(opencpAPI.DELETE("/namespaces/{namespace}/databases/{database}").To(c.Database.Delete).
		// Doc
		Doc("Delete a database in a namespace").Operation("DatabaseDeleteByNamespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{"opencpIo_v1alpha1"}).
		Param(opencpAPI.PathParameter("namespace", "namespace of the database").DataType("string")).
		Param(opencpAPI.PathParameter("firewall", "name of database").DataType("string")).
		AddExtension("x-kubernetes-action", "delete").
		AddExtension("x-kubernetes-group-version-kind", schema.GroupVersionKind{Group: "opencp.io", Version: "v1alpha1", Kind: "Database"}).
		Returns(http.StatusOK, "OK", metav1.Status{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
}
