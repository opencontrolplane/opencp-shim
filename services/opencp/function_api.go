package opencp

import (
	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	restful "github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *OpenCP) apiResourceListv1alpha1(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)
	
	APIResourcesList := []metav1.APIResource{}
	for _, resource := range app.Config.ApiResource {
		APIResourcesList = append(APIResourcesList, metav1.APIResource{
			Kind:         resource.Kind,
			SingularName: resource.SingularName,
			Name:         resource.Name,
			Version:      resource.Version,
			Verbs:        resource.Verbs,
			Namespaced:   resource.Namespaced,
			ShortNames:   resource.ShortNames,
		})
	}

	resourceList := metav1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResourceList",
			APIVersion: "v1",
		},
		GroupVersion: "opencp.io/v1alpha1",
		APIResources: APIResourcesList,
	}

	w.WriteAsJson(resourceList)
}
