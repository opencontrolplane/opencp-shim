package apis

import (
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Apis struct {
	APIGroup ApisInterface
}

func NewAPIGroup() *Apis {
	return &Apis{
		APIGroup: NewAPIGroupModel(),
	}
}

func (a Apis) APIS() []*restful.WebService {
	apis := new(restful.WebService).Path("/apis").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	apis.Route(apis.GET("").To(a.APIGroup.List).
		Doc("List all API groups").
		Operation("apisgroupList").
		Writes(v1.APIGroupList{}).
		Returns(http.StatusOK, "OK", v1.APIGroupList{}).
		Returns(http.StatusUnauthorized, "Unauthorized", nil))
	return []*restful.WebService{apis}
}
