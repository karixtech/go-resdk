package resthelper

import (
	"errors"
	"net/http"
)

type JsonableError interface {
	MarshalJSON() ([]byte, error)
}

type Inputable interface {
	Validate() error
}
type Authorizable interface {
	Authorize(u interface{}) error
}
type Outputable interface {
}

type Authenticatable interface {
	// Returns true if user was successfully authenticated
	// Use context to set user details for future use
	Authenticate(r *http.Request) (interface{}, error)
}
type Deserializable interface {
	// Deserilaizes a request and populate an Inputable object
	Deserialize(r *http.Request) Inputable
}
type Processable interface {
	Process(in Inputable) (Outputable, error)
}
type Serializable interface {
	Serialize(out Outputable, w http.ResponseWriter, r *http.Request)
}

// Handles GET requests for objects
type BaseHandler struct {
	// If not nil it enforces authentication
	Authenticator Authenticatable

	// Fetches validatable struct from request
	Deserializer Deserializable

	// Processes the input object to give authorizable output
	Processor Processable

	// Success response handler
	SuccessSerializer Serializable
	// Error response handlers
	ValidationErrorSerializer     Serializable
	AuthenticationErrorSerializer Serializable
	AuthorizationErrorSerializer  Serializable
	ProcessingErrorSerializer     Serializable
	NotFoundSerializer            Serializable
}

func GetAuthorizer(o Outputable) Authorizable {
	var authorizer Authorizable
	var ok bool
	if authorizer, ok = o.(Authorizable); ok {
		return authorizer
	}
	return nil
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
