package authapi

import (
	"time" // for time.Time
	"go.mongodb.org/mongo-driver/bson/primitive" // for primitive.ObjectID
)


type signUpIn struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type signUpOut struct {
	OK  bool   `json:"ok"`
	ID  string `json:"id,omitempty"`
	Err string `json:"error,omitempty"`
}

type loginIn struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type loginOut struct {
	OK  bool   `json:"ok"`
	ID  string `json:"id,omitempty"`
	Err string `json:"error,omitempty"`
}

type UserRecord struct {
	ID             primitive.ObjectID `bson:"_id"`
	Email          string             `bson:"email"`
	Username       string             `bson:"username"`
	PasswordHash   []byte             `bson:"password_hash"`
	FailedAttempts int                `bson:"failedAttempts"`
	LockUntil      time.Time          `bson:"lockUntil"`
}
