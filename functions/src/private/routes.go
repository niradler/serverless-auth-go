package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadRoutes(router *gin.Engine) {

	private := router.Group("/private")
	{
		private.GET("/me", func(context *gin.Context) {
			claims, validate := ValidateTokenMiddleware(context)
			if validate != true {
				context.AbortWithStatusJSON(http.StatusForbidden,
				gin.H{
					"error":   "Forbidden",
					"message": "Forbidden",
				})
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"claims": claims,
			})
		})
	}
}
