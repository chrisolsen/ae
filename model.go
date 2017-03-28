package ae

import "google.golang.org/appengine/datastore"

// Model has the common key property
type Model struct {
	Key *datastore.Key `json:"key" datastore:"-"`
}
