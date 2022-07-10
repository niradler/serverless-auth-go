package routes

import (
	"errors"
	"net/http"
	"os"
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
				Model:     "org",
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
				Model:     "orgUser",
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

		orgRouter.GET("/:orgId/users", func(context *gin.Context) {
			orgId := context.Param("orgId")
			id, _ := context.Get("id")
			isValid := auth.RoleCheck(orgId, id.(string), "admin")
			if !isValid {
				utils.HandlerError(context, errors.New("Only admins can get org users"), http.StatusForbidden)
				return
			}

			orgUsers, err := db.GetOrgUsers(orgId)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}

			context.JSON(http.StatusOK, orgUsers)
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
			id, _ := context.Get("email")
			isValid := auth.RoleCheck(orgId, id.(string), "admin")
			if !isValid {
				utils.HandlerError(context, errors.New("Only admins can invite users"), http.StatusForbidden)
				return
			}
			orgUser := types.OrgUser{
				PK:        db.ToKey("user", body.Email),
				SK:        db.GenerateKey("org", orgId),
				Role:      body.Role,
				Model:     "orgUser",
				CreatedAt: time.Now().UnixNano(),
			}

			err = db.CreateItem(orgUser)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}

			err = utils.SendGenericEmail(
				utils.EmailGenericRequest{
					To:      body.Email,
					Subject: "Invitation to join " + orgId,
					Args: utils.GenericEmailArgs{
						Content:    "You're invited to " + orgId + ". Please click the link below and signup to join.",
						SubContent: "To complete the process follow the link below:",
						Title:      "Invitation to join " + orgId,
						Label:      "Join",
						Logo:       os.Getenv("SLS_AUTH_APP_NAME"),
						Link:       os.Getenv("SLS_AUTH_DOMAIN"),
						Contact:    os.Getenv("SLS_AUTH_APP_CONTACT"),
						Company:    os.Getenv("SLS_AUTH_APP_NAME"),
					},
				})
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, "")
		})
	}

}
