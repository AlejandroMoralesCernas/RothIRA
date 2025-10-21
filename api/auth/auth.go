package authapi

import (
	"log" // logging
	"context" // manages cancelation signals
	"encoding/json" // encoding and decoding JSON
	"errors" // error handling
	"fmt" // formatted I/O
	"net/http" // HTTP client and server implementations
	"net/mail" // email address parsing
	"regexp" // regular expressions
	"strings" // string manipulation functions
	"time" // time-related functions
	// "strconv" string conversions for lockout countdown
	"os" // read JWT secret from .env
	"golang.org/x/crypto/bcrypt"             // secure password hashing
	"github.com/golang-jwt/jwt/v5"           // JWT generation and validation
	"go.mongodb.org/mongo-driver/bson"       // BSON encoding/decoding
	"go.mongodb.org/mongo-driver/bson/primitive" // mongo ObjectID type
	"go.mongodb.org/mongo-driver/mongo"      // MongoDB driver
	"go.mongodb.org/mongo-driver/mongo/options" // configure MongoDB queries, indexes, etc
)


// remembers where the users collection is in the database, so our functions can use it
type Handler struct {
	Users *mongo.Collection
}

// jwt secret should be set in environment variable JWT_SECRET
// calls EnsureUserIndexes to make sure the database prevents duplicates
// then creates the routes for signing up and logging in
func (h *Handler) Register(mux *http.ServeMux) error {
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("Warning: JWT_SECRET not set! Tokens will fail to generate.")
	}

	// ensure email and username indexes are unique, return error if index creation fails
	if err := EnsureUserIndexes(h.Users); err != nil {
		return err
	}

	// connects frontend URLs to backend Go functions that handle them
	mux.HandleFunc("/api/auth/create-user", h.SignUp)
	mux.HandleFunc("/api/auth/login-user", h.Login)
	mux.HandleFunc("/api/auth/verify-session", h.VerifySession)

	return nil
}

// VerifySession checks if the provided JWT token is valid and not expired.
func (h *Handler) VerifySession(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusUnauthorized, loginOut{OK: false, Err: "missing token"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := verifyJWT(token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, loginOut{OK: false, Err: "invalid or expired token"})
		return
	}

	writeJSON(w, http.StatusOK, bson.M{"ok": true, "user_id": userID})
}



// createJWT generates a signed JWT token valid for 2 minutes.
// takes userID string input, usually mongo user's ID, returns two strings: the token and an error
func createJWT(userID string) (string, error) {
	// we look for the JWT_SECRET environment variable to sign the token
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET not set")
	}

	// claims store token data like user_id and expiration time using MapClaims
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(2 * time.Minute).Unix(), // expires in 2 minutes
	}

	// create a new token object specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// we sign the token with our secret key, if successful we return the token string, otherwise an error
	return token.SignedString([]byte(secret))
}

// verifyJWT validates the token signature and expiration.
// function takes in the token string, returns userID string and error
func verifyJWT(tokenString string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET not set")
	}

	// Parse the token string to verify it's valid
	// jwt.Parse does three main things:
	//   1. Splits the token into its 3 parts: header, payload, and signature.
	//   2. Checks that the token was signed using the expected algorithm (HS256).
	//   3. Verifies the token’s signature using your secret key.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token was signed with the correct algorithm (HMAC SHA-256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		// Give the JWT library our secret key so it can verify the signature.
		// This secret key is like the password used to prove the token was made by us.
		return []byte(secret), nil
	})

	// checks if there was an error parsing or if the token is invalid/expired
	if err != nil || !token.Valid {
		return "", errors.New("invalid or expired token")
	}

	// taking the claims from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok {

		// look for the user_id field in the claims and return it
		if userID, ok := claims["user_id"].(string); ok {
			return userID, nil
		}
	}
	// if we reach here, something was wrong with the claims, so we return an error
	return "", errors.New("invalid claims")
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

// signUp handler
func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, signUpOut{OK: false, Err: "method not allowed"})
		return
	}

	in, err := parseSignUpRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: err.Error()})
		return
	}

	if err := validateSignUpInput(in); err != nil {
		writeJSON(w, http.StatusBadRequest, signUpOut{OK: false, Err: err.Error()})
		return
	}

	if err := h.insertUser(r.Context(), in); err != nil {
		writeJSON(w, http.StatusConflict, signUpOut{OK: false, Err: err.Error()})
		return
	}

	log.Printf("New user signed up: %s (%s)", in.Username, in.Email)
	writeJSON(w, http.StatusOK, signUpOut{OK: true})
}

