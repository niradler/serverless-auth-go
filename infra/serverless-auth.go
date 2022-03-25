package main

import (
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

	// The code that defines your stack goes here

	publicFunc := awslambda.NewFunction(stack, jsii.String("API-public-handler"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-public-api"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../api/build/public"), nil),
		Handler:      jsii.String("main"),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-table"),
		},
	})

	privateFunc := awslambda.NewFunction(stack, jsii.String("API-private-handler"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-private-api"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../api/build/private"), nil),
		Handler:      jsii.String("main"),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-table"),
		},
	})

	adminFunc := awslambda.NewFunction(stack, jsii.String("API-admin-handler"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-admin-api"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../api/build/admin"), nil),
		Handler:      jsii.String("main"),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-table"),
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
	rootRes.AddMethod(jsii.String("ANY"), awsapigateway.NewLambdaIntegration(publicFunc, nil), nil)
	publicRes := rootRes.AddResource(jsii.String("public"), nil)
	publicRes.AddMethod(jsii.String("ANY"), awsapigateway.NewLambdaIntegration(publicFunc, nil), nil)
	publicProxyRes := publicRes.AddResource(jsii.String("{proxy+}"), nil)
	publicProxyRes.AddMethod(jsii.String("ANY"), awsapigateway.NewLambdaIntegration(publicFunc, nil), nil)

	privateRes := rootRes.AddResource(jsii.String("private"), nil).AddResource(jsii.String("{proxy+}"), nil)
	privateRes.AddMethod(jsii.String("ANY"), awsapigateway.NewLambdaIntegration(privateFunc, nil), nil)

	adminRes := rootRes.AddResource(jsii.String("admin"), nil).AddResource(jsii.String("{proxy+}"), nil)
	adminRes.AddMethod(jsii.String("ANY"), awsapigateway.NewLambdaIntegration(adminFunc, nil), nil)

	usersTable := awsdynamodb.NewTable(stack, jsii.String("auth-table"), &awsdynamodb.TableProps{
		TableName:     jsii.String(*stack.StackName() + "-table"),
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("root_obj_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("sub_obj_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	usersTable.GrantWriteData(publicFunc)
	usersTable.GrantWriteData(privateFunc)
	usersTable.GrantWriteData(adminFunc)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewServerlessAuthStack(app, "ServerlessAuthStack", &ServerlessAuthStackProps{
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
