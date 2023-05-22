package brokers

import (
	"database/sql"
	"errors"

	"github.com/rwx-yxu/greenlight/internal/models"
)

type movie struct {
	db *sql.DB
}

type MovieReader interface {
	GetByID(id int64) (*models.Movie, error)
}

type MovieWriter interface {
	Update(m *models.Movie) error
}

type MovieDeleter interface {
	DeleteByID(id int64) error
}

type MovieReadWriteDeleter interface {
	MovieReader
	MovieWriter
	MovieDeleter
}

var (
	ErrRecordNotFound = errors.New("record not found")
)

func NewMovie(db *sql.DB) MovieReadWriteDeleter {
	return &movie{db: db}
}

func (m movie) GetByID(id int64) (*models.Movie, error) {
	return nil, nil
}

func (m movie) Update(movie *models.Movie) error {
	return nil
}

func (m movie) DeleteByID(id int64) error {
	return nil
}
