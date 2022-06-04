package main

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
)

var dump = spew.Dump

var Logger *zap.Logger

func InitializeLogger() {
	Logger, _ = zap.NewProduction()
}

func toHashId(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
