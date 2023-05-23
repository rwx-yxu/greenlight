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
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no movies will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	query := `
        SELECT id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE id = $1`

	// Declare a Movie struct to hold the data returned by the query.
	movie := new(models.Movie)

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// Movie struct. Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.
	err := m.db.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Otherwise, return a pointer to the Movie struct.
	return movie, nil

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
