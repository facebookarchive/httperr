package httperr_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/daaku/go.httperr"
)

const answer = "42"

type simpleRedacter string

func (s simpleRedacter) Redact(target string) string {
	return strings.Replace(target, string(s), answer, -1)
}

func TestRedactError(t *testing.T) {
	t.Parallel()
	redacter := simpleRedacter("world")
	template := "hello %s"
	originalErr := fmt.Errorf(template, redacter)
	redactedErr := httperr.RedactError(originalErr, redacter)

	expectedStr := fmt.Sprintf(template, answer)
	actualStr := redactedErr.Error()
	if actualStr != expectedStr {
		t.Fatalf(`was expecting "%s" but got "%s"`, expectedStr, actualStr)
	}

	actualErr := redactedErr.Actual()
	if originalErr != actualErr {
		t.Fatal("did not get expected Actual reference")
	}
}

func TestWrapWithoutResponse(t *testing.T) {
	t.Parallel()
	redacter := simpleRedacter("world")
	template := "hello %s"
	originalErr := fmt.Errorf(template, redacter)
	originalReq := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "daaku.org",
			Path:   "/bar/",
		},
	}
	wrapErr := httperr.WrapError(originalErr, redacter, originalReq, nil)

	expectedStr := `GET https://daaku.org/bar/ error hello 42`
	actualStr := wrapErr.Error()
	if actualStr != expectedStr {
		t.Fatalf(`was expecting "%s" but got "%s"`, expectedStr, actualStr)
	}

	actualErr := wrapErr.Actual()
	if originalErr != actualErr {
		t.Fatal("did not get expected Actual reference")
	}

	actualReq := wrapErr.Request()
	if actualReq != originalReq {
		t.Fatal("did not get expected Request reference")
	}
}

func TestWrapWithResponse(t *testing.T) {
	t.Parallel()
	redacter := simpleRedacter("world")
	template := "hello %s"
	originalErr := fmt.Errorf(template, redacter)
	originalReq := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "daaku.org",
			Path:   "/bar/",
		},
	}
	originalRes := &http.Response{
		StatusCode: http.StatusBadGateway,
		Status:     http.StatusText(http.StatusBadGateway),
	}
	wrapErr := httperr.WrapError(originalErr, redacter, originalReq, originalRes)

	expectedStr := `GET https://daaku.org/bar/ got 502 Bad Gateway error hello 42`
	actualStr := wrapErr.Error()
	if actualStr != expectedStr {
		t.Fatalf(`was expecting "%s" but got "%s"`, expectedStr, actualStr)
	}

	actualErr := wrapErr.Actual()
	if originalErr != actualErr {
		t.Fatal("did not get expected Actual reference")
	}

	actualReq := wrapErr.Request()
	if actualReq != originalReq {
		t.Fatal("did not get expected Request reference")
	}

	actualRes := wrapErr.Response()
	if actualRes != originalRes {
		t.Fatal("did not get expected Response reference")
	}
}
