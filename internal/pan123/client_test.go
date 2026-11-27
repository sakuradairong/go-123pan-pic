package pan123

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRawPUTErrorOmitsSignature(t *testing.T) {
	for _, transportError := range []bool{false, true} {
		c := NewClient("", "")
		c.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if transportError {
				return nil, fmt.Errorf("request failed at %s", r.URL)
			}
			return &http.Response{StatusCode: 403, Body: io.NopCloser(strings.NewReader(r.URL.String()))}, nil
		})
		err := c.DoRawPUT("https://store.example/upload?sig=private", strings.NewReader("x"), 1)
		if err == nil || strings.Contains(err.Error(), "private") {
			t.Fatalf("unexpected PUT result: %v", err)
		}
	}
}

func TestGetTokenRejectsInvalidResponse(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
	}{
		{403, `{"code":0,"data":{"accessToken":"ignored"}}`},
		{200, `{"code":0,"data":{}}`},
	} {
		c := NewClient("", "")
		c.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		if token, err := c.getToken(); err == nil || token != "" {
			t.Fatalf("status %d: unexpected token result (%q, %v)", tc.status, token, err)
		}
	}
}

func TestDeleteRequiresSuccessCode(t *testing.T) {
	for _, body := range []string{`{}`, `{"code":null}`, `{"code":1}`, `{"code":0}`} {
		c := NewClient("", "")
		c.token, c.expiredAt = "test", time.Now().Add(time.Hour)
		c.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		if err := c.DeleteFiles([]string{"id"}); (err == nil) != (body == `{"code":0}`) {
			t.Fatalf("delete response %s: %v", body, err)
		}
	}
}
