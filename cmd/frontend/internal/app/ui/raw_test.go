package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_serveRawWithHTTPRequestMethodHEAD(t *testing.T) {
	// mockNewCommon ensures that we do not need the repo-updater running for this unit test.
	mockNewCommon = func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHandler) (*Common, error) {
		return &Common{
			Repo: &types.Repo{
				Name: "test",
			},
			CommitID: api.CommitID("12345"),
		}, nil
	}

	t.Run("success response for HEAD request", func(t *testing.T) {
		// httptest server will return a 200 OK, so gitserver.DefaultClient.RepoInfo will not return an error.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("{}"))
			if err != nil {
				t.Fatalf("Failed to write to httptest server: %v", err)
			}
		}))
		t.Cleanup(func() { server.Close() })

		gitserver.DefaultClient.Addrs = func() []string {
			return []string{strings.TrimPrefix(server.URL, "http://")}
		}

		req := httptest.NewRequest("HEAD", "/github.com/sourcegraph/sourcegraph/-/raw", nil)
		w := httptest.NewRecorder()

		err := serveRaw(w, req)
		if err != nil {
			t.Fatal(err)
		}

		if w.Code != http.StatusOK {
			t.Fatalf("Want %d but got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("failure response for HEAD request", func(t *testing.T) {
		// httptest server will return a 404 Not Found, so gitserver.DefaultClient.RepoInfo will return an error.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("{}"))
			t.Fatalf("Failed to write to httptest server: %v", err)
		}))
		t.Cleanup(func() { server.Close() })

		gitserver.DefaultClient.Addrs = func() []string {
			return []string{strings.TrimPrefix(server.URL, "http://")}
		}

		req := httptest.NewRequest("HEAD", "/github.com/sourcegraph/sourcegraph/-/raw", nil)
		w := httptest.NewRecorder()

		err := serveRaw(w, req)
		if err == nil {
			t.Fatal("Want error but got nil")
		}

		if w.Code != http.StatusNotFound {
			t.Fatalf("Want %d but got %d", http.StatusNotFound, w.Code)
		}
	})
}
