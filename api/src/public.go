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

	r.POST("/public/login", func(context *gin.Context) {
		type Body struct {
			Email    uint `json:"email" binding:"required"`
			Password uint `json:"password" binding:"required"`
		}
		body := Body{}
		if context.ShouldBindJSON(&body) != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{
					"error":   "ValidationError",
					"message": "Invalid inputs. Please check your inputs"})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"message": "Login success",
		})
	})

	r.POST("/public/signup", func(context *gin.Context) {
		type Body struct {
			Email    uint `json:"email" binding:"required"`
			Password uint `json:"password" binding:"required"`
		}
		body := Body{}
		if context.ShouldBindJSON(&body) != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{
					"error":   "ValidationError",
					"message": "Invalid inputs. Please check your inputs"})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"message": "Signup success",
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
