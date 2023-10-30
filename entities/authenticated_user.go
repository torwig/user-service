package entities

type AuthenticatedUser struct {
	id              int64
	canCreate       bool
	canDelete       bool
	canUpdateOthers bool
	canViewOthers   bool
}

type UserAuthOption func(user *AuthenticatedUser)

func CreateUsersGranted() UserAuthOption {
	return func(au *AuthenticatedUser) {
		au.canCreate = true
	}
}

func DeleteUsersGranted() UserAuthOption {
	return func(au *AuthenticatedUser) {
		au.canDelete = true
	}
}

func UpdateUsersGranted() UserAuthOption {
	return func(au *AuthenticatedUser) {
		au.canUpdateOthers = true
	}
}

func ViewUsersGranted() UserAuthOption {
	return func(au *AuthenticatedUser) {
		au.canViewOthers = true
	}
}

func NewAuthenticatedUser(id int64, options ...UserAuthOption) *AuthenticatedUser {
	au := &AuthenticatedUser{id: id}

	for _, o := range options {
		o(au)
	}

	return au
}

func (au AuthenticatedUser) ID() int64 {
	return au.id
}

func (au AuthenticatedUser) CanCreate() bool {
	return au.canCreate
}

func (au AuthenticatedUser) CanDelete(id int64) bool {
	return au.canDelete && au.id != id
}

func (au AuthenticatedUser) CanUpdateUser(id int64) bool {
	return au.canUpdateOthers || au.id == id
}

func (au AuthenticatedUser) CanViewUser(id int64) bool {
	return au.canViewOthers || au.id == id
}
