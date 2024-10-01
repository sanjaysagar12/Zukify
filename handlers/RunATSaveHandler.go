package handlers

// import (
// 	"net/http"
// 	// "strconv"

// 	"github.com/labstack/echo/v4"
// 	"github.com/golang-jwt/jwt"
// 	"zukify.com/database"
// 	"zukify.com/services"
// 	"zukify.com/types"
// )

// func HandlerRunATSave(c echo.Context) error {
// 	// Extract UID from JWT
// 	user := c.Get("user")
// 	claims, ok := user.(jwt.MapClaims)
// 	if !ok {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
// 	}
	
// 	uidFloat, ok := claims["uid"].(float64)
// 	if !ok {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid UID in token"})
// 	}
// 	uid := int(uidFloat)

// 	// Get WID from request
// 	wid := c.QueryParam("wid")
// 	if wid == "" {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing WID parameter"})
// 	}

// 	// Get ID from request
// 	id := c.QueryParam("id")
// 	if id == "" {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing ID parameter"})
// 	}

// 	// Check if UID has access to WID
// 	hasAccess, err := database.UserHasAccessToWorkspace(uid, wid)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database query error"})
// 	}
// 	if !hasAccess {
// 		return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized access to WID"})
// 	}

// 	// Fetch AT data
// 	atData, err := database.FetchAllAT(wid, id)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch AT data"})
// 	}
// 	if atData == nil {
// 		return c.JSON(http.StatusNotFound, map[string]string{"error": "No AT data found for given ID"})
// 	}

// 	// Convert database.AllATData to types.ComplexATRequest
// 	complexATRequest, err := convertToComplexATRequest(atData)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process AT data"})
// 	}

// 	// Run the test endpoint
// 	testResponse, newEnv, endpointResponse := services.TestEndpoint(*complexATRequest)

// 	// Prepare the response
// 	response := types.ATResponse{
// 		Results:          testResponse.Results,
// 		AllImpPassed:     testResponse.AllImpPassed,
// 		NewEnv:           newEnv,
// 		EndpointResponse: endpointResponse,
// 	}

// 	// Save the test results back to the database
// 	err = database.SaveATResponse(wid, id, response)
// 	if err != nil {
// 		c.Logger().Error("Failed to save test results:", err)
// 	}

// 	return c.JSON(http.StatusOK, response)
// }