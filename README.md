# Serverless Auth

Provide simple authentication mechanism base on aws serverless services (Dynmodb, Lambda, ApiGateway)

## Tech Stack

- Using aws cdk (iac) to deploy the aws services. (golang)
- Scalable api using Gin framework. (golang)

## Deploy

```sh
npm i -g cdk
git clone https://github.com/niradler/serverless-auth-go.git
cd serverless-auth-go
$env:GOOS = 'linux' or export GOOS=linux
cd functions && go get && go build -o build/main ./src
cd infra && cdk deploy
```

## Contribution

Contribution is welcome.

(Deploy first)

`cd functions/src && go run .`
