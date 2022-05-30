package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
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
		var credentials = credentials.NewSharedCredentials("", "default")
		if os.Getenv("LAMBDA_TASK_ROOT") != "" {
			credentials = nil
		}
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String("us-east-1"),
			Credentials: credentials,
		})
		if err != nil {
			log.Println(err)
		}
		db = dynamodb.New(sess)
	}
	return db
}

func CreateItem[Payload User | Org | OrgUser](payload Payload) error {
	db := GetDB()

	item, err := dynamodbattribute.MarshalMap(payload)
	if err != nil {
		log.Println(err)
		errors.New("error when try to convert user data to dynamodbattribute")
		return err
	}
	params := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(usersTable),
	}
	if _, err := db.PutItem(params); err != nil {
		log.Println(err)
		return errors.New("error when try to save data to database")
	}
	return nil
}

func GetItem(pk string, sk string) (map[string]interface{}, error) {
	db := GetDB()
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(pk),
			},
			"sk": {
				S: aws.String(sk),
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

func GetItemByPK(pk string) ([]map[string]*dynamodb.AttributeValue, error) {
	db := GetDB()
	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("pk").Equal(expression.Value(pk))).
		Build()
	result, err := db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(usersTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return result.Items, nil
}

// func GetUserContext(email string) (interface{}, error) {
// 	result, err := GetItemByPK(email)
// 	var user map[string]interface{}
// 	for _, i := range result.Items {
// 		var item interface{}
// 		err = dynamodbattribute.UnmarshalMap(i, &item)
// 		if err == nil {
// 			itemData := item.(map[string]interface{})
// 			sk := strings.Split(itemData["sk"].(string), "#")

// 			switch sk[0] {
// 			case "org":
// 				log.Println("org")
// 			case "user":
// 				user =
// 			}
// 		}

// 	}

// 	return "test", nil
// }
func DeleteItem(pk string, sk string) error {
	db := GetDB()
	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(pk),
			},
			"sk": {
				S: aws.String(sk),
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

func CreateUser(userPayload UserPayload) (*UserCreated, error) {
	// db := GetDB()
	user := User{
		PK:        "user#" + userPayload.Email,
		SK:        "user#" + userPayload.Email,
		Email:     userPayload.Email,
		Password:  userPayload.Password,
		CreatedAt: time.Now().UnixNano(),
		Data:      userPayload.Data,
	}

	err := CreateItem(user)
	if err != nil {
		log.Println(err)
		errors.New("CreateItem - user - " + err.Error())
		return nil, err
	}

	// item, err := dynamodbattribute.MarshalMap(user)
	// if err != nil {
	// 	log.Println(err)
	// 	errors.New("error when try to convert user data to dynamodbattribute")
	// 	return nil, err
	// }
	// params := &dynamodb.PutItemInput{
	// 	Item:      item,
	// 	TableName: aws.String(usersTable),
	// }
	// if _, err := db.PutItem(params); err != nil {
	// 	log.Println(err)
	// 	return nil, errors.New("error when try to save data to database")
	// }
	userCreated := UserCreated{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	return &userCreated, nil
}
