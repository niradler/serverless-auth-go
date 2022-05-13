package main

import (
	"fmt"
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
			check, _ := VerifyPassword(body.Password, HashPassword("demo"))
			if body.Email == "demo@demo.com" && check {
				token, refreshToken, _ := GenerateToken(body.Email)
				context.JSON(http.StatusOK, gin.H{
					"token":         token,
					"refresh_token": refreshToken,
					"message":       "Login success",
				})
				return
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
				log.Printf("fmt")
				fmt.Println(err)
				context.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "ValidationError",
						"message": err.Error(),
					})
				return
			}
			password := HashPassword(body.Password)
			// db.CreateUser(db.UserPayload{
			// 	Email:    "demo@demo.com",
			// 	Password: "Password",
			// })
			context.JSON(http.StatusOK, gin.H{
				"message":  "Signup success",
				"password": password,
				"email":    body.Email,
			})
		})
	}
}
