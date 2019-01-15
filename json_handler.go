package resdk

import (
	"net/http"
)

// Extends BaseHandler with default serializers for Json responses
// Also assumes standard response status codes
// Use NewJsonHandler to use it properly
type JsonHandler struct {
	BaseHandler
}

// Creates a new JsonHandler from a BaseHandler with default serializers
func NewJsonHandler(base BaseHandler) JsonHandler {
	j := JsonHandler{
		BaseHandler: base,
	}
	j.setDefaults()
	return j
}

func (j *JsonHandler) setDefaults() {
	if j.SuccessSerializer == nil {
		j.SuccessSerializer = &JsonSerializer{StatusCode: http.StatusOK}
	}
	if j.DeserializationErrorSerializer == nil {
		j.DeserializationErrorSerializer = &JsonErrorSerializer{StatusCode: http.StatusBadRequest}
	}
	if j.ValidationErrorSerializer == nil {
		j.ValidationErrorSerializer = &JsonErrorSerializer{StatusCode: http.StatusBadRequest}
	}
	if j.AuthenticationErrorSerializer == nil {
		j.AuthenticationErrorSerializer = &JsonErrorSerializer{StatusCode: http.StatusUnauthorized}
	}
	if j.ProcessingErrorSerializer == nil {
		j.ProcessingErrorSerializer = &JsonErrorSerializer{StatusCode: http.StatusInternalServerError}
	}
	if j.NotFoundSerializer == nil {
		j.NotFoundSerializer = &JsonErrorSerializer{StatusCode: http.StatusNotFound}
	}
	if j.AuthorizationErrorSerializer == nil {
		j.AuthorizationErrorSerializer = &JsonErrorSerializer{StatusCode: http.StatusForbidden}
	}
}