// signUp helpers

// Reads and decodes the signup JSON request body.
func parseSignUpRequest(r *http.Request) (signUpIn, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	var in signUpIn
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return signUpIn{}, errors.New("invalid JSON")
	}

	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Username = strings.TrimSpace(in.Username)
	in.FirstName = normalizeName(in.FirstName)
	in.LastName = normalizeName(in.LastName)
	return in, nil
}

// Validates signup fields: email, username, and password.
func validateSignUpInput(in signUpIn) error {
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return errors.New("invalid email")
	}
	if !usernameRE.MatchString(in.Username) {
		return errors.New("username must be 3-32 chars [a-zA-Z0-9_]")
	}
	if len(in.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

// Inserts new user into MongoDB.
func (h *Handler) insertUser(ctx context.Context, in signUpIn) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	doc := bson.M{
		"email":          in.Email,
		"username":       in.Username,
		"password_hash":  hash,
		"failedAttempts": 0,
		"lockUntil":      time.Time{},
		"createdAt":      time.Now(),
		"updatedAt":      time.Now(),
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := h.Users.InsertOne(ctx, doc); err != nil {
		if isDup(err) {
			return errors.New("email or username already exists")
		}
		return errors.New("database error")
	}
	return nil
}

// login handler
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, loginOut{OK: false, Err: "method not allowed"})
		return
	}

	in, err := parseLoginRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, loginOut{OK: false, Err: err.Error()})
		return
	}

	user, err := h.findUserByIdentifier(r.Context(), in.Identifier)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, loginOut{OK: false, Err: "invalid credentials"})
		return
	}

	if err := h.handleLockout(user); err != nil {
		writeJSON(w, http.StatusTooManyRequests, loginOut{OK: false, Err: err.Error()})
		return
	}

	if err := h.verifyUserPassword(user, in.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, loginOut{OK: false, Err: err.Error()})
		return
	}

	tokenStr, err := createJWT(user.ID.Hex())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, loginOut{OK: false, Err: "failed to generate session token"})
		return
	}

	writeJSON(w, http.StatusOK, bson.M{
		"ok":        true,
		"id":        user.ID.Hex(),
		"token":     tokenStr,
		"expiresIn": 120,
	})
}

// login helpers

// Reads and decodes the login JSON request.
func parseLoginRequest(r *http.Request) (loginIn, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	var in loginIn
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return loginIn{}, errors.New("invalid JSON")
	}
	in.Identifier = strings.TrimSpace(in.Identifier)
	in.Password = strings.TrimSpace(in.Password)
	if in.Identifier == "" || in.Password == "" {
		return loginIn{}, errors.New("missing username or password")
	}
	return in, nil
}

// Finds a user by email or username.
func (h *Handler) findUserByIdentifier(ctx context.Context, identifier string) (*UserRecord, error) {
	emailCandidate := strings.ToLower(identifier)
	filter := bson.D{
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "email", Value: emailCandidate}},
			bson.D{{Key: "username", Value: identifier}},
		}},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user UserRecord
	if err := h.Users.FindOne(ctx, filter).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Checks whether the account is currently locked.
func (h *Handler) handleLockout(user *UserRecord) error {
	now := time.Now()
	if !user.LockUntil.IsZero() && user.LockUntil.After(now) {
		remaining := int(user.LockUntil.Sub(now).Seconds())
		return fmt.Errorf("account locked. Try again in %d seconds", remaining)
	}
	return nil
}

// Verifies password and updates lockout/attempts.
func (h *Handler) verifyUserPassword(user *UserRecord, password string) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		user.FailedAttempts++
		update := bson.M{"$set": bson.M{"updatedAt": now}}

		if user.FailedAttempts >= 3 {
			lockDuration := 1 * time.Minute
			update["$set"].(bson.M)["failedAttempts"] = 0
			update["$set"].(bson.M)["lockUntil"] = now.Add(lockDuration)
			h.Users.UpdateByID(ctx, user.ID, update)
			return errors.New("too many failed attempts. Account locked for 60 seconds.")
		}

		update["$set"].(bson.M)["failedAttempts"] = user.FailedAttempts
		h.Users.UpdateByID(ctx, user.ID, update)
		return fmt.Errorf("invalid credentials (%d/3)", user.FailedAttempts)
	}

	// successful login — reset failed attempts and lock
	h.Users.UpdateByID(ctx, user.ID, bson.M{
		"$set": bson.M{
			"failedAttempts": 0,
			"lockUntil":      time.Time{},
			"updatedAt":      now,
		},
	})
	return nil
}
