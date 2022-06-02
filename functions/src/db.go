package main

import (
	"errors"
	"os"
	"strings"
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
		}
		db = dynamodb.New(sess)
	}
	return db
}

func toKey(key string, id string) string {
	return key + "#" + id
}

func fromKey(keyString string) (string, string) {
	obj := strings.Split(keyString, "#")
	return obj[0], obj[1]
}

func CreateItem[Payload User | Org | OrgUser](payload Payload) error {
	db := GetDB()

	item, err := dynamodbattribute.MarshalMap(payload)
	if err != nil {
		errors.New("error when try to convert user data to dynamodbattribute")
		return err
	}
	params := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(usersTable),
	}
	if _, err := db.PutItem(params); err != nil {
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
		return nil, err
	}
	var item interface{}
	if err := dynamodbattribute.UnmarshalMap(resp.Item, &item); err != nil {
		return nil, err
	}
	if item != nil {
		value := item.(map[string]interface{})

		return value, nil
	}

	return nil, nil
}

func GetItemByPK(key string) ([]interface{}, error) {
	db := GetDB()
	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("pk").Equal(expression.Value(key))).
		Build()
	result, err := db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(usersTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var items interface{}
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}
	if items != nil {
		value := items.([]interface{})

		return value, nil
	}
	return nil, nil
}

func GetUserContext(email string) (*UserContext, error) {
	items, _ := GetItemByPK(toKey("user", email))
	if items == nil {
		return nil, errors.New("user not found")
	}
	var user UserContext
	var orgs []OrgContext
	for _, i := range items {
		obj := i.(map[string]interface{})
		_, pkId := fromKey(obj["pk"].(string))
		sk, skId := fromKey(obj["sk"].(string))
		switch sk {
		case "org":
			orgs = append(orgs, OrgContext{
				Id:   skId,
				Name: skId,
				Role: obj["role"].(string),
			})
		case "user":
			user = UserContext{
				Id:    pkId,
				Email: obj["email"].(string),
				Data:  obj["data"].(interface{}),
			}

		}
	}
	return &UserContext{
		Id:    user.Id,
		Email: user.Email,
		Data:  user.Data,
		Orgs:  orgs,
	}, nil
}
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
