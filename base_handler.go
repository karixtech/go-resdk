package resdk

import (
	"errors"
	"net/http"
)

// Set of functions which must be implemented by the deserialized
// struct containing incoming parameters.
type Inputable interface {
	// Validate the object and return an error in case of
	// validation failure. Return nil otherwise.
	// This function allows for validation of parameters in an
	// incoming request.
	Validate() error
}

// Set of functions which can be optionally implemented by an
// Outputable object which needs authorization check
type Authorizable interface {
	// Authorize access to the object based on details of an
	// authenticated user represented as auth_details.
	// auth_details is the same as returned by Authenticatable.Authenticate
	Authorize(auth_details interface{}) error
}

// Set of functions which must be implemented by the object
// being sent to the response serializer
type Outputable interface {
	// No special requirements. Any object must be serializable
	// by the serializers
}

// Manages first phase of the request lifecycle responsible for
// authentication
type Authenticatable interface {
	// Returns authentication details if user was successfully
	// authenticated. Otherwise returns an error to be sent as
	// the response.
	Authenticate(r *http.Request) (interface{}, error)
}

// Manages second phase of the request lifecycle responsible for
// deserialization.
type Deserializable interface {
	// Deserilaizes a request returns an object which should be
	// an implementation of Inputable.
	Deserialize(r *http.Request) Inputable
}

// Manages third phase of the request lifecycle responsible for
// processing.
type Processable interface {
	// Processes an Inputable which describes an object and
	// returns an Outputable object or an error. In case no object
	// is found being referred by Inputable (nil, nil) is returned.
	// Note: This is where most of business logic should go.
	Process(in Inputable) (Outputable, error)
}

// Manages fourth phase of the request lifecycle responsible for
// serialization and writing response.
type Serializable interface {
	// Serializes an Outputable object onto the response writer
	Serialize(out Outputable, w http.ResponseWriter, r *http.Request)
}

// Checks of an Outputable implements Authorizable interface
// and returns the Authorizor. Returns nil if it does not.
func GetAuthorizer(o Outputable) Authorizable {
	_ = o.(Authorizable)
	if authorizer, ok := o.(Authorizable); ok {
		return authorizer
	}
	return nil
}

// A net/http Handler implementation which sets up the basic request
// lifecycle.
type BaseHandler struct {
	// Phase I
	// Performs authentication of incoming request.
	// Set it to nil if no authentication is needed.
	Authenticator Authenticatable

	// Phase II
	// Performs deserialization of incoming request.
	// Required parameter.
	Deserializer Deserializable

	// Phase III
	// Performs the actual CRUD operations based on the
	// deserialized request and returns a serializable
	// object to be sent as response.
	Processor Processable

	// Phase IV
	// Serializes the output object returned by the Processor
	// in case of no errors.
	SuccessSerializer Serializable

	// Error response serializer in case of authentication failure
	AuthenticationErrorSerializer Serializable
	// Error response serializer in case of validation failure
	ValidationErrorSerializer Serializable
	// Error response serializer in case of processing failure
	ProcessingErrorSerializer Serializable
	// Error response serializer in case no output from Processor
	NotFoundSerializer Serializable
	// Error response serializer in case authenticated user has
	// no authority over processor output for this operation
	AuthorizationErrorSerializer Serializable
}

func (m *BaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var auth_details interface{} = nil
	// Authenticate if Authenticator was set
	if m.Authenticator != nil {
		auth_details, err = m.Authenticator.Authenticate(r)
		if err != nil {
			m.AuthenticationErrorSerializer.Serialize(err, w, r)
			return
		}
	}

	// Deserialize and validate the request
	in := m.Deserializer.Deserialize(r)
	verrors := in.Validate()
	if verrors != nil {
		m.ValidationErrorSerializer.Serialize(verrors, w, r)
		return
	}

	// Process the request to get an Outputable
	out, err := m.Processor.Process(in)
	if err != nil {
		m.ProcessingErrorSerializer.Serialize(err, w, r)
		return
	}
	if out == nil {
		// No output is treated as NotFound
		err = errors.New("Not found")
		m.NotFoundSerializer.Serialize(err, w, r)
		return
	}

	// If Outputable is also Authorizable then Authorize it
	authorizer := GetAuthorizer(out)
	if authorizer != nil {
		err = authorizer.Authorize(auth_details)
		if err != nil {
			m.AuthorizationErrorSerializer.Serialize(err, w, r)
			return
		}
	}

	m.SuccessSerializer.Serialize(out, w, r)
	return
}
