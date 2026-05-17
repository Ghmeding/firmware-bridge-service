import { StackProps, Stack, Duration } from 'aws-cdk-lib/core';
import { Construct } from 'constructs';
import { Function, Architecture, Runtime } from 'aws-cdk-lib/aws-lambda';
import { ITable } from 'aws-cdk-lib/aws-dynamodb';
import { IBucket } from 'aws-cdk-lib/aws-s3';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';

export interface LambdaStackProps extends StackProps {
  firmwareTable: ITable;
  firmwareBucket: IBucket;
  connectionsTable: ITable;
  updateProgressTable: ITable;
}

export class LambdaStack extends Stack {
  public readonly getLatestVersionFn: Function;
  public readonly uploadFirmwareFn: Function;
  public readonly getDownloadUrlFn: Function;
  public readonly wsConnectFn: Function;
  public readonly wsDisconnectFn: Function;
  public readonly wsSendProgressFn: Function;

  constructor(scope: Construct, id: string, props: LambdaStackProps) {
    super(scope, id, props);

    // ── Get Latest Version ─────────────────────────────────────
    this.getLatestVersionFn = new GoFunction(this, 'GetLatestVersionFn', {
      entry: 'lambdas/get-latest-version',
      architecture: Architecture.ARM_64,
      runtime: Runtime.PROVIDED_AL2023,
      environment: {
        FIRMWARE_TABLE_NAME: props.firmwareTable.tableName,
      },
      timeout: Duration.seconds(10),
      memorySize: 128,
    });

    // ── Get Latest Version ─────────────────────────────────────
    this.uploadFirmwareFn = new GoFunction(this, 'UploadFirmwareFn', {
      entry: 'lambdas/upload-firmware',
      architecture: Architecture.ARM_64,
      runtime: Runtime.PROVIDED_AL2023,
      environment: {
        FIRMWARE_TABLE_NAME: props.firmwareTable.tableName,
        FIRMWARE_BUCKET_NAME: props.firmwareBucket.bucketName,
      },
      timeout: Duration.seconds(10),
      memorySize: 128,
    });

    // ── Get Download URL ───────────────────────────────────────
    this.getDownloadUrlFn = new GoFunction(this, 'GetDownloadUrlFn', {
      entry: 'lambdas/get-download-url',
      architecture: Architecture.ARM_64,
      runtime: Runtime.PROVIDED_AL2023,
      environment: {
        FIRMWARE_TABLE_NAME: props.firmwareTable.tableName,
        FIRMWARE_BUCKET_NAME: props.firmwareBucket.bucketName,
      },
      timeout: Duration.seconds(10),
      memorySize: 128,
    });

    props.firmwareTable.grantReadData(this.getLatestVersionFn);
    props.firmwareTable.grantReadWriteData(this.uploadFirmwareFn);
    props.firmwareTable.grantReadData(this.getDownloadUrlFn);
    props.firmwareBucket.grantReadWrite(this.uploadFirmwareFn);
    props.firmwareBucket.grantRead(this.getDownloadUrlFn);

    // ── WebSocket Connect ──────────────────────────────────────
    this.wsConnectFn = new GoFunction(this, 'WsConnectFn', {
      entry: 'lambdas/ws-connect',
      architecture: Architecture.ARM_64,
      runtime: Runtime.PROVIDED_AL2023,
      environment: {
        CONNECTIONS_TABLE_NAME: props.connectionsTable.tableName,
      },
      timeout: Duration.seconds(10),
      memorySize: 128,
    });

    props.connectionsTable.grantWriteData(this.wsConnectFn);

    // ── WebSocket Disconnect ───────────────────────────────────
    this.wsDisconnectFn = new GoFunction(this, 'WsDisconnectFn', {
      entry: 'lambdas/ws-disconnect',
      architecture: Architecture.ARM_64,
      runtime: Runtime.PROVIDED_AL2023,
      environment: {
        CONNECTIONS_TABLE_NAME: props.connectionsTable.tableName,
      },
      timeout: Duration.seconds(10),
      memorySize: 128,
    });

    props.connectionsTable.grantWriteData(this.wsDisconnectFn);

    // ── WebSocket Send Progress ────────────────────────────────
    this.wsSendProgressFn = new GoFunction(this, 'WsSendProgressFn', {
      entry: 'lambdas/ws-send-progress',
      architecture: Architecture.ARM_64,
      runtime: Runtime.PROVIDED_AL2023,
      environment: {
        UPDATE_PROGRESS_TABLE_NAME: props.updateProgressTable.tableName,
      },
      timeout: Duration.seconds(10),
      memorySize: 128,
    });

    props.updateProgressTable.grantWriteData(this.wsSendProgressFn);
  }
}
