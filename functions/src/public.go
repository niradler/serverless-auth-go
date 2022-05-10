package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func LoadRoutes(router *gin.Engine) {

	public := router.Group("/public")
	{
		public.GET("/me", func(context *gin.Context) {

			context.JSON(http.StatusOK, gin.H{
				"message": "my details",
			})
		})

		public.POST("/login", func(context *gin.Context) {
			type Body struct {
				Email    string `json:"email" binding:"required"`
				Password string `json:"password" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if err != nil {
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": err.Error(),
					})
				return
			}

			context.JSON(http.StatusOK, gin.H{
				"message": "Login success",
			})
		})

		public.POST("/signup", func(context *gin.Context) {
			type Body struct {
				Email    string      `json:"email" binding:"required"`
				Password string      `json:"password" binding:"required"`
				Data     interface{} `json:"data"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if err != nil {
				log.Printf("fmt")
				fmt.Println(err)
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": err.Error(),
					})
				return
			}

			context.JSON(http.StatusOK, gin.H{
				"message": "Signup success",
				"email":   body.Email,
			})
		})
	}
}

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
	router.Run("localhost:8282")
}
