package auth

type contextKey string

func (c contextKey) String() string {
	return "ae-context-key" + string(c)
}
