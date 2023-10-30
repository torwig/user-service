package responses

import (
	"github.com/torwig/user-service/entities"
	"github.com/torwig/user-service/ports/http/generated"
)

func UserFromEntity(u entities.User) generated.User {
	return generated.User{
		Id:          u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		PhoneNumber: u.PhoneNumber,
		Address:     u.Address,
	}
}
