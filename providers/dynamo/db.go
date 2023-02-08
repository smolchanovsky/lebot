package dynamo

import (
	"awesomeProject/secrets"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

func GetDb() (*dynamo.DB, error) {
	keyId, err := secrets.GetSecret(secrets.AwsKeyIdPath)
	if err != nil {
		return nil, err
	}

	key, err := secrets.GetSecret(keyId)
	if err != nil {
		return nil, err
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     keyId,
			SecretAccessKey: key,
			SessionToken:    "",
			ProviderName:    "test",
		}),
	})

	return dynamo.New(sess), err
}
