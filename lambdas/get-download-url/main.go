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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/firmware-bridge-service/shared/models"
	"github.com/firmware-bridge-service/shared/responses"
)

type DownloadResponse struct {
	DownloadURL string `json:"downloadUrl"`
	Version     string `json:"version"`
	S3Key       string `json:"s3Key"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tableName := os.Getenv("FIRMWARE_TABLE_NAME")
	bucketName := os.Getenv("FIRMWARE_BUCKET_NAME")

	// Determine which version to download
	version := request.QueryStringParameters["version"]

	// Default to FIRMWARE#LATEST if no version specified
	pk := models.FirmwareLatestPK
	if version != "" {
		pk = fmt.Sprintf("FIRMWARE#%s", version)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to load AWS config"), nil
	}

	// Look up the firmware record in DynamoDB
	ddbClient := dynamodb.NewFromConfig(cfg)
	result, err := ddbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
		},
	})
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to query firmware table"), nil
	}

	if result.Item == nil {
		return responses.Error(http.StatusNotFound, "firmware version not found"), nil
	}

	var record models.FirmwareRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to parse firmware record"), nil
	}

	// Generate presigned GET URL
	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(record.S3Key),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to generate download URL"), nil
	}

	resp := DownloadResponse{
		DownloadURL: presignResult.URL,
		Version:     record.Version,
		S3Key:       record.S3Key,
	}
	body, _ := json.Marshal(resp)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
