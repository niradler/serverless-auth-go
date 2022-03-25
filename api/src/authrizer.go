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

	r.GET("/routes", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"/":            "status",
			"/routes":      "routes",
			"/admin/users": "users",
			"/admin/user":  "user",
			"/login":       "login",
			"/signup":      "signup",
			"/me":          "me",
			"/validate":    "validate",
		})
	})

	r.GET("/", func(context *gin.Context) {
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
