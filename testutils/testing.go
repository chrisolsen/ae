package testutils

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// Simplifies creating contexts for tests and allows tests to use the same
// context which will greatly speed up tests (by like a factor of 1000)
// Example Setup;
//  var T = testutils.T{}
//
//  func TestMain(m *testing.M) {
//  	os.Exit(func() int {
//  		code := m.Run()
//  		T.Close()
//  		return code
//  	}())
//  }
//
//  func TestSomething(t testing.T) {
//  	c := T.GetContext()
//  	k := ...
//  	datastore.Get(c, key, nil)
//  }

// T contains a reference to a aetest.Instance to allow for faster tests
// and closing of the test on completion
type T struct {
	inst aetest.Instance
}

// GetContext returns a new or cached context reference
func (t *T) GetContext() context.Context {
	inst := t.getInstance()
	r, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil
	}
	return appengine.NewContext(r)
}

func (t *T) getInstance() aetest.Instance {
	if t.inst == nil {
		t.inst, _ = aetest.NewInstance(nil)
	}

	return t.inst
}

// Close closes the testing instance
func (t *T) Close() {
	if t.inst != nil {
		t.inst.Close()
	}
}
