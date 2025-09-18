// api/auth/auth.go
package authapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"             // secure password hashing
	"go.mongodb.org/mongo-driver/bson"       // BSON encoding/decoding
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"      // MongoDB driver
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handler holds references to collections needed by the auth API.
type Handler struct {
	Users *mongo.Collection
}

// ensure db indexes exist and hook up sign up route
func (h *Handler) Register(mux *http.ServeMux) error {
	if err := EnsureUserIndexes(h.Users); err != nil {
		return err
	}
	// if someone asks for /api/auth/create-user, send to SignUp
	mux.HandleFunc("/api/auth/create-user", h.SignUp)
	return nil
}

// signUpIn is the shape of the JSON request body for sign-up.
type signUpIn struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// signUpOut is the uniform JSON response shape for success/errors.
type signUpOut struct {
	OK  bool   `json:"ok"`              // success flag
	ID  string `json:"id,omitempty"`    // new user's ID (on success)
	Err string `json:"error,omitempty"` // human-readable error (on failure)
}

// EnsureUserIndexes creates unique indexes for email and username.
func EnsureUserIndexes(col *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetName("unique_email").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetName("unique_username").SetUnique(true),
		},
	}
	_, err := col.Indexes().CreateMany(ctx, models)
	return err
}

// usernameRE enforces 3-32 chars from [a-zA-Z0-9_]
var usernameRE = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// writeJSON standardizes JSON responses (success & errors).
// when talking to the frontend
func writeJSON(w http.ResponseWriter, status int, v any) {
	// Set content-type, status code, and encode response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// creates a new encoder and writes to the response writer
	_ = json.NewEncoder(w).Encode(v)
}

// isDup detects Mongo duplicate key (11000) in a portable way.
func isDup(err error) bool {
	// no error passed in
	if err == nil {
		return false
	}
	var we *mongo.WriteException // detailed error type
	// check if the error can be cast to a WriteException
	if errors.As(err, &we) { // checks if err ism or wraps a WriteException
		for _, e := range we.WriteErrors { // loop through all write errors
			 // duplicate key error code is 11000
			if e.Code == 11000 { 
				return true
			}
		}
	}
	return mongo.IsDuplicateKeyError(err)
}

// SignUp handles POST /api/auth/create-user
// Flow: parse -> validate -> hash -> insert -> return ID
func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) { // r is a pointer to an http.Request struct
	// Enforce POST-only
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, signUpOut{OK: false, Err: "method not allowed"})
		return
	}

	// Cap request body to prevent abuse (1 MiB is generous for this payload)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// JSON decode (reject unknown fields)
	var in signUpIn
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "invalid JSON"})
		return
	}

	// lowecase inputs and remove spaces
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))     
	in.Username = strings.TrimSpace(in.Username)
	in.FirstName = normalizeName(in.FirstName)
	in.LastName = normalizeName(in.LastName)

	// Validate email syntax (ignore this value, return the error)
	if _, err := mail.ParseAddress(in.Email); err != nil { 
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "invalid email"})
		return
	}
	// Validate username policy
	if !usernameRE.MatchString(in.Username) {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "username must be 3-32 chars [a-zA-Z0-9_]"})
		return
	}
	// Validate password length (expand policy later if desired)
	if len(in.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "password must be at least 8 chars"})
		return
	}

	// Hash password with bcrypt (DefaultCost ~10)
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, signUpOut{OK: false, Err: "hash error"})
		return
	}

	// Build the document to insert.
	// Note: we store password as "password_hash" (binary []byte); clients never see it.
	doc := bson.M{
		"email":         in.Email,
		"username":      in.Username,
		"password_hash": hash,
		"createdAt":     time.Now(),
		"updatedAt":     time.Now(),
	}

	// Short, cancellable DB context for the insert.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// res holds the result of the insert operation
	res, err := h.Users.InsertOne(ctx, doc)
	if err != nil {
		// Clean handling of duplicate email/username (unique index violation)
		if isDup(err) {
			writeJSON(w, http.StatusConflict, signUpOut{OK: false, Err: "email or username already exists"})
			return
		}
		// Generic DB error
		writeJSON(w, http.StatusInternalServerError, signUpOut{OK: false, Err: "database error"})
		return
	}

	// Convert InsertedID to hex if frontend needs it
	idHex := ""
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		idHex = oid.Hex()
	}

	// Success
	writeJSON(w, http.StatusOK, signUpOut{OK: true, ID: idHex})
}

// normalizeName capitalizes the first letter and lowercases the rest for each word.
func normalizeName(s string) string {
	parts := strings.Fields(s) // split by whitespace
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, " ")
}
