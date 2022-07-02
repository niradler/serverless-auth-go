package utils

import (
	"crypto/md5"
	"encoding/hex"
	"os"

	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

var Dump = spew.Dump
var Logger *zap.Logger
var Debug = os.Getenv("SLS_AUTH_DEBUG") == "true"

func InitializeLogger() {
	Logger, _ = zap.NewProduction()
}

func ToHashId(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func HandlerError(context *gin.Context, err error, status int) bool {
	if err != nil {
		Logger.Info("handlerError", zap.Error(err))
		context.AbortWithStatusJSON(status,
			gin.H{
				"error":   "Error",
				"message": err.Error(),
			})
		return true
	}
	return false
}
