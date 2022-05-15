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
	if db == nil {
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String("us-east-1"),
			Credentials: credentials.NewSharedCredentials("", "default"),
		})
		if err != nil {
			log.Println(err)
		}
		db = dynamodb.New(sess)
	}
	return db
}

type User struct {
	ID        string      `json:"root_obj_id,omitempty"`
	Email     string      `json:"sub_obj_id"`
	Password  string      `json:"password"`
	UpdatedAt int64       `json:"updatedAt,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type UserPayload struct {
	Email    string
	Password string
	Data     interface{}
}

type UserCreated struct {
	ID        string `json:"id,omitempty"`
	Email     string `json:"email"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
}

func CreateUser(userPayload UserPayload) (*UserCreated, error) {
	db := GetDB()
	user := User{
		ID:        "org#default",
		Email:     "user#" + userPayload.Email,
		Password:  userPayload.Password,
		UpdatedAt: time.Now().UnixNano(),
		Data:      userPayload.Data,
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

func GetItem(rootObjId string, subObjId string) (map[string]interface{}, error) {
	db := GetDB()
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"root_obj_id": {
				S: aws.String(rootObjId),
			},
			"sub_obj_id": {
				S: aws.String(subObjId),
			},
		},
		TableName: aws.String(usersTable),
	}
	resp, err := db.GetItem(params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var item interface{}
	if err := dynamodbattribute.UnmarshalMap(resp.Item, &item); err != nil {
		log.Println(err)
		return nil, err
	}
	if item != nil {
		value := item.(map[string]interface{})

		return value, nil
	}

	return nil, nil
}

func DeleteItem(rootObjId string, subObjId string) error {
	db := GetDB()
	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"root_obj_id": {
				S: aws.String(rootObjId),
			},
			"sub_obj_id": {
				S: aws.String(subObjId),
			},
		},
		TableName: aws.String(usersTable),
	}
	_, err := db.DeleteItem(params)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
