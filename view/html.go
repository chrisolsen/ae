package view

// Page packages up any data and/or message that is
// needed to be rendered to the screen for the user
type Page struct {
	Error   string
	Success string
	Warning string
	Content interface{}
}
