package handlers
import (
	"net/http"

	"github.com/labstack/echo/v4"
	"zukify.com/services"
	"zukify.com/types"
)
func HandlePostAT(c echo.Context) error {
	var req types.ATRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	results := service.TestEndpoint(req)
	return c.JSON(http.StatusOK, results)
}