package middlewares

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

const (
	acceptTermsPath = "/quiz/accept-terms"
)

// Middleware for enforcing that user ahs accepted terms of service. Can either redirect to accepting page, or respond with 409 conflict.
// Admins do not need to accept terms of service.
type acceptedTerms struct {
	sharedData     *config.SharedData
	redirectToPage bool
}

// Creates new instance of acceptedTerms middleware
func NewAcceptedTerms(data *config.SharedData, redirectToLogin bool) *acceptedTerms {
	return &acceptedTerms{data, redirectToLogin}
}

// Only allows users who accepted the terms of service, otherwise either redirects to accepting page or returns 409 conflict. Admins are ignroed.
func (m *acceptedTerms) EncofreAcceptedTerms(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := users.GetUserByID(m.sharedData.DB, utils.GetUserIDFromCtx(c))
		if err != nil {
			return err
		}
		if !user.Role.IsAdministrator() && !user.AcceptedTerms && c.Path() != acceptTermsPath {
			if m.redirectToPage {
				return c.Redirect(http.StatusTemporaryRedirect, acceptTermsPath)
			} else {
				return echo.NewHTTPError(http.StatusConflict, "You must accept terms of service to proceed")
			}
		}

		return next(c)

	}
}
