FROM golang:alpine

VOLUME "~/.aws/:/root/.aws:ro"

RUN apk update && apk add --update nodejs npm
RUN npm install -g aws-cdk

ENV GOOS=linux

WORKDIR "/app"
COPY . .

WORKDIR "/app/server"
RUN go mod download
RUN go mod verify
RUN go build -o build/main ./src

WORKDIR "/app/infra"
RUN go mod download
RUN go mod verify
RUN cdk deploy
