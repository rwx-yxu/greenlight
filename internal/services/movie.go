package services

import (
	"time"

	"github.com/rwx-yxu/greenlight/internal/brokers"
	"github.com/rwx-yxu/greenlight/internal/models"
	"github.com/rwx-yxu/greenlight/internal/validator"
)

type movie struct {
	Broker brokers.MovieReadWriteDeleter
}

type MovieValidator interface {
	Validate(input models.Movie) validator.Validator
}

type MovieReader interface {
	FindByID(id int64) (*models.Movie, error)
}

type MovieWriter interface {
	Add(m *models.Movie) error
	Edit(m *models.Movie) error
}

type MovieDeleter interface {
	RemoveByID(id int64) error
}

type MovieReadWriteDeleter interface {
	MovieValidator
	MovieDeleter
	MovieWriter
	MovieReader
}

func NewMovie(b brokers.MovieReadWriteDeleter) MovieReadWriteDeleter {
	return &movie{
		Broker: b,
	}
}

func (movie) Validate(input models.Movie) validator.Validator {
	// Initialize a new Validator instance.
	v := validator.New()

	// Use the Check() method to execute our validation checks. This will add the
	// provided key and error message to the errors map if the check does not evaluate
	// to true. For example, in the first line here we "check that the title is not
	// equal to the empty string". In the second, we "check that the length of the title
	// is less than or equal to 500 bytes" and so on.
	v.Check(input.Title != "", "title", "must be provided")
	v.Check(len(input.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(input.Genres != nil, "genres", "must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genres")
	// Note that we're using the Unique helper in the line below to check that all
	// values in the input.Genres slice are unique.
	v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")
	return *v
}

func (m movie) FindByID(id int64) (*models.Movie, error) {
	return nil, nil
}

func (m movie) Add(movie *models.Movie) error {
	err := m.Broker.Insert(movie)
	if err != nil {
		return err
	}
	return nil
}

func (m movie) Edit(movie *models.Movie) error {
	return nil
}

func (m movie) RemoveByID(id int64) error {
	return nil
}
