package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func LoadPrivateRoutes(router *gin.RouterGroup) {

	usersRouter := router.Group("/users")

	usersRouter.Use(AuthenticationMiddleware())
	{
		usersRouter.GET("/me", func(context *gin.Context) {
			log.Println(context.GetString("email"))
			userData, _ := GetItemByPK("user#" + context.GetString("email"))
			if userData == nil {
				log.Println(userData)
				context.AbortWithStatusJSON(http.StatusNotFound,
					gin.H{
						"error":   "Error",
						"message": "Not found",
					})
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"data": userData,
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
			if err != nil {
				log.Println(err.Error())
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": err.Error(),
					})
				return
			}
			orgName := body.Name
			existingOrg, _ := GetItem("org#"+orgName, "org#"+orgName)
			if existingOrg != nil {
				log.Println("Already exists")
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": "Already exists",
					})
				return
			}
			org := Org{
				PK:        "org#" + orgName,
				SK:        "org#" + orgName,
				Name:      orgName,
				CreatedBy: context.GetString("email"),
				CreatedAt: time.Now().UnixNano(),
			}

			err = CreateItem(org)
			if err != nil {
				log.Println(err)
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "CreateError",
						"message": "Failed to create org",
					})
				return
			}
			orgUser := OrgUser{
				PK:        "user#" + context.GetString("email"),
				SK:        "org#" + orgName,
				Role:      "admin",
				CreatedAt: time.Now().UnixNano(),
			}

			err = CreateItem(orgUser)
			if err != nil {
				log.Println(err)
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "CreateError",
						"message": "Failed to create role",
					})
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"message": "Org created",
				"email":   orgName,
			})
		})
	}

}
