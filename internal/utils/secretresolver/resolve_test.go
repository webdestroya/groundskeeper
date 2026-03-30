package secretresolver_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
	"github.com/webdestroya/groundskeeper/internal/testutil"
	"github.com/webdestroya/groundskeeper/internal/utils/secretresolver"
)

func TestResolve(t *testing.T) {
	ctx := context.Background()

	t.Run("EmptyValue", func(t *testing.T) {
		_, err := secretresolver.Resolve(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not set")
	})

	t.Run("PlainString", func(t *testing.T) {
		val, err := secretresolver.Resolve(ctx, "postgres://localhost/db")
		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost/db", val)
	})

	t.Run("SsmPrefix", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "ssm",
					Action:  "GetParameter",
					JMESPathMatches: map[string]any{
						"Name": "/app/DATABASE_URL",
					},
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"Parameter": map[string]any{
							"Name":    "/app/DATABASE_URL",
							"Type":    "SecureString",
							"Value":   "postgres://host/db",
							"Version": 1,
						},
					},
				},
			},
		})

		val, err := secretresolver.Resolve(ctx, "ssm:/app/DATABASE_URL")
		require.NoError(t, err)
		assert.Equal(t, "postgres://host/db", val)
	})

	t.Run("SlashPrefix", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "ssm",
					Action:  "GetParameter",
					JMESPathMatches: map[string]any{
						"Name": "/app/DATABASE_URL",
					},
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"Parameter": map[string]any{
							"Name":    "/app/DATABASE_URL",
							"Type":    "String",
							"Value":   "postgres://host/db",
							"Version": 1,
						},
					},
				},
			},
		})

		val, err := secretresolver.Resolve(ctx, "/app/DATABASE_URL")
		require.NoError(t, err)
		assert.Equal(t, "postgres://host/db", val)
	})

	t.Run("SsmArn", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "ssm",
					Action:  "GetParameter",
					JMESPathMatches: map[string]any{
						"Name": "/app/DATABASE_URL",
					},
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"Parameter": map[string]any{
							"Name":    "/app/DATABASE_URL",
							"Type":    "String",
							"Value":   "postgres://host/db",
							"Version": 1,
						},
					},
				},
			},
		})

		val, err := secretresolver.Resolve(ctx, "arn:aws:ssm:us-east-1:555555555555:parameter/app/DATABASE_URL")
		require.NoError(t, err)
		assert.Equal(t, "postgres://host/db", val)
	})

	t.Run("SecretsManagerJsonString", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "secretsmanager",
					Action:  "GetSecretValue",
					JMESPathMatches: map[string]any{
						"SecretId": "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
					},
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"ARN":          "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
						"Name":         "mydb",
						"SecretString": `"postgres://host/db"`,
					},
				},
			},
		})

		val, err := secretresolver.Resolve(ctx, "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf")
		require.NoError(t, err)
		assert.Equal(t, "postgres://host/db", val)
	})

	t.Run("SecretsManagerRawString", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "secretsmanager",
					Action:  "GetSecretValue",
					JMESPathMatches: map[string]any{
						"SecretId": "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
					},
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"ARN":          "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
						"Name":         "mydb",
						"SecretString": "postgres://host/db",
					},
				},
			},
		})

		val, err := secretresolver.Resolve(ctx, "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf")
		require.NoError(t, err)
		assert.Equal(t, "postgres://host/db", val)
	})

	t.Run("SecretsManagerJsonKey", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "secretsmanager",
					Action:  "GetSecretValue",
					JMESPathMatches: map[string]any{
						"SecretId": "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
					},
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"ARN":          "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
						"Name":         "mydb",
						"SecretString": `{"password":"hunter2","host":"db.example.com"}`,
					},
				},
			},
		})

		val, err := secretresolver.Resolve(ctx, "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf:password::")
		require.NoError(t, err)
		assert.Equal(t, "hunter2", val)
	})

	t.Run("UnsupportedArnService", func(t *testing.T) {
		_, err := secretresolver.Resolve(ctx, "arn:aws:s3:::my-bucket")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported ARN service: s3")
	})
}

func TestResolveSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("MissingJsonKey", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "secretsmanager",
					Action:  "GetSecretValue",
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"ARN":          "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
						"Name":         "mydb",
						"SecretString": `{"host":"db.example.com"}`,
					},
				},
			},
		})

		_, err := secretresolver.ResolveSecret(ctx, "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf:badkey::")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `key "badkey" not found in secret`)
	})

	t.Run("SecretBinary", func(t *testing.T) {
		binaryData := []byte("binary-db-connection-string")

		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "secretsmanager",
					Action:  "GetSecretValue",
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"ARN":          "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
						"Name":         "mydb",
						"SecretBinary": binaryData,
					},
				},
			},
		})

		raw, err := secretresolver.ResolveSecret(ctx, "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf")
		require.NoError(t, err)
		assert.Equal(t, json.RawMessage(binaryData), raw)
	})

	t.Run("VersionStageAndId", func(t *testing.T) {
		testutil.StartMocker(t, []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "secretsmanager",
					Action:  "GetSecretValue",
					JMESPathMatches: map[string]any{
						"SecretId":     "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
						"VersionStage": "AWSPREVIOUS",
						"VersionId":    "ver-123",
					},
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]any{
						"ARN":          "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf",
						"Name":         "mydb",
						"SecretString": "old-db-url",
						"VersionId":    "ver-123",
					},
				},
			},
		})

		raw, err := secretresolver.ResolveSecret(ctx, "arn:aws:secretsmanager:us-east-1:555555555555:secret:mydb-AbCdEf::AWSPREVIOUS:ver-123")
		require.NoError(t, err)
		assert.Equal(t, json.RawMessage("old-db-url"), raw)
	})
}
