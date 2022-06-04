package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func LoadOrgsRoutes(router *gin.RouterGroup) {

	orgRouter := router.Group("/orgs")

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
			existingOrg, _ := GetItem(toKey("org", orgName), toKey("org", orgName))
			if existingOrg != nil {
				if handlerError(context, errors.New("Already exists"), http.StatusBadRequest) {
					return
				}
			}
			org := Org{
				PK:        toKey("org", orgName),
				SK:        toKey("org", orgName),
				Name:      orgName,
				CreatedBy: context.GetString("id"),
				CreatedAt: time.Now().UnixNano(),
			}

			err = CreateItem(org)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			orgUser := OrgUser{
				PK:        toKey("user", context.GetString("email")),
				SK:        toKey("org", orgName),
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

		orgRouter.POST("/:orgId/invite", func(context *gin.Context) {
			type Body struct {
				Email string `json:"email" binding:"required"`
				Role  string `json:"role" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			orgId := context.Param("orgId")
			isValid := roleCheck(context, orgId, "admin")
			if !isValid {
				handlerError(context, errors.New("Forbidden"), http.StatusForbidden)
				return
			}
			orgUser := OrgUser{
				PK:        toKey("user", body.Email),
				SK:        generateKey("org", orgId),
				Role:      body.Role,
				CreatedAt: time.Now().UnixNano(),
			}

			err = CreateItem(orgUser)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, "")
		})
	}

}