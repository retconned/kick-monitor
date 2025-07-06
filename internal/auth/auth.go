package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http" // Required for HTTP status codes

	"os"   // For reading environment variables
	"time" // For JWT expiration

	"github.com/retconned/kick-monitor/internal/db"     // Your database connection
	"github.com/retconned/kick-monitor/internal/models" // Your user model

	"github.com/golang-jwt/jwt/v5"            // The JWT library
	"github.com/google/uuid"                  // For UUID generation
	echojwt "github.com/labstack/echo-jwt/v4" // <--- CRITICAL CHANGE: NEW IMPORT AND ALIAS
	"github.com/labstack/echo/v4"             // Echo framework
	// "github.com/labstack/echo/v4/middleware" // <--- CRITICAL CHANGE: ADD 'middleware' ALIAS HERE
	"golang.org/x/crypto/bcrypt" // For password hashing
	"gorm.io/gorm"               // GORM for database operations
)

// JwtCustomClaims defines the custom claims for your JWT token.
type JwtCustomClaims struct {
	ID                   string `json:"id"`
	Email                string `json:"email"`
	jwt.RegisteredClaims        // Embed standard JWT claims
}

var jwtSecret []byte // Stores the JWT secret key as a byte slice

// InitAuth initializes the authentication system by loading the JWT secret.
func InitAuth() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable not set. Please configure it.")
	}
	jwtSecret = []byte(secret) // Convert string secret to byte slice
}

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate bcrypt hash: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a hashed password with a plain-text password.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // Returns true if passwords match, false otherwise
}

// GenerateToken generates a JWT token for a given user.
func GenerateToken(user *models.User) (string, error) {
	claims := &JwtCustomClaims{
		ID:    user.ID.String(),
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)), // Token valid for 72 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()), // Token is valid immediately
		},
	}

	// Create the token with the signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}
	return signedToken, nil
}

// --- API Handlers for Authentication ---

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterHandler handles user registration.
func RegisterHandler(c echo.Context) error {
	req := new(RegisterRequest)
	if err := c.Bind(req); err != nil { // Bind request body to struct
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	// Basic input validation
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Email and password are required"})
	}

	// Hash the user's password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to hash password"})
	}

	// Create a new user model
	user := models.User{
		ID:           uuid.New(), // Generate a new UUID for the user ID
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	// Save the user to the database
	if err := db.DB.Create(&user).Error; err != nil {
		// Check for unique constraint violation (email must be unique)
		if errors.Is(err, gorm.ErrDuplicatedKey) { // This correctly checks for unique constraint violation
			return c.JSON(http.StatusConflict, map[string]string{"message": "User with this email already exists"})
		}
		log.Printf("Database error during user registration: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to register user"})
	}

	// Return success response
	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully", "id": user.ID.String()})
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginHandler handles user login and JWT issuance.
func LoginHandler(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Email and password are required"})
	}

	// Find the user by email
	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"}) // User not found
		}
		log.Printf("Database error during user login: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
	}

	// Check if the provided password matches the stored hash
	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"}) // Password mismatch
	}

	// Generate a JWT token
	token, err := GenerateToken(&user)
	if err != nil {
		log.Printf("Error generating token for user %s: %v", user.Email, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to generate token"})
	}

	// Return success response with the token
	return c.JSON(http.StatusOK, map[string]string{"message": "Login successful", "token": token})
}

// AuthMiddleware provides JWT authentication middleware for Echo.
func AuthMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  jwtSecret,
		TokenLookup: "header:Authorization:Bearer ",
		ErrorHandler: func(c echo.Context, err error) error {
			log.Printf("JWT authentication error: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token. Please log in again.")
		},
		ContextKey: "user",
		Skipper:    nil,
	})
}
