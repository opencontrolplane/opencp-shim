package middleware

import (
	"net/http"
	"strings"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	log "github.com/sirupsen/logrus"
)

type responseRecorder struct {
	http.ResponseWriter
	Status int
	Length int
}

func Logging(r *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	recorder := &responseRecorder{
		ResponseWriter: resp,
		Length:         0,
	}

	fields := log.Fields{
		"remote_address": requestGetRemoteAddress(r.Request),
		"user_agent":     r.Request.UserAgent(),
		"request_id":     r.Request.Header.Get("X-Request-Id"),
	}

	log.WithFields(fields).Infof("Request received for %s %s", r.Request.Method, r.Request.RequestURI)

	fields = log.Fields{
		"status":     recorder.Status,
		"size":       recorder.Length,
		"time_taken": time.Since(start),
	}

	log.WithFields(fields).Infof("Request completed for %s %s", r.Request.Method, r.Request.RequestURI)
	chain.ProcessFilter(r, resp)
}


// RequestGetRemoteAddress returns ip address of the client making the request,
// taking into account http proxies
func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts[0]
	}
	return hdrRealIP
}

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}
