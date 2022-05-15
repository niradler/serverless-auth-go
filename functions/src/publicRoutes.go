package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadPublicRoutes(router *gin.RouterGroup) {

	public := router.Group("/auth")
	{
		public.POST("/login", func(context *gin.Context) {
			type Body struct {
				Email    string `json:"email" binding:"required"`
				Password string `json:"password" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if err != nil {
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": err.Error(),
					})
				return
			}
			user, err := GetItem("org#default", "user#"+body.Email)
			if err != nil {
				log.Println(err)
			}
			if user != nil {
				log.Println(user)
				check, _ := VerifyPassword(body.Password, user["password"].(string))
				if check {
					token, refreshToken, _ := GenerateToken(body.Email)
					context.JSON(http.StatusOK, gin.H{
						"token":         token,
						"refresh_token": refreshToken,
						"message":       "Login success",
					})
					return
				}
			}

			context.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{
					"error":   "ValidationError",
					"message": "Failed to validate",
				})
		})

		public.POST("/signup", func(context *gin.Context) {
			type Body struct {
				Email    string      `json:"email" binding:"required"`
				Password string      `json:"password" binding:"required"`
				Data     interface{} `json:"data"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if err != nil {
				log.Println(err)
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": err.Error(),
					})
				return
			}
			user, _ := GetItem("org#default", "user#"+body.Email)
			if user != nil {
				log.Println("Already exists")
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": "Already exists",
					})
				return
			}
			password := HashPassword(body.Password)
			_, err = CreateUser(UserPayload{
				Email:    body.Email,
				Password: password,
				Data:     body.Data,
			})
			if err != nil {
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": err.Error(),
					})
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"message": "Signup success",
				"email":   body.Email,
			})
		})
	}
}
