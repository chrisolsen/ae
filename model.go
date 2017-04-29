package ae

import (
	"google.golang.org/appengine/datastore"
)

type ErrModelValidation struct {
	Message string
}

func NewValidationError(msg string) ErrModelValidation {
	return ErrModelValidation{Message: msg}
}

func (mv ErrModelValidation) Error() string {
	return mv.Message
}

// Model has the common key property
type Model struct {
	Key *datastore.Key `json:"key" datastore:"-"`
}
