package authapi

import (
	"log" // logging
	"context" // manages cancelation signals
	"encoding/json" // encoding and decoding JSON
	"errors" // error handling
	"net/http" // HTTP client and server implementations
	"net/mail" // email address parsing
	"regexp" // regular expressions
	"strings" // string manipulation functions
	"time" // time-related functions
	"golang.org/x/crypto/bcrypt"             // secure password hashing
	"go.mongodb.org/mongo-driver/bson"       // BSON encoding/decoding
	"go.mongodb.org/mongo-driver/bson/primitive" // mongo ObjectID type
	"go.mongodb.org/mongo-driver/mongo"      // MongoDB driver
	"go.mongodb.org/mongo-driver/mongo/options" // configure MongoDB queries, indexes, etc
)

// remembers where the users collection is in the database, so our functions can use it
type Handler struct {
	Users *mongo.Collection
}

// first calls EnsureUserIndexes to make sure the database prevents duplicates
// then creates the routes for signing up and logging in
func (h *Handler) Register(mux *http.ServeMux) error {
	if err := EnsureUserIndexes(h.Users); err != nil {
		return err
	}
	mux.HandleFunc("/api/auth/create-user", h.SignUp)
	mux.HandleFunc("/api/auth/login-user", h.Login)
	return nil
}

// what the user sends us (email, username, password, first/last name)
type signUpIn struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// what we send back 
type signUpOut struct {
	OK  bool   `json:"ok"`              // success flag
	ID  string `json:"id,omitempty"`    // new user's ID (on success)
	Err string `json:"error,omitempty"` // error (on failure)
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
	if err == nil {
		return nil
	}

	var ce mongo.CommandError
	if errors.As(err, &ce) {
		if ce.Code == 85 || ce.Code == 86 {
			return nil
		}
	}
	return err
}

// usernameRE enforces a simple username policy: 3-32 chars, alphanumeric + underscore
var usernameRE = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// writeJSON standardizes JSON responses (success & errors).
func writeJSON(w http.ResponseWriter, status int, v any) { // int is status code, v is any value to encode as JSON
	w.Header().Set("Content-Type", "application/json")	    // write header to say we are sending JSON so that client knows how to interpret it
	w.WriteHeader(status)									// sets the HTTP status code
	_ = json.NewEncoder(w).Encode(v)					   // encode v as JSON and write to response, _ for error we ignore
}

// detects Mongo duplicate key (11000) in a portable way.
func isDup(err error) bool {
	// no error passed in
	if err == nil {
		return false
	}
	var we *mongo.WriteException // detailed error type
	// check if the error can be cast to a WriteException
	if errors.As(err, &we) { 
		for _, e := range we.WriteErrors { 
			if e.Code == 11000 { 
				return true
			}
		}
	}
	return mongo.IsDuplicateKeyError(err)
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

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) { // method on handler, w is http response writer, r is incoming http request
	// Enforce POST-only
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, signUpOut{OK: false, Err: "method not allowed"})
		return
	}

	// cap request body to prevent abuse (1 MiB is generous for this payload)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var in signUpIn // input intostruct
	dec := json.NewDecoder(r.Body) // make a decoder that reads JSON from r.Body (the request stream)
	dec.DisallowUnknownFields() // reject any fields in the JSON that do not map to fields in the struct
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "invalid JSON"})
		return
	}

	// lowecase inputs and remove spaces
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))     
	in.Username = strings.TrimSpace(in.Username)
	in.FirstName = normalizeName(in.FirstName)
	in.LastName = normalizeName(in.LastName)

	// validate required fields
	if _, err := mail.ParseAddress(in.Email); err != nil { 
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "invalid email"})
		return
	}
	// username policy
	if !usernameRE.MatchString(in.Username) {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "username must be 3-32 chars [a-zA-Z0-9_]"})
		return
	}
	// password length
	if len(in.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: "password must be at least 8 chars"})
		return
	}

	// hash password to store securely
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, signUpOut{OK: false, Err: "hash error"})
		return
	}

	// build the document to insert.
	doc := bson.M{
		"email":         in.Email,
		"username":      in.Username,
		"password_hash": hash,
		"createdAt":     time.Now(),
		"updatedAt":     time.Now(),
	}

	// cancellable DB context for the insert.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// res holds the result of the insert operation
	if _, err := h.Users.InsertOne(ctx, doc); err != nil {
		if isDup(err) {
			writeJSON(w, http.StatusConflict, signUpOut{OK: false, Err: "email or username already exists"})
			log.Printf("Duplicate sign-up attempt")
			return
		}
		writeJSON(w, http.StatusInternalServerError, signUpOut{OK: false, Err: "database error"})
		return
	}

	// success
	log.Printf("New user signed up: %s (%s)", in.Username, in.Email)
	writeJSON(w, http.StatusOK, signUpOut{OK: true})
}


// JSON request body for login.
type loginIn struct {
  Identifier string `json:"identifier"`
  Password   string `json:"password"`
}

// JSON response shape for success/errors.
type loginOut struct {
	OK  bool   `json:"ok"`              // success flag
	ID  string `json:"id,omitempty"`    // user's ID (on success)
	Err string `json:"error,omitempty"` // human-readable error (on failure)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, loginOut{OK: false, Err: "method not allowed"})
		return
	}

	// cap request body to prevent abuse (1 MiB is generous for this payload)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// JSON decode (reject unknown fields)
	var in loginIn
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, loginOut{OK: false, Err: "invalid JSON"})
		return
	}

	in.Identifier = strings.TrimSpace(in.Identifier)
	pw := strings.TrimSpace(in.Password)
	if in.Identifier == "" || pw == "" {
		writeJSON(w, http.StatusBadRequest, loginOut{OK: false, Err: "missing username or password"})
		return
	}

	// find user by email or username (case-insensitive for email)
	emailCandidate := strings.ToLower(in.Identifier)
	filter := bson.D{
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "email", Value: emailCandidate}},
			bson.D{{Key: "username", Value: in.Identifier}},
		}},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// fetch only the fields we need
	var user struct {
		ID           primitive.ObjectID `bson:"_id"`
		Email        string             `bson:"email"`
		Username     string             `bson:"username"`
		PasswordHash []byte             `bson:"password_hash"` 
	}
	if err := h.Users.FindOne(ctx, filter).Decode(&user); err != nil {
		writeJSON(w, http.StatusUnauthorized, loginOut{OK: false, Err: "invalid credentials"})
		return
	}

	// compare bcrypt
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(pw)); err != nil {
		writeJSON(w, http.StatusUnauthorized, loginOut{OK: false, Err: "invalid credentials"})
		return
	}
	// success
	writeJSON(w, http.StatusOK, loginOut{OK: true, ID: user.ID.Hex()})
}