package gofb2

import "strings"

type parseErrors []error

func (pe parseErrors) Error() string {
	var b strings.Builder
	s := make([]string, len(pe))
	for i, e := range pe {
		s[i] = e.Error()
	}
	b.WriteString("[")
	b.WriteString(strings.Join(s, ", "))
	b.WriteString("]")
	return b.String()
}
