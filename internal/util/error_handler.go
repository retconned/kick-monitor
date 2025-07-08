package util

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	report, ok := err.(*echo.HTTPError)
	if !ok {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Send JSON response
	if !c.Response().Committed {
		if report.Internal != nil {
			c.Logger().Error("Internal error:", report.Internal) // Log internal errors
		}

		// Map common errors to client-friendly messages or structure
		message := report.Message
		if msgStr, isString := message.(string); !isString {
			message = "An unexpected error occurred" // Generic message if original is not string
		} else if msgStr == http.StatusText(http.StatusInternalServerError) {
			message = "An internal server error occurred" // More friendly for 500
		}

		errorResponse := map[string]any{
			"message": message,
			"code":    report.Code,
		}

		if err := c.JSON(report.Code, errorResponse); err != nil {
			c.Logger().Error("Failed to send error response:", err)
		}
	}
}
