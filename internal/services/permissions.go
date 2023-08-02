package services

import (
	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
)

type permission struct {
	Broker brokers.PermissionReader
}

type PermissionReader interface {
	FindAllForUser(userID int64) (models.Permissions, error)
}

func NewPermission(b brokers.PermissionReader) PermissionReader {
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
