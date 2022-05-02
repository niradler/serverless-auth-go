package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func LoadRoutes(r *gin.Engine) {

	r.GET("/admin/users", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Auth server is up!",
		})
	})

	r.GET("/admin/user", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Auth server is up!",
		})
	})

	// validate user token
	r.GET("/admin/validate", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Auth server is up!",
		})
	})

	// create token for user
	r.GET("/admin/token", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Auth server is up!",
		})
	})

}

func Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	log.Printf("Gin cold start")
	r := gin.Default()
	r.Use(gin.Logger())
	LoadRoutes(r)

	ginLambda = ginadapter.New(r)

	if os.Getenv("LAMBDA_TASK_ROOT") != "" {
		lambda.Start(Handler)
		return
	}
	r.Run("localhost:8282")
}
