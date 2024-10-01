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
		headers[k] = replaceVariables(fmt.Sprintf("%v", v), data.Variables, env)
	}
	body := make(map[string]interface{})
	for k, v := range data.Body {
		body[k] = v
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
			if setEnv, ok := tc.SetEnv.(map[string]interface{}); ok {
				for k, v := range setEnv {
					strValue := fmt.Sprintf("%v", v)
					if strings.HasPrefix(strValue, "(response[") && strings.HasSuffix(strValue, "])") {
						field := strings.TrimSuffix(strings.TrimPrefix(strValue, "(response["), "])")
						var responseJSON map[string]interface{}
						if err := json.Unmarshal(body, &responseJSON); err == nil {
							if value, ok := responseJSON[field]; ok {
								newEnv[k] = fmt.Sprintf("%v", value)
							}
						}
					} else {
						newEnv[k] = strValue
					}
				}
			}
		}

		results = append(results, result)
	}

	return results, newEnv
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
			value, exists := jsonResp[tc.Data.(map[string]interface{})["field"].(string)]
			return exists && value == tc.Data.(map[string]interface{})["value"]
		}
		return false
	case "check_response_time":
		return duration.Milliseconds() <= int64(tc.Data.(float64))
	case "check_header_exists":
		_, exists := resp.Header[tc.Data.(string)]
		return exists
	// case "check_header_value":
	// 	headerData := tc.Data.(map[string]string)
	// 	headerValue := resp.Header.Get(headerData["name"])
	// 	return headerValue == headerData["value"]
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
		// This is a simplified check. For robust XML parsing, consider using encoding/xml package
		xmlData := tc.Data.(map[string]string)
		return strings.Contains(string(body), "<"+xmlData["field"]+">"+xmlData["value"]+"</"+xmlData["field"]+">")
	case "check_specific_string_in_html":
		return strings.Contains(string(body), tc.Data.(string))
	case "check_json_array_contains_value":
		var jsonResp []interface{}
		if err := json.Unmarshal(body, &jsonResp); err == nil {
			for _, item := range jsonResp {
				if item == tc.Data {
					return true
				}
			}
		}
		return false
	case "check_non_empty_response":
		return len(body) > 0
	// case "check_specific_cookie":
	// 	cookieData := tc.Data.(map[string]string)
	// 	for _, cookie := range resp.Cookies() {
	// 		if cookie.Name == cookieData["name"] && cookie.Value == cookieData["value"] {
	// 			return true
	// 		}
	// 	}
	// 	return false
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