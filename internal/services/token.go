package services

import (
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
	"github.com/rwx-yxu/greenlight/internal/validator"
)

type token struct {
	Broker brokers.TokenWriteDeleter
}

type TokenValidator interface {
	ValidatePlainText(v *validator.Validator, tokenPlaintext string)
}

type TokenWriter interface {
	Add(token *models.Token) (*validator.Validator, error)
}

type TokenDeleter interface {
	RemoveAllForUser(scope string, userID int64) error
}

type TokenWriteDeleter interface {
	TokenValidator
	TokenWriter
	TokenDeleter
}

func NewToken(b brokers.TokenWriteDeleter) TokenWriteDeleter {
	return &token{
		Broker: b,
	}
}

func (t token) ValidatePlainText(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func (t token) Add(token *models.Token) (*validator.Validator, error) {
	v := validator.New()
	t.ValidatePlainText(v, token.Plaintext)
	if v.Valid() {
		err := t.Broker.Insert(token)
		if err != nil {
			return v, err
		}
	}
	return v, nil
}

func (t token) RemoveAllForUser(scope string, userID int64) error {
	err := t.Broker.DeleteAllForUser(scope, userID)
	if err != nil {
		return err
	}
	return nil
}
