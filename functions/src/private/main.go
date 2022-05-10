package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	log.Printf("Gin cold start")
	router := gin.Default()
	router.Use(gin.Logger())
	LoadRoutes(router)

	ginLambda = ginadapter.New(router)

	if os.Getenv("LAMBDA_TASK_ROOT") != "" {
		lambda.Start(Handler)
		return
	}
	router.Run("localhost:8281")
}
