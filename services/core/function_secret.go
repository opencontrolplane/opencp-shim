package core

// import (
// 	"encoding/json"
// 	"errors"
// 	"log"
// 	"net/http"
// 	"strings"

// 	// "git.civo.com/alejandro/api-v3/pkg"
// 	// "github.com/civo/civogo"
// 	restful "github.com/emicklei/go-restful/v3"
// 	clientv3 "go.etcd.io/etcd/client/v3"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/types"
// )

// type SecretInterface interface {
// 	List(r *restful.Request, w *restful.Response)
// 	Delete(r *restful.Request, w *restful.Response)
// 	Create(r *restful.Request, w *restful.Response)
// 	Get(r *restful.Request, w *restful.Response)
// }

// type Secret struct {
// 	EtcdClient *clientv3.Client
// 	DB         pkg.SecretInterface
// }

// func NewSecret(etcdClient *clientv3.Client, db pkg.SecretInterface) SecretInterface {
// 	return &Secret{EtcdClient: etcdClient, DB: db}
// }

// // List all secrets
// func (s *Secret) List(r *restful.Request, w *restful.Response) {
// 	ctx := r.Request.Context()
// 	client := ctx.Value("value").(pkg.Values).GetValue("client").(*civogo.Client)

// 	resolver := pkg.RequestInfoResolver()
// 	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	// Check if we need filter the list
// 	fileds := r.QueryParameter("fieldSelector")
// 	allFields := make(map[string]string)
// 	if fileds != "" {
// 		filedsList := strings.Split(fileds, ",")
// 		for _, field := range filedsList {
// 			fieldSplit := strings.Split(field, "=")
// 			allFields[fieldSplit[0]] = fieldSplit[1]
// 		}
// 	}

// 	// Get all the networks again and return them
// 	var allSecrets []corev1.Secret
// 	var secrets []pkg.Secret

// 	if len(allFields) > 0 {
// 		err := s.DB.Find(secrets, allFields["metadata.name"], allFields["metadata.namespace"])
// 		if err != nil {
// 			if errors.Is(err, civogo.ZeroMatchesError) {
// 				allSecrets = []corev1.Secret{}
// 			}
// 		}
// 	} else {
// 		secrets, err = s.DB.List()
// 		if err != nil {
// 			respondStatus := pkg.RespondError(apiRequestInfo, "error listing secrets")
// 			w.WriteAsJson(respondStatus)
// 			return
// 		}
// 	}

// 	for _, secret := range secrets {
// 		// convert to corev1.Secret
// 		s := corev1.Secret{}
// 		json.Unmarshal([]byte(secret.Data), &s)
// 		allSecrets = append(allSecrets, s)
// 	}

// 	if pkg.CheckHeader(r) {
// 		// get the network that the user is looking for
// 		network, err := client.FindNetwork(apiRequestInfo.Namespace)
// 		if err != nil {
// 			log.Println(err)
// 		}

// 		tableRow := []metav1.TableRow{}
// 		for _, secret := range allSecrets {

// 			if apiRequestInfo.Namespace != "" {
// 				if secret.Namespace == network.Label {
// 					cell := metav1.TableRow{Cells: []interface{}{secret.Name, secret.Type, len(secret.Data), pkg.TimeDiff(secret.CreationTimestamp.Time)}}
// 					tableRow = append(tableRow, cell)
// 				}
// 			} else {
// 				cell := metav1.TableRow{
// 					Cells: []interface{}{secret.Name, secret.Type, len(secret.Data), pkg.TimeDiff(secret.CreationTimestamp.Time)},
// 					Object: runtime.RawExtension{
// 						Object: &metav1.PartialObjectMetadata{
// 							TypeMeta: metav1.TypeMeta{
// 								Kind:       "Secret",
// 								APIVersion: "v1",
// 							},
// 							ObjectMeta: metav1.ObjectMeta{
// 								Name:      secret.Name,
// 								UID:       types.UID(secret.UID),
// 								Namespace: secret.Namespace,
// 							},
// 						},
// 					},
// 				}
// 				tableRow = append(tableRow, cell)
// 			}
// 		}

