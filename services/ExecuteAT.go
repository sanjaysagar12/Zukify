package services

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"fmt"
	"zukify.com/types"
)

func TestEndpoint(req types.ComplexATRequest) (types.TestResponse, map[string]string, types.EndpointResponse) {
	client := &http.Client{}
	fmt.Println("Request Data:",req.EndpointData, req.Env)
	httpReq, err := prepareRequest(req.EndpointData, req.Env)
	if err != nil {
		return types.TestResponse{
			Results: []types.TestResult{{Case: "request_creation", Passed: false, Imp: true}},
			AllImpPassed: false,
		}, req.Env, types.EndpointResponse{}
	}

	start := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return types.TestResponse{
			Results: []types.TestResult{{Case: "request_execution", Passed: false, Imp: true}},
			AllImpPassed: false,
		}, req.Env, types.EndpointResponse{}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return types.TestResponse{
			Results: []types.TestResult{{Case: "response_reading", Passed: false, Imp: true}},
			AllImpPassed: false,
		}, req.Env, types.EndpointResponse{}
	}

	duration := time.Since(start)

	endpointResponse := types.EndpointResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(body),
	}

	results, newEnv := runTestCases(req.EndpointData.TestCases, resp, body, duration, req.Env)
	allImpPassed := checkAllImpPassed(results)

	return types.TestResponse{
		Results:      results,
		AllImpPassed: allImpPassed,
	}, newEnv, endpointResponse
}


func prepareRequest(data types.ATRequest, env map[string]string) (*http.Request, error) {
	// Replace variables in URL, headers, and body
	url := replaceVariables(data.URL, data.Variables, env)
	headers := make(map[string]string)
	for k, v := range data.Headers {
		headers[k] = replaceVariables(v, data.Variables, env)
	}
	body := make(map[string]string)
	for k, v := range data.Body {
		body[k] = replaceVariables(v, data.Variables, env)
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(data.Method, url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

func replaceVariables(input string, variables map[string]string, env map[string]string) string {
	for k, v := range variables {
		input = strings.ReplaceAll(input, "<<"+k+">>", v)
	}
	for k, v := range env {
		input = strings.ReplaceAll(input, "<<"+k+">>", v)
	}
	return input
}

func runTestCases(testCases []types.TestCase, resp *http.Response, body []byte, duration time.Duration, env map[string]string) ([]types.TestResult, map[string]string) {
	var results []types.TestResult
	newEnv := make(map[string]string)
	for k, v := range env {
		newEnv[k] = v
	}

	for _, tc := range testCases {
		result := types.TestResult{Case: tc.Case, Imp: tc.Imp}
		result.Passed = runTestCase(tc, resp, body, duration)

		if result.Passed && tc.SetEnv != nil {
			for k, v := range tc.SetEnv {
				if strings.HasPrefix(v, "(response[") && strings.HasSuffix(v, "])") {
					field := strings.TrimSuffix(strings.TrimPrefix(v, "(response["), "])")
					var responseJSON map[string]interface{}
					if err := json.Unmarshal(body, &responseJSON); err == nil {
						if value, ok := responseJSON[field]; ok {
							newEnv[k] = fmt.Sprintf("%v", value)
						}
					}
				} else {
					newEnv[k] = v
				}
			}
		}

		results = append(results, result)
	}

	return results, newEnv
}

func runTestCase(tc types.TestCase, resp *http.Response, body []byte, duration time.Duration) bool {
	switch tc.Case {
	case "check_status_200":
		return resp.StatusCode == http.StatusOK
	case "check_response_contains":
		return strings.Contains(string(body), tc.Data.(string))
	case "check_json_field_exists":
		var jsonResp map[string]interface{}
		if err := json.Unmarshal(body, &jsonResp); err == nil {
			_, exists := jsonResp[tc.Data.(string)]
			return exists
		}
		return false
	case "check_response_time":
		return duration.Milliseconds() <= int64(tc.Data.(float64))
	// Add other test cases as needed
	default:
		return false
	}
}

func checkAllImpPassed(results []types.TestResult) bool {
	for _, result := range results {
		if result.Imp && !result.Passed {
			return false
		}
	}
	return true
}