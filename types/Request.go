package types
import (
	"net/http"
)



type ComplexATRequest struct {
	EndpointData ATRequest 
	Env          map[string]string
}

type ATRequest struct {
	Method     string
	URL        string
	Headers    map[string]string
	Body       map[string]interface{}
	Variables  map[string]string
	TestCases  []TestCase
}

type TestCase struct {
	Case   string      `json:"case"`
	Data   interface{} `json:"data"`
	Imp    bool        `json:"imp"`
	SetEnv interface{} `json:"set_env"`
}

type TestResponse struct {
	Results      []TestResult `json:"results"`
	AllImpPassed bool         `json:"allImpPassed"`
}

type TestResult struct {
	Case   string `json:"case"`
	Passed bool   `json:"passed"`
	Imp    bool   `json:"imp"`
}

type EndpointResponse struct {
	StatusCode int
	Headers    http.Header
	Body       string
}

type LoginRequest struct{
	Username	string `json:"username"`
	Password	string	`json:"password"`
}


type ATResponse struct {
	Results          []TestResult     `json:"results"`
	AllImpPassed     bool             `json:"all_imp_passed"`
	NewEnv           map[string]string `json:"new_env"`
	EndpointResponse EndpointResponse `json:"endpoint_response"`
}
