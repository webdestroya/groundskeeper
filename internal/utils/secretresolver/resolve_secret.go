package secretresolver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/webdestroya/groundskeeper/internal/awsclients/secretsclient"
)

// ResolveSecret fetches a secret from AWS Secrets Manager using an extended ARN format:
//
//	arn:aws:secretsmanager:region:account_id:secret:secret-name:json-key:version-stage:version-id
//
// If json-key is provided, the secret is parsed as a JSON map and the value at
// that key is returned as a json.RawMessage. If json-key is omitted, the raw
// secret value is returned. Version-stage and version-id are optional and fall
// back to Secrets Manager defaults (AWSCURRENT) when not specified.
func ResolveSecret(ctx context.Context, secretArn string) (json.RawMessage, error) {

	av, err := arn.Parse(secretArn)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(av.Resource, ":")
	// parts[0] = "secret"
	// parts[1] = secret-name (with random suffix)
	// parts[2] = json-key (optional)
	// parts[3] = version-stage (optional)
	// parts[4] = version-id (optional)

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid secret ARN resource: %s", av.Resource)
	}

	// Reconstruct the base ARN for the API call (strip json-key, version-stage, version-id)
	av.Resource = parts[0] + ":" + parts[1]
	baseArn := av.String()

	// Extract optional components
	var jsonKey string
	var versionStage *string
	var versionId *string

	if len(parts) >= 3 && parts[2] != "" {
		jsonKey = parts[2]
	}
	if len(parts) >= 4 && parts[3] != "" {
		versionStage = &parts[3]
	}
	if len(parts) >= 5 && parts[4] != "" {
		versionId = &parts[4]
	}

	client, err := secretsclient.New(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     &baseArn,
		VersionStage: versionStage,
		VersionId:    versionId,
	})
	if err != nil {
		return nil, err
	}

	secretData := resp.SecretBinary
	if len(secretData) == 0 && resp.SecretString != nil {
		secretData = []byte(*resp.SecretString)
	}

	// No JSON key specified, return the raw secret
	if jsonKey == "" {
		return json.RawMessage(secretData), nil
	}

	// Parse as JSON map and extract the requested key
	var m map[string]json.RawMessage
	if err := json.Unmarshal(secretData, &m); err != nil {
		return nil, fmt.Errorf("failed to parse secret as JSON: %w", err)
	}

	val, ok := m[jsonKey]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret", jsonKey)
	}

	return val, nil
}
