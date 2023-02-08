package drive

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"lebot/secrets"
)

func NewService() (*drive.Service, error) {
	ctx := context.Background()

	token, err := secrets.GetSecret(secrets.DriveTokenPath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON([]byte(token), drive.DriveScope)
	if err != nil {
		return nil, err
	}

	client := config.Client(context.Background())
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))

	return srv, err
}
