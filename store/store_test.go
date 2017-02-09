package store

import (
	"os"
	"testing"

	"github.com/chrisolsen/ae/testutils"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

var T = testutils.T{}

func TestMain(m *testing.M) {
	os.Exit(func() int {
		code := m.Run()
		T.Close()
		return code
	}())
}

func TestDeleteNonExistingRecord(t *testing.T) {
	c := T.GetContext()
	b := Base{}
	key := datastore.NewIncompleteKey(c, "users", nil)
	err := b.Delete(c, key)
	if err == nil {
		t.Errorf("ErrNoSuchEntity expsted")
		return
	}
}

func TestDeleteFromDatastore(t *testing.T) {
	c := T.GetContext()
	b := Base{TableName: "users"}

	type person struct {
		Name string
	}

	// create person
	p := person{Name: "Jim"}
	key, err := b.Create(c, &p, nil)
	if err != nil {
		t.Errorf("failed to create person: %v", err)
		return
	}

	// verify creatio
	var pverify person
	_, err = b.Get(c, key, &pverify)
	if err != nil {
		t.Errorf("can't get person: %v", err)
		return
	}

	// test exists in memcache
	var cperson person
	_, err = memcache.Gob.Get(c, key.Encode(), &cperson)
	if err != nil {
		t.Errorf("failed to fetch from memcache: %v", err)
		return
	}
	if cperson.Name != p.Name {
		t.Errorf("invalid cache match: %v", err)
		return
	}

	// delete user
	err = b.Delete(c, key)
	if err != nil {
		t.Errorf("failed to delete person: %v", err)
		return
	}

	// test cleared from db
	err = datastore.Get(c, key, &pverify)
	if err != datastore.ErrNoSuchEntity {
		t.Errorf("person still exists in database: %v", err)
		return
	}

	// test cleared from memcache
	_, err = memcache.Get(c, key.Encode())
	if err != memcache.ErrCacheMiss {
		t.Errorf("should no longer exist in memcache: %v", err)
		return
	}
}

func TestUpdate(t *testing.T) {
	c := T.GetContext()
	b := Base{TableName: "users"}

	type person struct {
		Name string
	}

	// create person
	p1 := person{Name: "Jim"}
	key, err := b.Create(c, &p1, nil)
	if err != nil {
		t.Errorf("failed to create person: %v", err)
		return
	}

	p1.Name = "Sam"
	err = b.Update(c, key, &p1)
	if err != nil {
		t.Errorf("failed to update person: %v", err)
		return
	}

	// test deleted from memcache
	var p2 person
	_, err = memcache.Gob.Get(c, key.Encode(), &p2)
	if err != memcache.ErrCacheMiss {
		t.Errorf("failed to delete from memcache: %v", err)
		return
	}

	// test updated in db
	var p3 person
	err = datastore.Get(c, key, &p3)
	if err != nil {
		t.Errorf("person should still exist in db: %v", err)
		return
	}
	if p3.Name != "Sam" {
		t.Errorf("person not updated in db")
		return
	}
}

func TestGetNilVals(t *testing.T) {
	c := T.GetContext()
	s := Base{TableName: "users"}
	key := datastore.NewIncompleteKey(c, "users", nil)

	type user struct {
		Name string
	}
	var u user
	_, err := s.Get(c, key, &u)
	if err == nil {
		t.Error("ErrNoSuchEntity expected")
		return
	}
}

func TestGetFromMemcache(t *testing.T) {
	c := T.GetContext()
	s := Base{TableName: "users"}
	key := datastore.NewIncompleteKey(c, "users", nil)

	type user struct {
		Name string
	}
	var u user
	key, err := s.Create(c, &u, nil)
	if err != nil {
		t.Errorf("error on create: %v", err)
		return
	}

	// prime memcache
	_, err = s.Get(c, key, &u)
	if err != nil {
		t.Errorf("error on priming get: %v", err)
		return
	}

	// should now retrieve from memcache
	_, err = s.Get(c, key, &u)
	if err != nil {
		t.Errorf("error on memcache get: %v", err)
		return
	}
}
