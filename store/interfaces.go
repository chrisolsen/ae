package store

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Getter interface {
	Get(c context.Context, key *datastore.Key, dst Model) error
}

type Updater interface {
	Update(c context.Context, key *datastore.Key, model interface{}) error
}

type Creater interface {
	Create(c context.Context, model Model, parent *datastore.Key) (*datastore.Key, error)
}

type Copier interface {
	Copy(c context.Context, srcKey *datastore.Key, dst Model) error
}

type Deleter interface {
	Delete(c context.Context, key *datastore.Key) error
}
