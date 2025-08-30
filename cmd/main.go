package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"rothira/api/health"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CalculationRequest struct {
	Income float64 `json:"income"`
}

type CalculationResponse struct {
	Outcome float64 `json:"outcome"`
	Message string  `json:"message"`
}

type App struct { // [NEW] holds dependencies for handlers
	UsersColl *mongo.Collection
}

type User struct { // [NEW]
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func main() {
	_ = godotenv.Load()

	fmt.Print("Starting up the Golang Roth IRA Backend...\n")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// ---------- MongoDB setup (NEW) ----------
	mongoURI := getenv("MONGODB_URI", "mongodb://localhost:27017") // [NEW]
	dbName := getenv("MONGODB_DB", "rothira")                      // [NEW]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // [NEW]
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI)) // [NEW]
	if err != nil {
		log.Fatalf("mongo connect failed: %v", err) // [NEW]
	}
	if err := client.Ping(ctx, nil); err != nil { // [NEW]
		log.Fatalf("mongo ping failed: %v", err)
	}

	db := client.Database(dbName) // [NEW]
	app := &App{
		UsersColl: db.Collection("users"), // [NEW]
	}
	fmt.Println("Connected to MongoDB") // [NEW]

	e.GET("/hello", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})

	mux.HandleFunc("/health", health.HealthHandler)

	// [NEW] DB health endpoint (pings Mongo)
	e.GET("/db-health", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := client.Ping(ctx, nil); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "down",
				"error":  err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "up"})
	})

	e.GET("/random-number", func(c echo.Context) error {
		randomValue := rand.Intn(100)
		return c.String(http.StatusOK, fmt.Sprintf("Your random value is: %d", randomValue))
	})

	e.POST("/calculate-interest", func(c echo.Context) error {
		type InterestRequest struct {
			Income   float64 `json:"income"`
			Interest float64 `json:"interest"`
		}
		type InterestResponse struct {
			Total float64 `json:"total"`
		}

		req := new(InterestRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
		}
		return c.JSON(http.StatusOK, InterestResponse{
			Total: req.Income * (1.00 + req.Interest),
		})
	})

	// [NEW] Example write route: insert a user
	e.POST("/users", app.CreateUser)

	// ---- Serve the React SPA at root (/) ----
	// 1) Serve static assets from the built folder
	e.Static("/", "frontend-build")

	// 2) SPA fallback ONLY when a GET request 404s and the client accepts HTML
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok &&
			he.Code == http.StatusNotFound &&
			c.Request().Method == http.MethodGet &&
			strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
			_ = c.File("frontend-build/index.html")
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
	// -----------------------------------------

	// Port
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = ":8080"
	}
	e.Logger.Fatal(e.Start(":" + httpPort))
}

func (a *App) CreateUser(c echo.Context) error { // [NEW]
	var user User
	if err := c.Bind(&user); err != nil || user.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload (need email, optional name)"})
	}
	user.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	res, err := a.UsersColl.InsertOne(ctx, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "insert failed"})
	}
	return c.JSON(http.StatusCreated, map[string]any{"insertedId": res.InsertedID})
}

// ---------- Helpers (NEW) ----------

func getenv(k, def string) string { // [NEW] env with default
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
