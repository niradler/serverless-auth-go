package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	// "strings"
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
		expectedResponse := `{"email":"demo@demo.com","message":"Signup success","password":"$2a$14$v81Dt.1cPN5zkv5D3N4DP.8dPYlTPXYlypj9Wu0Gmt8QVeQrV7bXa"}`
		response := string(p)
		fmt.Println(reflect.TypeOf(response))
		fmt.Println(reflect.TypeOf(expectedResponse))
		fmt.Println(response == expectedResponse)
		fmt.Println(response)
		fmt.Println(expectedResponse)

		return statusOK
	})
}
