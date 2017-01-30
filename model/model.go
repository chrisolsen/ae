package model

import "google.golang.org/appengine/datastore"

// Base has the common key property
type Base struct {
	Key *datastore.Key `json:"key" datastore:"-"`
}

// SetKey allows the key of a model to be auto set
func (b *Base) SetKey(k *datastore.Key) {
	b.Key = k
}
