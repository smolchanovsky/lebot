package googledialogflow

import (
	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"cloud.google.com/go/dialogflow/apiv2/dialogflowpb"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/option"
	"lebot/internal/secret"
)

func NewClient() (*dialogflow.SessionsClient, error) {
	ctx := context.Background()

	token, err := secret.GetSecret(secret.DriveTokenPath)
	sessionClient, err := dialogflow.NewSessionsClient(ctx, option.WithCredentialsJSON([]byte(token)))
	if err != nil {
		return nil, err
	}
	//defer sessionClient.Close()

	return sessionClient, nil
}

func DetectIntentText(client *dialogflow.SessionsClient, projectID, sessionID, text string) (*dialogflowpb.QueryResult, error) {
	ctx := context.Background()

	if projectID == "" || sessionID == "" {
		return nil, errors.New(fmt.Sprintf("Received empty project (%s) or session (%s)", projectID, sessionID))
	}

	sessionPath := fmt.Sprintf("projects/%s/agent/sessions/%s", projectID, sessionID)
	textInput := dialogflowpb.TextInput{Text: text, LanguageCode: "en"}
	queryTextInput := dialogflowpb.QueryInput_Text{Text: &textInput}
	queryInput := dialogflowpb.QueryInput{Input: &queryTextInput}
	request := dialogflowpb.DetectIntentRequest{Session: sessionPath, QueryInput: &queryInput}

	response, err := client.DetectIntent(ctx, &request)
	if err != nil {
		return nil, err
	}

	queryResult := response.GetQueryResult()
	return queryResult, nil
}
