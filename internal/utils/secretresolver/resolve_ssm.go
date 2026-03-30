package secretresolver

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/webdestroya/groundskeeper/internal/awsclients/ssmclient"
)

func ResolveParameter(ctx context.Context, paramName string) (string, error) {
	client, err := ssmclient.New(ctx)
	if err != nil {
		return "", fmt.Errorf(`unable to resolve SSM parameter: %w`, err)
	}
	resp, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: new(true),
	})
	if err != nil {
		return "", err
	}

	return *resp.Parameter.Value, nil
}
