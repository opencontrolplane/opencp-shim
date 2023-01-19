package apis

import (
	restful "github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApisInterface interface {
	List(r *restful.Request, w *restful.Response)
}

type APIGroupModel struct {
}

func NewAPIGroupModel() *APIGroupModel {
	return &APIGroupModel{}
}

func (a APIGroupModel) List(r *restful.Request, w *restful.Response) {
	groupList := metav1.APIGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIGroupList",
			APIVersion: "v1",
		},
		Groups: []metav1.APIGroup{
			{
				Name: "opencp.io",
				Versions: []metav1.GroupVersionForDiscovery{
					{
						GroupVersion: "opencp.io/v1alpha1",
						Version:      "v1alpha1",
					},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{
					GroupVersion: "opencp.io/v1alpha1",
					Version:      "v1alpha1",
				},
			},
		},
	}

	w.WriteAsJson(groupList)
}
