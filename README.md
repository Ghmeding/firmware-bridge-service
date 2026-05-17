# Firmware Bridge Service

A serverless AWS backend built to manage firmware updates for hardware devices. It acts as a bridge between the hardware engineering team (who uploads the firmware) and an iOS app (which downloads it and transfers it to the device via BLE). 

The infrastructure is defined using **AWS CDK (TypeScript)**, and all compute logic is powered by **AWS Lambda** written in **Go (ARM64)** for optimal performance and cost-efficiency.

## 🏗 Architecture Overview

The system uses standard serverless AWS components:
- **API Gateway (REST)** for HTTP requests (upload, check version, get download URL).
- **API Gateway (WebSocket)** for real-time progress events from the devices.
- **AWS Lambda (Go)** for all business logic.
- **Amazon S3** for secure storage of the physical `.bin` firmware files (using presigned URLs for direct upload/download).
- **Amazon DynamoDB** for state and metadata (Firmware Versions, WebSocket Connections, and Update Progress).

For a detailed look at the data flows and architecture diagram, see the [Architecture Documentation](documentation/architecture.md).

## 🧰 Prerequisites

To build and deploy this project, you need:
- [Node.js](https://nodejs.org/) (for AWS CDK)
- [AWS CDK CLI](https://docs.aws.amazon.com/cdk/v2/guide/cli.html) (`npm install -g aws-cdk`)
- [Go](https://go.dev/) (1.23+ recommended) to compile the Lambda functions.
- AWS CLI configured with administrator credentials for your target account.

## 🚀 Deployment

1. **Install CDK dependencies:**
   ```bash
   npm install
   ```

2. **Ensure Go modules are tidy:**
   Navigate into the `lambdas` directory and ensure packages are downloaded. The project uses a shared internal Go module.
   *(Note: The CDK `GoFunction` construct will automatically compile the Go code using provided AL2023 during deployment.)*

3. **Deploy the stacks:**
   ```bash
   npx cdk deploy --all
   ```

   The CDK will deploy four stacks: `DatabaseStack`, `StorageStack`, `LambdaStack`, and `ApiGatewayStack`.

## 🔌 API Endpoints

Once deployed, the CDK outputs will provide your REST API URL and WebSocket Endpoint.

### REST API

- **POST** `/firmware/upload`
  *Hardware Team:* Generates a presigned S3 PUT URL to upload a new firmware binary, and logs the new version.
- **GET** `/firmware/latest`
  *iOS App:* Retrieves the latest available firmware version metadata.
- **GET** `/firmware/download?version=x.y.z`
  *iOS App:* Generates a short-lived presigned S3 GET URL to download the actual firmware binary directly from S3. *(Omitting `version` defaults to the latest).*

### WebSocket API

- **Connect:** Connect to `wss://<api-id>.execute-api.<region>.amazonaws.com/prod?deviceId=<id>`
- **Action: `sendProgress`**
  *Device/App:* While performing the BLE update, send a JSON payload to report progress.
  ```json
  {
    "action": "sendProgress",
    "deviceId": "test-device-001",
    "percent": 45,
    "status": "in_progress"
  }
  ```

## 📁 Project Structure

```text
├── bin/
│   └── firmware-bridge-service.ts  # CDK app entry point
├── lib/
│   ├── database-stack.ts           # DynamoDB tables
│   ├── storage-stack.ts            # S3 Buckets
│   ├── lambda-stack.ts             # Go Lambda function definitions
│   └── api-gateway-stack.ts        # REST and WebSocket APIs
├── lambdas/
│   ├── shared/                     # Shared Go structs and utilities
│   ├── get-latest-version/         # GET /firmware/latest
│   ├── get-download-url/           # GET /firmware/download
│   ├── upload-firmware/            # POST /firmware/upload
│   ├── ws-connect/                 # WSS $connect
│   ├── ws-disconnect/              # WSS $disconnect
│   └── ws-send-progress/           # WSS sendProgress Action
└── documentation/
    ├── architecture.md             # Detailed architecture docs
    └── image.png                   # Architecture diagram
```
