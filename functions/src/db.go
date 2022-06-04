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

var usersTable = os.Getenv("USERS_TABLE")
var db *dynamodb.DynamoDB

func Init() {
	if usersTable == "" {
		usersTable = "ServerlessAuthStack-table"
	}
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

func generateKey(key string, id string) string {
	return key + "#" + id
}

func toKey(key string, id string) string {
	return generateKey(key, toHashId(id))
}

func fromKey(keyString string) (string, string) {
	obj := strings.Split(keyString, "#")
	return obj[0], obj[1]
}

func CreateItem[Payload User | Org | OrgUser](payload Payload) error {
	db := GetDB()

	item, err := dynamodbattribute.MarshalMap(payload)
	if err != nil {
		dump(err)
		errors.New("error when try to convert user data to dynamodbattribute")
		return err
	}
	params := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(usersTable),
	}
	if _, err := db.PutItem(params); err != nil {
		dump(err)
		return errors.New("error when try to save data to database")
	}
	return nil
}

func UpdateUser(id string, data interface{}) error {
	db := GetDB()
	pk := generateKey("user", id)
	upd := expression.
		Set(expression.Name("data"), expression.Value(data))
	expr, err := expression.NewBuilder().WithUpdate(upd).Build()
	if err != nil {
		return err
	}
	params := &dynamodb.UpdateItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(pk),
			},
			"sk": {
				S: aws.String(pk),
			},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	}
	if _, err := db.UpdateItem(params); err != nil {
		dump(err)
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

	user := User{
		PK:        toKey("user", userPayload.Email),
		SK:        toKey("user", userPayload.Email),
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

	userCreated := UserCreated{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	return &userCreated, nil
}
