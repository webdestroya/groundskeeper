package globals_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/webdestroya/groundskeeper/internal/globals"
	"github.com/webdestroya/groundskeeper/internal/testutil"
)

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func TestGithubSanity(t *testing.T) {
	mux := testutil.MockGithub(t)

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})

	ctx := context.Background()

	client, err := globals.GithubClient(ctx)
	require.NoError(t, err)

	user, resp, err := client.Users.Get(ctx, "")
	require.NoError(t, err)
	_ = user
	_ = resp
	_ = mux

}
