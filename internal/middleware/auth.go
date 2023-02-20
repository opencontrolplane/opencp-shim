package middleware

import (
	"regexp"
	"strings"

	restful "github.com/emicklei/go-restful/v3"
	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	grpcMetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Authenticate(r *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// Check the header User-Agent to see if it's kubectl
	if !strings.Contains(r.HeaderParameter("User-Agent"), "kubectl") {
		resp.WriteErrorString(401, "401: Not Authorized or not a valid client")
		return
	}

	var apiKey string
	tokens, ok := r.Request.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		re := regexp.MustCompile(`(?i)bearer\s+`)
		apiKey = re.ReplaceAllString(tokens[0], "")
	}

	// Init the auth client
	// get the app from the request attribute
	app := r.Attribute("app").(*setup.OpenCPApp)

	// Modify the ctx to add the token to check if the token is valid
	ctx := r.Request.Context()
	ctx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "bearer "+apiKey)

	//Call the auth service to check if the token is valid
	validToken, err := app.LoginClient.Check(ctx, &opencpspec.LoginRequest{Token: apiKey})
	if err != nil {
		s, _ := status.FromError(err)
		resp.WriteErrorString(500, s.Message())
		return
	}

	if !validToken.Valid {
		resp.WriteErrorString(401, "401: Not Authorized")
		return
	}

	r.Request = r.Request.WithContext(ctx)
	app.Token = apiKey
	// r.SetAttribute("app", app)
	chain.ProcessFilter(r, resp)
}
