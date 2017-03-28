package ae

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

// Store is the include common attrs and methods for other *model types
type Store struct {
	TableName string
}

// NewStore is a helper to create a base store
func NewStore(tableName string) Store {
	return Store{TableName: tableName}
}

// Delete deletes the record and clears the memcached record
func (s Store) Delete(c context.Context, key *datastore.Key) error {
	err := datastore.Delete(c, key)
	if err != nil {
		return err
	}
	memcache.Delete(c, key.Encode())
	return nil
}

// Create creates the model
func (s Store) Create(c context.Context, data interface{}, parentKey *datastore.Key) (*datastore.Key, error) {
	key := datastore.NewIncompleteKey(c, s.TableName, parentKey)
	return datastore.Put(c, key, data)
}

// Update updates the model and clears the memcached data
func (s Store) Update(c context.Context, key *datastore.Key, data interface{}) error {
	_, err := datastore.Put(c, key, data)
	memcache.Delete(c, key.Encode())
	return err
}

// Get attempts to return the cached model, if no cached data exists, it then
// fetches the data from the database and caches the data
func (s Store) Get(c context.Context, key *datastore.Key, dst interface{}) (*datastore.Key, error) {
	encodedKey := key.Encode()
	_, err := memcache.Gob.Get(c, encodedKey, dst)
	if err == nil {
		return key, nil
	}
	if err != memcache.ErrCacheMiss {
		return nil, fmt.Errorf("memcache get: %v", err)
	}

	err = datastore.Get(c, key, dst)
	if err != nil {
		return nil, err
	}

	memcache.Gob.Set(c, &memcache.Item{Key: encodedKey, Object: dst})
	return key, nil
}
