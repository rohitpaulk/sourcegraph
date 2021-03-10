package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_serveRaw(t *testing.T) {
	// git.Mocks.ReadDir = func(commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error)
	// serverRaw(mockedRequest)


	mockNewCommon = func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHandler) (*Common, error) {
		return &Common{
			Repo: &types.Repo{
				Name: "test",
			},
			CommitID: api.CommitID("12345"),
		}, nil
	}



	// req := httptest.NewRequest("HEAD", "/github.com/sourcegraph/sourcegraph/-/raw", nil)
	// w := httptest.NewRecorder()

	// httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
	// }))


	fmt.Println("hitting serveRaw")
	err := serveRaw(w, req)
	fmt.Println("returned serveRaw")

	if err == nil {
		t.Errorf("Expected error. But got none.")
	}

	
}
