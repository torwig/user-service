package requests

import (
	"encoding/json"
	"net/http"

	"github.com/torwig/user-service/entities"
	"github.com/torwig/user-service/ports/http/generated"
)

type UpdateUser struct {
	generated.UpdateUserJSONRequestBody
}

func NewUpdateUser(r *http.Request) (UpdateUser, error) {
	var req UpdateUser

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, ErrRequestBodyDecodingFailed
	}

	return req, nil
}

func (r UpdateUser) ToUpdateUserParams() entities.UpdateUserParams {
	return entities.UpdateUserParams{
		FirstName:   r.FirstName,
		LastName:    r.LastName,
		PhoneNumber: r.PhoneNumber,
		Address:     r.Address,
	}
}
