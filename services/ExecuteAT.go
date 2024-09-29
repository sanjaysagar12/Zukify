package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"zukify.com/types"
)

func TestEndpoint(req types.ComplexATRequest) (types.TestResponse, map[string]string, types.EndpointResponse) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	fmt.Println("Request Data:", req.EndpointData, req.Env)
	
	httpReq, err := prepareRequest(req.EndpointData, req.Env)
	if err != nil {
		fmt.Printf("Error preparing request: %v\n", err)
		return runAllTestCases(req.EndpointData.TestCases, nil, nil, 0, req.Env, "request_preparation")
	}

	start := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		return runAllTestCases(req.EndpointData.TestCases, nil, nil, 0, req.Env, "request_execution")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return runAllTestCases(req.EndpointData.TestCases, resp, nil, time.Since(start), req.Env, "response_reading")
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
	url := replaceVariables(data.URL, data.Variables, env)
	
	httpReq, err := http.NewRequest(data.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	for k, v := range data.Headers {
		httpReq.Header.Set(k, replaceVariables(v, data.Variables, env))
	}

	// For GET requests, we don't need to set the body
	if data.Method != "GET" && len(data.Body) > 0 {
		bodyJSON, err := json.Marshal(data.Body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling body: %v", err)
		}
		httpReq.Body = ioutil.NopCloser(bytes.NewBuffer(bodyJSON))
		httpReq.ContentLength = int64(len(bodyJSON))
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
	case "check_json_field_value":
		var jsonResp map[string]interface{}
		if err := json.Unmarshal(body, &jsonResp); err == nil {
			data := tc.Data.(map[string]interface{})
			value, exists := jsonResp[data["field"].(string)]
			return exists && fmt.Sprintf("%v", value) == fmt.Sprintf("%v", data["value"])
		}
		return false
	case "check_response_time":
		return duration.Milliseconds() <= int64(tc.Data.(float64))
	case "check_header_exists":
		_, exists := resp.Header[tc.Data.(string)]
		return exists
	case "check_header_value":
		headerData := tc.Data.(map[string]string)
		return resp.Header.Get(headerData["name"]) == headerData["value"]
	case "check_response_non_empty":
		return len(body) > 0
	case "check_content_type":
		return resp.Header.Get("Content-Type") == tc.Data.(string)
	case "check_response_body_length":
		return len(body) == int(tc.Data.(float64))
	case "check_response_is_valid_json":
		var js json.RawMessage
		return json.Unmarshal(body, &js) == nil
	case "check_xml_field_value":
		// This is a simplified check. For proper XML parsing, you'd need to use an XML parsing library.
		return strings.Contains(string(body), tc.Data.(string))
	case "check_json_array_contains_value":
		var jsonResp []interface{}
		if err := json.Unmarshal(body, &jsonResp); err == nil {
			for _, item := range jsonResp {
				if fmt.Sprintf("%v", item) == fmt.Sprintf("%v", tc.Data) {
					return true
				}
			}
		}
		return false
	case "check_specific_cookie":
		cookieData := tc.Data.(map[string]string)
		for _, cookie := range resp.Cookies() {
			if cookie.Name == cookieData["name"] && cookie.Value == cookieData["value"] {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func runAllTestCases(testCases []types.TestCase, resp *http.Response, body []byte, duration time.Duration, env map[string]string, failedCase string) (types.TestResponse, map[string]string, types.EndpointResponse) {
	results := []types.TestResult{{Case: failedCase, Passed: false, Imp: true}}
	for _, tc := range testCases {
		if tc.Case != failedCase {
			results = append(results, types.TestResult{Case: tc.Case, Passed: false, Imp: tc.Imp})
		}
	}
	return types.TestResponse{
		Results:      results,
		AllImpPassed: false,
	}, env, types.EndpointResponse{}
}

func checkAllImpPassed(results []types.TestResult) bool {
	for _, result := range results {
		if result.Imp && !result.Passed {
			return false
		}
	}
	return true
}