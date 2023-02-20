package middleware

import (
	restful "github.com/emicklei/go-restful/v3"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	gorestfulmiddleware "github.com/slok/go-http-metrics/middleware/gorestful"
)

func Metrics() restful.FilterFunction {
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	return gorestfulmiddleware.Handler("", mdlw)
}
