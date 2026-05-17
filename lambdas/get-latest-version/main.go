package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/firmware-bridge-service/shared/models"
	"github.com/firmware-bridge-service/shared/responses"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tableName := os.Getenv("FIRMWARE_TABLE_NAME")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to load AWS config"), nil
	}

	client := dynamodb.NewFromConfig(cfg)

	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: models.FirmwareLatestPK},
		},
	})
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to query firmware table"), nil
	}

	if result.Item == nil {
		return responses.Error(http.StatusNotFound, "no firmware version found"), nil
	}

	var record models.FirmwareRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to parse firmware record"), nil
	}

	body, _ := json.Marshal(record)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
