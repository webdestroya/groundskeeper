package secretresolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

// Resolve resolves a value that may be a plain string, an SSM parameter path,
// or an ARN pointing to either Secrets Manager or SSM Parameter Store.
func Resolve(ctx context.Context, value string) (string, error) {
	if value == "" {
		return "", errors.New("not set")
	}

	// ssm: prefix -> SSM parameter
	if after, ok := strings.CutPrefix(value, "ssm:"); ok {
		return ResolveParameter(ctx, after)
	}

	// Starts with / -> SSM parameter path
	if strings.HasPrefix(value, "/") {
		return ResolveParameter(ctx, value)
	}

	// Not an ARN -> return as-is
	if !arn.IsARN(value) {
		return value, nil
	}

	av, err := arn.Parse(value)
	if err != nil {
		return value, err
	}

	switch av.Service {
	case "secretsmanager":
		raw, err := ResolveSecret(ctx, value)
		if err != nil {
			return "", err
		}
		// Try to unmarshal as a JSON string (strips quotes).
		// If it's not a JSON string, use the raw bytes as-is.
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s, nil
		}
		return string(raw), nil

	case "ssm":
		// Resource format: "parameter/path/to/param"
		name := strings.TrimPrefix(av.Resource, "parameter")
		return ResolveParameter(ctx, name)

	default:
		return "", fmt.Errorf("unsupported ARN service: %s", av.Service)
	}
}
