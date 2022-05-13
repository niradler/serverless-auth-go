package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadPrivateRoutes(router *gin.RouterGroup) {

	private := router.Group("/user")
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
