package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
)

func main() {

	fileName := "email_login.html"

	tmpl := template.Must(template.ParseFiles(fileName))
	buffer := new(bytes.Buffer)
	err := tmpl.Execute(buffer, map[string]string{
		"URL": "http://localhost:8280/v1/auth/login",
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
	body := buffer.String()
	os.WriteFile("./compiled/"+fileName, []byte(body), 0644)

}
