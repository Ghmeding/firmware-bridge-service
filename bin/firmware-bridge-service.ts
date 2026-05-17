#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib/core';
import { DatabaseStack } from '../lib/database-stack';
import { StorageStack } from '../lib/storage-stack';
import { LambdaStack } from '../lib/lambda-stack';
import { ApiGatewayStack } from '../lib/api-gateway-stack';

const app = new cdk.App();

const database = new DatabaseStack(app, 'DatabaseStack');

const storage = new StorageStack(app, 'StorageStack');

const lambdas = new LambdaStack(app, 'LambdaStack', {
  firmwareTable: database.firmwareTable,
  firmwareBucket: storage.firmwareBucket,
  connectionsTable: database.connectionsTable,
  updateProgressTable: database.updateProgressTable,
});

new ApiGatewayStack(app, 'ApiGatewayStack', {
  getLatestVersionHandler: lambdas.getLatestVersionFn,
  uploadFirmwareHandler: lambdas.uploadFirmwareFn,
  getDownloadUrlHandler: lambdas.getDownloadUrlFn,
  wsConnectHandler: lambdas.wsConnectFn,
  wsDisconnectHandler: lambdas.wsDisconnectFn,
  wsSendProgressHandler: lambdas.wsSendProgressFn,
});
