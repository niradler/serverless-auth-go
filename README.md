# Serverless Auth

Simple authentication mechanism base on aws serverless services (Dynmodb, Lambda, ApiGateway)

## Usage

## Deploy

use deploy.sh script to setup and customize the deployment

## Develop

```sh
npm i -g cdk
git clone https://github.com/niradler/serverless-auth-go.git
cd serverless-auth-go
$env:GOOS = 'linux' or export GOOS=linux
cd functions && go get && go build -o build/main ./src
cd infra && cdk deploy
# run
cd functions
go run ./src/.
```

## Tech Stack

- Using aws cdk (iac) to deploy the aws services. (golang)
- Scalable api using Gin framework. (golang)
- goth for providers logins. (github/bitbucket/gitlab/facebook/google) (only google/github is tested)
-

## Contribution

Contribution is welcome.
