package main

import (
	"context"
	"net/http"
	"os"

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

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to load AWS config"), nil
	}

	client := dynamodb.NewFromConfig(cfg)

	_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"connectionId": &types.AttributeValueMemberS{Value: connectionId},
		},
	})
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to remove connection"), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
