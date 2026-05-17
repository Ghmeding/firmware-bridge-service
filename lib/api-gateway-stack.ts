import * as cdk from 'aws-cdk-lib/core';
import { Construct } from 'constructs';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as lambda from 'aws-cdk-lib/aws-lambda';

export interface ApiGatewayStackProps extends cdk.StackProps {
  getLatestVersionHandler?: lambda.IFunction;
  uploadFirmwareHandler?: lambda.IFunction;
  getDownloadUrlHandler?: lambda.IFunction;
  wsConnectHandler?: lambda.IFunction;
  wsDisconnectHandler?: lambda.IFunction;
  wsSendProgressHandler?: lambda.IFunction;
}

export class ApiGatewayStack extends cdk.Stack {
  public readonly restApi: apigateway.RestApi;
  public readonly webSocketApi: apigatewayv2.CfnApi;
  public readonly webSocketStage: apigatewayv2.CfnStage;

  constructor(scope: Construct, id: string, props: ApiGatewayStackProps = {}) {
    super(scope, id, props);

    // ── REST API ────────────────────────────────────────────────
    this.restApi = new apigateway.RestApi(this, 'FirmwareRestApi', {
      restApiName: 'FirmwareBridgeRestApi',
      description: 'REST API for firmware version queries and download URLs',
      deployOptions: {
        stageName: 'prod',
      },
    });

    // /firmware resource
    const firmware = this.restApi.root.addResource('firmware');

    // GET /firmware/latest
    const latest = firmware.addResource('latest');
    latest.addMethod('GET',
      props.getLatestVersionHandler
        ? new apigateway.LambdaIntegration(props.getLatestVersionHandler)
        : undefined,
      { operationName: 'GetLatestVersion' },
    );

    // GET /firmware/download?version=x.y.z
    const download = firmware.addResource('download');
    download.addMethod('GET',
      props.getDownloadUrlHandler
        ? new apigateway.LambdaIntegration(props.getDownloadUrlHandler)
        : undefined,
      { operationName: 'GetDownloadUrl' },
    );

    // POST /firmware/upload  – hardware team trigger
    const upload = firmware.addResource('upload');
    upload.addMethod('POST', 
      props.uploadFirmwareHandler
      ? new apigateway.LambdaIntegration(props.uploadFirmwareHandler)
      : undefined,
       { operationName: 'UploadFirmware',}
      );

    // ── WebSocket API ──────────────────────────────────────────
    this.webSocketApi = new apigatewayv2.CfnApi(this, 'FirmwareWebSocketApi', {
      name: 'FirmwareBridgeWebSocketApi',
      protocolType: 'WEBSOCKET',
      routeSelectionExpression: '$request.body.action',
    });

    this.webSocketStage = new apigatewayv2.CfnStage(this, 'WebSocketProdStage', {
      apiId: this.webSocketApi.ref,
      stageName: 'prod',
      autoDeploy: true,
    });

    // ── WebSocket $connect route ────────────────────────────────
    if (props.wsConnectHandler) {
      const connectIntegration = new apigatewayv2.CfnIntegration(this, 'ConnectIntegration', {
        apiId: this.webSocketApi.ref,
        integrationType: 'AWS_PROXY',
        integrationUri: `arn:aws:apigateway:${this.region}:lambda:path/2015-03-31/functions/${props.wsConnectHandler.functionArn}/invocations`,
      });

      new apigatewayv2.CfnRoute(this, 'ConnectRoute', {
        apiId: this.webSocketApi.ref,
        routeKey: '$connect',
        target: `integrations/${connectIntegration.ref}`,
      });

      new lambda.CfnPermission(this, 'AllowApiGatewayInvokeConnect', {
        action: 'lambda:InvokeFunction',
        functionName: props.wsConnectHandler.functionArn,
        principal: 'apigateway.amazonaws.com',
        sourceArn: `arn:aws:execute-api:${this.region}:${this.account}:${this.webSocketApi.ref}/*/$connect`,
      });
    }

    // ── WebSocket $disconnect route ─────────────────────────────
    if (props.wsDisconnectHandler) {
      const disconnectIntegration = new apigatewayv2.CfnIntegration(this, 'DisconnectIntegration', {
        apiId: this.webSocketApi.ref,
        integrationType: 'AWS_PROXY',
        integrationUri: `arn:aws:apigateway:${this.region}:lambda:path/2015-03-31/functions/${props.wsDisconnectHandler.functionArn}/invocations`,
      });

      new apigatewayv2.CfnRoute(this, 'DisconnectRoute', {
        apiId: this.webSocketApi.ref,
        routeKey: '$disconnect',
        target: `integrations/${disconnectIntegration.ref}`,
      });

      new lambda.CfnPermission(this, 'AllowApiGatewayInvokeDisconnect', {
        action: 'lambda:InvokeFunction',
        functionName: props.wsDisconnectHandler.functionArn,
        principal: 'apigateway.amazonaws.com',
        sourceArn: `arn:aws:execute-api:${this.region}:${this.account}:${this.webSocketApi.ref}/*/$disconnect`,
      });
    }

    // ── WebSocket sendProgress route ────────────────────────────
    if (props.wsSendProgressHandler) {
      const sendProgressIntegration = new apigatewayv2.CfnIntegration(this, 'SendProgressIntegration', {
        apiId: this.webSocketApi.ref,
        integrationType: 'AWS_PROXY',
        integrationUri: `arn:aws:apigateway:${this.region}:lambda:path/2015-03-31/functions/${props.wsSendProgressHandler.functionArn}/invocations`,
      });

      new apigatewayv2.CfnRoute(this, 'SendProgressRoute', {
        apiId: this.webSocketApi.ref,
        routeKey: 'sendProgress',
        target: `integrations/${sendProgressIntegration.ref}`,
      });

      new lambda.CfnPermission(this, 'AllowApiGatewayInvokeSendProgress', {
        action: 'lambda:InvokeFunction',
        functionName: props.wsSendProgressHandler.functionArn,
        principal: 'apigateway.amazonaws.com',
        sourceArn: `arn:aws:execute-api:${this.region}:${this.account}:${this.webSocketApi.ref}/*/sendProgress`,
      });
    }

    // ── Outputs ────────────────────────────────────────────────
    new cdk.CfnOutput(this, 'RestApiUrl', {
      value: this.restApi.url,
      description: 'REST API base URL',
    });

    new cdk.CfnOutput(this, 'WebSocketApiEndpoint', {
      value: `wss://${this.webSocketApi.ref}.execute-api.${this.region}.amazonaws.com/${this.webSocketStage.stageName}`,
      description: 'WebSocket API endpoint',
    });
  }
}
