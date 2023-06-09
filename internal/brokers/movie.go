package brokers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/rwx-yxu/greenlight/internal/filter"
	"github.com/rwx-yxu/greenlight/internal/models"
)

type movie struct {
	db *sql.DB
}

type MovieReader interface {
	GetByID(id int64) (*models.Movie, error)
	GetAll(title string, genres []string, f filter.Filter) ([]*models.Movie, filter.Metadata, error)
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
	ErrEditConflict   = errors.New("edit conflict")
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
	// Use the context.WithTimeout() function to create a context.Context which carries a
	// 3-second timeout deadline. Note that we're using the empty context.Background()
	// as the 'parent' context.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// Movie struct. Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.
	err := m.db.QueryRowContext(ctx, query, id).Scan(
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
	// Declare the SQL query for updating the record and returning the new version
	// number.
	query := `
        UPDATE movies
        SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version`

	// Create an args slice containing the values for the placeholder parameters.
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the SQL query. If no matching row could be found, we know the movie
	// version has changed (or the record has been deleted) and we return our custom
	// ErrEditConflict error.
	err := m.db.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m movie) DeleteByID(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
        DELETE FROM movies
        WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the SQL query using the Exec() method, passing in the id variable as
	// the value for the placeholder parameter. The Exec() method returns a sql.Result
	// object.
	result, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number of rows
	// affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, we know that the movies table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system-
	// generated id, created_at and version values into the movie struct.
	return m.db.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)

}

func (m movie) GetAll(title string, genres []string, f filter.Filter) ([]*models.Movie, filter.Metadata, error) {
	// Construct the SQL query to retrieve all movie records.
	query := fmt.Sprintf(`
        SELECT count(*) OVER(),id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
        AND (genres @> $2 OR $2 = '{}')
        ORDER BY %s %s, id ASC
				LIMIT $3 OFFSET $4`, f.SortColumn(), f.SortDirection())

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{title, pq.Array(genres), f.Limit(), f.Offset()}
	// Use QueryContext() to execute the query. This returns a sql.Rows resultset
	// containing the result.
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, filter.Metadata{}, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the movie data.
	movies := []*models.Movie{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var movie models.Movie

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, filter.Metadata{}, err
		}

		// Add the Movie struct to the slice.
		movies = append(movies, &movie)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, filter.Metadata{}, err
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := filter.CalculateMetadata(totalRecords, f.Page, f.PageSize)

	// If everything went OK, then return the slice of movies.
	return movies, metadata, nil
}
