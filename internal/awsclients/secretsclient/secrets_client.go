package secretsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/webdestroya/groundskeeper/internal/awsclients"
)

type SecretsClienter interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

var _ SecretsClienter = (*secretsmanager.Client)(nil)

func New(ctx context.Context) (SecretsClienter, error) {
	cfg, err := awsclients.UseDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return secretsmanager.NewFromConfig(cfg), nil
}

func NewFromConfig(cfg aws.Config) SecretsClienter {
	return secretsmanager.NewFromConfig(cfg)
}
