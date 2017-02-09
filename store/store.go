package store

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

// Base is the include common attrs and methods for other *model types
type Base struct {
	TableName string
}

// Delete deletes the record and clears the memcached record
func (b Base) Delete(c context.Context, key *datastore.Key) error {
	err := datastore.Delete(c, key)
	if err != nil {
		return err
	}
	memcache.Delete(c, key.Encode())
	return nil
}

// Create creates the model
func (b Base) Create(c context.Context, data interface{}, parentKey *datastore.Key) (*datastore.Key, error) {
	key := datastore.NewIncompleteKey(c, b.TableName, parentKey)
	return datastore.Put(c, key, data)
}

// Update updates the model and clears the memcached data
func (b Base) Update(c context.Context, key *datastore.Key, data interface{}) error {
	_, err := datastore.Put(c, key, data)
	memcache.Delete(c, key.Encode())
	return err
}

// Get attempts to return the cached model, if no cached data exists, it then
// fetches the data from the database and caches the data
func (b Base) Get(c context.Context, key *datastore.Key, dst interface{}) (*datastore.Key, error) {
	encodedKey := key.Encode()
	_, err := memcache.Gob.Get(c, encodedKey, dst)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			return nil, fmt.Errorf("memcache get: %v", err)
		}
	} else {
		return key, nil
	}

	err = datastore.Get(c, key, dst)
	if err != nil {
		return nil, err
	}

	memcache.Gob.Set(c, &memcache.Item{Key: encodedKey, Object: dst})
	return key, nil
}
