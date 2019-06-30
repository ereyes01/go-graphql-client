package gqlclient

import "fmt"

type GraphqlErrLoc struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type GraphqlError struct {
	Status    int             `json:"status"`
	Message   string          `json:"message"`
	Locations []GraphqlErrLoc `json:"locations"`
}

func (e GraphqlError) String() string {
	var s string

	for idx, loc := range e.Locations {
		if idx != 0 {
			s += ","
		}
		s += fmt.Sprintf("%d:%d", loc.Line, loc.Column)
	}

	if len(e.Locations) > 0 {
		return fmt.Sprintf("(status:%d) %s: %s", e.Status, s, e.Message)
	}
	return fmt.Sprintf("(status:%d) %s", e.Status, e.Message)
}

type GraphqlErrors []GraphqlError

func (e GraphqlErrors) Error() string {
	if len(e) == 0 {
		return "nil"
	}
	if len(e) == 1 {
		return e[0].String()
	}

	var s string

	for idx, err := range e {
		s += fmt.Sprintf("[%d] %s\n", idx, err.String())
	}

	return s
}
