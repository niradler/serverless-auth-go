package types

type User struct {
	PK        string      `json:"pk,omitempty"`
	SK        string      `json:"sk,omitempty"`
	Model     string      `json:"model,omitempty"`
	Email     string      `json:"email,omitempty"`
	Password  string      `json:"password,omitempty"`
	CreatedAt int64       `json:"createdAt,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type Org struct {
	PK        string `json:"pk,omitempty"`
	SK        string `json:"sk,omitempty"`
	Model     string `json:"model,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

type OrgUser struct {
	PK        string `json:"pk,omitempty"`
	SK        string `json:"sk,omitempty"`
	Model     string `json:"model,omitempty"`
	Role      string `json:"role,omitempty"`
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

type OrgContext struct {
	Id   string `json:"id,omitempty"`
	Role string `json:"role,omitempty"`
}
type UserContext struct {
	Id    string       `json:"id,omitempty"`
	Email string       `json:"email,omitempty"`
	Orgs  []OrgContext `json:"orgs,omitempty"`
	Data  interface{}  `json:"data,omitempty"`
}
