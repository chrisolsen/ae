package page

type page map[string]interface{}

// New creates a new page
func New() page {
	return page(make(map[string]interface{}))
}

// SetError sets any error that needs to be shown
func (p page) SetError(err interface{}) {
	switch err.(type) {
	case string:
		p["Error"] = err.(string)
	case error:
		p["Error"] = err.(error).Error()
	default:
		p["Error"] = ""
	}
}

// SetUser sets the current user
func (p page) SetUser(user interface{}) {
	p["CurrentUser"] = user
}

// Set sets the key and value
func (p page) Set(key string, val interface{}) {
	p[key] = val
}
