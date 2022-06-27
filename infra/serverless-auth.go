package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/joho/godotenv"
)

type ServerlessAuthStackProps struct {
	awscdk.StackProps
}

type Maps map[string]string

func ReduceItem(m1, m2 Maps) Maps {
	for key, value := range m2 {
		m1[key] = value

	}
	return m1
}

func GetAppEnv(merged map[string]string) map[string]string {
	allEnvMap := make(map[string]string)
	allEnv := os.Environ()
	prefix := "SLS_AUTH_"
	for _, envVar := range allEnv {
		if i := strings.Index(envVar, "="); i >= 0 {
			key := envVar[:i]
			value := envVar[i+1:]
			if strings.HasPrefix(key, prefix) {
				allEnvMap[key] = value
			}
		}

	}

	return ReduceItem(merged, allEnvMap)
}

func NewServerlessAuthStack(scope constructs.Construct, id string, props *ServerlessAuthStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	envMap := GetAppEnv(map[string]string{
		"GIN_MODE":       "release",
		"AUTH_APP_TABLE": *stack.StackName() + "-table",
	})

	var cdkEnv *map[string]*string
	data, _ := json.Marshal(envMap)
	json.Unmarshal(data, &cdkEnv)

	authFunc := awslambda.NewFunction(stack, jsii.String("API-public-handler"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-auth-api"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(512),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../functions/build"), nil),
		Handler:      jsii.String("main"),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Environment:  cdkEnv,
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

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

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