// 		list := metav1.Table{
// 			TypeMeta: metav1.TypeMeta{
// 				Kind:       "Table",
// 				APIVersion: "meta.k8s.io/v1",
// 			},
// 			ColumnDefinitions: []metav1.TableColumnDefinition{
// 				{Name: "Name", Type: "string", Format: "name", Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names", Priority: 0},
// 				{Name: "Type", Type: "string", Format: "", Description: "Used to facilitate programmatic handling of secret data.", Priority: 0},
// 				{Name: "Data", Type: "string", Format: "", Description: "Data contains the secret data. Each key must consist of alphanumeric characters, '-', '_' or '.'. The serialized form of the secret data is a base64 encoded string, representing the arbitrary (possibly non-string) data value here. Described in https://tools.ietf.org/html/rfc4648#section-4", Priority: 0},
// 				{Name: "Age", Type: "string", Format: "", Description: "CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata", Priority: 0},
// 			},
// 			Rows: tableRow,
// 		}

// 		w.WriteAsJson(list)
// 		return
// 	}

// 	list := corev1.SecretList{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "SecretList",
// 			APIVersion: "v1",
// 		},
// 		Items: allSecrets,
// 	}

// 	// print the request method and path
// 	w.WriteAsJson(list)
// }

// func (s *Secret) Get(r *restful.Request, w *restful.Response) {
// 	ctx := r.Request.Context()
// 	client := ctx.Value("value").(pkg.Values).GetValue("client").(*civogo.Client)

// 	resolver := pkg.RequestInfoResolver()
// 	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	secret := &pkg.Secret{}
// 	err = s.DB.Get(secret, apiRequestInfo.Name, apiRequestInfo.Namespace)
// 	if err != nil {
// 		log.Println(err)
// 		w.WriteAsJson(pkg.RespondNotFound(apiRequestInfo))
// 		return
// 	}

// 	coreSecret := &corev1.Secret{}
// 	json.Unmarshal([]byte(secret.Data), coreSecret)

// 	if coreSecret.Name == "" {
// 		respondStatus := pkg.RespondNotFound(apiRequestInfo)
// 		// print the request method and path
// 		w.ResponseWriter.WriteHeader(http.StatusNotFound)
// 		w.WriteAsJson(respondStatus)
// 		return
// 	}

// 	// get the network that the user is looking for
// 	network, err := client.FindNetwork(apiRequestInfo.Namespace)
// 	if err != nil {
// 		respondStatus := pkg.RespondError(apiRequestInfo, "error finding network")
// 		w.WriteAsJson(respondStatus)
// 		return
// 	}

// 	if pkg.CheckHeader(r) {
// 		tableRow := []metav1.TableRow{}
// 		if coreSecret.Namespace == network.Label {
// 			cell := metav1.TableRow{Cells: []interface{}{coreSecret.Name, coreSecret.Type, len(secret.Data), pkg.TimeDiff(coreSecret.CreationTimestamp.Time)}}
// 			tableRow = append(tableRow, cell)
// 		}

// 		list := metav1.Table{
// 			TypeMeta: metav1.TypeMeta{
// 				Kind:       "Table",
// 				APIVersion: "meta.k8s.io/v1",
// 			},
// 			ColumnDefinitions: []metav1.TableColumnDefinition{
// 				{Name: "Name", Type: "string", Format: "name", Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names", Priority: 0},
// 				{Name: "Type", Type: "string", Format: "", Description: "Used to facilitate programmatic handling of secret data.", Priority: 0},
// 				{Name: "Data", Type: "string", Format: "", Description: "Data contains the secret data. Each key must consist of alphanumeric characters, '-', '_' or '.'. The serialized form of the secret data is a base64 encoded string, representing the arbitrary (possibly non-string) data value here. Described in https://tools.ietf.org/html/rfc4648#section-4", Priority: 0},
// 				{Name: "Age", Type: "string", Format: "", Description: "CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata", Priority: 0},
// 			},
// 			Rows: tableRow,
// 		}

// 		// print the request method and path
// 		w.WriteAsJson(list)
// 		return
// 	}

// 	// print the request method and path
// 	w.WriteAsJson(coreSecret)
// }

// func (s *Secret) Delete(r *restful.Request, w *restful.Response) {

// }

// func (s *Secret) Create(r *restful.Request, w *restful.Response) {

// }
