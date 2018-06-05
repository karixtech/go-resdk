package resdk

import (
	"encoding/json"
	"errors"
	"net/http"
)

// A serializer for response in json
type JsonSerializer struct {
	// HTTP Status Code to be returned
	StatusCode int
}

// Serializes Outputable to a ResponseWriter
func (j JsonSerializer) Serialize(out Outputable, w http.ResponseWriter, r *http.Request) {
	out_b, _ := json.Marshal(out)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(j.StatusCode)
	w.Write(out_b)
	return
}

// A serializer for error response in json
type JsonErrorSerializer struct {
	// HTTP Status Code to be returned
	StatusCode int
	// If set it overrides the error message in response
	Error Outputable
}

// Serializes out to a ResponseWriter in standard error format
// If out is not Json Marshalable but is an error, its error message
// is used instead.
// Error response format: {"error": <object or error message>}
func (j JsonErrorSerializer) Serialize(out Outputable, w http.ResponseWriter, r *http.Request) {
	if j.Error != nil {
		out = j.Error
	}

	var out_obj interface{} = out

	if out_err, ok := out.(error); ok {
		if _, ok = out.(json.Marshaler); !ok {
			// If out type is error but out is not a marshaller use error string
			out_obj = map[string]interface{}{
				"error": out_err.Error(),
			}
		}
	}

	out_b, _ := json.Marshal(out_obj)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(j.StatusCode)
	w.Write(out_b)
	return
}

// A JsonErrorSerializer which serializes Not found error
type JsonNotFoundSerializer struct {
	JsonErrorSerializer
}

func (j JsonNotFoundSerializer) Serialize(out Outputable, w http.ResponseWriter, r *http.Request) {
	j.StatusCode = http.StatusNotFound
	j.Error = errors.New("Not found")
	j.JsonErrorSerializer.Serialize(out, w, r)
	return
}
