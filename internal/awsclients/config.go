package awsclients

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsHttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
)

// var (
// 	AwsConfig aws.Config
// )

// func init() {
// 	_ = UseDefaultConfig(context.Background())
// }

// func SetupWithConfig(cfg aws.Config) {
// 	AwsConfig = cfg
// }

func UseDefaultConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {

	cfg, err := config.LoadDefaultConfig(
		ctx,
		WithDefaultHTTPClient(),
		WithRetryer(),
	)
	if err != nil {
		return aws.Config{}, err
	}
	// SetupWithConfig(cfg)
	// return nil
	return cfg, nil
}

func WithDefaultHTTPClient() config.LoadOptionsFunc {
	return config.WithHTTPClient(DefaultHTTPClient())
}

func WithRetryer() config.LoadOptionsFunc {
	return config.WithRetryer(func() aws.Retryer {
		return retry.NewStandard(func(o *retry.StandardOptions) {
			o.MaxAttempts = 3
			o.MaxBackoff = 3 * time.Second
		})
	})
}

func DefaultHTTPClient() *awsHttp.BuildableClient {
	return awsHttp.NewBuildableClient().WithTransportOptions(func(t *http.Transport) {
		t.ResponseHeaderTimeout = 1 * time.Second
		t.MaxIdleConns = 100
		t.IdleConnTimeout = 90 * time.Second
		t.TLSHandshakeTimeout = 1 * time.Second
		t.ExpectContinueTimeout = 1 * time.Second
	}).WithDialerOptions(func(d *net.Dialer) {
		d.KeepAlive = 0 // 0 = use default (15sec)
		d.Timeout = time.Millisecond * 500
		// d.FallbackDelay: 100 * time.Millisecond
	})
}
