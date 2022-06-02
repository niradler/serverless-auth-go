package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func LoadPrivateRoutes(router *gin.RouterGroup) {

	usersRouter := router.Group("/users")

	usersRouter.Use(AuthenticationMiddleware())
	{
		usersRouter.GET("/me", func(context *gin.Context) {
			userContext, err := GetUserContext(context.GetString("email"))
			if handlerError(context, err, http.StatusNotFound) {
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"context": userContext,
			})
		})
	}
	orgRouter := router.Group("/org")

	orgRouter.Use(AuthenticationMiddleware())
	{
		orgRouter.POST("/", func(context *gin.Context) {

			type Body struct {
				Name string `json:"name" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)

			if handlerError(context, err, http.StatusBadRequest) {
				return
			}

			orgName := body.Name
			existingOrg, _ := GetItem("org#"+orgName, "org#"+orgName)
			if existingOrg != nil {
				if handlerError(context, errors.New("Already exists"), http.StatusBadRequest) {
					return
				}
			}
			org := Org{
				PK:        "org#" + orgName,
				SK:        "org#" + orgName,
				Name:      orgName,
				CreatedBy: context.GetString("email"),
				CreatedAt: time.Now().UnixNano(),
			}

			err = CreateItem(org)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			orgUser := OrgUser{
				PK:        "user#" + context.GetString("email"),
				SK:        "org#" + orgName,
				Role:      "admin",
				CreatedAt: time.Now().UnixNano(),
			}

			err = CreateItem(orgUser)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"message": "Org created",
				"email":   orgName,
			})
		})
	}

}
