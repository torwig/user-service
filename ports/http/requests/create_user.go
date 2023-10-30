package requests

import (
	"encoding/json"
	"net/http"

	"github.com/torwig/user-service/entities"

	"github.com/torwig/user-service/ports/http/generated"
)

type CreateUser struct {
	generated.CreateUserJSONRequestBody
}

func NewCreateUser(r *http.Request) (CreateUser, error) {
	var req CreateUser

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, ErrRequestBodyDecodingFailed
	}

	return req, req.Validate()
}

func (r CreateUser) Validate() error {
	if r.FirstName == "" || r.LastName == "" || r.PhoneNumber == "" || r.Address == "" {
		return ErrEmptyRequestField
	}

	return nil
}

func (r CreateUser) ToCreateUserParams() entities.CreateUserParams {
	return entities.CreateUserParams{
		FirstName:   r.FirstName,
		LastName:    r.LastName,
		PhoneNumber: r.PhoneNumber,
		Address:     r.Address,
	}
}
