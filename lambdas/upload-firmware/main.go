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
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/firmware-bridge-service/shared/models"
	"github.com/firmware-bridge-service/shared/responses"
)

type UploadRequest struct {
	Version      string `json:"version"`
	ReleaseNotes string `json:"releaseNotes,omitempty"`
}

type UploadResponse struct {
	UploadURL string `json:"uploadUrl"`
	S3Key     string `json:"s3Key"`
	Version   string `json:"version"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tableName := os.Getenv("FIRMWARE_TABLE_NAME")
	bucketName := os.Getenv("FIRMWARE_BUCKET_NAME")

	// Parse request body
	var req UploadRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return responses.Error(http.StatusBadRequest, "invalid request body"), nil
	}

	if req.Version == "" {
		return responses.Error(http.StatusBadRequest, "version is required"), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to load AWS config"), nil
	}

	// Generate S3 key and presigned upload URL
	s3Key := fmt.Sprintf("firmware/%s/firmware-%s.bin", req.Version, req.Version)

	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)

	presignResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to generate upload URL"), nil
	}

	// Write firmware record to DynamoDB
	now := time.Now().UTC().Format(time.RFC3339)
	record := models.FirmwareRecord{
		Version:      req.Version,
		S3Key:        s3Key,
		ReleaseNotes: req.ReleaseNotes,
		CreatedAt:    now,
	}

	ddbClient := dynamodb.NewFromConfig(cfg)

	// Write FIRMWARE#LATEST (overwrite)
	record.PK = models.FirmwareLatestPK
	latestItem, err := attributevalue.MarshalMap(record)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to marshal record"), nil
	}
	_, err = ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      latestItem,
	})
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to write latest record"), nil
	}

	// Write FIRMWARE#<version> (history)
	record.PK = fmt.Sprintf("FIRMWARE#%s", req.Version)
	versionItem, err := attributevalue.MarshalMap(record)
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to marshal version record"), nil
	}
	_, err = ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      versionItem,
	})
	if err != nil {
		return responses.Error(http.StatusInternalServerError, "failed to write version record"), nil
	}

	// Return presigned URL
	resp := UploadResponse{
		UploadURL: presignResult.URL,
		S3Key:     s3Key,
		Version:   req.Version,
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
