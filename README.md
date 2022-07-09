# Serverless Auth

Simple authentication mechanism base on aws serverless services (Dynmodb, Lambda, ApiGateway)

## Features

- Basic user authentication with email/password
- Passwordless user authentication with email only (need to configure email provider)
- Login with providers like Google, Facebook, etc.
- Authentication and Authorization roles.
- Multi tenants

## Usage

| method | route                                | payload             | Role  | public | description             |
| ------ | ------------------------------------ | ------------------- | ----- | ------ | ----------------------- |
| POST   | /v1/auth/login                       | email,password      |       | true   | Login                   |
| POST   | /v1/auth/login/email                 | email               |       | true   | Passwordless Login      |
| POST   | /v1/auth/signup                      | email,password,data |       | true   | Signup                  |
| GET    | /v1/auth/validate                    |                     |       | true   | ValidateToken           |
| POST   | /v1/auth/renew                       |                     |       | false  | Get new Token           |
| GET    | /v1/auth/provider/:provider          |                     |       | true   | Login with provider     |
| GET    | /v1/auth/provider/:provider/callback |                     |       | true   | Validate provider login |
| GET    | /v1/users/me                         |                     |       | true   | Health check            |
| PUT    | /v1/users/me                         | data                |       | false  | Update user data        |
| POST   | /v1/orgs                             | name                |       | false  | Create Org              |
| POST   | /v1/orgs/:orgId/invite               | email,role          | admin | false  | Invite user to me org   |
| GET    | /v1/orgs/:orgId/users                |                     | admin | false  | Get org users           |

## Deploy

use deploy.sh script to setup and customize the deployment

to customize the deployment, you can use the following environment variables:

```.env
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=
SLS_AUTH_GOOGLE_CLIENT_ID=
SLS_AUTH_GOOGLE_CLIENT_SECRET=
SLS_AUTH_GOOGLE_CALLBACK=
SLS_AUTH_CLIENT_CALLBACK=
SLS_AUTH_SESSION_SECRET=
SLS_AUTH_JWT_SECRET=
SLS_AUTH_APP_NAME=
SLS_AUTH_SENDGRID_API_KEY=
SLS_AUTH_FROM_EMAIL=
SLS_AUTH_EMAIL_TEMPLATES_FOLDER=
SLS_AUTH_APP_NAME=
SLS_AUTH_APP_CONTACT=
```

## Develop

```sh
npm i -g cdk
git clone https://github.com/niradler/serverless-auth-go.git
cd serverless-auth-go
$env:GOOS = 'linux' or export GOOS=linux
cd server && go get && go build -o build/main ./src
cd infra && cdk deploy
# run
cd server
go run ./src/.
```

## Tech Stack

- Using aws cdk (iac) to deploy the aws services. (golang)
- Scalable api using Gin framework. (golang)
- goth for providers logins. (github/bitbucket/gitlab/facebook/google) (only google/github is tested)

## Contribution

Contribution is welcome.
