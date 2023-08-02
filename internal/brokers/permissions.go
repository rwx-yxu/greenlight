package brokers

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/rwx-yxu/greenlight/internal/models"
)

type permission struct {
	db *sql.DB
}

type PermissionWriter interface {
	InsertForUser(userId int64, codes ...string) error
}

type PermissionReader interface {
	GetAllForUser(userID int64) (models.Permissions, error)
}

type PermissionReadWriter interface {
	PermissionWriter
	PermissionReader
}

func NewPermission(db *sql.DB) PermissionReadWriter {
	return &permission{db: db}
}

// The GetAllForUser() method returns all permission codes for a specific user in a
// Permissions slice. The code in this method should feel very familiar --- it uses the
// standard pattern that we've already seen before for retrieving multiple data rows in
// an SQL query.
func (p permission) GetAllForUser(userID int64) (models.Permissions, error) {
	query := `
        SELECT permissions.code
        FROM permissions
        INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
        INNER JOIN users ON users_permissions.user_id = users.id
        WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions models.Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// Add the provided permission codes for a specific user. Notice that we're using a
// variadic parameter for the codes so that we can assign multiple permissions in a
// single call.
func (p permission) InsertForUser(userID int64, codes ...string) error {
	query := `
        INSERT INTO users_permissions
        SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.db.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
