package main

import (
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// var userTable = os.Getenv("USERS_TABLE")
var usersTable = "ServerlessAuthStack-table"

var db *dynamodb.DynamoDB

func Init() {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
	if err != nil {
		log.Println(err)
		return
	}
	db = dynamodb.New(sess)
}

func GetDB() *dynamodb.DynamoDB {
	return db
}

type User struct {
	ID        string `json:"root_obj_id,omitempty"`
	Email     string `json:"sub_obj_id"`
	Password  string `json:"password"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
}

type UserPayload struct {
	Email    string
	Password string
}

type UserCreated struct {
	ID        string `json:"id,omitempty"`
	Email     string `json:"email"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
}

func CreateUser(userPayload UserPayload) (*UserCreated, error) {
	db := GetDB()
	user := User{
		ID:        "org#test",
		Email:     "user#" + userPayload.Email,
		Password:  userPayload.Password,
		UpdatedAt: time.Now().UnixNano(),
	}
	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Println(err)
		errors.New("error when try to convert user data to dynamodbattribute")
		return nil, err
	}
	params := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(usersTable),
	}
	if _, err := db.PutItem(params); err != nil {
		log.Println(err)
		return nil, errors.New("error when try to save data to database")
	}
	userCreated := UserCreated{
		ID:        user.ID,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
	}
	return &userCreated, nil
}

func main() {
	log.Println("create user")
	Init()
	userCreated, err := CreateUser(UserPayload{
		Email:    "demo@demo.com",
		Password: "Password",
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(userCreated)
}
