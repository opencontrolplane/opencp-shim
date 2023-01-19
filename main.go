package main

import (
	"log"
	"net/http"
	"os"

	// "git.civo.com/alejandro/api-v3/pkg"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	middleware "github.com/opencontrolplane/opencp-shim/internal/middleware"
	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	openapi "github.com/opencontrolplane/opencp-shim/internal/openapi"

	// API
	apis "github.com/opencontrolplane/opencp-shim/services/apis"
	core "github.com/opencontrolplane/opencp-shim/services/core"
	opencp "github.com/opencontrolplane/opencp-shim/services/opencp"
	// "k8s.io/apiserver/pkg/storage/storagebackend"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	app := setup.NewAOpenCP()

	// Config
	app.Config = setup.Config("config.yaml")

	// Disable for now as we are not using it
	// etcdClient, err := setup.Etcd(app.Config)
	// if err != nil {
	// 	log.Println(err)
	// }
	// app.EtcdClient = etcdClient

	app.LoginClient = setup.Login(app.Config)
	app.VirtualMachine = setup.VirtualMachine(app.Config)
	app.KubernetesCluster = setup.KubernetesCluster(app.Config)
	app.Namespace = setup.Namespace(app.Config)
	app.Domain = setup.Domain(app.Config)
	app.SSHkey = setup.SSHKey(app.Config)
	app.Firewall = setup.Firewall(app.Config)

	// We add the app as attribute to the request
	restful.DefaultContainer.Filter(func(r *restful.Request, w *restful.Response, chain *restful.FilterChain) {
		r.SetAttribute("app", app)
		chain.ProcessFilter(r, w)
	})

	// Service
	coreService := core.NewCore()
	apisService := apis.NewAPIGroup()
	opencpService := opencp.NewOpenCP()

	allWebservice := []*restful.WebService{}
	allWebservice = append(allWebservice, coreService.API()...)
	allWebservice = append(allWebservice, coreService.Version()...)
	allWebservice = append(allWebservice, apisService.APIS()...)
	allWebservice = append(allWebservice, opencpService.OpenCP()...)

	// Register the API
	for _, ws := range allWebservice {
		restful.DefaultContainer.Add(ws)
	}

	// OPENAPI
	config := restfulspec.Config{
		WebServices:                   restful.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/openapi/v2",
		DisableCORS:                   true,
		PostBuildSwaggerObjectHandler: openapi.SwaggerObject,
	}
	openAPIv2 := openapi.NewOpenAPIService(config)
	restful.DefaultContainer.Add(openAPIv2)

	// Added the filter
	restful.DefaultContainer.Filter(middleware.Metrics())
	restful.DefaultContainer.Filter(middleware.Authenticate)
	restful.DefaultContainer.Filter(middleware.AddHeaders)
	restful.DefaultContainer.Filter(middleware.Logging)

	// Serve our metrics.
	go func() {
		log.Printf("metrics listening at %s", "8081")
		if err := http.ListenAndServe(":8081", promhttp.Handler()); err != nil {
			log.Panicf("error while serving metrics: %s", err)
		}
	}()

	// This ssl is just for development
	if os.Getenv("SSL") == "true" {
		log.Println("Starting server with ssl on :4000")
		err := http.ListenAndServeTLS(":4000", "ssl/server.crt", "ssl/server.key", nil)
		log.Fatal(err)
	} else {
		log.Println("Starting server on :4000")
		err := http.ListenAndServe(":4000", nil)
		log.Fatal(err)
	}
}
