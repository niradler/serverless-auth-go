# tools for building and deploying
go version       # > 1.18.2
npm i -g aws-cdk # > 2.28.1 (build d035432)

# This script will deploy the serverless-auth stack to aws.

# setup
# aws configuration for deployment
# $env:AWS_ACCESS_KEY_ID = ''
# $env:AWS_SECRET_ACCESS_KEY = ''
# $env:AWS_DEFAULT_REGION = ''
# stack name
# $env:SLS_AUTH_APP_NAME = ''
# supported providers is facebook/google/github/bitbucket/gitlab
# $env:SLS_AUTH_[provider]_CLIENT_ID = ''
# $env:SLS_AUTH_GOOGLE_CLIENT_ID = ''
# $env:SLS_AUTH_GOOGLE_CLIENT_SECRET = ''
# $env:SLS_AUTH_GOOGLE_CALLBACK = ''
# $env:SLS_AUTH_CLIENT_CALLBACK = ''
# encypetion keys, keep in a safe place.
# $env:SLS_AUTH_SESSION_SECRET = ''
# $env:SLS_AUTH_JWT_SECRET = ''

# deploy
cd server && go get && export $env:GOOS='linux' && go build -o build/main ./src && cd .. && cd infra && cdk deploy
