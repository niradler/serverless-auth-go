package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	"github.com/niradler/social-lab/src/auth"
	"github.com/niradler/social-lab/src/routes"
	"github.com/niradler/social-lab/src/utils"
)

var ginLambda *ginadapter.GinLambda

func Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	utils.InitializeLogger()
	utils.Logger.Info("Gin cold start")
	auth.GothInit()
	router := gin.Default()

	router.Static("src/assets", "./assets")
	router.LoadHTMLGlob("src/ui/*.tmpl")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")

	router.Use(gin.Logger())

	v1 := router.Group("/v1")
	routes.LoadUsersRoutes(v1)
	routes.LoadOrgsRoutes(v1)
	routes.LoadPublicRoutes(v1)

	ginLambda = ginadapter.New(router)

	if os.Getenv("LAMBDA_TASK_ROOT") != "" {
		lambda.Start(Handler)
		return
	}
	router.Run("localhost:8280")
}
