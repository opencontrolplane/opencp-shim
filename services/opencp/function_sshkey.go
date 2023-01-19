package opencp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	setup "github.com/opencontrolplane/opencp-shim/internal/setup"
	"github.com/opencontrolplane/opencp-shim/pkg"
	"github.com/opencontrolplane/opencp-spec/apis/v1alpha1"
	opencpgrpc "github.com/opencontrolplane/opencp-spec/grpc"

	restful "github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SSHKeyInterface interface {
	List(r *restful.Request, w *restful.Response)
	Get(r *restful.Request, w *restful.Response)
	Create(r *restful.Request, w *restful.Response)
	// Update(r *restful.Request, w *restful.Response)
	Delete(r *restful.Request, w *restful.Response)
}

type SSHKey struct {
}

func NewSSHKey() SSHKeyInterface {
	return &SSHKey{}
}

// List - List all sshkeys
func (s *SSHKey) List(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Check if we need filter the list
	fileds := r.QueryParameter("fieldSelector")
	allFields := make(map[string]string)
	if fileds != "" {
		filedsList := strings.Split(fileds, ",")
		for _, field := range filedsList {
			fieldSplit := strings.Split(field, "=")
			allFields[fieldSplit[0]] = fieldSplit[1]
		}
	}

	// Get all the networks again and return them
	var allSSHKey *opencpgrpc.SSHKeyList
	if len(allFields) > 0 {
		q := allFields["metadata.name"]
		sshkey, err := app.SSHkey.GetSSHKey(r.Request.Context(), &opencpgrpc.FilterOptions{
			Name: &q,
		})
		if err != nil {
			allSSHKey = &opencpgrpc.SSHKeyList{}
		}

		if sshkey != nil {
			allSSHKey.Items = append(allSSHKey.Items, sshkey)
		}
	} else {
		allSSHKey, err = app.SSHkey.ListSSHKey(r.Request.Context(), &opencpgrpc.FilterOptions{})
		if err != nil {
			respondStatus := pkg.RespondError(apiRequestInfo, "error listing sshkeys")
			w.WriteAsJson(respondStatus)
			return
		}
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		for _, sshkey := range allSSHKey.Items {
			tableRow = append(tableRow, metav1.TableRow{
				Cells: []interface{}{sshkey.Metadata.Name, sshkey.Metadata.UID, sshkey.Metadata.CreationTimestamp, sshkey.Status.State},
			})
		}

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the sshkey", Priority: 0},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the ssh key (from metadata)"},
				{Name: "Creaated", Type: "date", Format: "date", Description: "Created of the ssh key (from metadata)"},
				{Name: "Status", Type: "string", Format: "string", Description: "Status of the ssh key"},
			},
			Rows: tableRow,
		}

		w.WriteAsJson(list)
		return
	}

	sshkeyList := []v1alpha1.SSHKey{}
	for _, ssh := range allSSHKey.Items {
		sshKeySpec := v1alpha1.SSHKeySpec{}
		sshKeyStatus := v1alpha1.SSHKeyStatus{}

		pkg.CopyTo(ssh.Spec, &sshKeySpec)
		pkg.CopyTo(ssh.Status, &sshKeyStatus)

		sshkey := v1alpha1.SSHKey{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SSHKey",
				APIVersion: "opencp.io/v1alpha1",
			},
			ObjectMeta: *ssh.Metadata,
			Spec:       sshKeySpec,
			Status:     sshKeyStatus,
		}
		sshkeyList = append(sshkeyList, sshkey)
	}

	list := v1alpha1.SSHKeyList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SSHKeyList",
			APIVersion: "opencp.io/v1alpha1",
		},
		Items: sshkeyList,
	}

	// print the request method and path
	w.WriteAsJson(list)
}

