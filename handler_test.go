package ae

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerDefaultsWithNoParams(t *testing.T) {
	c := HandlerConfig{}
	_ = NewHandler(&c)

	if c.LayoutFileName != defaultHandlerConfig.LayoutFileName {
		t.Error("LayoutFileName config setting not being set to default")
	}
	if c.LayoutPath != defaultHandlerConfig.LayoutPath {
		t.Error("LayoutPath config setting not being set to default")
	}
	if c.ParentLayoutName != defaultHandlerConfig.ParentLayoutName {
		t.Error("ParentLayoutName config setting not being set to default")
	}
}

func TestConfigCopy(t *testing.T) {
	c := HandlerConfig{}
	h := NewHandler(&c)
	c.LayoutFileName = "foobar.html"
	if h.config.LayoutFileName == "foobar.html" {
		t.Errorf("config params are not being duplicated")
	}
}

func TestDefaultHandler(t *testing.T) {
	dh := DefaultHandler()
	if dh.config.LayoutFileName != defaultHandlerConfig.LayoutFileName {
		t.Error("DefaultHandler not being set with default layoutFileName")
	}
	if dh.config.LayoutPath != defaultHandlerConfig.LayoutPath {
		t.Error("DefaultHandler not being set with default layoutPath")
	}
	if dh.config.ParentLayoutName != defaultHandlerConfig.ParentLayoutName {
		t.Error("DefaultHandler not being set with default parentLayoutName")
	}
	if dh.config.ViewPath != defaultHandlerConfig.ViewPath {
		t.Error("DefaultHandler not being set with default viewPath")
	}
}

func TestAddHelpersDuplicates(t *testing.T) {
	helpers := map[string]interface{}{
		"foo": "bar",
	}

	h := DefaultHandler()
	h.AddHelpers(helpers)
	if h.templateHelpers["foo"] != "bar" {
		t.Error("AddHelpers fail")
	}

	helpers["foo"] = func() string { return "bar" }
	if h.templateHelpers["foo"] != "bar" {
		t.Error("AddHelpers isn't duplicating the helper map")
	}
}

func TestAddHelper(t *testing.T) {
	helpers := map[string]interface{}{
		"foo": "bar",
	}

	h := DefaultHandler()
	h.AddHelpers(helpers)
	h.AddHelper("barMethod", func() string { return "bar" })

	fn := h.templateHelpers["barMethod"].(func() string)
	if fn() != "bar" {
		t.Error("AddHelper method fail")
	}
}

func TestToJSON(t *testing.T) {
	w := httptest.NewRecorder()
	h := Handler{Res: w}
	h.ToJSON(map[string]interface{}{"Foo": "bar"})
	if strings.Index(w.Body.String(), `{"Foo":"bar"}`) != 0 {
		t.Errorf("unexpected json response: %v", w.Body.String())
	}
	if h.Res.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type not being set to application/json")
		return
	}
}

func TestToJSONWithStatus(t *testing.T) {
	w := httptest.NewRecorder()
	h := Handler{Res: w}
	h.ToJSONWithStatus(map[string]interface{}{"Foo": "bar"}, http.StatusCreated)
	if h.Res.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type not being set to application/json")
		return
	}
	if w.Code != http.StatusCreated {
		t.Errorf("%s status received", h.Res.Header().Get("Status-Line"))
	}
}

func TestWritStatus(t *testing.T) {
	w := httptest.NewRecorder()
	h := Handler{Res: w}
	h.SendStatus(http.StatusCreated)
	if w.Code != http.StatusCreated {
		t.Error("status not written")
	}
}

func TestGetSetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	h := Handler{Res: w}
	h.SetHeader("x-Foo", "bar")
	if w.Header().Get("x-Foo") != "bar" {
		t.Error("header not written")
	}
}
