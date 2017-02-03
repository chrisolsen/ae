package model

import "google.golang.org/appengine/datastore"

// Base has the common key property
type Base struct {
	Key *datastore.Key `json:"key" datastore:"-"`
}
