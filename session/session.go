package session

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// Specifying of CacheType will make an initial cache lookup before calling
// the datstore
const (
	CacheTypeNone = iota
	CacheTypeJSON
	CacheTypeGob
)

// New returns a Session instance. A zero expiry duration will prevent expiration
func New(cacheType int, expires time.Duration) Session {
	return Session{CacheType: cacheType, Duration: expires}
}

// Session provides helper methods to get and set the account key within the request context
type Session struct {
	CacheType int
	Duration  time.Duration
}

// Keyer allows the Account() method to set the key on the passed in dst
type Keyer interface {
	SetKey(key *datastore.Key)
}

// Account fetches the account either from the cache or datastore
func (s *Session) Account(c context.Context, dst Keyer) error {
	key, err := s.AccountKey(c)
	if err != nil {
		return err
	}

	switch s.CacheType {
	case CacheTypeJSON:
		_, err = memcache.JSON.Get(c, key.Encode(), dst)
	case CacheTypeGob:
		_, err = memcache.Gob.Get(c, key.Encode(), dst)
	}

	if err == memcache.ErrCacheMiss {
		err = datastore.Get(c, key, dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		return nil
	}

	if err == nil {
		var cacheErr error
		item := memcache.Item{Key: key.Encode(), Object: dst, Expiration: s.Duration}
		switch s.CacheType {
		case CacheTypeJSON:
			cacheErr = memcache.JSON.Set(c, &item)
		case CacheTypeGob:
			cacheErr = memcache.Gob.Set(c, &item)
		}
		if cacheErr != nil {
			// log error, but don't return it
			log.Errorf(c, "cache account key: %v", err)
		}
	}

	// if err is nil account was found, otherwise return err
	return err
}

// AccountKey return the *datastore.Key value for the account
func (s *Session) AccountKey(c context.Context) (*datastore.Key, error) {
	var err error
	val := c.Value(s)
	if val == nil {
		return nil, fmt.Errorf("missing context account key: %v", err)
	}

	key, err := datastore.DecodeKey(val.(string))
	if err != nil {
		return nil, fmt.Errorf("decoding the context account key: %v", err)
	}

	return key, nil
}

// SetAccountKey sets the key in the request context to allow for later access
func (s *Session) SetAccountKey(c context.Context, key *datastore.Key) context.Context {
	return context.WithValue(c, s, key.Encode())
}
