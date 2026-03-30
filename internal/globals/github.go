package globals

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/utils/secretresolver"
)

var (
	ghTokenMu  sync.Mutex
	ghTokenVal string
	ghTokenErr error
)

func GithubToken(ctx context.Context) (string, error) {
	ghTokenMu.Lock()
	defer ghTokenMu.Unlock()

	if ghTokenVal != "" || ghTokenErr != nil {
		return ghTokenVal, ghTokenErr
	}

	ghTokenVal, ghTokenErr = resolveGithubToken(ctx)

	return ghTokenVal, ghTokenErr
}

func resolveGithubToken(ctx context.Context) (string, error) {

	value := config.GithubToken()

	if value == "" {
		return "", errors.New("github token is not set")
	}

	return secretresolver.Resolve(ctx, value)
}

var BaseGithubHTTPClient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives:   false,
		MaxConnsPerHost:     10,
		MaxIdleConnsPerHost: 5,
		MaxIdleConns:        20,
		IdleConnTimeout:     20 * time.Second,
	},
	Timeout: 30 * time.Second,
}

func GithubClient(ctx context.Context) (*github.Client, error) {
	token, err := GithubToken(ctx)
	if err != nil {
		return nil, err
	}
	return github.NewClient(BaseGithubHTTPClient).WithAuthToken(token), nil
}
