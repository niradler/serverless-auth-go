#!/usr/bin/env bash

cd functions && go build -o build/main ./src
cd infra && cdk deploy
