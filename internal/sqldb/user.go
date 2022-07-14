package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	sq "github.com/Masterminds/squirrel"
)

// User contains the database entry
type User struct {
	ID          int       `db:"id"`
	Email       string    `db:"email"`
	Password    string    `json:"-" db:"password"`
	Description string    `db:"description"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// UserByEmail loads User with given email, returns nil if not found
func (db *Client) FindOneUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}
	stmt, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").
		From("users").Where(sq.Eq{"email": email}).Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	db.logQuery(stmt, args...)

	err = db.GetContext(ctx, &u, stmt, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// Insert sanitizes and inserts a User in database
func (db *Client) InsertUser(ctx context.Context, u *User) error {
	if u.Password == "" {
		return errors.New("password is required")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	// is the email already in the database? it must be unique
	user, err := db.FindOneUserByEmail(ctx, u.Email)
	if err != nil {
		return err
	}

	if user != nil {
		return fmt.Errorf("user with email %v does already exist", u.Email)
	}

	// insert to database
	stmt, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("users").
		Columns("email", "password", "description", "first_name", "last_name").
		Values(u.Email, u.Password, u.Description, u.FirstName, u.LastName).Suffix(" RETURNING * ").ToSql()
	if err != nil {
		return err
	}
	db.logQuery(stmt, args...)

	err = db.GetContext(ctx, u, stmt, args...)
	if err != nil {
		return err
	}

	return nil
}
