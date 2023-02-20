package middleware

import (
	"time"

	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func AddHeaders(r *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	resp.Header().Set("Cache-Control", "no-cache, private")
	resp.Header().Set("Date", time.Now().Format(time.RFC1123))
	resp.Header().Set("X-Kubernetes-Pf-Flowschema-Uid", string(uuid.NewUUID()))
	resp.Header().Set("X-Kubernetes-Pf-Prioritylevel-Uid", string(uuid.NewUUID()))
	resp.PrettyPrint(false)

	chain.ProcessFilter(r, resp)
}
