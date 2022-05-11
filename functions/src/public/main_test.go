package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignup(test *testing.T) {
	router := gin.Default()
	LoadRoutes(router)
	var userData = []byte(`{
		"email": "demo@demo.com",
		"password": "demo"
	}`)
	req, _ := http.NewRequest("POST", "/public/signup", bytes.NewBuffer(userData))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")

	testHTTPResponse(test, router, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusOK

		p, _ := ioutil.ReadAll(w.Body)

		response := string(p)

		type ResponseData struct {
			Email string
		}
		var responseData ResponseData
		json.Unmarshal([]byte(response), &responseData)
		checkEmail := responseData.Email == "demo@demo.com"

		return statusOK && checkEmail
	})
}
