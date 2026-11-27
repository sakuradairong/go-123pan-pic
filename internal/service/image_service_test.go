package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"imagehost/internal/pan123"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGetImageItemsFollowsCursor(t *testing.T) {
	for _, end := range []string{"-1", "", "cursor"} {
		t.Run("end="+end, func(t *testing.T) {
			oldTransport := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = oldTransport })
			pages := 0
			http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				var body string
				switch r.URL.Path {
				case "/api/v1/access_token":
					body = `{"code":0,"data":{"accessToken":"test","expiredAt":"2099-01-01T00:00:00Z"}}`
				case "/api/v1/oss/file/list":
					var req pan123.FileListReq
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						t.Fatal(err)
					}
					pages++
					switch {
					case pages == 1 && req.LastFileID == "":
						body = `{"code":0,"data":{"fileList":[{"filename":"a#1%.png"}],"lastFileId":"cursor"}}`
					case pages == 2 && req.LastFileID == "cursor":
						body = `{"code":0,"data":{"fileList":[{"userSelfURL":"https://cdn.example/b.png"}],"lastFileId":"` + end + `"}}`
					default:
						t.Fatalf("unexpected page %d with cursor %q", pages, req.LastFileID)
					}
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})

			s := NewImageService(pan123.NewClient("id", "secret"), "", "https://img.example/")
			items, err := s.GetImageItems()
			if end == "cursor" {
				if err == nil || items != nil {
					t.Fatal("expected a pagination error without partial results")
				}
				return
			}
			if err != nil || len(items) != 2 {
				t.Fatalf("got %d items, err=%v", len(items), err)
			}
			if items[0].UserSelfURL != "https://img.example/a%231%25.png" || items[1].UserSelfURL != "https://cdn.example/b.png" {
				t.Fatalf("unexpected links: %+v", items)
			}
		})
	}
}
