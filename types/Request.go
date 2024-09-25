package types
import (
	"net/http"
)
type LoginRequest struct{
	Username	string `json:"username"`
	Password	string	`json:"password"`
}

type EndpointResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    http.Header       `json:"headers"`
	Body       string            `json:"body"`
}


type ComplexATRequest struct {
	Env          map[string]string `json:"env"`
	EndpointData ATRequest         `json:"endpoint_data"`
}


type ATRequest struct {
	Tag       string            `json:"tag"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Variables map[string]string `json:"variables"`
	Headers   map[string]string `json:"headers"`
	Body      map[string]string `json:"body"`
	TestCases []TestCase        `json:"test_cases"`
}

type TestCase struct {
	Case   string                 `json:"case"`
	Data   interface{}            `json:"data"`
	SetEnv map[string]string      `json:"set_env,omitempty"`
	Imp    bool                   `json:"imp"`
}

type TestResult struct {
	Case   string `json:"case"`
	Passed bool   `json:"passed"`
	Imp    bool   `json:"imp"`
}

type TestResponse struct {
	Results      []TestResult `json:"results"`
	AllImpPassed bool         `json:"all_imp_passed"`
}

type ATResponse struct {
	Results          []TestResult     `json:"results"`
	AllImpPassed     bool             `json:"all_imp_passed"`
	NewEnv           map[string]string `json:"new_env"`
	EndpointResponse EndpointResponse `json:"endpoint_response"`
}
