package routes

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/niradler/social-lab/src/auth"
	"github.com/niradler/social-lab/src/db"
	"github.com/niradler/social-lab/src/types"
	"github.com/niradler/social-lab/src/utils"
)

func LoadOrgsRoutes(router *gin.RouterGroup) {

	orgRouter := router.Group("/orgs")

	orgRouter.Use(auth.AuthenticationMiddleware())
	{
		orgRouter.POST("/", func(context *gin.Context) {

			type Body struct {
				Name string `json:"name" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)

			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}

			orgName := body.Name
			existingOrg, _ := db.GetItem(db.ToKey("org", orgName), "#")
			if existingOrg != nil {
				if utils.HandlerError(context, errors.New("Already exists"), http.StatusBadRequest) {
					return
				}
			}
			org := types.Org{
				PK:        db.ToKey("org", orgName),
				SK:        "#",
				Name:      orgName,
				CreatedBy: context.GetString("id"),
				CreatedAt: time.Now().UnixNano(),
			}

			err = db.CreateItem(org)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			orgUser := types.OrgUser{
				PK:        db.ToKey("user", context.GetString("email")),
				SK:        db.ToKey("org", orgName),
				Role:      "admin",
				CreatedAt: time.Now().UnixNano(),
			}

			err = db.CreateItem(orgUser)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
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
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			orgId := context.Param("orgId")
			isValid := auth.RoleCheck(context, orgId, "admin")
			if !isValid {
				utils.HandlerError(context, errors.New("Forbidden"), http.StatusForbidden)
				return
			}
			orgUser := types.OrgUser{
				PK:        db.ToKey("user", body.Email),
				SK:        db.GenerateKey("org", orgId),
				Role:      body.Role,
				CreatedAt: time.Now().UnixNano(),
			}

			err = db.CreateItem(orgUser)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, "")
		})
	}

}
