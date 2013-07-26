package httperr_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/daaku/go.httperr"
)

const question = "world"
const answer = "42"

var redactor = httperr.RedactString(question, answer)

func TestRedactError(t *testing.T) {
	t.Parallel()
	template := "hello %s"
	originalErr := fmt.Errorf(template, question)
	redactedErr := httperr.RedactError(originalErr, redactor)

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
	template := "hello %s"
	originalErr := fmt.Errorf(template, question)
	originalReq := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "daaku.org",
			Path:   "/bar/",
		},
	}
	wrapErr := httperr.NewError(originalErr, redactor, originalReq, nil)

	expectedStr := `GET https://daaku.org/bar/ failed with hello 42`
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
	template := "hello %s"
	originalErr := fmt.Errorf(template, question)
	originalReq := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "daaku.org",
			Path:   "/bar/",
		},
	}
	originalRes := &http.Response{
		Status: http.StatusText(http.StatusBadGateway),
	}
	wrapErr := httperr.NewError(originalErr, redactor, originalReq, originalRes)

	expectedStr := `GET https://daaku.org/bar/ got Bad Gateway failed with hello 42`
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

func TestRedactNoOp(t *testing.T) {
	t.Parallel()
	if httperr.RedactNoOp().Redact(answer) != answer {
		t.Fatal("no op did something")
	}
}
