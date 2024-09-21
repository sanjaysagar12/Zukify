package services

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"zukify.com/types"
)

type TestResponse struct {
	Results       []types.TestResult `json:"results"`
	AllImpPassed  bool               `json:"all_imp_passed"`
}

func TestEndpoint(req types.ATRequest) TestResponse {
	client := &http.Client{}
	httpReq, err := prepareRequest(req)
	if err != nil {
		return TestResponse{
			Results: []types.TestResult{{Case: "request_creation", Passed: false, Imp: true}},
			AllImpPassed: false,
		}
	}

	start := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return TestResponse{
			Results: []types.TestResult{{Case: "request_execution", Passed: false, Imp: true}},
			AllImpPassed: false,
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TestResponse{
			Results: []types.TestResult{{Case: "response_reading", Passed: false, Imp: true}},
			AllImpPassed: false,
		}
	}

	duration := time.Since(start)

	results := runTestCases(req.TestCases, resp, body, duration)
	allImpPassed := checkAllImpPassed(results)

	return TestResponse{
		Results:      results,
		AllImpPassed: allImpPassed,
	}
}

func prepareRequest(req types.ATRequest) (*http.Request, error) {
	bodyJSON, err := json.Marshal(req.Body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

func runTestCases(testCases []types.TestCase, resp *http.Response, body []byte, duration time.Duration) []types.TestResult {
	var results []types.TestResult

	for _, tc := range testCases {
		result := types.TestResult{Case: tc.Case, Imp: tc.Imp}

		switch tc.Case {
		case "check_status_200":
			result.Passed = resp.StatusCode == http.StatusOK

		case "check_response_contains":
			result.Passed = strings.Contains(string(body), tc.Data.(string))

		case "check_json_field_exists":
			var jsonResp map[string]interface{}
			if err := json.Unmarshal(body, &jsonResp); err == nil {
				_, result.Passed = jsonResp[tc.Data.(string)]
			}

		case "check_response_time":
			result.Passed = duration.Milliseconds() <= int64(tc.Data.(float64))

		case "check_response_body_length":
			result.Passed = len(body) == int(tc.Data.(float64))

		case "check_response_is_valid_json":
			var js json.RawMessage
			result.Passed = json.Unmarshal(body, &js) == nil

		case "check_xml_field_value":
			result.Passed = checkXMLFieldValue(body, tc.Data.(map[string]interface{}))

		case "check_specific_string_in_html":
			result.Passed = strings.Contains(string(body), tc.Data.(string))

		case "check_json_array_contains_value":
			result.Passed = checkJSONArrayContainsValue(body, tc.Data.(map[string]interface{}))

		case "check_non_empty_response":
			result.Passed = len(body) > 0

		// case "check_specific_cookie":
		// 	cookieName := tc.Data.(map[string]interface{})["cookie"].(string)
		// 	_, err := resp.Cookie(cookieName)
		// 	result.Passed = err == nil

		default:
			result.Passed = false
		}

		results = append(results, result)
	}

	return results
}

func checkXMLFieldValue(body []byte, data map[string]interface{}) bool {
	var xmlDoc map[string]interface{}
	if err := xml.Unmarshal(body, &xmlDoc); err != nil {
		return false
	}

	field := data["field"].(string)
	value := data["value"].(string)

	return xmlDoc[field] == value
}

func checkJSONArrayContainsValue(body []byte, data map[string]interface{}) bool {
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return false
	}

	field := data["field"].(string)
	value := data["value"]

	if arr, ok := jsonResp[field].([]interface{}); ok {
		for _, v := range arr {
			if v == value {
				return true
			}
		}
	}

	return false
}

func checkAllImpPassed(results []types.TestResult) bool {
	for _, result := range results {
		if result.Imp && !result.Passed {
			return false
		}
	}
	return true
}