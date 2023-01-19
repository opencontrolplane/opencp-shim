package pkg

import (
	"bytes"
	// "crypto/des"
	"encoding/json"
	"fmt"

	// "log"
	"math"
	"net/http"
	"strings"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/request"
)

// CheckHeader to see if contains table
func CheckHeader(request *restful.Request) bool {
	getHeader := request.HeaderParameter("Accept")
	return strings.Contains(getHeader, "Table")
}

func RespondNotFound(requestInfo *request.RequestInfo) metav1.Status {
	notFound := metav1.Status{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Status",
			APIVersion: "v1",
		},
		Status:  metav1.StatusFailure,
		Message: fmt.Sprintf("%s.%s \"%s\" not found", requestInfo.Resource, requestInfo.APIGroup, requestInfo.Name),
		Reason:  metav1.StatusReasonNotFound,
		Details: &metav1.StatusDetails{
			Name:  requestInfo.Name,
			Group: requestInfo.APIGroup,
			Kind:  requestInfo.Resource,
		},
		Code: http.StatusNotFound,
	}

	return notFound
}

func RespondError(requestInfo *request.RequestInfo, reason string) metav1.Status {
	notFound := metav1.Status{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Status",
			APIVersion: "v1",
		},
		Status:  string(metav1.StatusReasonInternalError),
		Message: fmt.Sprintf("%s.%s \"%s\"", requestInfo.Resource, requestInfo.APIGroup, requestInfo.Name),
		Reason:  metav1.StatusReason(reason),
		Details: &metav1.StatusDetails{
			Name:  requestInfo.Name,
			Group: requestInfo.APIGroup,
			Kind:  requestInfo.Resource,
		},
		Code: http.StatusInternalServerError,
	}

	return notFound
}

// RequestInfoResolver is a function that returns a RequestInfo object
func RequestInfoResolver() *request.RequestInfoFactory {
	return &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis"),
		GrouplessAPIPrefixes: sets.NewString("api"),
	}
}

func TimeDiff(t time.Time) string {
	diff := time.Since(t)
	days := diff / (24 * time.Hour)
	hours := diff % (24 * time.Hour)
	minutes := hours % time.Hour
	seconds := math.Mod(minutes.Seconds(), 60)
	var buffer bytes.Buffer
	if days > 0 {
		buffer.WriteString(fmt.Sprintf("%dd", days))
		return buffer.String()
	}
	if hours/time.Hour > 0 {
		buffer.WriteString(fmt.Sprintf("%dh", hours/time.Hour))
		return buffer.String()
	}
	if minutes/time.Minute > 0 {
		buffer.WriteString(fmt.Sprintf("%dm", minutes/time.Minute))
		return buffer.String()
	}
	if seconds > 0 {
		buffer.WriteString(fmt.Sprintf("%.1fs", seconds))
		return buffer.String()
	}
	return "0s"
}

// CopyTo is a helper function to copy a struct to another struct using the json marshal and unmarshal
func CopyTo(src protoreflect.ProtoMessage, dst interface{}) error {
	exportable, err := protojson.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(exportable), dst); err != nil {
		return err
	}

	return nil
}
