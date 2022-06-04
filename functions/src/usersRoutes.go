package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadUsersRoutes(router *gin.RouterGroup) {

	usersRouter := router.Group("/users")

	usersRouter.Use(AuthenticationMiddleware())
	{
		usersRouter.GET("/validate", func(context *gin.Context) {

			orgs, _ := context.Get("orgs")
			data, _ := context.Get("data")

			context.JSON(http.StatusOK, gin.H{
				"email": context.GetString("email"),
				"id":    context.GetString("id"),
				"data":  data,
				"orgs":  orgs,
			})
		})

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
