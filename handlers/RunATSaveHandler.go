package handlers

import (
    "net/http"
    "log"
    "encoding/json"
    "github.com/labstack/echo/v4"
    "github.com/golang-jwt/jwt"
)

func HandlerExtractJWT(c echo.Context) error {
    // Get user from context with type assertion safety check
    user := c.Get("user")
    if user == nil {
        log.Printf("No JWT token found in context")
        return echo.NewHTTPError(http.StatusUnauthorized, "Missing authentication token")
    }

    // Type assert with safety check for jwt.MapClaims
    claims, ok := user.(jwt.MapClaims)
    if !ok {
        log.Printf("Failed to convert user to jwt.MapClaims. Type: %T, Value: %v", user, user)
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token format")
    }

    // Pretty print the claims (if they exist)
    prettyJSON, err := json.MarshalIndent(claims, "", "    ")
    if err != nil {
        log.Printf("Failed to marshal claims: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process token")
    }
    
    log.Printf("JWT Claims:\n%s", string(prettyJSON))

    // Extract UID with type safety
    uidRaw, exists := claims["uid"]
    if !exists {
        log.Printf("No UID found in token claims")
        return echo.NewHTTPError(http.StatusBadRequest, "Token missing required claim: uid")
    }

    // Handle different numeric types that might come from JWT
    var uid float64
    switch v := uidRaw.(type) {
    case float64:
        uid = v
    case float32:
        uid = float64(v)
    case int:
        uid = float64(v)
    case int64:
        uid = float64(v)
    default:
        log.Printf("UID is not a number. Type: %T, Value: %v", uidRaw, uidRaw)
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid UID format in token")
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "message": "JWT extracted successfully",
        "claims":  claims,
        "uid":     uid,
    })
}