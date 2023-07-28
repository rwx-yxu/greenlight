package services

import (
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
	"github.com/rwx-yxu/greenlight/internal/validator"
)

type user struct {
	Broker brokers.UserReadWriter
}

type UserReader interface {
	//FindByEmail(email string) (*models.User, error)
	FindByToken(scope, tokenPlainText string) (*models.User, error)
}

type UserWriter interface {
	Add(user *models.User) (*validator.Validator, error)
	Edit(user *models.User) (*validator.Validator, error)
}

type UserReadWriter interface {
	UserReader
	UserWriter
}

func NewUser(b brokers.UserReadWriter) UserReadWriter {
	return &user{
		Broker: b,
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *models.User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, user.Email)

	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.
	if user.Password.Plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.Plaintext)
	}

	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set a password for the user). It's a
	// useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map we
	// raise a panic instead.
	if user.Password.Hash == nil {
		panic("missing password hash for user")
	}
}

func (u user) Add(user *models.User) (*validator.Validator, error) {
	v := validator.New()
	ValidateUser(v, user)
	if v.Valid() {
		err := u.Broker.Insert(user)
		if err != nil {
			return v, err
		}
	}

	return v, nil
}

func (u user) Edit(user *models.User) (*validator.Validator, error) {
	v := validator.New()
	ValidateUser(v, user)
	if v.Valid() {
		err := u.Broker.Update(user)
		if err != nil {
			return v, err
		}
	}
	return v, nil
}

func (u user) FindByToken(scope, tokenPlainText string) (*models.User, error) {
	user, err := u.Broker.GetByToken(scope, tokenPlainText)
	if err != nil {
		return nil, err
	}
	return user, nil
}
