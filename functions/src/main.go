package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var ginLambda *ginadapter.GinLambda

func Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func handlerError(context *gin.Context, err error, status int) bool {
	if err != nil {
		Logger.Info("handlerError", zap.Error(err))
		context.AbortWithStatusJSON(status,
			gin.H{
				"error":   "Error",
				"message": err.Error(),
			})
		return true
	}
	return false
}

func main() {
	InitializeLogger()
	Logger.Info("Gin cold start")
	router := gin.Default()
	router.Use(gin.Logger())
	v1 := router.Group("/v1")
	LoadUsersRoutes(v1)
	LoadOrgsRoutes(v1)
	LoadPublicRoutes(v1)

	ginLambda = ginadapter.New(router)

	if os.Getenv("LAMBDA_TASK_ROOT") != "" {
		lambda.Start(Handler)
		return
	}
	router.Run("localhost:8280")
}
