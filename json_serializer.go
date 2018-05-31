package resdk

import (
	"encoding/json"
	"net/http"
)

// A serializer for response in json
type JsonSerializer struct {
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
	StatusCode int
}

// Serializes out to a ResponseWriter in standard error format
// If out is not Json Marshalable but is an error, its error message
// is used instead.
// Error response format: {"error": <object or error message>}
func (j JsonErrorSerializer) Serialize(out Outputable, w http.ResponseWriter, r *http.Request) {
	var out_obj interface{} = out
	if out_err, ok := out.(error); ok {
		if _, ok = out.(json.Marshaler); !ok {
			// If out type is error but out is not a marshaller use error string
			out_obj = out_err.Error()
		}
	}

	out_b, _ := json.Marshal(map[string]interface{}{
		"error": out_obj,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(j.StatusCode)
	w.Write(out_b)
	return
}
