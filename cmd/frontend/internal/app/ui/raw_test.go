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

// initHTTPTestGitServer instantiates an httptest.Server to make it return an HTTP response as set
// by httpStatusCode and a body as set by resp. It also ensures that the server is closed during
// test cleanup, thus ensuring that the caller does not have to remember to close the server.
//
// Finally, initHTTPTestGitServer patches the gitserver.DefaultClient.Addrs to the URL of the test
// HTTP server, so that API calls to the gitserver are received by the test HTTP server.
//
// TL;DR: This function helps us to mock the gitserver without having to define mock functions for
// each of the gitserver client methods.
func initHTTPTestGitServer(t *testing.T, httpStatusCode int, resp string) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(httpStatusCode)
		_, err := w.Write([]byte(resp))
		if err != nil {
			t.Fatalf("Failed to write to httptest server: %v", err)
		}
	}))

	t.Cleanup(func() { s.Close() })

	// Strip the protocol from the URI while patching the gitserver client's addres, since the
	// gitsever implementation does not want the protocol in the address.
	gitserver.DefaultClient.Addrs = func() []string {
		return []string{strings.TrimPrefix(s.URL, "http://")}
	}
}

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
		// httptest server will return a 200 OK, so gitserver.DefaultClient.RepoInfo will not return
		// an error.
		initHTTPTestGitServer(t, http.StatusOK, "{}")

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
		// httptest server will return a 404 Not Found, so gitserver.DefaultClient.RepoInfo will
		// return an error.
		initHTTPTestGitServer(t, http.StatusNotFound, "{}")

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

func Test_serveRawWithContentTypeZip(t *testing.T) {
	mockNewCommon = func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHandler) (*Common, error) {
		return &Common{
			Repo: &types.Repo{
				Name: "test",
			},
			CommitID: api.CommitID("12345"),
		}, nil
	}

	t.Run("TODO", func(t *testing.T) {
		// httptest server will return a 200 OK, so gitserver.DefaultClient.RepoInfo will not return an error.
		initHTTPTestGitServer(t, http.StatusOK, "{}")

		req := httptest.NewRequest("GET", "/github.com/sourcegraph/sourcegraph/-/raw?format=zip", nil)
		w := httptest.NewRecorder()

		err := serveRaw(w, req)
		if err != nil {
			t.Fatalf("Failed to invoke serveRaw: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Fatalf("Want %d but got %d", http.StatusOK, w.Code)
		}
	})
}
