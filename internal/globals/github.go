package globals

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/go-github/v84/github"
	"github.com/webdestroya/groundskeeper/internal/awsclients/secretsclient"
	"github.com/webdestroya/groundskeeper/internal/config"
)

type gkSecret struct {
	GithubToken string `json:"githubToken,omitempty"`
}

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

	if !arn.IsARN(value) {
		return value, nil
	}

	// Secret is an ARN. Assume it's a secret
	client, err := secretsclient.New(ctx)
	if err != nil {
		return "", err
	}

	resp, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &value,
	})
	if err != nil {
		return "", err
	}

	secretData := resp.SecretBinary
	if len(secretData) == 0 && resp.SecretString != nil {
		secretData = []byte(*resp.SecretString)
	}
	var secret gkSecret
	if err := json.Unmarshal(secretData, &secret); err != nil {
		return "", err
	}

	return secret.GithubToken, nil
}

var BaseGithubHTTPClient = &http.Client{
	Transport: &http.Transport{
		// Controls connection reuse - false allows reuse, true forces new connections for each request
		DisableKeepAlives: false,
		// Maximum concurrent connections per host (active + idle)
		MaxConnsPerHost: 10,
		// Maximum idle connections maintained per host for reuse
		MaxIdleConnsPerHost: 5,
		// Maximum total idle connections across all hosts
		MaxIdleConns: 20,
		// How long an idle connection remains in the pool before being closed
		IdleConnTimeout: 20 * time.Second,
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
