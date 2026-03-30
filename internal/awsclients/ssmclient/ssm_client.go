package ssmclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/webdestroya/groundskeeper/internal/awsclients"
)

type SSMClienter interface {
	GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

var _ SSMClienter = (*ssm.Client)(nil)

func New(ctx context.Context) (SSMClienter, error) {
	cfg, err := awsclients.UseDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return ssm.NewFromConfig(cfg), nil
}

func NewFromConfig(cfg aws.Config) SSMClienter {
	return ssm.NewFromConfig(cfg)
}

// func GetParameter(ctx context.Context, input *ssm.GetParameterInput, opts ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
// 	return Client.GetParameter(ctx, input, opts...)
// }
