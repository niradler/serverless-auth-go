#!/usr/bin/env bash

npm i -g aws-cdk

# This script will deploy the serverless-auth stack to aws.

# setup
# aws configuration for deployment
# export AWS_ACCESS_KEY_ID=
# export AWS_SECRET_ACCESS_KEY=
# export AWS_DEFAULT_REGION=
# stack name
# export SLS_AUTH_APP_NAME=
# supported providers is facebook/google/github/bitbucket/gitlab
# export SLS_AUTH_[provider]_CLIENT_ID=
# export SLS_AUTH_GOOGLE_CLIENT_ID=
# export SLS_AUTH_GOOGLE_CLIENT_SECRET=
# export SLS_AUTH_GOOGLE_CALLBACK=
# export SLS_AUTH_CLIENT_CALLBACK=
# encypetion keys, keep in a safe place.
# export SLS_AUTH_SESSION_SECRET=
# export SLS_AUTH_JWT_SECRET=

# deploy
cd functions && go get && go build -o build/main ./src && cd .. && cd infra && cdk deploy
