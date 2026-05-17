package main

import (
	"context"
	"encoding/json"
	"fmt"
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

type ProgressMessage struct {
	Action   string `json:"action"`
	DeviceId string `json:"deviceId"`
	Percent  int    `json:"percent"`
	Status   string `json:"status,omitempty"`
}

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	tableName := os.Getenv("UPDATE_PROGRESS_TABLE_NAME")
	connectionId := request.RequestContext.ConnectionID

	var msg ProgressMessage
	if err := json.Unmarshal([]byte(request.Body), &msg); err != nil {
		return responses.Error(http.StatusBadRequest, "invalid message body"), nil
	}

	if msg.DeviceId == "" {
		return responses.Error(http.StatusBadRequest, "deviceId is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to load AWS config"), nil
	}

	client := dynamodb.NewFromConfig(cfg)

	status := msg.Status
	if status == "" {
		status = "in_progress"
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"deviceId":     &types.AttributeValueMemberS{Value: msg.DeviceId},
			"connectionId": &types.AttributeValueMemberS{Value: connectionId},
			"percent":      &types.AttributeValueMemberN{Value: itoa(msg.Percent)},
			"status":       &types.AttributeValueMemberS{Value: status},
			"updatedAt":    &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to write progress"), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func main() {
	lambda.Start(handler)
}
