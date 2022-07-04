package auth

import (
	"github.com/niradler/social-lab/src/types"
	"testing"
)

func TestPasswordValidation(test *testing.T) {
	hash := HashPassword("Password")
	valid, _ := VerifyPassword("Password", hash)
	if valid != true {
		test.Fail()
	}
}

func TestGenerateToken(test *testing.T) {
	_, _, err := GenerateToken(types.UserContext{
		Id:    "1",
		Email: "test@test.com",
		Data:  nil,
		Orgs: []types.OrgContext{
			types.OrgContext{
				Id:   "1",
				Role: "admin",
			},
		},
	})

	if err != nil {
		test.Fail()
	}
}

func TestValidateToken(test *testing.T) {
	token, _, err := GenerateToken(types.UserContext{
		Id:    "1",
		Email: "test@test.com",
		Data:  nil,
		Orgs: []types.OrgContext{
			types.OrgContext{
				Id:   "1",
				Role: "admin",
			},
		},
	})

	if err != nil {
		test.Fail()
	}

	claims, err := ValidateToken(token)

	if err != nil {
		test.Fail()
	}

	if claims.Email != "test@test.com" || claims.Id != "1" {
		test.Fail()
	}
}

func TestValidateRefreshToken(test *testing.T) {
	_, refreshToken, err := GenerateToken(types.UserContext{
		Id:    "1",
		Email: "test@test.com",
		Data:  nil,
		Orgs: []types.OrgContext{
			types.OrgContext{
				Id:   "1",
				Role: "admin",
			},
		},
	})

	if err != nil {
		test.Fail()
	}

	claims, err := ValidateRefreshToken(refreshToken)

	if err != nil {
		test.Fail()
	}

	if claims.Email != "test@test.com" || claims.Id != "1" {
		test.Fail()
	}
}
