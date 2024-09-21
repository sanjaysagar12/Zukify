package handlers

import (
	"net/http"
	"os"
	"time"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"zukify.com/database"

)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func HandlerPostRegister(c echo.Context) error {
	user := new(database.User)
	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate required fields
	if user.Username == "" || user.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Username and password are required")
	}

	// Check if user already exists
	exists, err := database.UserExists(user.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check user existence")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}

	uid, err := database.CreateUser(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	// Set the UID and remove password from response
	user.UID = uid
	user.Password = ""

	return c.JSON(http.StatusCreated, user)
}

func HandlerPostLogin(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	user, err := database.GetUserByUsername(username)
	fmt.Println("Username is",username,"Password is",password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["uid"] = user.UID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return err
	}

	// Remove password from response
	user.Password = ""

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Logged in successfully",
		"user":    user,
		"token":   t,
	})
}

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
		}

		// Remove 'Bearer ' prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}
			return jwtSecret, nil
		})

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user", claims)
			return next(c)
		}

		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}
}

// New handler function for token verification
func HandlerVerifyToken(c echo.Context) error {
    fmt.Println("HandlerVerifyToken called")
    
    user := c.Get("user")
    if user == nil {
        fmt.Println("No user found in context")
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No user found in context"})
    }

    claims, ok := user.(jwt.MapClaims)
    if !ok {
        fmt.Println("User is not jwt.MapClaims")
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User is not jwt.MapClaims"})
    }

    uid, ok := claims["uid"]
    if !ok {
        fmt.Println("UID not found in claims")
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "UID not found in claims"})
    }

    fmt.Printf("UID found: %v\n", uid)

    return c.JSON(http.StatusOK, map[string]interface{}{
        "uid": uid,
    })
}