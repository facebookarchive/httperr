// Package httperr provides HTTP errors and utilities.
package httperr

import (
	"bytes"
	"fmt"
	"net/http"
)

// Wraps another error.
type ErrorWrapper interface {
	error

	// Return the error being wrapped.
	Actual() error
}

// An HTTP Error.
type Error interface {
	error

	// The associated HTTP Request. This must always be available.
	Request() *http.Request

	// The associated HTTP Response. This may be nil.
	Response() *http.Response

	// The associated error being wrapped. This must always be available.
	Actual() error
}

// Pass-thru to redact sensitive information from an error.
type Redacter interface {
	// This will redact known sensitive information from the given string.
	Redact(s string) string
}

// Wrap an Error along with the associated request & response. The Redacter
// will also be applied to the final Error string.
func WrapError(e error, r Redacter, req *http.Request, res *http.Response) Error {
	return &wrapError{
		actual:   e,
		redacter: r,
		request:  req,
		response: res,
	}
}

type wrapError struct {
	actual   error
	request  *http.Request
	response *http.Response
	redacter Redacter
}

func (e *wrapError) Request() *http.Request {
	return e.request
}

func (e *wrapError) Response() *http.Response {
	return e.response
}

func (e *wrapError) Actual() error {
	return e.actual
}

func (e *wrapError) Error() string {
	var buf bytes.Buffer
	fmt.Fprintf(
		&buf,
		"%s %s",
		e.request.Method,
		e.redacter.Redact(e.request.URL.String()),
	)

	if e.response != nil {
		fmt.Fprintf(
			&buf,
			" got %d %s",
			e.response.StatusCode,
			e.response.Status,
		)
	}

	fmt.Fprintf(&buf, " error %s", e.redacter.Redact(e.actual.Error()))
	return buf.String()
}

type redactError struct {
	actual   error
	redacter Redacter
}

func (e *redactError) Actual() error {
	return e.actual
}

func (e *redactError) Error() string {
	return e.redacter.Redact(e.actual.Error())
}

// Apply the Redacter to the given error.
func RedactError(e error, r Redacter) ErrorWrapper {
	return &redactError{actual: e, redacter: r}
}
