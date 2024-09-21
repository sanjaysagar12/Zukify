package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"encoding/json"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/url"
	"zukify.com/database"
	"zukify.com/handlers"
	"mime/multipart"
    "os"
    "path/filepath"
)

type Environment map[string]string

type Variables map[string]string

type TestCase struct {
	Case   string            `json:"case"`
	Data   interface{}       `json:"data"`
	SetEnv map[string]string `json:"set_env"`
	Imp    bool              `json:"imp"`
}

type EndpointData struct {
	Tag       string            `json:"tag"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Variables Variables         `json:"variables"`
	Headers   map[string]string `json:"headers"`
	Body      map[string]string `json:"body"`
	TestCases []TestCase        `json:"test_cases"`
	Files     map[string]string `json:"files"`
}

type RequestData struct {
	Env          Environment  `json:"env"`
	EndpointData EndpointData `json:"endpoint_data"`
}

type TestResult struct {
	Case   string `json:"case"`
	Passed bool   `json:"passed"`
}

type APIResponse struct {
	EndpointResponse string            `json:"endpoint_response"`
	TestResults      []TestResult      `json:"test_results"`
	AllPassed        bool              `json:"all_passed"`
	UpdatedEnv       map[string]string `json:"updated_env"`
}

func handleTest(c echo.Context) error {
	var reqData RequestData
	if err := c.Bind(&reqData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	response, err := performAPIRequest(reqData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	testResults, allPassed := runTestCases(reqData.EndpointData.TestCases, response, &reqData.Env)

	apiResponse := APIResponse{
		EndpointResponse: string(response),
		TestResults:      testResults,
		AllPassed:        allPassed,
		UpdatedEnv:       reqData.Env,
	}

	return c.JSON(http.StatusOK, apiResponse)
}

func performAPIRequest(reqData RequestData) ([]byte, error) {
	endpoint_url := substituteVariables(reqData.EndpointData.URL, reqData.EndpointData.Variables, reqData.Env)
	method := reqData.EndpointData.Method
	headers := substituteMapVariables(reqData.EndpointData.Headers, reqData.EndpointData.Variables, reqData.Env)
	body := substituteMapVariables(reqData.EndpointData.Body, reqData.EndpointData.Variables, reqData.Env)

	log.Printf("URL: %s", endpoint_url)
	log.Printf("Method: %s", method)
	log.Printf("Headers: %v", headers)
	log.Printf("Body before processing: %v", body)

	var req *http.Request
	var err error

	if method == "GET" || method == "DELETE" {
		req, err = http.NewRequest(method, endpoint_url, nil)
	} else {
		contentType := headers["Content-Type"]
		if contentType == "" {
			contentType = "application/json" // Default to JSON if not specified
		}

		var reqBody io.Reader

		switch contentType {
		case "application/json":
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("error marshaling request body to JSON: %v", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		case "application/x-www-form-urlencoded":
			formData := url.Values{}
			for key, value := range body {
				formData.Add(key, value)
			}
			reqBody = strings.NewReader(formData.Encode())
		case "multipart/form-data":
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add regular form fields
			for key, value := range reqData.EndpointData.Body {
				writer.WriteField(key, value)
			}

			// Add file fields
			for fieldName, filePath := range reqData.EndpointData.Files {
				file, err := os.Open(filePath)
				if err != nil {
					return nil, fmt.Errorf("error opening file %s: %v", filePath, err)
				}
				defer file.Close()

				part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
				if err != nil {
					return nil, fmt.Errorf("error creating form file: %v", err)
				}
				_, err = io.Copy(part, file)
				if err != nil {
					return nil, fmt.Errorf("error copying file contents: %v", err)
				}
			}

			err = writer.Close()
			if err != nil {
				return nil, fmt.Errorf("error closing multipart writer: %v", err)
			}

			req, err = http.NewRequest(method, endpoint_url, body)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

		default:
			// For other content types, just join the key-value pairs
			pairs := []string{}
			for key, value := range body {
				pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
			}
			reqBody = strings.NewReader(strings.Join(pairs, "&"))
		}

		req, err = http.NewRequest(method, endpoint_url, reqBody)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default Content-Type if not provided
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	log.Printf("Final request headers: %v", req.Header)
	log.Printf("Final request body: %v", req.Body)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %s", resp.Status)

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf("Response Body: %s", string(responseBody))

	return responseBody, nil
}

func runTestCases(testCases []TestCase, response []byte, env *Environment) ([]TestResult, bool) {
	var results []TestResult
	allPassed := true

	for _, tc := range testCases {
		passed := runTestCase(tc, response, env)
		results = append(results, TestResult{Case: tc.Case, Passed: passed})
		if tc.Imp && !passed {
			allPassed = false
		}
	}

	return results, allPassed
}

func runTestCase(tc TestCase, response []byte, env *Environment) bool {
	passed := false

	switch tc.Case {
	case "check_status_200":
		passed = strings.Contains(string(response), "200 OK")
	case "check_response_contains":
		passed = strings.Contains(string(response), tc.Data.(string))
	case "check_json_field_exists":
		var jsonResponse map[string]interface{}
		json.Unmarshal(response, &jsonResponse)
		_, passed = jsonResponse[tc.Data.(string)]
	}

	if tc.SetEnv != nil {
		var jsonResponse map[string]interface{}
		err := json.Unmarshal(response, &jsonResponse)
		if err != nil {
			log.Printf("Error unmarshaling JSON response: %v", err)
			return passed
		}

		log.Printf("JSON Response: %+v", jsonResponse)

		for key, value := range tc.SetEnv {
			log.Printf("Processing SetEnv: key=%s, value=%s", key, value)
			if strings.HasPrefix(value, "(response") && strings.HasSuffix(value, ")") {
				path := strings.Trim(value, "(response)")
				log.Printf("Extracted path: %s", path)
				nestedKeys := strings.Split(path, ".")
				log.Printf("Nested keys: %v", nestedKeys)
				nestedValue := getNestedValue(jsonResponse, nestedKeys)
				if nestedValue != nil {
					(*env)[key] = fmt.Sprintf("%v", nestedValue)
					log.Printf("Set env variable %s to %v", key, nestedValue)
				} else {
					log.Printf("Failed to set env variable %s, value not found in response", key)
				}
			} else {
				(*env)[key] = value
				log.Printf("Set env variable %s to %s", key, value)
			}
		}
	}

	return passed
}

func getNestedValue(data interface{}, keys []string) interface{} {

	log.Printf("getNestedValue called with keys: %v", keys)
	for _, key := range keys {
		log.Printf("Processing key: %s", key)
		switch v := data.(type) {
		case map[string]interface{}:
			data = v[key]
		case []interface{}:
			index := 0
			fmt.Sscanf(key, "%d", &index)
			if index >= 0 && index < len(v) {
				data = v[index]
			} else {
				log.Printf("Array index out of bounds: %d", index)
				return nil
			}
		default:
			log.Printf("Unexpected type at key %s: %T", key, v)
			return nil
		}
		if data == nil {
			log.Printf("Data is nil at key: %s", key)
			return nil
		}
		log.Printf("After key %s, data: %+v", key, data)
	}
	return data
}

func substituteVariables(input string, variables Variables, env Environment) string {
	for key, value := range variables {
		input = strings.ReplaceAll(input, "<<"+key+">>", value)
	}
	for key, value := range env {
		input = strings.ReplaceAll(input, "<<"+key+">>", value)
	}
	return input
}

func substituteMapVariables(input map[string]string, variables Variables, env Environment) map[string]string {
	result := make(map[string]string)
	for key, value := range input {
		result[key] = substituteVariables(value, variables, env)
	}
	return result
}

func main() {
	// Initialize database connections
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.UserDB.Close()
	defer database.WorkspaceDB.Close()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS middleware configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://zukify.portos.site"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// Public routes
	e.POST("/register", handlers.HandlerPostRegister)
	e.POST("/login", handlers.HandlerPostLogin)

	// Protected routes
	r := e.Group("/api")
	r.Use(handlers.JWTMiddleware)
	r.GET("/verify", (handlers.HandlerVerifyToken))
	r.POST("/workspace", handlers.HandlerCreateWorkspace) // New route for workspace creation
	r.GET("/getworkspace", handlers.HandlerGetWorkspaces)
	r.POST("/workspace/saveat", handlers.HandlerSaveAT)
	r.POST("/workspace/saveflow", handlers.HandlerSaveFlow)
	r.GET("/workspace/fetchpathat", handlers.HandlerFetchPathAT)
	r.GET("/workspace/fetchallat", handlers.HandlerFetchAllAT)
	r.GET("/workspace/fetchpathflow", handlers.HandlerFetchPathFlow)
	r.GET("/workspace/fetchallflow", handlers.HandlerFetchAllFlow)
	r.POST("/collaborator", handlers.HandlerAddCollaborator)
	e.POST("/runAT", handleTest)
	// e.POST("runFlow",handleTest)
	// Start server
	e.Logger.Fatal(e.Start(":80"))
}
