package forms

type errors map[string][]string

// Add appends to e the input message to e[field]
func (e errors) Add(field string, message string) {
	e[field] = append(e[field], message)
}

// Get retrieves the first value of errors[field]
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
