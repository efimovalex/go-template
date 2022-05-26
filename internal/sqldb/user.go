package sqldb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iconimpact/go-core/errors"
	"github.com/iconimpact/go-core/structs"
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
func (db *Client) FindOneUserByEmail(email string) (*User, error) {
	u := User{}
	err := db.psql().Select("*").
		From("users").Where(sq.Eq{"email": email}).Limit(1).
		RunWith(db).QueryRow().
		Scan(&u.ID, &u.Password, &u.Email, &u.Description, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.E(err, errors.Internal)
	}
	db.Exec("SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));")
	return &u, nil
}

// Insert sanitizes and inserts a User in database
func (db *Client) InsertUser(u *User) error {
	// removes all leading and trailing white spaces from string fields
	err := u.Sanitize()
	if err != nil {
		return errors.E(err)
	}

	if u.Password == "" {
		msg := "password is required"
		return errors.E(fmt.Errorf(msg), errors.BadRequest, msg)
	}
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.E(err, errors.Internal)
	}
	u.Password = string(hashedPassword)

	// is the email already in the database? it must be unique
	user, err := db.FindOneUserByEmail(u.Email)
	if err != nil {
		return errors.E(err, errors.Internal, "error finding user")
	}
	if user != nil {
		msg := fmt.Sprintf("user with email %v does already exist", u.Email)
		return errors.E(fmt.Errorf(msg), errors.Conflict, msg)
	}

	// insert to database
	var createdUser User
	err = db.psql().Insert("users").
		Columns("email", "password", "description", "first_name", "last_name").
		Values(u.Email, u.Password, u.Description, u.FirstName, u.LastName).Suffix(" RETURNING * ").
		RunWith(db).QueryRow().
		Scan(&createdUser.ID, &createdUser.Password, &createdUser.Email, &createdUser.Description, &createdUser.FirstName, &createdUser.LastName, &createdUser.CreatedAt, &createdUser.UpdatedAt)
	if err != nil {
		return errors.E(err, errors.Internal, "error finding user")
	}

	*u = createdUser
	return nil
}

// Sanitize removes all leading and trailing white spaces from string fields
func (u *User) Sanitize() error {
	err := structs.Sanitize(u)
	if err != nil {
		return errors.E(err, errors.Internal)
	}

	return nil
}