// Get - Get a sshkey
func (s *SSHKey) Get(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Get the ssh key
	sshkey, err := app.SSHkey.GetSSHKey(r.Request.Context(), &opencpgrpc.FilterOptions{
		Name: &apiRequestInfo.Name,
	})
	if err != nil {
		// print the log
		log.Println(err)
	}

	if sshkey == nil {
		respondStatus := pkg.RespondNotFound(apiRequestInfo)
		// print the request method and path
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		w.WriteAsJson(respondStatus)
		return
	}

	if pkg.CheckHeader(r) {
		tableRow := []metav1.TableRow{}
		cell := metav1.TableRow{Cells: []interface{}{sshkey.Metadata.Name, sshkey.Metadata.UID, sshkey.Metadata.CreationTimestamp, sshkey.Status.State}}
		tableRow = append(tableRow, cell)

		list := metav1.Table{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Table",
				APIVersion: "meta.k8s.io/v1",
			},
			ColumnDefinitions: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name", Description: "Name of the sshkey", Priority: 0},
				{Name: "UID", Type: "string", Format: "string", Description: "UID of the ssh key (from metadata)"},
				{Name: "Creaated", Type: "date", Format: "date", Description: "Created of the ssh key (from metadata)"},
				{Name: "Status", Type: "string", Format: "string", Description: "Status of the ssh key"},
			},
			Rows: tableRow,
		}

		// print the request method and path
		w.WriteAsJson(list)
		return
	}

	sshkeySpec := v1alpha1.SSHKeySpec{}
	sshkeyStatus := v1alpha1.SSHKeyStatus{}

	pkg.CopyTo(sshkey.Spec, &sshkeySpec)
	pkg.CopyTo(sshkey.Status, &sshkeyStatus)

	sshKey := v1alpha1.SSHKey{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SSHKey",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *sshkey.Metadata,
		Spec:       sshkeySpec,
		Status:     sshkeyStatus,
	}

	// print the request method and path
	w.WriteAsJson(sshKey)
}

// DomainCreate - Create a domain
func (s *SSHKey) Create(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	// Real all the body of the request and unmarshal it
	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}

	sshkeyOpenCP := opencpgrpc.SSHKey{}
	err = json.Unmarshal(body, &sshkeyOpenCP)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Createn the sshkey
	sshKey, err := app.SSHkey.CreateSSHKey(r.Request.Context(), &sshkeyOpenCP)
	if err != nil {
		respondStatus := pkg.RespondError(apiRequestInfo, "error creating ssh key")
		w.WriteAsJson(respondStatus)
		return
	}

	sshkeySpec := v1alpha1.SSHKeySpec{}
	sshkeyStatus := v1alpha1.SSHKeyStatus{}

	pkg.CopyTo(sshKey.Spec, &sshkeySpec)
	pkg.CopyTo(sshKey.Status, &sshkeyStatus)

	sshkeyRespond := v1alpha1.SSHKey{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SSHKey",
			APIVersion: "opencp.io/v1alpha1",
		},
		ObjectMeta: *sshKey.Metadata,
		Spec:       sshkeySpec,
		Status:     sshkeyStatus,
	}
	w.WriteAsJson(sshkeyRespond)
}

// Delete - Delete a sshkey
func (s *SSHKey) Delete(r *restful.Request, w *restful.Response) {
	// Get the app config
	app := r.Attribute("app").(*setup.OpenCPApp)

	resolver := pkg.RequestInfoResolver()
	apiRequestInfo, err := resolver.NewRequestInfo(r.Request)
	if err != nil {
		log.Println(err)
	}

	respondStatus := metav1.Status{}
	q := apiRequestInfo.Name
	sshkey, err := app.SSHkey.GetSSHKey(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &q})
	if err != nil {
		respondStatus = pkg.RespondError(apiRequestInfo, "error finding sshkey")
		w.WriteAsJson(respondStatus)
		return
	}

	if sshkey != nil {
		_, err = app.SSHkey.DeleteSSHKey(r.Request.Context(), &opencpgrpc.FilterOptions{Name: &sshkey.Metadata.Name})
		if err != nil {
			respondStatus = pkg.RespondError(apiRequestInfo, "error deleting sshkey")
			w.WriteAsJson(respondStatus)
			return
		}

		respondStatus = metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: "Success",
			Details: &metav1.StatusDetails{
				Name:  sshkey.Metadata.Name,
				Group: apiRequestInfo.APIGroup,
				Kind:  apiRequestInfo.Resource,
				UID:   sshkey.Metadata.UID,
			},
		}
	}

	if sshkey == nil {
		respondStatus = pkg.RespondNotFound(apiRequestInfo)
	}

	// print the request method and path
	w.WriteAsJson(respondStatus)
}
