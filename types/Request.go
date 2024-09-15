package types
type LoginRequest struct{
	Username	string `json:"username"`
	Password	string	`json:"password"`
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
	Case string      `json:"case"`
	Data interface{} `json:"data"`
	Imp  bool        `json:"imp"`
}

type TestResult struct {
	Case   string `json:"case"`
	Passed bool   `json:"passed"`
	Imp    bool   `json:"imp"`
}