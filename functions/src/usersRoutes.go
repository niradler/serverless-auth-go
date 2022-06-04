package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadUsersRoutes(router *gin.RouterGroup) {

	usersRouter := router.Group("/users")

	usersRouter.Use(AuthenticationMiddleware())
	{
		usersRouter.GET("/me", func(context *gin.Context) {
			userContext, err := GetUserContext(context.GetString("email"))
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, userContext)
		})

		usersRouter.PUT("/me", func(context *gin.Context) {
			type Body struct {
				Data interface{} `json:"data"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			err = UpdateUser(context.GetString("id"), body.Data)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, "")
		})
	}
}
