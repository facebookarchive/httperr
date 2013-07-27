// Package httperr provides HTTP errors and utilities.
package httperr

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
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

// Pass-thru to redact sensitive information from an error. strings.Replacer
// works well here.
type Redactor interface {
	// This will redact known sensitive information from the given string.
	Replace(s string) string
}

// Wrap an Error along with the associated request & response. The Redactor
// will also be applied to the final Error string.
func NewError(e error, r Redactor, req *http.Request, res *http.Response) Error {
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
	redacter Redactor
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
		e.redacter.Replace(e.request.URL.String()),
	)

	if e.response != nil {
		fmt.Fprintf(&buf, " got %s", e.response.Status)
	}

	fmt.Fprintf(&buf, " failed with %s", e.redacter.Replace(e.actual.Error()))
	return buf.String()
}

type redactError struct {
	actual   error
	redacter Redactor
}

func (e *redactError) Actual() error {
	return e.actual
}

func (e *redactError) Error() string {
	return e.redacter.Replace(e.actual.Error())
}

// Apply the Redactor to the given error.
func RedactError(e error, r Redactor) ErrorWrapper {
	return &redactError{actual: e, redacter: r}
}

type nilRedact struct{}

func (r nilRedact) Replace(s string) string {
	return s
}

var nilR nilRedact

// A no-op Redactor.
func RedactNoOp() Redactor {
	return nilR
}

type regexpRedactor struct {
	regexp *regexp.Regexp
	repl   string
}

func (r *regexpRedactor) Replace(s string) string {
	return r.regexp.ReplaceAllString(s, r.repl)
}

func RedactRegexp(re *regexp.Regexp, repl string) Redactor {
	return &regexpRedactor{regexp: re, repl: repl}
}
