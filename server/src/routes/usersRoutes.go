package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/niradler/social-lab/src/auth"
	"github.com/niradler/social-lab/src/db"
	"github.com/niradler/social-lab/src/utils"
)

func LoadUsersRoutes(router *gin.RouterGroup) {

	usersRouter := router.Group("/users")

	usersRouter.Use(auth.AuthenticationMiddleware())
	{
		usersRouter.GET("/me", func(context *gin.Context) {
			userContext, err := db.GetUserContext(context.GetString("email"))
			if utils.HandlerError(context, err, http.StatusBadRequest) {
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
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			err = db.UpdateUser(context.GetString("id"), body.Data)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, "")
		})
	}
}
