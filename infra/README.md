# Welcome to Serverless-Auth CDK GO project

For first timer using CDK read the [documentation](https://docs.aws.amazon.com/cdk/index.html)

The CDK project will deploy all the necessary infrastructure for the Serverless-Auth.

## Useful commands

- `cdk deploy` deploy this stack to your default AWS account/region
- `cdk diff` compare deployed stack with current state
- `cdk synth` emits the synthesized CloudFormation template
- `go test` run unit tests

### For cost optimization we dont add indexes by default

```
	appTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("gsi-sk"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("sk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		ProjectionType: awsdynamodb.ProjectionType_ALL,
	})
```
