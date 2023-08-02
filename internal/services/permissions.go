package services

import (
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
)

type permission struct {
	Broker brokers.PermissionReadWriter
}

type PermissionReader interface {
	FindAllForUser(userID int64) (models.Permissions, error)
}

type PermissionWriter interface {
	AddForUser(userID int64, codes ...string) error
}

type PermissionReadWriter interface {
	PermissionReader
	PermissionWriter
}

func NewPermission(b brokers.PermissionReadWriter) PermissionReadWriter {
	return &permission{
		Broker: b,
	}
}

func (p permission) FindAllForUser(userID int64) (models.Permissions, error) {
	perms, err := p.Broker.GetAllForUser(userID)
	if err != nil {
		return models.Permissions{}, err
	}

	return perms, nil
}

func (p permission) AddForUser(userID int64, codes ...string) error {
	err := p.Broker.InsertForUser(userID, codes...)
	if err != nil {
		return err
	}
	return nil
}
