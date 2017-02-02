package testutils

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

type T struct {
	inst aetest.Instance
}

func (t *T) GetContext() context.Context {
	inst := t.GetInstance()
	r, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil
	}
	return appengine.NewContext(r)
}

func (t *T) GetInstance() aetest.Instance {
	if t.inst == nil {
		t.inst, _ = aetest.NewInstance(nil)
	}

	return t.inst
}

func (t *T) Close() {
	if t.inst != nil {
		t.inst.Close()
	}
}
