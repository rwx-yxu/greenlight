package brokers

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
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
	Insert(movie *models.Movie) error
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

func (m movie) Insert(movie *models.Movie) error {

	// Define the SQL query for inserting a new record in the movies table and returning
	// the system-generated data.
	query := `
        INSERT INTO movies (title, year, runtime, genres)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters from
	// the movie struct. Declaring this slice immediately next to our SQL query helps to
	// make it nice and clear *what values are being used where* in the query.
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system-
	// generated id, created_at and version values into the movie struct.
	return m.db.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)

}
