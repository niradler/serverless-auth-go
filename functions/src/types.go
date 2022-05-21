package main

type User struct {
	PK        string      `json:"pk,omitempty"`
	SK        string      `json:"sk,omitempty"`
	Email     string      `json:"email,omitempty"`
	Password  string      `json:"password,omitempty"`
	CreatedAt int64       `json:"createdAt,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type Org struct {
	PK        string `json:"pk,omitempty"`
	SK        string `json:"sk,omitempty"`
	name      string `json:"name,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

type OrgUser struct {
	PK        string `json:"pk,omitempty"`
	SK        string `json:"sk,omitempty"`
	role      string `json:"role,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

type UserPayload struct {
	Email    string
	Password string
	Data     interface{}
}

type UserCreated struct {
	Email     string `json:"email,omitempty"`
	CreatedAt int64  `json:"updatedAt,omitempty"`
}
