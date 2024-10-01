package handlers

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"zukify.com/services"
	"zukify.com/types"
	"fmt"
	"encoding/json"
)

func HandlePostAT(c echo.Context) error {
	var req types.ComplexATRequest
	if err := c.Bind(&req); err != nil {
		fmt.Println("400 Error")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	fmt.Println("ComplexATRequest:",req)
	results, newEnv, endpointResponse := services.TestEndpoint(req)
	response := types.ATResponse{
		Results:          results.Results,
		AllImpPassed:     results.AllImpPassed,
		NewEnv:           newEnv,
		EndpointResponse: endpointResponse,
	}
	b, err := json.MarshalIndent(response, "", "  ")
    if err != nil {
        fmt.Println(err)
    }
    fmt.Print(string(b))
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
