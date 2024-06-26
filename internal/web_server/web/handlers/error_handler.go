package handlers

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/public_pages"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	_API_CONTEXT = "api_error_context"
)

// Custom error handler.
//
// If the error caught is an echo.HTTPError, the status code provided will be used
// under constraints specified below. The message may be a string.
//
// If the error is of different type status code 500 will be returned.
//
// Status codes <400: the error will be logged, response code will be replaced with 500.
// Codes under 400 are not expected as errors.
//
// Status codes >=500: the error will be logged. Error message to user
// will be the status text for the status code. This is to prevent leaking internal
// error messages to the user.
//
// 4XX status codes will be returned to the user as is, with the provided message.
// No additional logging will be done.
func HTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	errorMessage := "Internal server error"

	he, ok := err.(*echo.HTTPError)
	if ok {
		code = he.Code
		errorMessage = fmt.Sprintf("%v", he.Message)
	} else {
		errorMessage = err.Error()
	}

	if code < 400 {
		code = http.StatusInternalServerError
		errorMessage = http.StatusText(code)
		log.Errorf("Unexpected http status code in error. Responding with 500. Original error: %v", err)
	} else if code >= 500 {
		log.Errorf("Server error: %v", err)
		errorMessage = http.StatusText(code)
	}

	if c.Get(_API_CONTEXT) != nil {
		utils.Render(c, code, components.ErrorText("", errorMessage))
	} else {
		utils.Render(c, code, public_pages.ErrorPage(code, errorMessage))
	}
}

// Sets the contextk for API error display.
//
// If this middleware is used, value is added to the context and errors can be rendered differently from full page errors.
func SetApiErrorDisplay(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set(_API_CONTEXT, true)
		return next(c)
	}
}
