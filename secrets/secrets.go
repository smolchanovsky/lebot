package secrets

import (
	"os"
)

const TgmTokenPath = "~/secrets/telegram-api"
const DriveTokenPath = "~/secrets/google-api"
const AwsKeyIdPath = "~/secrets/aws-key-id"
const AwsKeyPath = "~/secrets/aws-key"

func GetSecret(filePath string) (string, error) {
	data, err := os.ReadFile("~/secrets/telegram-api")
	return string(data), err
}
