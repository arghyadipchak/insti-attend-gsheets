package errors

import (
	"fmt"
	"sort"
	"strings"
)

type Error struct {
	context map[string]string
	err     string
	details error
}

func NewError(err string, details error) *Error {
	return &Error{
		context: make(map[string]string),
		err:     err,
		details: details,
	}
}

func (e *Error) sortedContextKeys() []string {
	keys := make([]string, 0, len(e.context))
	for key := range e.context {
		keys = append(keys, key)
	}
	if len(keys) > 1 {
		sort.Strings(keys)
	}
	return keys
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	var b strings.Builder

	if e.err != "" {
		b.WriteString(e.err)
	}

	if keys := e.sortedContextKeys(); len(keys) > 0 {
		ctxParts := make([]string, len(keys))
		for i, key := range keys {
			ctxParts[i] = fmt.Sprintf("%s=%s", key, e.context[key])
		}
		b.WriteString(" [")
		b.WriteString(strings.Join(ctxParts, " "))
		b.WriteString("]")
	}

	if e.details != nil {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		b.WriteString(e.details.Error())
	}

	return b.String()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.details
}

func (e *Error) AddContext(key, value string) *Error {
	if e.context == nil {
		e.context = make(map[string]string)
	}
	e.context[key] = value
	return e
}

func (e *Error) ToArgs() []any {
	if e == nil {
		return []any{"error", "<nil>"}
	}

	args := make([]any, 0, len(e.context)*2+4)

	for _, key := range e.sortedContextKeys() {
		args = append(args, key, e.context[key])
	}

	if e.err != "" {
		args = append(args, "error", e.err)
	}

	if e.details != nil {
		args = append(args, "details", e.details)
	}

	return args
}
