package core

import (
	"log"
	"strings"

	goruntime "runtime"
	"time"

	"github.com/opencontrolplane/opencp-shim/pkg"
	// "github.com/civo/civogo"
	restful "github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

type APIInterface interface {
	Version(r *restful.Request, w *restful.Response)
	APIServer(r *restful.Request, w *restful.Response)
	ResourceList(r *restful.Request, w *restful.Response)
	Events(r *restful.Request, w *restful.Response)
}

type APIServer struct {
}

func NewAPIServer() APIInterface {
	return &APIServer{}
}

func (a APIServer) Version(r *restful.Request, w *restful.Response) {
	version := version.Info{
		Major:        "1",
		Minor:        "24",
		GitVersion:   "v1.24.0",
		GitCommit:    "c2b5237ccd9c0f1d600d3072634ca66cefdf272f",
		GitTreeState: "clean",
		BuildDate:    time.Now().Format(time.RFC3339),
		GoVersion:    goruntime.Version(),
		Compiler:     "gc",
		Platform:     "linux/amd64",
	}

	w.WriteAsJson(version)
}

func (a APIServer) APIServer(r *restful.Request, w *restful.Response) {
	resolver := pkg.RequestInfoResolver()
	// print the request method and path
	log.Println(r.Request)
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	log.Println(apiRequestInfo)

	apiVersion := metav1.APIVersions{
		TypeMeta: metav1.TypeMeta{
			Kind: "APIVersions",
		},
		Versions: []string{"v1"},
		ServerAddressByClientCIDRs: []metav1.ServerAddressByClientCIDR{
			{
				ClientCIDR:    "0.0.0.0/0",
				ServerAddress: "172.17.0.3:6443",
			},
		},
	}

	w.WriteAsJson(apiVersion)
}

func (a APIServer) ResourceList(r *restful.Request, w *restful.Response) {
	resourceList := metav1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResourceList",
			APIVersion: "v1",
		},
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{
				Kind:         "Namespace",
				SingularName: "",
				Name:         "namespaces",
				Verbs:        []string{"create", "delete", "get", "list"},
				Namespaced:   false,
				ShortNames:   []string{"ns"},
			},
			{
				Kind:         "Namespace",
				SingularName: "",
				Name:         "namespaces/status",
				Verbs:        []string{"get"},
				Namespaced:   false,
			},
			{
				Kind:         "Secret",
				SingularName: "",
				Name:         "secrets",
				// Verbs:        []string{"create", "delete", "get", "list"},
				Verbs:      []string{"get", "list"},
				Namespaced: true,
				ShortNames: []string{"secret"},
			},
			{
				Kind:         "Secret",
				SingularName: "",
				Name:         "secrets/status",
				Verbs:        []string{"get"},
				Namespaced:   true,
			},
		},
	}

	// print the request method and path
	w.WriteAsJson(resourceList)
}

func (a APIServer) Events(r *restful.Request, w *restful.Response) {
	
	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	log.Println(apiRequestInfo)

	fileds := r.QueryParameter("fieldSelector")
	allFileds := strings.Split(fileds, ",")

	allEvents := make(map[string]string)
	for _, field := range allFileds {
		fieldSplit := strings.Split(field, "=")
		allEvents[fieldSplit[0]] = fieldSplit[1]
	}

	// actionFilter := &civogo.ActionListRequest{
	// 	PerPage:   100,
	// 	RelatedID: allEvents["involvedObject.uid"],
	// 	Details:   "pvc-fd5f94d7-eb92-45b1-9951-17326ee50001",
	// }
	// allActions, err := client.ListActions(actionFilter)
	// if err != nil {
	// 	log.Println(err)
	// }

	// allEvent := make([]corev1.Event, len(allActions.Items))
	// for i, action := range allActions.Items {
	// 	allEvent[i] = corev1.Event{
	// 		TypeMeta: metav1.TypeMeta{
	// 			Kind:       "Event",
	// 			APIVersion: "v1",
	// 		},
	// 		ObjectMeta: metav1.ObjectMeta{
	// 			Name: apiRequestInfo.Resource,
	// 		},
	// 		Source: corev1.EventSource{
	// 			Component: action.RelatedType,
	// 		},
	// 		Reason:  action.Type,
	// 		Message: action.Details,
	// 		FirstTimestamp: metav1.Time{
	// 			Time: action.CreatedAt,
	// 		},
	// 		LastTimestamp: metav1.Time{
	// 			Time: action.UpdatedAt,
	// 		},
	// 		Type: "Normal",
	// 	}
	// }

	// Build the Events response
	eventRespond := &corev1.EventList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventList",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{
			SelfLink:        "/api/v1/events",
			ResourceVersion: "1",
		},
		// Items: allEvent,
		Items: []corev1.Event{},
	}

	// print the request method and path
	w.WriteAsJson(eventRespond)
}
