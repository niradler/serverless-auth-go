package main

import (
	"log"
	"testing"
)

func TestDB(test *testing.T) {
	Init()
	userCreated, err := CreateUser(UserPayload{
		Email:    "demo@demo.com",
		Password: "Password",
	})
	if err != nil {
		log.Println(err)
		test.Fail()
		return
	}
	log.Println("CreateUser")
	log.Println(userCreated)
	item, err := GetItem("org#test", "user#demo@demo.com")
	if err != nil {
		log.Println(err)
		test.Fail()
		return
	}
	log.Println("GetItem")
	log.Println(item)
	err = DeleteItem("org#test", "user#demo@demo.com")
	if err != nil {
		log.Println(err)
		test.Fail()
		return
	}
	log.Println("DeleteItem")
}
