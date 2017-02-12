package store

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Getter interface {
	Get(c context.Context, key *datastore.Key, dst interface{}) (*datastore.Key, error)
}

type Updater interface {
	Update(c context.Context, key *datastore.Key, src interface{}) error
}

type Creater interface {
	Create(c context.Context, src interface{}, parent *datastore.Key) (*datastore.Key, error)
}

type Copier interface {
	Copy(c context.Context, srcKey *datastore.Key, dst interface{}) error
}

type Deleter interface {
	Delete(c context.Context, key *datastore.Key) error
}
