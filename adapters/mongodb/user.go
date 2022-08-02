package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const userCollection = "users"

// User contains the database entry
type User struct {
	ID          primitive.ObjectID `bson:"_id"`
	Email       string             `bson:"email"`
	Password    string             `json:"-" bson:"password"`
	Description string             `bson:"description"`
	FirstName   string             `bson:"first_name"`
	LastName    string             `bson:"last_name"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

// UserByEmail loads User with given email, returns nil if not found
func (db *Client) FindOneUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}

	filter := bson.M{"email": email}
	err := db.Users.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return &u, nil
}

// Insert sanitizes and inserts a User in database
func (db *Client) InsertUser(ctx context.Context, u *User) error {
	if u == nil {
		return errors.New("user is nil")
	}

	if u.Password == "" {
		return errors.New("password is required")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	u.ID = primitive.NewObjectID()

	// is the email already in the database? it must be unique
	user, err := db.FindOneUserByEmail(ctx, u.Email)
	if err != nil {
		return err
	}

	if user != nil {
		return fmt.Errorf("user with email %v does already exist", u.Email)
	}

	ins, err := db.Users.InsertOne(ctx, *u)
	if err != nil {
		return err
	}

	u.ID = ins.InsertedID.(primitive.ObjectID)

	return nil
}
