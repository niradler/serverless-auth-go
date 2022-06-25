package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ServerlessAuthStackProps struct {
	awscdk.StackProps
}

func NewServerlessAuthStack(scope constructs.Construct, id string, props *ServerlessAuthStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	jwtSecret := os.Getenv("SLS_AUTH_JWT_SECRET")
	sessSecret := os.Getenv("SLS_AUTH_SESSION_SECRET")
	clientCallback := os.Getenv("SLS_AUTH_CLIENT_CALLBACK")
	googleCallback := os.Getenv("SLS_AUTH_GOOGLE_CALLBACK")
	googleId := os.Getenv("SLS_AUTH_GOOGLE_CLIENT_ID")
	googleSecret := os.Getenv("SLS_AUTH_GOOGLE_CLIENT_SECRET")

	authFunc := awslambda.NewFunction(stack, jsii.String("API-public-handler"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-auth-api"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(512),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../functions/build"), nil),
		Handler:      jsii.String("main"),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"AUTH_APP_TABLE":                jsii.String(*stack.StackName() + "-table"),
			"SLS_AUTH_JWT_SECRET":           jsii.String(jwtSecret),
			"SLS_AUTH_SESSION_SECRET":       jsii.String(sessSecret),
			"SLS_AUTH_GOOGLE_CALLBACK":      jsii.String(googleCallback),
			"SLS_AUTH_GOOGLE_CLIENT_ID":     jsii.String(googleId),
			"SLS_AUTH_GOOGLE_CLIENT_SECRET": jsii.String(googleSecret),
			"SLS_AUTH_CLIENT_CALLBACK":      jsii.String(clientCallback),
		},
	})

	restApi := awsapigateway.NewRestApi(stack, jsii.String("RestApi"), &awsapigateway.RestApiProps{
		RestApiName:        jsii.String(*stack.StackName() + "-RestApi"),
		RetainDeployments:  jsii.Bool(false),
		EndpointExportName: jsii.String("RestApiUrl"),
		Deploy:             jsii.Bool(true),
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String("v1"),
		},
	})

	rootRes := restApi.Root()
	rootRes.AddMethod(jsii.String("ANY"), awsapigateway.NewLambdaIntegration(authFunc, nil), nil)
	proxyRootRes := rootRes.AddResource(jsii.String("{proxy+}"), nil)
	proxyRootRes.AddMethod(jsii.String("ANY"), awsapigateway.NewLambdaIntegration(authFunc, nil), nil)

	appTable := awsdynamodb.NewTable(stack, jsii.String("auth-table"), &awsdynamodb.TableProps{
		TableName:     jsii.String(*stack.StackName() + "-table"),
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("pk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("sk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	appTable.GrantReadWriteData(authFunc)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)
	appName := os.Getenv("SLS_AUTH_APP_NAME")
	if appName == "" {
		appName = "ServerlessAuthStack"
	}
	NewServerlessAuthStack(app, appName, &ServerlessAuthStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
