package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/firmware-bridge-service/shared/responses"
)

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	tableName := os.Getenv("CONNECTIONS_TABLE_NAME")
	connectionId := request.RequestContext.ConnectionID

	// Read optional deviceId from query string (e.g. wss://...?deviceId=abc-123)
	deviceId := request.QueryStringParameters["deviceId"]

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to load AWS config"), nil
	}

	client := dynamodb.NewFromConfig(cfg)

	item := map[string]types.AttributeValue{
		"connectionId": &types.AttributeValueMemberS{Value: connectionId},
		"connectedAt":  &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}

	if deviceId != "" {
		item["deviceId"] = &types.AttributeValueMemberS{Value: deviceId}
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to store connection"), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
