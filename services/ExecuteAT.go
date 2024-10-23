package services

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"io"
	"bytes"
	"mime/multipart"
	"encoding/xml"
	"net/url"
	"fmt"
	"zukify.com/types"
	"regexp"
	"strconv"
	
)

func TestEndpoint(req types.ComplexATRequest) (types.TestResponse, map[string]string, types.EndpointResponse) {
	client := &http.Client{}
	fmt.Println("Request Data:", req.EndpointData, req.Env)
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

	// Convert req.Env to map[string]interface{}
	envInterface := make(map[string]interface{})
	for k, v := range req.Env {
		envInterface[k] = v
	}

	results, newEnv := runTestCases(req.EndpointData.TestCases, resp, body, duration, envInterface)

	// Convert newEnv back to map[string]string
	newEnvString := make(map[string]string)
	for k, v := range newEnv {
		if strVal, ok := v.(string); ok {
			newEnvString[k] = strVal
		} else {
			// Handle other types if necessary, or skip
		}
	}

	allImpPassed := checkAllImpPassed(results)

	return types.TestResponse{
		Results:      results,
		AllImpPassed: allImpPassed,
	}, newEnvString, endpointResponse
}



func prepareRequest(data types.ATRequest, env map[string]string) (*http.Request, error) {
	endpoint_url := replaceVariables(data.URL, data.Variables, env)
	headers := make(map[string]string)
	for k, v := range data.Headers {
		headers[k] = replaceVariables(fmt.Sprintf("%v", v), data.Variables, env)
	}

	var bodyReader io.Reader
	contentType := headers["Content-Type"]

	// Process based on Content-Type header
	switch {
	case strings.Contains(contentType, "multipart/form-data"):
		// Handle multipart/form-data (used for file uploads)
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		for k, v := range data.Body {
			switch v := v.(type) {
			case string:
				// Handle text fields
				_ = writer.WriteField(k, v)
			case []byte:
				// Handle binary files
				part, err := writer.CreateFormFile(k, "filename")
				if err != nil {
					return nil, err
				}
				_, err = part.Write(v)
				if err != nil {
					return nil, err
				}
			}
		}

		writer.Close()
		bodyReader = &b
		headers["Content-Type"] = writer.FormDataContentType()

	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		// Handle application/x-www-form-urlencoded
		formData := url.Values{}
		for k, v := range data.Body {
			formData.Set(k, fmt.Sprintf("%v", v))
		}
		bodyReader = strings.NewReader(formData.Encode())

	case strings.Contains(contentType, "application/json"):
		// Handle application/json
		bodyJSON, err := json.Marshal(data.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyJSON))

	case strings.Contains(contentType, "text/plain"):
		// Handle text/plain
		bodyString := fmt.Sprintf("%v", data.Body["text_field"])
		bodyReader = strings.NewReader(bodyString)

	case strings.Contains(contentType, "application/xml"):
		// Handle application/xml
		bodyXML, err := xml.Marshal(data.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyXML))

	case strings.Contains(contentType, "application/octet-stream"):
		// Handle application/octet-stream (binary data)
		bodyReader = bytes.NewReader(data.Body["file"].([]byte))

	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Create the request with the body
	httpReq, err := http.NewRequest(data.Method, endpoint_url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set the headers
	fmt.Println("Setting headers:")
	for k, v := range headers {
		fmt.Println(k, ":", v)
		httpReq.Header.Set(k, v)
	}
	fmt.Println("--------------")

	return httpReq, nil
}


func extractData(data interface{}, slicePattern string) (interface{}, error) {
    // Remove the "response" part and clean the pattern
    slicePattern = strings.TrimPrefix(slicePattern, "response")
    
    // Use regular expressions to parse the slice pattern (keys and indices)
    re := regexp.MustCompile(`\[(\w+|\d+|\:\d*)\]`)
    matches := re.FindAllStringSubmatch(slicePattern, -1)

    current := data

    for _, match := range matches {
        key := match[1]

        switch v := current.(type) {
        case map[string]interface{}:
            // If it's a map, look for the key
            current = v[key]
        case []interface{}:
            // If it's a slice, handle indexing and slicing
            if idx, err := strconv.Atoi(key); err == nil {
                current = v[idx]
            } else if strings.Contains(key, ":") {
                // Handle slicing like [1:5]
                parts := strings.Split(key, ":")
                start, _ := strconv.Atoi(parts[0])
                var end int
                if len(parts) > 1 && parts[1] != "" {
                    end, _ = strconv.Atoi(parts[1])
                } else {
                    end = len(v) // if no end is provided, take till the end
                }

                if start < 0 || end > len(v) {
                    return nil, fmt.Errorf("slice out of range")
                }
                current = v[start:end]
            }
        default:
            return nil, fmt.Errorf("invalid type: %v", v)
        }
    }

    return current, nil
}

func runTestCases(testCases []types.TestCase, resp *http.Response, body []byte, duration time.Duration, env map[string]interface{}) ([]types.TestResult, map[string]interface{}) {
    var results []types.TestResult
    newEnv := make(map[string]interface{}) // Now, values in newEnv can be of any type

    // Copy the existing env to newEnv
    for k, v := range env {
        newEnv[k] = v
    }

    for _, tc := range testCases {
        result := types.TestResult{Case: tc.Case, Imp: tc.Imp}
        result.Passed = runTestCase(tc, resp, body, duration)

        if result.Passed && tc.SetEnv != nil {
            // First, type assert tc.SetEnv to map[string]interface{}
            setEnvMap, ok := tc.SetEnv.(map[string]interface{})
            if !ok {
                fmt.Println("SetEnv is not a map, skipping...")
                continue
            }

            for key, value := range setEnvMap {
                strVal, ok := value.(string)
                if ok && strings.HasPrefix(strVal, "(response[") {
                    // This means the value contains an expression to extract data
                    var data map[string]interface{}
                    if err := json.Unmarshal(body, &data); err != nil {
                        panic(err)
                    }

                    extractedValue, err := extractData(data, strVal)
                    fmt.Println("body:", data, "\nextractedValue ", extractedValue, "\nstrVal:", strVal)
                    if err != nil {
                        fmt.Printf("Error extracting data for key %s: %v\n", key, err)
                        continue
                    }

                    // Store extractedValue directly into newEnv
                    newEnv[key] = extractedValue
                } else {
                    // If it's a simple key-value pair, add it directly to newEnv
                    newEnv[key] = value
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