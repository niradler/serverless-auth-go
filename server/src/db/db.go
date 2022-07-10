package db

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

	"github.com/niradler/social-lab/src/types"
	"github.com/niradler/social-lab/src/utils"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
)

var appTable = os.Getenv("AUTH_APP_TABLE")
var db *dynamodb.DynamoDB

func GetDB() *dynamodb.DynamoDB {

	if appTable == "" {
		appTable = "ServerlessAuthStack-table"
	}

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

func GenerateKey(key string, id string) string {
	return key + "#" + id
}

func ToKey(key string, id string) string {
	return GenerateKey(key, utils.ToHashId(id))
}

func fromKey(keyString string) (string, string) {
	obj := strings.Split(keyString, "#")
	return obj[0], obj[1]
}

func CreateItem[Payload types.User | types.Org | types.OrgUser](payload Payload) error {
	db := GetDB()

	item, err := dynamodbattribute.MarshalMap(payload)
	if err != nil {
		utils.Dump(err)
		errors.New("error when try to convert user data to dynamodbattribute")
		return err
	}
	params := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(appTable),
	}
	if _, err := db.PutItem(params); err != nil {
		utils.Dump(err)
		return errors.New("error when try to save data to database")
	}
	return nil
}

func UpdateUser(id string, data interface{}) error {
	db := GetDB()
	pk := GenerateKey("user", id)
	upd := expression.
		Set(expression.Name("data"), expression.Value(data))
	expr, err := expression.NewBuilder().WithUpdate(upd).Build()
	if err != nil {
		return err
	}
	params := &dynamodb.UpdateItemInput{
		TableName: aws.String(appTable),
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
		utils.Dump(err)
		return errors.New("error when try to save data to database")
	}
	return nil
}

func UpdateUserPassword(id string, password string) error {
	db := GetDB()
	pk := GenerateKey("user", id)
	upd := expression.
		Set(expression.Name("password"), expression.Value(password))
	expr, err := expression.NewBuilder().WithUpdate(upd).Build()
	if err != nil {
		return err
	}
	params := &dynamodb.UpdateItemInput{
		TableName: aws.String(appTable),
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
		utils.Dump(err)
		return errors.New("error when try to save data to database")
	}
	return nil
}

func GetItem(pk string, sk string) (map[string]interface{}, error) {
	utils.Logger.Info("GetItem", zap.String("pk", pk), zap.String("sk", sk))
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
		TableName: aws.String(appTable),
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
		TableName:                 aws.String(appTable),
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

func GetUserContext(email string) (*types.UserContext, error) {
	items, err := GetItemByPK(ToKey("user", email))

	if err != nil {
		return nil, err
	}

	if items == nil {
		return nil, errors.New("user not found")
	}
	var user types.UserContext
	var orgs []types.OrgContext
	for _, i := range items {
		obj := i.(map[string]interface{})
		_, pkId := fromKey(obj["pk"].(string))
		_, skId := fromKey(obj["sk"].(string))
		switch obj["model"].(string) {
		case "orgUser":
			orgs = append(orgs, types.OrgContext{
				Id:   skId,
				Role: obj["role"].(string),
			})
		case "user":
			data := obj["data"]
			if data != nil {
				user = types.UserContext{
					Id:    pkId,
					Email: obj["email"].(string),
					Data:  data.(interface{}),
				}
			} else {
				user = types.UserContext{
					Id:    pkId,
					Email: obj["email"].(string),
					Data:  "",
				}
			}
		}
	}

	if user.Id == "" {
		return nil, errors.New("user error")
	}

	return &types.UserContext{
		Id:    user.Id,
		Email: user.Email,
		Data:  user.Data,
		Orgs:  orgs,
	}, nil
}

func GetOrgUsers(orgId string) ([]types.OrgUser, error) {
	db := GetDB()
	orgKey := GenerateKey("org", orgId)
	result, err := db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(appTable),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var items []types.OrgUser
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}
	orgUsers := funk.Filter(items, func(orgUser types.OrgUser) bool {
		utils.Dump(orgUser)
		return orgUser.SK == orgKey
	})

	return orgUsers.([]types.OrgUser), nil
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
		TableName: aws.String(appTable),
	}
	_, err := db.DeleteItem(params)
	if err != nil {
		return err
	}

	return nil
}

func CreateUser(userPayload types.UserPayload) (*types.UserCreated, error) {

	user := types.User{
		PK:        ToKey("user", userPayload.Email),
		SK:        "#",
		Model:     "user",
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

	userCreated := types.UserCreated{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	return &userCreated, nil
}
