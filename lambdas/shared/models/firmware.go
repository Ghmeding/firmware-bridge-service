package models

const FirmwareLatestPK = "FIRMWARE#LATEST"

type FirmwareRecord struct {
	PK           string `dynamodbav:"PK" json:"-"`
	Version      string `dynamodbav:"version" json:"version"`
	S3Key        string `dynamodbav:"s3Key" json:"s3Key"`
	ReleaseNotes string `dynamodbav:"releaseNotes,omitempty" json:"releaseNotes,omitempty"`
	CreatedAt    string `dynamodbav:"createdAt" json:"createdAt"`
}
