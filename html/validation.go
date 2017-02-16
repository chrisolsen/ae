package html

import (
	"errors"
	"fmt"
	"html/template"
	"strings"
)

type Errors struct {
	list []error
}

func (e Errors) Add(err error) {
	e.list = append(e.list, err)
}

func (e Errors) Error() error {
	if len(e.list) == 1 {
		return e.list[0]
	}

	var list []string
	for _, err := range e.list {
		list = append(list, err.Error())
	}
	return errors.New(strings.Join(list, "---"))
}

// ToErrorList can be added to the template FuncMap to format
// the error output.
//  // handler
//  myTemplate.executeTemplate(w, "layout", template.FuncMap{
//      "error": ToErrorList,
//  })
//
//  <!-- HTML -->
//  <form ...>
//      {{ .Error | error }}
//  </form>
func ToErrorList(in interface{}) interface{} {
	var list []string

	switch in.(type) {
	case string:
		err := in.(string)
		list = strings.Split(err, "---")
	case error:
		err := in.(error)
		list = strings.Split(err.Error(), "---")
	default:
		return "Invalid type"
	}

	var lis []string
	for _, e := range list {
		lis = append(lis, fmt.Sprintf("<li>%s</li>", e))
	}

	out := fmt.Sprintf(`<ul class="errors">%s<ul>`, strings.Join(lis, "\n"))
	return template.HTML(out)
}
