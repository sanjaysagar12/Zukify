package handlers

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"zukify.com/services"

	"github.com/golang-jwt/jwt"

	"zukify.com/database"
	"log"
	"zukify.com/types"
	"fmt"
	"encoding/json"
	"strings"
)

func HandlePostAT(c echo.Context) error {
	var req types.ComplexATRequest
	if err := c.Bind(&req); err != nil {
		fmt.Println("400 Error")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	// fmt.Println("ComplexATRequest:",req)
	results, newEnv, endpointResponse := services.TestEndpoint(req)
	response := types.ATResponse{
		Results:          results.Results,
		AllImpPassed:     results.AllImpPassed,
		NewEnv:           newEnv,
		EndpointResponse: endpointResponse,
	}
	fmt.Println("New Env: ",newEnv)
	// b, err := json.MarshalIndent(response, "", "  ")
    // if err != nil {
    //     fmt.Println(err)
    // }
    // fmt.Print(string(b))
	return c.JSON(http.StatusOK, response)
}

func HandlePostATFromSaved(c echo.Context) error {
	wid := c.QueryParam("wid")
	id := c.QueryParam("id")
	fmt.Println(wid,id)
	response := map[string]interface{}{
		"results":        nil,
		"all_imp_passed": true,
		"new_env":        map[string]interface{}{},
		"endpoint_response": map[string]interface{}{
			"status_code": 404,
			"headers": map[string]interface{}{
				"Access-Control-Allow-Origin": []string{"*"},
				"Cf-Cache-Status":             []string{"DYNAMIC"},
				"Cf-Ray":                      []string{"8c8b94535b6840a7-SIN"},
				"Date":                        []string{"Wed, 25 Sep 2024 14:07:14 GMT"},
				"Nel":                         []string{`{"success_fraction":0,"report_to":"cf-nel","max_age":604800}`},
				"Report-To":                   []string{`{"endpoints":[{"url":"https://a.nel.cloudflare.com/report/v4?s=P7SC7EzP9NEavgsTu%2BS7c7wvZHRTDNeEPU%2B6T1MFd%2BSRBixXnQpIv9T6BSpikKEz0U7v%2Fls4zuN083ME5S0L1n4cSDMj61QtSiK1lUcL0Xi1IHqd9k0KxKFr24Fl0BQ097fddas%3D"}],"group":"cf-nel","max_age":604800}`},
				"Server":                      []string{"cloudflare"},
			},
			"body": "",
		},
	}

	return c.JSON(http.StatusOK, response)
}


func HandlerRunSavedAT(c echo.Context) error {
    // Get user from context
    user := c.Get("user").(jwt.MapClaims)
    uid, ok := user["uid"].(float64)
    if !ok {
        log.Printf("Failed to extract UID from token: %v", user)
        return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
    }

    // Get query parameters
    wid := c.QueryParam("wid")
    if wid == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID (wid) is required")
    }

    id := c.QueryParam("id")
    if id == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
    }

    // Check workspace access
    hasAccess, err := database.UserHasAccessToWorkspace(int(uid), wid)
    if err != nil {
        log.Printf("Failed to check workspace access: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify workspace access")
    }
    if !hasAccess {
        return echo.NewHTTPError(http.StatusForbidden, "You don't have access to this workspace")
    }

    // Fetch AT data
    atData, err := database.FetchAllAT(wid, id)
    if err != nil {
        log.Printf("Failed to fetch AT data: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch AT data")
    }

    if atData == nil {
        return echo.NewHTTPError(http.StatusNotFound, "AT data not found")
    }

    // Convert atData to ComplexATRequest
    req, err := convertATDataToRequest(atData)
    if err != nil {
        log.Printf("Failed to convert AT data: %v", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process AT data")
    }

    // Run the test endpoint
    results, newEnv, endpointResponse := services.TestEndpoint(req)

    // Prepare response
    response := types.ATResponse{
        Results:          results.Results,
        AllImpPassed:     results.AllImpPassed,
        NewEnv:           newEnv,
        EndpointResponse: endpointResponse,
    }

    return c.JSON(http.StatusOK, response)
}

// Header structure for JSON parsing
type HeaderItem struct {
    IsInUse bool   `json:"is_inuse"`
    Key     string `json:"key"`
    Value   string `json:"value"`
    Desc    string `json:"desc"`
}

func convertATDataToRequest(atData *database.AllATData) (types.ComplexATRequest, error) {
    var req types.ComplexATRequest

    // Initialize the required maps
    req.EndpointData = types.ATRequest{
        Headers:    make(map[string]string),
        Body:      make(map[string]interface{}),
        Variables: make(map[string]string),
    }
    req.Env = make(map[string]string)

    // Set Method and URL
    req.EndpointData.Method = atData.Method
    req.EndpointData.URL = atData.URL

    // Parse Headers
    var headers []HeaderItem
    if err := json.Unmarshal([]byte(atData.Header), &headers); err != nil {
        return req, fmt.Errorf("failed to parse headers: %v", err)
    }

    // Add headers that are in use
    for _, header := range headers {
        if header.IsInUse {
            req.EndpointData.Headers[header.Key] = header.Value
        }
    }

    // Parse Body
    var body map[string]interface{}
    if err := json.Unmarshal([]byte(atData.Body), &body); err != nil {
        return req, fmt.Errorf("failed to parse body: %v", err)
    }
    req.EndpointData.Body = body

    // Parse TestCases
    // Remove escaped quotes first if present
    testcasesStr := strings.ReplaceAll(atData.Testcases, "\\\"", "\"")
    // Remove surrounding quotes if present
    testcasesStr = strings.Trim(testcasesStr, "\"")
    
    var testCases []types.TestCase
    if err := json.Unmarshal([]byte(testcasesStr), &testCases); err != nil {
        return req, fmt.Errorf("failed to parse test cases: %v", err)
    }
    req.EndpointData.TestCases = testCases

    // Add any default environment variables if needed
    req.Env["workspace_id"] = atData.Path // You might want to modify this based on your needs

    return req, nil
}