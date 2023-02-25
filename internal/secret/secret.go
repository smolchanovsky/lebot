package secret

import (
	"os"
	"path"
	"strings"
)

const TgmTokenPath = "secrets/telegram-api"
const DriveTokenPath = "secrets/google-api"
const AwsKeyIdPath = "secrets/aws-key-id"
const AwsKeyPath = "secrets/aws-key"

func GetSecret(filePath string) (string, error) {
	fullPath := path.Join("tmp", filePath)
	data, err := os.ReadFile(fullPath)
	return strings.TrimSpace(string(data)), err
}
