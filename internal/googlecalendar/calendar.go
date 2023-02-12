package googlecalendar

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"lebot/internal/secret"
)

func NewService() (*calendar.Service, error) {
	ctx := context.Background()

	token, err := secret.GetSecret(secret.DriveTokenPath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON([]byte(token), calendar.CalendarScope, calendar.CalendarEventsScope)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}
